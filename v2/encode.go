package bsonpb

import (
	"errors"
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/romnnn/bsonpb/v2/internal/genid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	protoLegacy   = false
	defaultIndent = "  "
	extensionName = "message_set_extension"
)

// IsMessageSet returns whether the message uses the MessageSet wire format.
func IsMessageSet(md pref.MessageDescriptor) bool {
	xmd, ok := md.(interface{ IsMessageSet() bool })
	return ok && xmd.IsMessageSet()
}

// IsMessageSetExtension reports this field extends a MessageSet.
func IsMessageSetExtension(fd pref.FieldDescriptor) bool {
	if fd.Name() != extensionName {
		return false
	}
	if fd.FullName().Parent() != fd.Message().FullName() {
		return false
	}
	return IsMessageSet(fd.ContainingMessage())
}

// Marshal writes the given proto.Message in JSON format using default options.
// Do not depend on the output being stable. It may change over time across
// different versions of the program.
func Marshal(m proto.Message) (bson.D, error) {
	result, err := MarshalOptions{}.Marshal(m)
	return result.(bson.D), err
}

// NoUnkeyedLiterals can be embedded in a struct to prevent unkeyed literals.
type NoUnkeyedLiterals struct{}

// MarshalOptions is a configurable JSON format marshaler.
type MarshalOptions struct {
	NoUnkeyedLiterals

	// Multiline specifies whether the marshaler should format the output in
	// indented-form with every textual element on a new line.
	// If Indent is an empty string, then an arbitrary indent is chosen.
	Multiline bool

	// Indent specifies the set of indentation characters to use in a multiline
	// formatted output such that every entry is preceded by Indent and
	// terminated by a newline. If non-empty, then Multiline is treated as true.
	// Indent can only be composed of space or tab characters.
	Indent string

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false (the default),
	// Marshal will return error if there are any missing required fields.
	AllowPartial bool

	// UseProtoNames uses proto field name instead of lowerCamelCase name in JSON
	// field names.
	UseProtoNames bool

	// UseEnumNumbers emits enum values as numbers.
	UseEnumNumbers bool

	// EmitUnpopulated specifies whether to emit unpopulated fields. It does not
	// emit unpopulated oneof fields or unpopulated extension fields.
	// The JSON value emitted for unpopulated fields are as follows:
	//  ╔═══════╤════════════════════════════╗
	//  ║ JSON  │ Protobuf field             ║
	//  ╠═══════╪════════════════════════════╣
	//  ║ false │ proto3 boolean fields      ║
	//  ║ 0     │ proto3 numeric fields      ║
	//  ║ ""    │ proto3 string/bytes fields ║
	//  ║ null  │ proto2 scalar fields       ║
	//  ║ null  │ message fields             ║
	//  ║ []    │ list fields                ║
	//  ║ {}    │ map fields                 ║
	//  ╚═══════╧════════════════════════════╝
	EmitUnpopulated bool

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		protoregistry.ExtensionTypeResolver
		protoregistry.MessageTypeResolver
	}
}

// Marshal marshals the given proto.Message in the JSON format using options in
// MarshalOptions. Do not depend on the output being stable. It may change over
// time across different versions of the program.
func (o MarshalOptions) Marshal(m proto.Message) (interface{}, error) {
	return o.marshal(m)
}

// marshal is a centralized function that all marshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for marshal that do not go through this.
func (o MarshalOptions) marshal(m proto.Message) (interface{}, error) {
	if o.Multiline && o.Indent == "" {
		o.Indent = defaultIndent
	}
	if o.Resolver == nil {
		o.Resolver = protoregistry.GlobalTypes
	}

	// Treat nil message interface as an empty message,
	// in which case the output in an empty JSON object.
	if m == nil {
		return bson.D{}, nil
	}

	enc := encoder{o}
	result, err := enc.marshalMessage(m.ProtoReflect())
	if err != nil {
		return bson.D{}, err
	}
	if o.AllowPartial {
		return result, nil
	}
	return result, proto.CheckInitialized(m)
}

type encoder struct {
	opts MarshalOptions
}

// marshalMessage marshals the given protoreflect.Message.
func (e encoder) marshalMessage(m pref.Message) (interface{}, error) {
	if marshal := wellKnownTypeMarshaler(m.Descriptor().FullName()); marshal != nil {
		return marshal(e, m)
	}

	// This is the entry
	result, err := e.marshalFields(m)
	if err != nil {
		return bson.D{}, err
	}

	return result, nil
}

// marshalFields marshals the fields in the given protoreflect.Message.
func (e encoder) marshalFields(m pref.Message) (bson.D, error) {
	result := bson.D{}
	messageDesc := m.Descriptor()
	if !protoLegacy && IsMessageSet(messageDesc) {
		return result, errors.New("no support for proto1 MessageSets")
	}

	// Marshal out known fields.
	fieldDescs := messageDesc.Fields()
	for i := 0; i < fieldDescs.Len(); {
		fd := fieldDescs.Get(i)
		if od := fd.ContainingOneof(); od != nil {
			fd = m.WhichOneof(od)
			i += od.Fields().Len()
			if fd == nil {
				continue // unpopulated oneofs are not affected by EmitUnpopulated
			}
		} else {
			i++
		}

		val := m.Get(fd)
		if !m.Has(fd) {
			if !e.opts.EmitUnpopulated {
				continue
			}
			isProto2Scalar := fd.Syntax() == pref.Proto2 && fd.Default().IsValid()
			isSingularMessage := fd.Cardinality() != pref.Repeated && fd.Message() != nil
			if isProto2Scalar || isSingularMessage {
				// Use invalid value to emit null.
				val = pref.Value{}
			}
		}

		name := fd.JSONName()
		if e.opts.UseProtoNames {
			name = string(fd.Name())
			// Use type name for group field name.
			if fd.Kind() == pref.GroupKind {
				name = string(fd.Message().Name())
			}
		}

		marshaled, err := e.marshalValue(val, fd)
		if err != nil {
			return bson.D{}, err
		}
		result = append(result, bson.E{Key: name, Value: marshaled})
	}

	// Marshal out extensions.
	extensions, err := e.marshalExtensions(m)
	if err != nil {
		return bson.D{}, err
	}
	result = append(result, extensions...)
	return result, nil
}

func (e encoder) marshalValue(val pref.Value, fd pref.FieldDescriptor) (interface{}, error) {
	// fmt.Printf("Marshal Value: %s: %v\n", name, val)
	switch {
	case fd.IsList():
		return e.marshalList(val.List(), fd)
	case fd.IsMap():
		return e.marshalMap(val.Map(), fd)
	default:
		return e.marshalSingular(val, fd)
	}
}

func (e encoder) marshalSingular(val pref.Value, fd pref.FieldDescriptor) (interface{}, error) {
	if !val.IsValid() {
		return primitive.Null{}, nil
	}

	switch kind := fd.Kind(); kind {
	case pref.BoolKind:
		return val.Bool(), nil

	case pref.StringKind:
		if valid := utf8.Valid([]byte(val.String())); valid {
			return val.String(), nil
		}
		return "", fmt.Errorf("InvalidUTF8: %s", val.String())

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind:
		return int32(val.Int()), nil

	case pref.Uint32Kind, pref.Fixed32Kind:
		return uint32(val.Uint()), nil

	case pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
		return val.Int(), nil

	case pref.Uint64Kind, pref.Fixed64Kind:
		return val.Uint(), nil

	case pref.FloatKind:
		return float32(val.Float()), nil

	case pref.DoubleKind:
		return val.Float(), nil

	case pref.BytesKind:
		return primitive.Binary{Data: val.Bytes()}, nil

	case pref.EnumKind:
		if fd.Enum().FullName() == genid.NullValue_enum_fullname {
			return primitive.Null{}, nil
		}
		desc := fd.Enum().Values().ByNumber(val.Enum())
		if e.opts.UseEnumNumbers || desc == nil {
			return int64(val.Enum()), nil
		}
		return string(desc.Name()), nil

	case pref.MessageKind, pref.GroupKind:
		marshaled, err := e.marshalMessage(val.Message())
		if err != nil {
			return bson.D{}, err
		}
		return marshaled, nil

	default:
		return bson.D{}, fmt.Errorf("%v has unknown kind: %v", fd.FullName(), kind)
	}
}

// marshalList marshals the given protoreflect.List.
func (e encoder) marshalList(list pref.List, fd pref.FieldDescriptor) (interface{}, error) {
	result := bson.A{}
	for i := 0; i < list.Len(); i++ {
		item := list.Get(i)
		val, err := e.marshalSingular(item, fd)
		if err != nil {
			return bson.A{}, err
		}
		result = append(result, val)
	}
	return result, nil
}

type mapEntry struct {
	key   pref.MapKey
	value pref.Value
}

// marshalMap marshals given protoreflect.Map.
func (e encoder) marshalMap(mmap pref.Map, fd pref.FieldDescriptor) (interface{}, error) {
	result := bson.D{}
	// Get a sorted list based on keyType first.
	entries := make([]mapEntry, 0, mmap.Len())
	mmap.Range(func(key pref.MapKey, val pref.Value) bool {
		entries = append(entries, mapEntry{key: key, value: val})
		return true
	})
	sortMap(fd.MapKey().Kind(), entries)

	// Write out sorted list.
	for _, entry := range entries {
		val, err := e.marshalSingular(entry.value, fd.MapValue())
		if err != nil {
			return nil, err
		}
		result = append(result, bson.E{Key: entry.key.String(), Value: val})
	}
	return result, nil
}

// sortMap orders list based on value of key field for deterministic ordering.
func sortMap(keyKind pref.Kind, values []mapEntry) {
	sort.Slice(values, func(i, j int) bool {
		switch keyKind {
		case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind,
			pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
			return values[i].key.Int() < values[j].key.Int()

		case pref.Uint32Kind, pref.Fixed32Kind,
			pref.Uint64Kind, pref.Fixed64Kind:
			return values[i].key.Uint() < values[j].key.Uint()
		}
		return values[i].key.String() < values[j].key.String()
	})
}

// marshalExtensions marshals extension fields.
func (e encoder) marshalExtensions(m pref.Message) ([]bson.E, error) {
	result := []bson.E{}
	type entry struct {
		key   string
		value pref.Value
		desc  pref.FieldDescriptor
	}

	// Get a sorted list based on field key first.
	var entries []entry
	m.Range(func(fd pref.FieldDescriptor, v pref.Value) bool {
		if !fd.IsExtension() {
			return true
		}

		// For MessageSet extensions, the name used is the parent message.
		name := fd.FullName()
		if IsMessageSetExtension(fd) {
			name = name.Parent()
		}

		// Use [name] format for JSON field name.
		entries = append(entries, entry{
			key:   string(name),
			value: v,
			desc:  fd,
		})
		return true
	})

	// Sort extensions lexicographically.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	// Write out sorted list.
	for _, entry := range entries {
		// JSON field name is the proto field name enclosed in [], similar to
		// textproto. This is consistent with Go v1 lib. C++ lib v3.7.0 does not
		// marshal out extension fields.
		marshaled, err := e.marshalValue(entry.value, entry.desc)
		if err != nil {
			return result, err
		}
		result = append(result, bson.E{Key: "[" + entry.key + "]", Value: marshaled})
	}
	return result, nil
}
