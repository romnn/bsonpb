package bsonpb

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/lunemec/as"
	"github.com/romnnn/bsonpb/v2/internal/genid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Unmarshal reads the given bson.D into the given proto.Message.
func Unmarshal(doc interface{}, m proto.Message) error {
	return UnmarshalOptions{}.Unmarshal(doc, m)
}

// UnmarshalOptions is a configurable JSON format parser.
type UnmarshalOptions struct {
	NoUnkeyedLiterals

	// If AllowPartial is set, input for messages that will result in missing
	// required fields will not return an error.
	AllowPartial bool

	// If DiscardUnknown is set, unknown fields are ignored.
	DiscardUnknown bool

	// Resolver is used for looking up types when unmarshaling
	// google.protobuf.Any messages or extension fields.
	// If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}
}

// Unmarshal reads the given []byte and populates the given proto.Message using
// options in UnmarshalOptions object. It will clear the message first before
// setting the fields. If it returns an error, the given message may be
// partially set.
func (o UnmarshalOptions) Unmarshal(doc interface{}, m proto.Message) error {
	return o.unmarshal(doc, m)
}

// UnmarshalBytes ...
func (o UnmarshalOptions) UnmarshalBytes(b []byte, m proto.Message) error {
	reader := bsonrw.NewBSONDocumentReader(b)
	bsonDec, err := bson.NewDecoder(reader)
	if err != nil {
		return fmt.Errorf("Failed to create decoder for BSON stream: %s", err.Error())
	}

	var inputValue bson.D
	if err := bsonDec.Decode(&inputValue); err != nil {
		return fmt.Errorf("Failed to decode bson into interface: %s", err.Error())
	}
	return o.unmarshal(inputValue, m)
}

// unmarshal is a centralized function that all unmarshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for unmarshal that do not go through this.
func (o UnmarshalOptions) unmarshal(doc interface{}, m proto.Message) error {
	proto.Reset(m)

	if o.Resolver == nil {
		o.Resolver = protoregistry.GlobalTypes
	}

	dec := decoder{o}
	if err := dec.unmarshalMessage(doc, m.ProtoReflect(), false); err != nil {
		return err
	}

	if o.AllowPartial {
		return nil
	}
	return proto.CheckInitialized(m)
}

type decoder struct {
	opts UnmarshalOptions
}

// unmarshalMessage unmarshals a message into the given protoreflect.Message.
func (d decoder) unmarshalMessage(doc interface{}, m pref.Message, skipTypeURL bool) error {
	if unmarshalFunc := wellKnownTypeUnmarshaler(m.Descriptor().FullName()); unmarshalFunc != nil {
		return unmarshalFunc(d, doc, m)
	}

	_, isNullPrimitive := doc.(primitive.Null)
	if (isNullPrimitive || doc == nil) && m.Descriptor().FullName() == genid.Value_message_fullname {
		return nil
	}

	messageDesc := m.Descriptor()
	if !protoLegacy && IsMessageSet(messageDesc) {
		return errors.New("no support for proto1 MessageSets")
	}

	var seenNums Ints
	var seenOneofs Ints
	fieldDescs := messageDesc.Fields()

	docD, ok := doc.(bson.D)
	if !ok {
		return fmt.Errorf("unexpected message value: %v", doc)
	}
	for _, item := range docD {
		name := item.Key
		val := item.Value
		var fd pref.FieldDescriptor
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			// Only extension names are in [name] format.
			extName := pref.FullName(name[1 : len(name)-1])
			extType, err := d.opts.Resolver.FindExtensionByName(extName)
			if err != nil && err != protoregistry.NotFound {
				return fmt.Errorf("unable to resolve %v: %v", val, err)
			}
			if extType != nil {
				fd = extType.TypeDescriptor()
				if !messageDesc.ExtensionRanges().Has(fd.Number()) || fd.ContainingMessage().FullName() != messageDesc.FullName() {
					return fmt.Errorf("message %v cannot be extended by %v", messageDesc.FullName(), fd.FullName())
				}
			}
		} else {
			// The name can either be the JSON name or the proto field name.
			fd = fieldDescs.ByJSONName(name)
			/*
				// TODO: Coming in v1.25+
				if fd == nil {
					fd = fieldDescs.ByTextName(name)
				}
			*/
			if fd == nil {
				fd = fieldDescs.ByName(pref.Name(name))
				if fd == nil {
					// The proto name of a group field is in all lowercase,
					// while the textual field name is the group message name.
					gd := fieldDescs.ByName(pref.Name(strings.ToLower(name)))
					if gd != nil && gd.Kind() == pref.GroupKind && gd.Message().Name() == pref.Name(name) {
						fd = gd
					}
				} else if fd.Kind() == pref.GroupKind && fd.Message().Name() != pref.Name(name) {
					fd = nil // reset since field name is actually the message name
				}
			}
		}
		if protoLegacy {
			if fd != nil && fd.IsWeak() && fd.Message().IsPlaceholder() {
				fd = nil // reset since the weak reference is not linked in
			}
		}

		if fd == nil {
			// Field is unknown.
			if d.opts.DiscardUnknown {
				continue
			}
			return fmt.Errorf("unknown field %q", name)
		}

		// Do not allow duplicate fields.
		num := uint64(fd.Number())
		if seenNums.Has(num) {
			return fmt.Errorf("duplicate field %q", name)
		}
		seenNums.Set(num)

		// No need to set values for JSON null unless the field type is
		// google.protobuf.Value or google.protobuf.NullValue.
		_, isNullPrimitive := val.(primitive.Null)
		if (isNullPrimitive || val == nil) && !isKnownValue(fd) && !isNullValue(fd) {
			continue
		}

		switch {
		case fd.IsList():
			list := m.Mutable(fd).List()
			nested, ok := val.(bson.A)
			if ok {
				if err := d.unmarshalList(nested, list, fd); err != nil {
					return err
				}
			}
		case fd.IsMap():
			mmap := m.Mutable(fd).Map()
			if err := d.unmarshalMap(val.(bson.D), mmap, fd); err != nil {
				return err
			}
		default:
			// If field is a oneof, check if it has already been set.
			if od := fd.ContainingOneof(); od != nil {
				idx := uint64(od.Index())
				if seenOneofs.Has(idx) {
					return fmt.Errorf("error parsing %q, oneof %v is already set", name, od.FullName())
				}
				seenOneofs.Set(idx)
			}

			// Required or optional fields.
			if err := d.unmarshalSingular(val, m, fd); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d decoder) unmarshalMap(doc bson.D, mmap pref.Map, fd pref.FieldDescriptor) error {
	// Determine ahead whether map entry is a scalar type or a message type in
	// order to call the appropriate unmarshalMapValue func inside the for loop
	// below.
	var unmarshalMapValue func(val interface{}) (pref.Value, error)
	switch fd.MapValue().Kind() {
	case pref.MessageKind, pref.GroupKind:
		unmarshalMapValue = func(val interface{}) (pref.Value, error) {
			mapVal := mmap.NewValue()
			if err := d.unmarshalMessage(val, mapVal.Message(), false); err != nil {
				return pref.Value{}, err
			}
			return mapVal, nil
		}
	default:
		unmarshalMapValue = func(val interface{}) (pref.Value, error) {
			return d.unmarshalScalar(val, fd.MapValue())
		}
	}

	for _, item := range doc {
		name := item.Key
		val := item.Value
		// Unmarshal field name.
		pkey, err := d.unmarshalMapKey(name, fd.MapKey())
		if err != nil {
			return err
		}

		// Check for duplicate field name.
		if mmap.Has(pkey) {
			return fmt.Errorf("duplicate map key %v", name)
		}

		// Read and unmarshal field value.
		pval, err := unmarshalMapValue(val)
		if err != nil {
			return err
		}

		mmap.Set(pkey, pval)
	}

	return nil
}

// unmarshalMapKey converts given token of Name kind into a protoreflect.MapKey.
// A map key type is any integral or string type.
func (d decoder) unmarshalMapKey(name string, fd pref.FieldDescriptor) (pref.MapKey, error) {
	const b32 = 32
	const b64 = 64
	const base10 = 10

	kind := fd.Kind()
	switch kind {
	case pref.StringKind:
		return pref.ValueOfString(name).MapKey(), nil

	case pref.BoolKind:
		switch name {
		case "true":
			return pref.ValueOfBool(true).MapKey(), nil
		case "false":
			return pref.ValueOfBool(false).MapKey(), nil
		}

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind:
		if n, err := strconv.ParseInt(name, base10, b32); err == nil {
			return pref.ValueOfInt32(int32(n)).MapKey(), nil
		}

	case pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
		if n, err := strconv.ParseInt(name, base10, b64); err == nil {
			return pref.ValueOfInt64(int64(n)).MapKey(), nil
		}

	case pref.Uint32Kind, pref.Fixed32Kind:
		if n, err := strconv.ParseUint(name, base10, b32); err == nil {
			return pref.ValueOfUint32(uint32(n)).MapKey(), nil
		}

	case pref.Uint64Kind, pref.Fixed64Kind:
		if n, err := strconv.ParseUint(name, base10, b64); err == nil {
			return pref.ValueOfUint64(uint64(n)).MapKey(), nil
		}

	default:
		panic(fmt.Sprintf("invalid kind for map key: %v", kind))
	}

	return pref.MapKey{}, fmt.Errorf("invalid value for %v key: %q", kind, name)
}

func (d decoder) unmarshalList(doc bson.A, list pref.List, fd pref.FieldDescriptor) error {
	switch fd.Kind() {
	case pref.MessageKind, pref.GroupKind:
		for _, item := range doc {
			val := list.NewElement()
			if err := d.unmarshalMessage(item, val.Message(), false); err != nil {
				return err
			}
			list.Append(val)
		}
	default:
		for _, item := range doc {
			val, err := d.unmarshalScalar(item, fd)
			if err != nil {
				return err
			}
			list.Append(val)
		}
	}
	return nil
}

// unmarshalSingular unmarshals to the non-repeated field specified
// by the given FieldDescriptor.
func (d decoder) unmarshalSingular(doc interface{}, m pref.Message, fd pref.FieldDescriptor) error {
	var val pref.Value
	var err error
	switch fd.Kind() {
	case pref.MessageKind, pref.GroupKind:
		val = m.NewField(fd)
		err = d.unmarshalMessage(doc, val.Message(), false)
	default:
		val, err = d.unmarshalScalar(doc, fd)
	}

	if err != nil {
		return err
	}
	m.Set(fd, val)
	return nil
}

// unmarshalScalar unmarshals to a scalar/enum protoreflect.Value specified by
// the given FieldDescriptor.
func (d decoder) unmarshalScalar(doc interface{}, fd pref.FieldDescriptor) (pref.Value, error) {
	kind := fd.Kind()

	if doc == nil {
		return pref.Value{}, fmt.Errorf(`invalid value for %v type: %v (has type %T)`, kind, doc, doc)
	}

	vdoc := reflect.ValueOf(doc)
	docType := vdoc.Type()
	switch kind {
	case pref.BoolKind:
		if docType.Kind() == reflect.Bool {
			return pref.ValueOfBool(vdoc.Bool()), nil
		}

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind:
		switch docType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if i32, err := as.Int32(vdoc.Int()); err == nil {
				return pref.ValueOfInt32(i32), nil
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			if i32, err := as.Int32(vdoc.Uint()); err == nil {
				return pref.ValueOfInt32(i32), nil
			}
		}

	case pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
		switch docType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if i64, err := as.Int64(vdoc.Int()); err == nil {
				return pref.ValueOfInt64(i64), nil
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			if i64, err := as.Int64(vdoc.Uint()); err == nil {
				return pref.ValueOfInt64(i64), nil
			}
		}

	case pref.Uint32Kind, pref.Fixed32Kind:
		switch docType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if ui32, err := as.Uint32(vdoc.Int()); err == nil {
				return pref.ValueOfUint32(ui32), nil
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			if ui32, err := as.Uint32(vdoc.Uint()); err == nil {
				return pref.ValueOfUint32(ui32), nil
			}
		}

	case pref.Uint64Kind, pref.Fixed64Kind:
		switch docType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if ui64, err := as.Uint64(vdoc.Int()); err == nil {
				return pref.ValueOfUint64(ui64), nil
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			if ui64, err := as.Uint64(vdoc.Uint()); err == nil {
				return pref.ValueOfUint64(ui64), nil
			}
		}

	case pref.FloatKind:
		switch docType.Kind() {
		case reflect.Float32, reflect.Float64:
			isInf := math.IsInf(vdoc.Float(), +1) || math.IsInf(vdoc.Float(), -1)
			isSafe := vdoc.Float() == 0 || (math.SmallestNonzeroFloat32 <= math.Abs(vdoc.Float()) && math.Abs(vdoc.Float()) <= math.MaxFloat32)
			if isInf || isSafe {
				return pref.ValueOfFloat32(float32(vdoc.Float())), nil
			}
		}

	case pref.DoubleKind:
		switch docType.Kind() {
		case reflect.Float32, reflect.Float64:
			return pref.ValueOfFloat64(float64(vdoc.Float())), nil
		}

	case pref.StringKind:
		if docType.Kind() == reflect.String {
			if valid := utf8.Valid([]byte(vdoc.String())); !valid {
				return pref.Value{}, fmt.Errorf("invalid UTF-8: %s", vdoc.String())
			}
			return pref.ValueOfString(vdoc.String()), nil
		}

	case pref.BytesKind:
		if binary, ok := doc.(primitive.Binary); ok {
			return pref.ValueOfBytes(binary.Data), nil
		}
	case pref.EnumKind:
		// Check for null value first
		if _, ok := doc.(primitive.Null); ok {
			// This is only valid for google.protobuf.NullValue.
			if isNullValue(fd) {
				return pref.ValueOfEnum(0), nil
			}
		}

		switch docType.Kind() {
		case reflect.String:
			// Lookup EnumNumber based on name.
			if enumVal := fd.Enum().Values().ByName(pref.Name(vdoc.String())); enumVal != nil {
				return pref.ValueOfEnum(enumVal.Number()), nil
			}

		case reflect.Int, reflect.Int32, reflect.Int64:
			if i32, err := as.Int32(vdoc.Int()); err == nil {
				return pref.ValueOfEnum(pref.EnumNumber(i32)), nil
			}
		}

	default:
		panic(fmt.Sprintf("unmarshalScalar: invalid scalar kind %v", kind))
	}
	return pref.Value{}, fmt.Errorf(`invalid value for %v type: %s (has type %T)`, kind, quoted(doc), doc)
}

func quoted(i interface{}) string {
	quoted := fmt.Sprintf(`%v`, i)
	if reflect.TypeOf(i).Kind() == reflect.String {
		quoted = fmt.Sprintf(`"%s"`, quoted)
	}
	return quoted
}

func isNullValue(fd pref.FieldDescriptor) bool {
	ed := fd.Enum()
	return ed != nil && ed.FullName() == genid.NullValue_enum_fullname
}

func isKnownValue(fd pref.FieldDescriptor) bool {
	md := fd.Message()
	return md != nil && md.FullName() == genid.Value_message_fullname
}
