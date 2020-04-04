package bsonpb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
)

// Version is incremented using bump2version
const Version = "0.0.1"

const secondInNanos = int32(time.Second / time.Nanosecond)
const maxSecondsInDuration = 315576000000

// AnyResolver takes a type URL, present in an Any message, and resolves it into
// an instance of the associated message.
type AnyResolver interface {
	Resolve(typeURL string) (proto.Message, error)
}

func defaultResolveAny(typeURL string) (proto.Message, error) {
	// Only the part of typeURL after the last slash is relevant.
	mname := typeURL
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return nil, fmt.Errorf("unknown message type %q", mname)
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

func unquote(s string) (string, error) {
	var ret string
	err := bson.Unmarshal([]byte(s), &ret)
	return ret, err
}

// bsonProperties returns parsed proto.Properties for the field and corrects JSONName attribute.
func bsonProperties(f reflect.StructField, origName bool) *proto.Properties {
	var prop proto.Properties
	prop.Init(f.Type, f.Name, f.Tag.Get("protobuf"), &f)
	if origName || prop.JSONName == "" {
		prop.JSONName = prop.OrigName
	}
	return &prop
}

type fieldNames struct {
	orig, camel string
}

func acceptedBSONFieldNames(prop *proto.Properties) fieldNames {
	opts := fieldNames{orig: prop.OrigName, camel: prop.OrigName}
	if prop.JSONName != "" {
		opts.camel = prop.JSONName
	}
	return opts
}

// Map fields may have key types of non-float scalars, strings and enums.
// The easiest way to sort them in some deterministic order is to use fmt.
// If this turns out to be inefficient we can always consider other options,
// such as doing a Schwartzian transform.
//
// Numeric keys are sorted in numeric order per
// https://developers.google.com/protocol-buffers/docs/proto#maps.
type mapKeys []reflect.Value

func (s mapKeys) Len() int      { return len(s) }
func (s mapKeys) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s mapKeys) Less(i, j int) bool {
	if k := s[i].Kind(); k == s[j].Kind() {
		switch k {
		case reflect.String:
			return s[i].String() < s[j].String()
		case reflect.Int32, reflect.Int64:
			return s[i].Int() < s[j].Int()
		case reflect.Uint32, reflect.Uint64:
			return s[i].Uint() < s[j].Uint()
		}
	}
	return fmt.Sprint(s[i].Interface()) < fmt.Sprint(s[j].Interface())
}

// checkRequiredFields returns an error if any required field in the given proto message is not set.
// This function is used by both Marshal and Unmarshal.  While required fields only exist in a
// proto2 message, a proto3 message can contain proto2 message(s).
func checkRequiredFields(pb proto.Message) error {
	// Most well-known type messages do not contain required fields.  The "Any" type may contain
	// a message that has required fields.
	//
	// When an Any message is being marshaled, the code will invoked proto.Unmarshal on Any.Value
	// field in order to transform that into BSON, and that should have returned an error if a
	// required field is not set in the embedded message.
	//
	// When an Any message is being unmarshaled, the code will have invoked proto.Marshal on the
	// embedded message to store the serialized message in Any.Value field, and that should have
	// returned an error if a required field is not set.
	if _, ok := pb.(wkt); ok {
		return nil
	}

	v := reflect.ValueOf(pb)
	// Skip message if it is not a struct pointer.
	if v.Kind() != reflect.Ptr {
		return nil
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		sfield := v.Type().Field(i)

		if sfield.PkgPath != "" {
			// blank PkgPath means the field is exported; skip if not exported
			continue
		}

		if strings.HasPrefix(sfield.Name, "XXX_") {
			continue
		}

		// Oneof field is an interface implemented by wrapper structs containing the actual oneof
		// field, i.e. an interface containing &T{real_value}.
		if sfield.Tag.Get("protobuf_oneof") != "" {
			if field.Kind() != reflect.Interface {
				continue
			}
			v := field.Elem()
			if v.Kind() != reflect.Ptr || v.IsNil() {
				continue
			}
			v = v.Elem()
			if v.Kind() != reflect.Struct || v.NumField() < 1 {
				continue
			}
			field = v.Field(0)
			sfield = v.Type().Field(0)
		}

		protoTag := sfield.Tag.Get("protobuf")
		if protoTag == "" {
			continue
		}
		var prop proto.Properties
		prop.Init(sfield.Type, sfield.Name, protoTag, &sfield)

		switch field.Kind() {
		case reflect.Map:
			if field.IsNil() {
				continue
			}
			// Check each map value.
			keys := field.MapKeys()
			for _, k := range keys {
				v := field.MapIndex(k)
				if err := checkRequiredFieldsInValue(v); err != nil {
					return err
				}
			}
		case reflect.Slice:
			// Handle non-repeated type, e.g. bytes.
			if !prop.Repeated {
				if prop.Required && field.IsNil() {
					return fmt.Errorf("required field %q is not set", prop.Name)
				}
				continue
			}

			// Handle repeated type.
			if field.IsNil() {
				continue
			}
			// Check each slice item.
			for i := 0; i < field.Len(); i++ {
				v := field.Index(i)
				if err := checkRequiredFieldsInValue(v); err != nil {
					return err
				}
			}
		case reflect.Ptr:
			if field.IsNil() {
				if prop.Required {
					return fmt.Errorf("required field %q is not set", prop.Name)
				}
				continue
			}
			if err := checkRequiredFieldsInValue(field); err != nil {
				return err
			}
		}
	}

	// Handle proto2 extensions.
	for _, ext := range proto.RegisteredExtensions(pb) {
		if !proto.HasExtension(pb, ext) {
			continue
		}
		ep, err := proto.GetExtension(pb, ext)
		if err != nil {
			return err
		}
		err = checkRequiredFieldsInValue(reflect.ValueOf(ep))
		if err != nil {
			return err
		}
	}

	return nil
}

func checkRequiredFieldsInValue(v reflect.Value) error {
	if v.Type().Implements(messageType) {
		return checkRequiredFields(v.Interface().(proto.Message))
	}
	return nil
}

const (
	dynamicMessageName = "google.protobuf.bsonpb.testing.dynamicMessage"
)

func newDynamicMessage(rawBson interface{}) *dynamicMessage {
	b, err := bson.Marshal(rawBson)
	if err != nil {
		panic(err)
	}
	return &dynamicMessage{RawBson: b}
}

func registerDynamicMessage() {
	// we register the custom type below so that we can use it in Any types
	proto.RegisterType((*dynamicMessage)(nil), dynamicMessageName)
}

type ptrFieldMessage struct {
	StringField *stringField `protobuf:"bytes,1,opt,name=stringField"`
}

func (m *ptrFieldMessage) Reset() {
}

func (m *ptrFieldMessage) String() string {
	return m.StringField.StringValue
}

func (m *ptrFieldMessage) ProtoMessage() {
}

type stringField struct {
	IsSet       bool   `protobuf:"varint,1,opt,name=isSet"`
	StringValue string `protobuf:"bytes,2,opt,name=stringValue"`
}

func (s *stringField) Reset() {
}

func (s *stringField) String() string {
	return s.StringValue
}

func (s *stringField) ProtoMessage() {
}

func (s *stringField) UnmarshalBSONPB(jum *Unmarshaler, js []byte) error {
	s.IsSet = true
	s.StringValue = string(js)
	return nil
}

// dynamicMessage implements protobuf.Message but is not a normal generated message type.
// It provides implementations of JSONPBMarshaler and JSONPBUnmarshaler for JSON support.
type dynamicMessage struct {
	// RawBson bson.D `protobuf:"bytes,1,opt,name=rawJson"`
	RawBson []byte `protobuf:"bytes,1,opt,name=rawBson"`

	// an unexported nested message is present just to ensure that it
	// won't result in a panic (see issue #509)
	Dummy *dynamicMessage `protobuf:"bytes,2,opt,name=dummy"`
}

func (m *dynamicMessage) Reset() {
	m.RawBson = []byte{} // bson.D{}
}

func (m *dynamicMessage) String() string {
	return string(m.RawBson)
}

func (m *dynamicMessage) ProtoMessage() {
}

func (m *dynamicMessage) MarshalBSONPB(jm *Marshaler) (interface{}, error) {
	var out bson.D
	err := bson.Unmarshal(m.RawBson, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *dynamicMessage) UnmarshalBSONPB(jum *Unmarshaler, data bson.M) error {
	b, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	m.RawBson = b
	return nil
}
