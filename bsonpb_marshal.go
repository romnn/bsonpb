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
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Marshaler is a configurable object for converting between
// protocol buffer objects and a BSON representation for them.
type Marshaler struct {
	// Whether to render enum values as integers, as opposed to string values.
	EnumsAsInts bool

	// Whether to use the original (.proto) name for fields.
	OrigName bool

	// Whether to use the original (.proto) name for fields.
	Omit OmitOptions

	// A custom URL resolver to use when marshaling Any messages to BSON.
	// If unset, the default resolution strategy is to extract the
	// fully-qualified type name from the type URL and pass that to
	// proto.MessageType(string).
	AnyResolver AnyResolver
}

// Whether to use the original (.proto) name for fields.
type OmitOptions struct {
	All      bool
	Bools    bool
	Ints     bool
	UInts    bool
	Floats   bool
	Strings  bool
	Maps     bool
	Pointers bool
	Slices   bool
}

// BSONPBMarshaler is implemented by protobuf messages that customize the
// way they are marshaled to BSON. Messages that implement this should
// also implement BSONPBUnmarshaler so that the custom format can be
// parsed.
//
// The BSON marshaling must follow the proto to BSON specification
type BSONPBMarshaler interface {
	MarshalBSONPB(*Marshaler) (interface{}, error)
}

// Marshal marshals a protocol buffer into BSON.
func (m *Marshaler) Marshal(pb proto.Message) (interface{}, error) {

	v := reflect.ValueOf(pb)
	if pb == nil || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return nil, errors.New("Marshal called with nil")
	}

	// Check for unset required fields first.
	if err := checkRequiredFields(pb); err != nil {
		return nil, err
	}

	marshaled, err := m.marshalObject(pb, "")
	if err != nil {
		return nil, err
	}
	if d, ok := marshaled.(bson.D); ok {
		return d, nil
	}
	return nil, fmt.Errorf("%v is not a valid bson document as it has no top level map!")
}

type int32Slice []int32

// For sorting extensions ids to ensure stable output.
func (s int32Slice) Len() int           { return len(s) }
func (s int32Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s int32Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type wkt interface {
	XXX_WellKnownType() string
}

var (
	wktType     = reflect.TypeOf((*wkt)(nil)).Elem()
	messageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

func (m *Marshaler) marshalObject(v proto.Message, typeURL string) (interface{}, error) {
	out := bson.D{}
	if jsm, ok := v.(BSONPBMarshaler); ok {
		marshaled, err := jsm.MarshalBSONPB(m)
		if err != nil {
			return nil, err
		}
		if typeURL != "" {
			// we are marshaling this object to an Any type
			// TODO: How to sort the keys again?
			turl, err := m.marshalTypeURL(typeURL)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal type URL %q to BSON: %v", typeURL, err)
			}

			if _marshaled, ok := marshaled.(bson.D); ok {
				marshaled = append(_marshaled, turl)
			}
		}

		return marshaled, nil

	}

	s := reflect.ValueOf(v).Elem()

	// Handle well-known types.
	if wkt, ok := v.(wkt); ok {
		switch wkt.XXX_WellKnownType() {
		case "DoubleValue", "FloatValue", "Int64Value", "UInt64Value",
			"Int32Value", "UInt32Value", "BoolValue", "StringValue", "BytesValue":
			// "Wrappers use the same representation in BSON
			//  as the wrapped primitive type, ..."

			sprop := proto.GetProperties(s.Type())
			marshaledValue, err := m.marshalValue(sprop.Prop[0], s.Field(0))
			if err != nil {
				return nil, err
			}
			return marshaledValue, err

		case "Any":
			// Any is a bit more involved.
			return m.marshalAny(v)

		case "Duration":
			sdur, ok := v.(*duration.Duration)
			if !ok {
				return nil, errors.New("Not a duration")
			}
			sec, ns := sdur.Seconds, sdur.Nanos
			if sec < -maxSecondsInDuration || sec > maxSecondsInDuration {
				return nil, fmt.Errorf("seconds out of range %v", s)
			}
			if ns <= -secondInNanos || ns >= secondInNanos {
				return nil, fmt.Errorf("ns out of range (%v, %v)", -secondInNanos, secondInNanos)
			}
			if (sec > 0 && ns < 0) || (sec < 0 && ns > 0) {
				return nil, errors.New("signs of seconds and nanos do not match")
			}
			native, err := ptypes.Duration(sdur)
			if err != nil {
				return nil, err
			}
			converted := native.Seconds()
			return converted, nil

		case "Struct", "ListValue":
			// Let marshalValue handle the `Struct.fields` map or the `ListValue.values` slice.
			// TODO: pass the correct Properties if needed.
			return m.marshalValue(&proto.Properties{}, s.Field(0))
		case "Timestamp":
			// Convert proto timestamp to golang time.Time
			vv, ok := v.(*timestamp.Timestamp)
			if !ok {
				return nil, fmt.Errorf("Timestamp: Have %v but need a protobuf timestamp", reflect.TypeOf(v))
			}
			nativeTime, err := ptypes.Timestamp(vv)
			if err != nil {
				return nil, err
			}
			return primitive.NewDateTimeFromTime(nativeTime), nil

		case "Value":
			// Value has a single oneof.
			kind := s.Field(0)
			if kind.IsNil() {
				// "absence of any variant indicates an error"
				return nil, errors.New("nil Value")
			}
			// oneof -> *T -> T -> T.F
			x := kind.Elem().Elem().Field(0)
			// TODO: pass the correct Properties if needed.
			return m.marshalValue(&proto.Properties{}, x)
		}
	}

	if typeURL != "" {
		tURL, err := m.marshalTypeURL(typeURL)
		if err != nil {
			return nil, err
		}
		out = append(out, tURL)
	}

	// Handling struct fields
	for i := 0; i < s.NumField(); i++ {
		value := s.Field(i)
		valueField := s.Type().Field(i)
		if strings.HasPrefix(valueField.Name, "XXX_") {
			continue
		}

		// IsNil will panic on most value kinds.
		switch value.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface:
			if value.IsNil() {
				continue
			}
		}

		if m.Omit != (OmitOptions{}) {
			switch value.Kind() {
			case reflect.Bool:
				if !value.Bool() && (m.Omit.All || m.Omit.Bools) {
					continue
				}
			case reflect.Int32, reflect.Int64:
				if value.Int() == 0 && (m.Omit.All || m.Omit.Ints) {
					continue
				}
			case reflect.Uint32, reflect.Uint64:
				if value.Uint() == 0 && (m.Omit.All || m.Omit.UInts) {
					continue
				}
			case reflect.Float32, reflect.Float64:
				if value.Float() == 0 && (m.Omit.All || m.Omit.Floats) {
					continue
				}
			case reflect.String:
				if value.Len() == 0 && (m.Omit.All || m.Omit.Strings) {
					continue
				}
			case reflect.Map:
				if value.IsNil() && (m.Omit.All || m.Omit.Maps) {
					continue
				}
			case reflect.Ptr:
				if value.IsNil() && (m.Omit.All || m.Omit.Pointers) {
					continue
				}
			case reflect.Slice:
				if value.IsNil() && (m.Omit.All || m.Omit.Slices) {
					continue
				}
			}
		}

		// Oneof fields need special handling.
		if valueField.Tag.Get("protobuf_oneof") != "" {
			// value is an interface containing &T{real_value}.
			sv := value.Elem().Elem() // interface -> *T -> T
			value = sv.Field(0)
			valueField = sv.Type().Field(0)
		}
		prop := bsonProperties(valueField, m.OrigName)
		marshaledValue, err := m.marshalValue(prop, value)
		if err != nil {
			return out, err
		}
		out = append(out, bson.E{prop.JSONName, marshaledValue})
	}

	// Handle proto2 extensions.
	if ep, ok := v.(proto.Message); ok {
		extensions := proto.RegisteredExtensions(v)
		// Sort extensions for stable output.
		ids := make([]int32, 0, len(extensions))
		for id, desc := range extensions {
			if !proto.HasExtension(ep, desc) {
				continue
			}
			ids = append(ids, id)
		}
		sort.Sort(int32Slice(ids))
		for _, id := range ids {
			desc := extensions[id]
			if desc == nil {
				// unknown extension
				continue
			}
			ext, extErr := proto.GetExtension(ep, desc)
			if extErr != nil {
				return nil, extErr
			}
			value := reflect.ValueOf(ext)
			var prop proto.Properties
			prop.Parse(desc.Tag)
			prop.JSONName = fmt.Sprintf("[%s]", desc.Name)
			marshaledField, err := m.marshalField(&prop, value)
			if err != nil {
				return nil, err
			}
			out = append(out, marshaledField)
		}
	}

	return out, nil
}

func (m *Marshaler) marshalToSafeString(v reflect.Value) (string, error) {
	// TODO: Use JSONPB here?
	var s string
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return "", err
	}
	s = string(b)

	// Fix double quotes for terminal values not wrapped in {} or []
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		s = s[1 : len(s)-1]
	}
	return s, nil
}

func (m *Marshaler) marshalAny(any proto.Message) (interface{}, error) {
	// "If the Any contains a value that has a special BSON mapping,
	//  it will be converted as follows: {"@type": xxx, "value": yyy}.
	//  Otherwise, the value will be converted into a BSON object,
	//  and the "@type" field will be inserted to indicate the actual data type."
	v := reflect.ValueOf(any).Elem()
	turl := v.Field(0).String()
	val := v.Field(1).Bytes()

	var msg proto.Message
	var err error
	if m.AnyResolver != nil {
		msg, err = m.AnyResolver.Resolve(turl)
	} else {
		msg, err = defaultResolveAny(turl)
	}
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(val, msg); err != nil {
		return nil, err
	}

	// Is well known
	if _, ok := msg.(wkt); ok {
		out := bson.D{}
		marshaledTURL, err := m.marshalTypeURL(turl)
		if err != nil {
			return nil, err
		}
		out = append(out, marshaledTURL)

		marshaledObj, err := m.marshalObject(msg, "")
		if err != nil {
			return nil, err
		}
		out = append(out, bson.E{"value", marshaledObj})
		return out, nil
	}

	return m.marshalObject(msg, turl)
}

func (m *Marshaler) marshalTypeURL(typeURL string) (bson.E, error) {
	turl, err := m.marshalToSafeString(reflect.ValueOf(typeURL))
	if err != nil {
		return bson.E{}, err
	}
	return bson.E{"@type", turl}, nil
}

// marshalField writes field description and value to the Writer.
func (m *Marshaler) marshalField(prop *proto.Properties, v reflect.Value) (bson.E, error) {
	marshaledValue, err := m.marshalValue(prop, v)
	if err != nil {
		return bson.E{}, err
	}
	return bson.E{prop.JSONName, marshaledValue}, nil
}

// marshalValue writes the value to the Writer.
func (m *Marshaler) marshalValue(prop *proto.Properties, v reflect.Value) (interface{}, error) {
	v = reflect.Indirect(v)

	// Handle nil pointer
	if v.Kind() == reflect.Invalid {
		return primitive.Null{}, nil
	}

	// Handle repeated elements.
	if v.Kind() == reflect.Slice {
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Handle raw bytes
			if b, ok := v.Interface().([]byte); ok {
				return primitive.Binary{Data: b}, nil
			} else {
				return nil, fmt.Errorf("Is not bytes")
			}
		} else {
			repeated := bson.A{}
			for i := 0; i < v.Len(); i++ {
				sliceVal := v.Index(i)
				marshaled, err := m.marshalValue(prop, sliceVal)
				if err != nil {
					return repeated, err
				}
				repeated = append(repeated, marshaled)
			}
			return repeated, nil
		}
	}

	// Handle well-known types.
	// Most are handled up in marshalObject (because 99% are messages).
	if v.Type().Implements(wktType) {
		wkt := v.Interface().(wkt)
		switch wkt.XXX_WellKnownType() {
		case "NullValue":
			return primitive.Null{}, nil
		}
	}

	// Handle enumerations.
	if !m.EnumsAsInts && prop.Enum != "" {
		// Unknown enum values will are stringified by the proto library as their
		// value. Such values should _not_ be quoted or they will be interpreted
		// as an enum string instead of their value.

		// Assume the following enum:
		/*
			type Widget_Color int32

			const (
				Widget_RED   Widget_Color = 0
				Widget_GREEN Widget_Color = 1
				Widget_BLUE  Widget_Color = 2
			)
		*/

		// (valStr, enumStr) for Widget(0) (does exist) would be (0, "RED")
		// (valStr, enumStr) for Widget(1000) (does NOT exist) would be (0, "0")
		var valStr string
		if v.Kind() == reflect.Ptr {
			valStr = strconv.Itoa(int(v.Elem().Int()))
		} else {
			valStr = strconv.Itoa(int(v.Int()))
		}
		enumStr := v.Interface().(fmt.Stringer).String()
		isValidEnum := enumStr != valStr
		if isValidEnum {
			return enumStr, nil
		} else {
			return v.Interface(), nil
		}
	}

	// Handle nested messages.
	if v.Kind() == reflect.Struct {
		return m.marshalObject(v.Addr().Interface().(proto.Message), "")
	}

	// Handle maps.
	// Since Go randomizes map iteration, we sort keys for stable output.
	if v.Kind() == reflect.Map {
		bsonMap := bson.D{}
		keys := v.MapKeys()
		sort.Sort(mapKeys(keys))
		for _, k := range keys {
			// TODO handle map key prop properly

			s := "undefined"

			// Check for string interface first
			ownKeyStr, ok := k.Interface().(fmt.Stringer)
			if ok {
				s = ownKeyStr.String()
			} else if ss, err := m.marshalToSafeString(k); err == nil {
				s = ss
			} else {
				return bsonMap, fmt.Errorf("Cannot marshal key %v to string", k.Interface())
			}

			vprop := prop
			if prop != nil && prop.MapValProp != nil {
				vprop = prop.MapValProp
			}
			marshaled, err := m.marshalValue(vprop, v.MapIndex(k))
			if err != nil {
				return bsonMap, err
			}
			bsonMap = append(bsonMap, bson.E{s, marshaled})
		}
		return bsonMap, nil
	}

	return v.Interface(), nil
}
