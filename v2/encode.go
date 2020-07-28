// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsonpb

import (
	"encoding/base64"
	"fmt"
	"errors"

	// "google.golang.org/protobuf/internal/encoding/json"
	"github.com/romnnn/bsonpb/v2/internal/json"
	// "google.golang.org/protobuf/internal/encoding/messageset"
	// "google.golang.org/protobuf/internal/errors"
	// "google.golang.org/protobuf/internal/filedesc"
	// "google.golang.org/protobuf/internal/flags"
	// "google.golang.org/protobuf/internal/genid"
	"github.com/romnnn/bsonpb/v2/internal/genid"
	// "google.golang.org/protobuf/internal/order"
	"github.com/romnnn/bsonpb/v2/internal/order"
	// "google.golang.org/protobuf/internal/pragma"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	pref "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	defaultIndent = "  "
	protoLegacy = false
)

// Format formats the message as a multiline string.
// This function is only intended for human consumption and ignores errors.
// Do not depend on the output being stable. It may change over time across
// different versions of the program.
func Format(m proto.Message) string {
	return MarshalOptions{Multiline: true}.Format(m)
}

// Marshal writes the given proto.Message in JSON format using default options.
// Do not depend on the output being stable. It may change over time across
// different versions of the program.
func Marshal(m proto.Message) ([]byte, error) {
	return MarshalOptions{}.Marshal(m)
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

// Format formats the message as a string.
// This method is only intended for human consumption and ignores errors.
// Do not depend on the output being stable. It may change over time across
// different versions of the program.
func (o MarshalOptions) Format(m proto.Message) string {
	if m == nil || !m.ProtoReflect().IsValid() {
		return "<nil>" // invalid syntax, but okay since this is for debugging
	}
	o.AllowPartial = true
	b, _ := o.Marshal(m)
	return string(b)
}

// Marshal marshals the given proto.Message in the JSON format using options in
// MarshalOptions. Do not depend on the output being stable. It may change over
// time across different versions of the program.
func (o MarshalOptions) Marshal(m proto.Message) ([]byte, error) {
	return o.marshal(m)
}

// marshal is a centralized function that all marshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for marshal that do not go through this.
func (o MarshalOptions) marshal(m proto.Message) ([]byte, error) {
	if o.Multiline && o.Indent == "" {
		o.Indent = defaultIndent
	}
	if o.Resolver == nil {
		o.Resolver = protoregistry.GlobalTypes
	}

	internalEnc, err := json.NewEncoder(o.Indent)
	if err != nil {
		return nil, err
	}

	// Treat nil message interface as an empty message,
	// in which case the output in an empty JSON object.
	if m == nil {
		return []byte("{}"), nil
	}

	enc := encoder{internalEnc, o}
	if err := enc.marshalMessage(m.ProtoReflect(), ""); err != nil {
		return nil, err
	}
	if o.AllowPartial {
		return enc.Bytes(), nil
	}
	return enc.Bytes(), proto.CheckInitialized(m)
}

/*
type (
	File struct {
		fileRaw
		L1 FileL1

		once uint32     // atomically set if L2 is valid
		mu   sync.Mutex // protects L2
		L2   *FileL2
	}
	FileL1 struct {
		Syntax  pref.Syntax
		Path    string
		Package pref.FullName

		Enums      Enums
		Messages   Messages
		Extensions Extensions
		Services   Services
	}
	FileL2 struct {
		Options   func() pref.ProtoMessage
		Imports   FileImports
		Locations SourceLocations
	}
)

func (fd *File) ParentFile() pref.FileDescriptor { return fd }
func (fd *File) Parent() pref.Descriptor         { return nil }
func (fd *File) Index() int                      { return 0 }
func (fd *File) Syntax() pref.Syntax             { return fd.L1.Syntax }
func (fd *File) Name() pref.Name                 { return fd.L1.Package.Name() }
func (fd *File) FullName() pref.FullName         { return fd.L1.Package }
func (fd *File) IsPlaceholder() bool             { return false }
func (fd *File) Options() pref.ProtoMessage {
	if f := fd.lazyInit().Options; f != nil {
		return f()
	}
	return descopts.File
}
func (fd *File) Path() string                          { return fd.L1.Path }
func (fd *File) Package() pref.FullName                { return fd.L1.Package }
func (fd *File) Imports() pref.FileImports             { return &fd.lazyInit().Imports }
func (fd *File) Enums() pref.EnumDescriptors           { return &fd.L1.Enums }
func (fd *File) Messages() pref.MessageDescriptors     { return &fd.L1.Messages }
func (fd *File) Extensions() pref.ExtensionDescriptors { return &fd.L1.Extensions }
func (fd *File) Services() pref.ServiceDescriptors     { return &fd.L1.Services }
func (fd *File) SourceLocations() pref.SourceLocations { return &fd.lazyInit().Locations }
func (fd *File) Format(s fmt.State, r rune)            { descfmt.FormatDesc(s, r, fd) }
func (fd *File) ProtoType(pref.FileDescriptor)         {}
func (fd *File) ProtoInternal(pragma.DoNotImplement)   {}

func (fd *File) lazyInit() *FileL2 {
	if atomic.LoadUint32(&fd.once) == 0 {
		fd.lazyInitOnce()
	}
	return fd.L2
}

func (fd *File) lazyInitOnce() {
	fd.mu.Lock()
	if fd.L2 == nil {
		fd.lazyRawInit() // recursively initializes all L2 structures
	}
	atomic.StoreUint32(&fd.once, 1)
	fd.mu.Unlock()
}

// ProtoLegacyRawDesc is a pseudo-internal API for allowing the v1 code
// to be able to retrieve the raw descriptor.
//
// WARNING: This method is exempt from the compatibility promise and may be
// removed in the future without warning.
func (fd *File) ProtoLegacyRawDesc() []byte {
	return fd.builder.RawDescriptor
}

// GoPackagePath is a pseudo-internal API for determining the Go package path
// that this file descriptor is declared in.
//
// WARNING: This method is exempt from the compatibility promise and may be
// removed in the future without warning.
func (fd *File) GoPackagePath() string {
	return fd.builder.GoPackagePath
}

type Base struct {
	L0 BaseL0
}
type BaseL0 struct {
	FullName   pref.FullName // must be populated
	ParentFile *File         // must be populated
	Parent     pref.Descriptor
	Index      int
}

func (d *Base) Name() pref.Name         { return d.L0.FullName.Name() }
func (d *Base) FullName() pref.FullName { return d.L0.FullName }
func (d *Base) ParentFile() pref.FileDescriptor {
	if d.L0.ParentFile == SurrogateProto2 || d.L0.ParentFile == SurrogateProto3 {
		return nil // surrogate files are not real parents
	}
	return d.L0.ParentFile
}
func (d *Base) Parent() pref.Descriptor             { return d.L0.Parent }
func (d *Base) Index() int                          { return d.L0.Index }
func (d *Base) Syntax() pref.Syntax                 { return d.L0.ParentFile.Syntax() }
func (d *Base) IsPlaceholder() bool                 { return false }
func (d *Base) ProtoInternal(pragma.DoNotImplement) {}

type Field struct {
	Base
	L1 FieldL1
}

type FieldL1 struct {
	Options          func() pref.ProtoMessage
	Number           pref.FieldNumber
	Cardinality      pref.Cardinality // must be consistent with Message.RequiredNumbers
	Kind             pref.Kind
	StringName       stringName
	IsProto3Optional bool // promoted from google.protobuf.FieldDescriptorProto
	IsWeak           bool // promoted from google.protobuf.FieldOptions
	HasPacked        bool // promoted from google.protobuf.FieldOptions
	IsPacked         bool // promoted from google.protobuf.FieldOptions
	HasEnforceUTF8   bool // promoted from google.protobuf.FieldOptions
	EnforceUTF8      bool // promoted from google.protobuf.FieldOptions
	Default          defaultValue
	ContainingOneof  pref.OneofDescriptor // must be consistent with Message.Oneofs.Fields
	Enum             pref.EnumDescriptor
	Message          pref.MessageDescriptor
}
*/

type encoder struct {
	*json.Encoder
	opts MarshalOptions
}

// typeFieldDesc is a synthetic field descriptor used for the "@type" field.
var typeFieldDesc = func() protoreflect.FieldDescriptor {
	// TODO
	var fd Field
	fd.L0.FullName = "@type"
	fd.L0.Index = -1
	fd.L1.Cardinality = protoreflect.Optional
	fd.L1.Kind = protoreflect.StringKind
	/*
	return &protoreflect.FieldDescriptor{
		L0: struct{ FullName string; Index int }{ FullName: "@type", Index: -1 },
		L1: struct{ Cardinality protoreflect.Cardinality; Kind protoreflect.Kind }{ Cardinality: protoreflect.Optional, Kind: protoreflect.StringKind },
	}
	*/
}()

// typeURLFieldRanger wraps a protoreflect.Message and modifies its Range method
// to additionally iterate over a synthetic field for the type URL.
type typeURLFieldRanger struct {
	order.FieldRanger
	typeURL string
}

func (m typeURLFieldRanger) Range(f func(pref.FieldDescriptor, pref.Value) bool) {
	if !f(typeFieldDesc, pref.ValueOfString(m.typeURL)) {
		return
	}
	m.FieldRanger.Range(f)
}

// unpopulatedFieldRanger wraps a protoreflect.Message and modifies its Range
// method to additionally iterate over unpopulated fields.
type unpopulatedFieldRanger struct{ pref.Message }

func (m unpopulatedFieldRanger) Range(f func(pref.FieldDescriptor, pref.Value) bool) {
	fds := m.Descriptor().Fields()
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if m.Has(fd) || fd.ContainingOneof() != nil {
			continue // ignore populated fields and fields within a oneofs
		}

		v := m.Get(fd)
		isProto2Scalar := fd.Syntax() == pref.Proto2 && fd.Default().IsValid()
		isSingularMessage := fd.Cardinality() != pref.Repeated && fd.Message() != nil
		if isProto2Scalar || isSingularMessage {
			v = pref.Value{} // use invalid value to emit null
		}
		if !f(fd, v) {
			return
		}
	}
	m.Message.Range(f)
}

// from https://github.com/protocolbuffers/protobuf-go/blob/master/internal/encoding/messageset/messageset.go
// IsMessageSet returns whether the message uses the MessageSet wire format.
func IsMessageSet(md pref.MessageDescriptor) bool {
	xmd, ok := md.(interface{ IsMessageSet() bool })
	return ok && xmd.IsMessageSet()
}

// marshalMessage marshals the fields in the given protoreflect.Message.
// If the typeURL is non-empty, then a synthetic "@type" field is injected
// containing the URL as the value.
func (e encoder) marshalMessage(m pref.Message, typeURL string) error {
	if !protoLegacy && IsMessageSet(m.Descriptor()) {
		return errors.New("no support for proto1 MessageSets")
	}

	if marshal := wellKnownTypeMarshaler(m.Descriptor().FullName()); marshal != nil {
		return marshal(e, m)
	}

	e.StartObject()
	defer e.EndObject()

	var fields FieldRanger = m
	if e.opts.EmitUnpopulated {
		fields = unpopulatedFieldRanger{m}
	}
	if typeURL != "" {
		fields = typeURLFieldRanger{fields, typeURL}
	}

	var err error
	RangeFields(fields, IndexNameFieldOrder, func(fd pref.FieldDescriptor, v pref.Value) bool {
		name := fd.JSONName()
		fmt.Printf("fd.JSONName()=%v\n", name)
		if e.opts.UseProtoNames {
			name = fd.TextName()
		}

		if err = e.WriteName(name); err != nil {
			return false
		}
		if err = e.marshalValue(v, fd); err != nil {
			return false
		}
		fmt.Printf("e.marshalValue(v, fd)=%v\n", e.marshalValue(v, fd))
		return true
	})
	return err
}

// marshalValue marshals the given protoreflect.Value.
func (e encoder) marshalValue(val pref.Value, fd pref.FieldDescriptor) error {
	switch {
	case fd.IsList():
		return e.marshalList(val.List(), fd)
	case fd.IsMap():
		return e.marshalMap(val.Map(), fd)
	default:
		return e.marshalSingular(val, fd)
	}
}

// marshalSingular marshals the given non-repeated field value. This includes
// all scalar types, enums, messages, and groups.
func (e encoder) marshalSingular(val pref.Value, fd pref.FieldDescriptor) error {
	if !val.IsValid() {
		e.WriteNull()
		return nil
	}

	switch kind := fd.Kind(); kind {
	case pref.BoolKind:
		e.WriteBool(val.Bool())

	case pref.StringKind:
		if e.WriteString(val.String()) != nil {
			return errors.New("Invalid UTF8: %s", string(fd.FullName()))
		}

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind:
		e.WriteInt(val.Int())

	case pref.Uint32Kind, pref.Fixed32Kind:
		e.WriteUint(val.Uint())

	case pref.Int64Kind, pref.Sint64Kind, pref.Uint64Kind,
		pref.Sfixed64Kind, pref.Fixed64Kind:
		// 64-bit integers are written out as JSON string.
		e.WriteString(val.String())

	case pref.FloatKind:
		// Encoder.WriteFloat handles the special numbers NaN and infinites.
		e.WriteFloat(val.Float(), 32)

	case pref.DoubleKind:
		// Encoder.WriteFloat handles the special numbers NaN and infinites.
		e.WriteFloat(val.Float(), 64)

	case pref.BytesKind:
		e.WriteString(base64.StdEncoding.EncodeToString(val.Bytes()))

	case pref.EnumKind:
		if fd.Enum().FullName() == genid.NullValue_enum_fullname {
			e.WriteNull()
		} else {
			desc := fd.Enum().Values().ByNumber(val.Enum())
			if e.opts.UseEnumNumbers || desc == nil {
				e.WriteInt(int64(val.Enum()))
			} else {
				e.WriteString(string(desc.Name()))
			}
		}

	case pref.MessageKind, pref.GroupKind:
		if err := e.marshalMessage(val.Message(), ""); err != nil {
			return err
		}

	default:
		panic(fmt.Sprintf("%v has unknown kind: %v", fd.FullName(), kind))
	}
	return nil
}

// marshalList marshals the given protoreflect.List.
func (e encoder) marshalList(list pref.List, fd pref.FieldDescriptor) error {
	e.StartArray()
	defer e.EndArray()

	for i := 0; i < list.Len(); i++ {
		item := list.Get(i)
		if err := e.marshalSingular(item, fd); err != nil {
			return err
		}
	}
	return nil
}

// marshalMap marshals given protoreflect.Map.
func (e encoder) marshalMap(mmap pref.Map, fd pref.FieldDescriptor) error {
	e.StartObject()
	defer e.EndObject()

	var err error
	RangeEntries(mmap, GenericKeyOrder, func(k pref.MapKey, v pref.Value) bool {
		if err = e.WriteName(k.String()); err != nil {
			return false
		}
		if err = e.marshalSingular(v, fd.MapValue()); err != nil {
			return false
		}
		return true
	})
	return err
}