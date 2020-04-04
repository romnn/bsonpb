/*
Package bsonpb provides marshaling and unmarshaling between protocol buffers and BSON.
It follows the specification at https://developers.google.com/protocol-buffers/docs/proto3#bson.
This package produces a different output than the standard "encoding/bson" package,
which does not operate correctly on protocol buffers.
*/
package bsonpb

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	stpb "github.com/golang/protobuf/ptypes/struct"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BSONPBUnmarshaler is implemented by protobuf messages that customize
// the way they are unmarshaled from BSON. Messages that implement this
// should also implement BSONPBMarshaler so that the custom format can be
// produced.
//
// The BSON unmarshaling must follow the BSON to proto specification:
type BSONPBUnmarshaler interface {
	// UnmarshalBSONPB(*Unmarshaler, bson.D, proto.Message) (protoerror // bson.D
	UnmarshalBSONPB(*Unmarshaler, bson.M) error
}

// Unmarshaler is a configurable object for converting from a BSON
// representation to a protocol buffer object.
type Unmarshaler struct {
	// Whether to allow messages to contain unknown fields, as opposed to
	// failing to unmarshal.
	AllowUnknownFields bool

	// A custom URL resolver to use when unmarshaling Any messages from BSON.
	// If unset, the default resolution strategy is to extract the
	// fully-qualified type name from the type URL and pass that to
	// proto.MessageType(string).
	AnyResolver AnyResolver

	dec *bson.Decoder
}

// UnmarshalNext unmarshals the next protocol buffer from a BSON object stream.
// This function is lenient and will decode any options permutations of the
// related Marshaler.
func (u *Unmarshaler) UnmarshalNext(pb proto.Message) error {
	var inputValue bson.M
	if err := u.dec.Decode(&inputValue); err != nil {
		return err
	}
	if err := u.unmarshalValue(reflect.ValueOf(pb).Elem(), inputValue, nil); err != nil {
		return err
	}
	return checkRequiredFields(pb)
}

// Unmarshal unmarshals a BSON object stream into a protocol
// buffer. This function is lenient and will decode any options
// permutations of the related Marshaler.
func (u *Unmarshaler) Unmarshal(data []byte, pb proto.Message) error {
	reader := bsonrw.NewBSONDocumentReader(data)
	dec, err := bson.NewDecoder(reader)
	if err != nil {
		return fmt.Errorf("Failed to create decoder for BSON stream: %s", err.Error())
	}
	u.dec = dec
	return u.UnmarshalNext(pb)
}

// UnmarshalBSON unmarshals a BSON
func (u *Unmarshaler) UnmarshalBSON(data bson.D, pb proto.Message) error {
	raw, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	return u.Unmarshal(raw, pb)
}

// UnmarshalNext unmarshals the next protocol buffer from a BSON object stream.
// This function is lenient and will decode any options permutations of the
// related Marshaler.
func UnmarshalNext(dec *bson.Decoder, pb proto.Message) error {
	u := &Unmarshaler{dec: dec}
	return u.UnmarshalNext(pb)
}

// Unmarshal unmarshals a BSON object stream into a protocol
// buffer. This function is lenient and will decode any options
// permutations of the related Marshaler.
func Unmarshal(data []byte, pb proto.Message) error {
	return new(Unmarshaler).Unmarshal(data, pb)
}

// UnmarshalBSON unmarshals a BSON object stream into a protocol
// buffer. This function is lenient and will decode any options
// permutations of the related Marshaler.
func UnmarshalBSON(data bson.D, pb proto.Message) error {
	return new(Unmarshaler).UnmarshalBSON(data, pb)
}

func (u *Unmarshaler) unmarshalFromSafeString(v string, dest interface{}) error {
	err := json.Unmarshal([]byte(v), dest)
	if err != nil {
		return err
	}
	return nil
}

func consumeField(prop *proto.Properties, bsonFields bson.M) (interface{}, string, bool) {
	// Be liberal in what names we accept; both orig_name and camelName are okay.
	fieldNames := acceptedBSONFieldNames(prop)

	vOrig, okOrig := bsonFields[fieldNames.orig]
	vCamel, okCamel := bsonFields[fieldNames.camel]
	if !okOrig && !okCamel {
		return nil, "", false
	}
	// If, for some reason, both are present in the data, favour the camelName.
	// var raw interface{} // bson.RawMessage
	if okOrig {
		return vOrig, fieldNames.orig, true
	}
	if okCamel {
		return vCamel, fieldNames.camel, true
	}
	return nil, "", false
}

// unmarshalValue converts/copies a value into the target.
// prop may be nil.
func (u *Unmarshaler) unmarshalValue(target reflect.Value, inputValue interface{}, prop *proto.Properties) error {

	targetType := target.Type()

	// Allocate memory for pointer fields.
	if targetType.Kind() == reflect.Ptr {
		// If input value is "null" and target is a pointer type, then the field should be treated as not set
		// UNLESS the target is structpb.Value, in which case it should be set to structpb.NullValue.
		_, isBSONPBUnmarshaler := target.Interface().(BSONPBUnmarshaler)
		if (inputValue == primitive.Null{}) && targetType != reflect.TypeOf(&stpb.Value{}) && !isBSONPBUnmarshaler {
			return nil
		}

		target.Set(reflect.New(targetType.Elem()))
		return u.unmarshalValue(target.Elem(), inputValue, prop)
	}

	if customUnmarshaler, ok := target.Addr().Interface().(BSONPBUnmarshaler); ok {
		inputD, ok := inputValue.(bson.M)
		if !ok {
			return errors.New("Not a bson.M")
		}
		return customUnmarshaler.UnmarshalBSONPB(u, inputD)
	}

	// Handle well-known types that are not pointers.
	if w, ok := target.Addr().Interface().(wkt); ok {
		switch w.XXX_WellKnownType() {
		case "DoubleValue", "FloatValue", "Int64Value", "UInt64Value",
			"Int32Value", "UInt32Value", "BoolValue", "StringValue", "BytesValue":
			return u.unmarshalValue(target.Field(0), inputValue, prop)
		case "Any":
			d, dok := inputValue.(bson.M)
			if !dok {
				return fmt.Errorf("Any: Have %v but need a map (bson.M)", reflect.TypeOf(inputValue))
			}
			val, tok := d["@type"]
			if !tok || val == nil {
				return errors.New("Any BSON doesn't have '@type'")
			}
			turl, sok := val.(string)
			if !sok {
				return errors.New("@type is not a string")
			}

			target.Field(0).SetString(turl)

			var m proto.Message
			var err error
			if u.AnyResolver != nil {
				m, err = u.AnyResolver.Resolve(turl)
			} else {
				m, err = defaultResolveAny(turl)
			}
			if err != nil {
				return err
			}

			if _, ok := m.(wkt); ok {
				val, ok := d["value"]
				if !ok {
					return errors.New("Any BSON doesn't have 'value'")
				}

				if err := u.unmarshalValue(reflect.ValueOf(m).Elem(), val, nil); err != nil {
					return fmt.Errorf("can't unmarshal Any nested proto %T: %v", m, err)
				}
			} else {
				delete(d, "@type")
				if err = u.unmarshalValue(reflect.ValueOf(m).Elem(), d, nil); err != nil {
					return fmt.Errorf("can't unmarshal Any nested proto %T: %v", m, err)
				}
			}

			b, err := proto.Marshal(m)
			if err != nil {
				return fmt.Errorf("can't marshal proto %T into Any.Value: %v", m, err)
			}
			target.Field(1).SetBytes(b)

			return nil
		case "Duration":
			floatSec, ok := inputValue.(float64)
			if !ok {
				return errors.New("Not a valid duration (float64 seconds)")
			}

			dur := time.Duration(floatSec * float64(time.Second))
			durProto := ptypes.DurationProto(dur)
			target.Set(reflect.ValueOf(durProto).Elem())
			return nil

		case "Timestamp":
			datetime, ok := inputValue.(primitive.DateTime)
			if !ok {
				return errors.New("Not a BSON datetime")
			}
			t := datetime.Time()
			target.Field(0).SetInt(t.Unix())
			target.Field(1).SetInt(int64(t.Nanosecond()))
			return nil

		case "Struct":
			m, ok := inputValue.(bson.M)
			if !ok {
				return fmt.Errorf("Struct: Have %v but need a map (bson.M)", reflect.TypeOf(inputValue))
			}
			if len(m) < 1 {
				return nil
			}
			target.Field(0).Set(reflect.ValueOf(map[string]*stpb.Value{}))
			for k, jv := range m {
				pv := &stpb.Value{}
				if err := u.unmarshalValue(reflect.ValueOf(pv).Elem(), jv, prop); err != nil {
					return fmt.Errorf("bad value in StructValue for key %q: %v", k, err)
				}
				target.Field(0).SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(pv))
			}
			return nil
		case "ListValue":
			a, ok := inputValue.(bson.A)
			if !ok {
				return fmt.Errorf("ListValue: Have %v but need an array (bson.A)", reflect.TypeOf(inputValue))
			}
			if len(a) < 1 {
				return nil
			}
			target.Field(0).Set(reflect.ValueOf(make([]*stpb.Value, len(a))))
			for i, av := range a {
				if err := u.unmarshalValue(target.Field(0).Index(i), av, prop); err != nil {
					return err
				}
			}
			return nil

		case "Value":
			inputVal := reflect.ValueOf(inputValue)
			switch inputVal.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				if !inputVal.Type().ConvertibleTo(reflect.TypeOf(float64(1))) {
					return errors.New("Cannot convert to number value")
				}
				converted := inputVal.Convert(reflect.TypeOf(float64(1)))
				target.Field(0).Set(reflect.ValueOf(&stpb.Value_NumberValue{converted.Float()}))
				return nil
			case reflect.Bool:
				target.Field(0).Set(reflect.ValueOf(&stpb.Value_BoolValue{inputVal.Bool()}))
				return nil
			case reflect.String:
				target.Field(0).Set(reflect.ValueOf(&stpb.Value_StringValue{inputVal.String()}))
				return nil
			case reflect.Invalid:
				target.Field(0).Set(reflect.ValueOf(&stpb.Value_NullValue{}))
				return nil
			case reflect.Map:
				if m, ok := inputValue.(bson.M); ok {
					sv := &stpb.Struct{}
					target.Field(0).Set(reflect.ValueOf(&stpb.Value_StructValue{sv}))
					return u.unmarshalValue(reflect.ValueOf(sv).Elem(), m, prop)
				}
				return fmt.Errorf("invalid type %v for map (should be bson.M)", inputVal.Type())
			case reflect.Slice:
				if a, ok := inputValue.(bson.A); ok {
					lv := &stpb.ListValue{}
					target.Field(0).Set(reflect.ValueOf(&stpb.Value_ListValue{lv}))
					return u.unmarshalValue(reflect.ValueOf(lv).Elem(), a, prop)
				}
				return fmt.Errorf("invalid type %v for slice (should be bson.A)", inputVal.Type())
			default:
				return fmt.Errorf("unrecognized type for Value %v (was %v)", inputValue, inputVal.Type())
			}
		}
	}

	// Handle special BSON Types
	if binary, ok := inputValue.(primitive.Binary); ok {
		if targetType.Kind() == reflect.Slice && targetType.Elem().Kind() == reflect.Uint8 {
			target.SetBytes(binary.Data)
			return nil
		} else {
			return errors.New("Cannot set binary to non byte fields")
		}
	}

	// Handle enums, which have an underlying type of int32,
	// and may appear as strings.
	// The case of an enum appearing as a number is handled
	// at the bottom of this function.
	// enumString[0] == '"'
	if enumString, ok := inputValue.(string); ok && prop != nil && prop.Enum != "" {
		vmap := proto.EnumValueMap(prop.Enum)
		// Don't need to do unquoting; valid enum names
		// are from a limited character set.
		n, ok := vmap[enumString]
		if !ok {
			return fmt.Errorf("Unknown value %q for enum %s", enumString, prop.Enum)
		}
		if target.Kind() == reflect.Ptr { // proto2
			target.Set(reflect.New(targetType.Elem()))
			target = target.Elem()
		}
		if targetType.Kind() != reflect.Int32 {
			return fmt.Errorf("Invalid target %q for enum %s", targetType.Kind(), prop.Enum)
		}
		target.SetInt(int64(n))
		return nil
	}

	// Handle nested messages.
	if targetType.Kind() == reflect.Struct {
		bsonFields, ok := inputValue.(bson.M)
		if !ok {
			return fmt.Errorf("Nested: Have %v but need a map (bson.M)", reflect.TypeOf(inputValue))
		}

		sprops := proto.GetProperties(targetType)
		for i := 0; i < target.NumField(); i++ {
			ft := target.Type().Field(i)
			if strings.HasPrefix(ft.Name, "XXX_") {
				continue
			}

			valueForField, remove, ok := consumeField(sprops.Prop[i], bsonFields)
			if !ok {
				continue
			}
			delete(bsonFields, remove)
			val := reflect.ValueOf(valueForField)

			if val.Kind() != reflect.Invalid || target.Field(i).Type() == reflect.TypeOf(&stpb.Value{}) {
				if err := u.unmarshalValue(target.Field(i), valueForField, sprops.Prop[i]); err != nil {
					return err
				}
			}
		}
		// Check for any oneof fields.
		if len(bsonFields) > 0 {
			for _, oop := range sprops.OneofTypes {
				raw, remove, ok := consumeField(oop.Prop, bsonFields)
				if !ok {
					continue
				}
				delete(bsonFields, remove)
				nv := reflect.New(oop.Type.Elem())
				target.Field(oop.Field).Set(nv)
				if err := u.unmarshalValue(nv.Elem().Field(0), raw, oop.Prop); err != nil {
					return err
				}
			}
		}
		// Handle proto2 extensions.
		if len(bsonFields) > 0 {
			if ep, ok := target.Addr().Interface().(proto.Message); ok {
				for _, ext := range proto.RegisteredExtensions(ep) {
					name := fmt.Sprintf("[%s]", ext.Name)
					raw, ok := bsonFields[name]
					if !ok {
						continue
					}
					delete(bsonFields, name)
					nv := reflect.New(reflect.TypeOf(ext.ExtensionType).Elem())
					if err := u.unmarshalValue(nv.Elem(), raw, nil); err != nil {
						return err
					}
					if err := proto.SetExtension(ep, ext, nv.Interface()); err != nil {
						return err
					}
				}
			}
		}
		if !u.AllowUnknownFields && len(bsonFields) > 0 {
			// Pick any field to be the scapegoat.
			var f string
			for fname := range bsonFields {
				f = fname
				break
			}
			return fmt.Errorf("unknown field %q in %v", f, targetType)
		}
		return nil
	}

	// Handle arrays (which aren't encoded bytes)
	if targetType.Kind() == reflect.Slice && targetType.Elem().Kind() != reflect.Uint8 {
		slc, ok := inputValue.(bson.A)
		if !ok {
			return fmt.Errorf("Slice: Have %v but need an array (bson.A)", reflect.TypeOf(inputValue))
		}

		l := len(slc)
		target.Set(reflect.MakeSlice(targetType, l, l))
		for i := 0; i < l; i++ {
			if err := u.unmarshalValue(target.Index(i), slc[i], prop); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle maps (whose keys are always strings)
	if targetType.Kind() == reflect.Map {
		mp, ok := inputValue.(bson.M)
		if !ok {
			return fmt.Errorf("Have %v but need a map (bson.M)", reflect.TypeOf(inputValue))
		}

		target.Set(reflect.MakeMap(targetType))
		for ks, raw := range mp {
			// Unmarshal map key. The core json library already decoded the key into a
			// string, so we handle that specially. Other types were quoted post-serialization.
			var k reflect.Value
			if targetType.Key().Kind() == reflect.String {
				k = reflect.ValueOf(ks)
			} else {
				kk := reflect.New(targetType.Key()).Interface() // .Elem()
				err := u.unmarshalFromSafeString(ks, &kk)
				if err != nil {
					return err
				}
				k = reflect.ValueOf(kk).Elem()
			}
			// Unmarshal map value.
			v := reflect.New(targetType.Elem()).Elem()
			var vprop *proto.Properties
			if prop != nil && prop.MapValProp != nil {
				vprop = prop.MapValProp
			}
			if err := u.unmarshalValue(v, raw, vprop); err != nil {
				return err
			}
			target.SetMapIndex(k, v)
		}
		return nil
	}

	/*
		The following special rules apply
		1. time.Time marshals to a BSON datetime.
		2. int8, int16, and int32 marshal to a BSON int32.
		3. int marshals to a BSON int32 if the value is between math.MinInt32 and math.MaxInt32, inclusive, and a BSON int64
		otherwise.
		4. int64 marshals to BSON int64.
		5. uint8 and uint16 marshal to a BSON int32.
		6. uint, uint32, and uint64 marshal to a BSON int32 if the value is between math.MinInt32 and math.MaxInt32,
		inclusive, and BSON int64 otherwise.
		7. BSON null values will unmarshal into the zero value of a field (e.g. unmarshalling a BSON null value into a string
		will yield the empty string.).
	*/

	// Check before converting BSON int32, int64 and 128 bit
	if isOneOfType(target.Kind(), []reflect.Kind{
		reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16}) {
		return fmt.Errorf("Not allowed as BSON does not support numbers with less than 32 bits")
	}

	if reflect.TypeOf(inputValue).ConvertibleTo(targetType) {
		target.Set(reflect.ValueOf(inputValue).Convert(targetType))
		return nil
	}

	return fmt.Errorf("Cannot fit %v into %v", reflect.TypeOf(inputValue), targetType)

	return nil
}

func isOneOfType(t reflect.Kind, types []reflect.Kind) bool {
	for _, tt := range types {
		if t == tt {
			return true
		}
	}
	return false
}
