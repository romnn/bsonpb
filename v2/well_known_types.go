package bsonpb

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/romnnn/bsonpb/v2/internal/genid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
)

type marshalFunc func(encoder, pref.Message) (interface{}, error)

// wellKnownTypeMarshaler returns a marshal function if the message type
// has specialized serialization behavior. It returns nil otherwise.
func wellKnownTypeMarshaler(name pref.FullName) marshalFunc {
	if name.Parent() == genid.GoogleProtobuf_package {
		switch name.Name() {
		case genid.Any_message_name:
			return encoder.marshalAny
		case genid.Timestamp_message_name:
			return encoder.marshalTimestamp
		case genid.Duration_message_name:
			return encoder.marshalDuration
		case genid.BoolValue_message_name,
			genid.Int32Value_message_name,
			genid.Int64Value_message_name,
			genid.UInt32Value_message_name,
			genid.UInt64Value_message_name,
			genid.FloatValue_message_name,
			genid.DoubleValue_message_name,
			genid.StringValue_message_name,
			genid.BytesValue_message_name:
			return encoder.marshalWrapperType
		case genid.Struct_message_name:
			return encoder.marshalStruct
		case genid.ListValue_message_name:
			return encoder.marshalListValue
		case genid.Value_message_name:
			return encoder.marshalKnownValue
		case genid.FieldMask_message_name:
			return encoder.marshalFieldMask
		case genid.Empty_message_name:
			return encoder.marshalEmpty
		}
	}
	return nil
}

type unmarshalFunc func(decoder, interface{}, pref.Message) error

// wellKnownTypeUnmarshaler returns a unmarshal function if the message type
// has specialized serialization behavior. It returns nil otherwise.
func wellKnownTypeUnmarshaler(name pref.FullName) unmarshalFunc {
	if name.Parent() == genid.GoogleProtobuf_package {
		switch name.Name() {
		case genid.Any_message_name:
			return decoder.unmarshalAny
		case genid.Timestamp_message_name:
			return decoder.unmarshalTimestamp
		case genid.Duration_message_name:
			return decoder.unmarshalDuration
		case genid.BoolValue_message_name,
			genid.Int32Value_message_name,
			genid.Int64Value_message_name,
			genid.UInt32Value_message_name,
			genid.UInt64Value_message_name,
			genid.FloatValue_message_name,
			genid.DoubleValue_message_name,
			genid.StringValue_message_name,
			genid.BytesValue_message_name:
			return decoder.unmarshalWrapperType
		case genid.Struct_message_name:
			return decoder.unmarshalStruct
		case genid.ListValue_message_name:
			return decoder.unmarshalListValue
		case genid.Value_message_name:
			return decoder.unmarshalKnownValue
		//case genid.FieldMask_message_name:
		//	return decoder.unmarshalFieldMask
		case genid.Empty_message_name:
			return decoder.unmarshalEmpty
		}
	}
	return nil
}

// The JSON representation of an Any message uses the regular representation of
// the deserialized, embedded message, with an additional field `@type` which
// contains the type URL. If the embedded message type is well-known and has a
// custom JSON representation, that representation will be embedded adding a
// field `value` which holds the custom JSON in addition to the `@type` field.

func (e encoder) marshalAny(m pref.Message) (interface{}, error) {
	result := bson.D{}
	fds := m.Descriptor().Fields()
	fdType := fds.ByNumber(genid.Any_TypeUrl_field_number)
	fdValue := fds.ByNumber(genid.Any_Value_field_number)

	if !m.Has(fdType) {
		if !m.Has(fdValue) {
			// If message is empty, marshal out empty JSON object.
			return bson.D{}, nil
		}
		// Return error if type_url field is not set, but value is set.
		return nil, fmt.Errorf("%s: %v is not set", genid.Any_message_fullname, genid.Any_TypeUrl_field_name)
	}

	typeVal := m.Get(fdType)
	valueVal := m.Get(fdValue)
	typeURL := typeVal.String()

	// Marshal out @type field.
	result = append(result, bson.E{Key: "@type", Value: typeURL})

	// Resolve the type in order to unmarshal value field.
	emt, err := e.opts.Resolver.FindMessageByURL(typeURL)
	if err != nil {
		return result, fmt.Errorf("%s: unable to resolve %q: %v", genid.Any_message_fullname, typeURL, err)
	}

	em := emt.New()
	err = proto.UnmarshalOptions{
		AllowPartial: true, // never check required fields inside an Any
		Resolver:     e.opts.Resolver,
	}.Unmarshal(valueVal.Bytes(), em.Interface())
	if err != nil {
		return result, fmt.Errorf("%s: unable to unmarshal %q: %v", genid.Any_message_fullname, typeURL, err)
	}

	// If type of value has custom JSON encoding, marshal out a field "value"
	// with corresponding custom JSON encoding of the embedded message as a
	// field.
	if marshal := wellKnownTypeMarshaler(emt.Descriptor().FullName()); marshal != nil {
		val, err := marshal(e, em)
		if err != nil {
			return result, err
		}
		result = append(result, bson.E{Key: "value", Value: val})
		return result, nil
	}

	// Else, marshal out the embedded message's fields in this Any object.
	marshaled, err := e.marshalFields(em)
	if err != nil {
		return result, err
	}

	result = append(result, marshaled...)

	return result, nil
}

func (d decoder) unmarshalAny(val interface{}, m pref.Message) error {
	// Use another decoder to parse the unread bytes for @type field. This
	// avoids advancing a read from current decoder because the current JSON
	// object may contain the fields of the embedded type.
	valD := val.(bson.D)
	var found, nonEmpty, ok bool
	var typeURL string
	for _, item := range valD {
		switch item.Key {
		case "@type":
			nonEmpty = true
			if found {
				// Duplicate
				return errors.New("duplicate @type field")
			}
			typeURL, ok = item.Value.(string)
			if !ok || typeURL == "" {
				return errors.New("@type field contains empty or invalid value")
			}
			found = true
		case "value":
			nonEmpty = true
		default:
			if !d.opts.DiscardUnknown {
				nonEmpty = true
			}
		}
	}
	if !nonEmpty || (!found && d.opts.DiscardUnknown) {
		return nil
	}
	if !found {
		return errors.New("missing @type field in non-empty message")
	}

	emt, err := d.opts.Resolver.FindMessageByURL(typeURL)
	if err != nil {
		return fmt.Errorf("unable to resolve %q: %s", typeURL, strings.Replace(err.Error(), "\u00a0", " ", -1))
	}

	// Create new message for the embedded message type and unmarshal into it.
	em := emt.New()
	if umFunc := wellKnownTypeUnmarshaler(emt.Descriptor().FullName()); umFunc != nil {
		// If embedded message is a custom type,
		// unmarshal the JSON "value" field into it.
		if err := d.unmarshalAnyValue(valD, umFunc, em); err != nil {
			return err
		}
	} else {
		// Else unmarshal the current JSON object into it.
		// Remove the type first
		valDNoType := bson.D{}
		for _, item := range valD {
			if item.Key != "@type" {
				valDNoType = append(valDNoType, item)
			}
		}
		if err := d.unmarshalMessage(valDNoType, em, true); err != nil {
			return err
		}
	}
	// Serialize the embedded message and assign the resulting bytes to the
	// proto value field.
	b, err := proto.MarshalOptions{
		AllowPartial:  true, // No need to check required fields inside an Any.
		Deterministic: true,
	}.Marshal(em.Interface())
	if err != nil {
		return fmt.Errorf("error in marshaling Any.value field: %v", err)
	}

	fds := m.Descriptor().Fields()
	fdType := fds.ByNumber(genid.Any_TypeUrl_field_number)
	fdValue := fds.ByNumber(genid.Any_Value_field_number)

	m.Set(fdType, pref.ValueOfString(typeURL))
	m.Set(fdValue, pref.ValueOfBytes(b))
	return nil
}

func (d decoder) unmarshalAnyValue(val bson.D, umFunc unmarshalFunc, m pref.Message) error {
	var found bool
	for _, item := range val {
		switch item.Key {
		case "@type":
			// Skip the value as this was previously parsed already.
		case "value":
			if found {
				return fmt.Errorf(`duplicate "value" field`)
			}
			// Unmarshal the field value into the given message.
			if err := umFunc(d, item.Value, m); err != nil {
				return err
			}
			found = true
		default:
			if d.opts.DiscardUnknown {
				continue
			}
			return fmt.Errorf("unknown field %q", item.Key)
		}
	}
	if !found {
		return fmt.Errorf(`missing "value" field`)
	}
	return nil
}

// Wrapper types are encoded as JSON primitives like string, number or boolean.

func (e encoder) marshalWrapperType(m pref.Message) (interface{}, error) {
	fd := m.Descriptor().Fields().ByNumber(genid.WrapperValue_Value_field_number)
	val := m.Get(fd)
	return e.marshalSingular(val, fd)
}

func (d decoder) unmarshalWrapperType(val interface{}, m pref.Message) error {
	fd := m.Descriptor().Fields().ByNumber(genid.WrapperValue_Value_field_number)
	marshaled, err := d.unmarshalScalar(val, fd)
	if err != nil {
		return err
	}
	m.Set(fd, marshaled)
	return nil
}

// The JSON representation for Empty is an empty JSON object.

func (e encoder) marshalEmpty(pref.Message) (interface{}, error) {
	return primitive.Null{}, nil
}

func (d decoder) unmarshalEmpty(val interface{}, m pref.Message) error {
	if valD, ok := val.(bson.D); ok {
		for _, item := range valD {
			if d.opts.DiscardUnknown {
				continue
			}
			return fmt.Errorf("unknown field %q", item.Key)
		}
	}
	return nil
}

// The JSON representation for Struct is a JSON object that contains the encoded
// Struct.fields map and follows the serialization rules for a map.

func (e encoder) marshalStruct(m pref.Message) (interface{}, error) {
	fd := m.Descriptor().Fields().ByNumber(genid.Struct_Fields_field_number)
	return e.marshalMap(m.Get(fd).Map(), fd)
}

func (d decoder) unmarshalStruct(val interface{}, m pref.Message) error {
	fd := m.Descriptor().Fields().ByNumber(genid.Struct_Fields_field_number)
	return d.unmarshalMap(val.(bson.D), m.Mutable(fd).Map(), fd)
}

// The JSON representation for ListValue is JSON array that contains the encoded
// ListValue.values repeated field and follows the serialization rules for a
// repeated field.

func (e encoder) marshalListValue(m pref.Message) (interface{}, error) {
	fd := m.Descriptor().Fields().ByNumber(genid.ListValue_Values_field_number)
	return e.marshalList(m.Get(fd).List(), fd)
}

func (d decoder) unmarshalListValue(val interface{}, m pref.Message) error {
	fd := m.Descriptor().Fields().ByNumber(genid.ListValue_Values_field_number)
	return d.unmarshalList(val.(bson.A), m.Mutable(fd).List(), fd)
}

// The JSON representation for a Value is dependent on the oneof field that is
// set. Each of the field in the oneof has its own custom serialization rule. A
// Value message needs to be a oneof field set, else it is an error.

func (e encoder) marshalKnownValue(m pref.Message) (interface{}, error) {
	od := m.Descriptor().Oneofs().ByName(genid.Value_Kind_oneof_name)
	fd := m.WhichOneof(od)
	if fd == nil {
		return nil, fmt.Errorf("%s: none of the oneof fields is set", genid.Value_message_fullname)
	}
	return e.marshalSingular(m.Get(fd), fd)
}

func (d decoder) unmarshalKnownValue(val interface{}, m pref.Message) error {
	var fd pref.FieldDescriptor
	var pval pref.Value
	valT := reflect.TypeOf(val)
	valV := reflect.ValueOf(val)
	switch valT.Kind() {
	case reflect.Bool:
		fd = m.Descriptor().Fields().ByNumber(genid.Value_BoolValue_field_number)
		pval = pref.ValueOfBool(valV.Bool())

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		fd = m.Descriptor().Fields().ByNumber(genid.Value_NumberValue_field_number)
		pval = pref.ValueOfFloat64(float64(valV.Int()))
	case reflect.Float32, reflect.Float64:
		fd = m.Descriptor().Fields().ByNumber(genid.Value_NumberValue_field_number)
		pval = pref.ValueOfFloat64(valV.Float())

	case reflect.String:
		// A JSON string may have been encoded from the number_value field,
		// e.g. "NaN", "Infinity", etc. Parsing a proto double type also allows
		// for it to be in JSON string form. Given this custom encoding spec,
		// however, there is no way to identify that and hence a JSON string is
		// always assigned to the string_value field, which means that certain
		// encoding cannot be parsed back to the same field.
		fd = m.Descriptor().Fields().ByNumber(genid.Value_StringValue_field_number)
		if valid := utf8.Valid([]byte(valV.String())); !valid {
			return fmt.Errorf("invalid UTF-8: %s", valV.String())
		}
		pval = pref.ValueOfString(valV.String())

	case reflect.Struct:
		// Check for null
		if _, null := val.(primitive.Null); null {
			fd = m.Descriptor().Fields().ByNumber(genid.Value_NullValue_field_number)
			pval = pref.ValueOfEnum(0)
		} else {
			fd = m.Descriptor().Fields().ByNumber(genid.Value_StructValue_field_number)
			pval = m.NewField(fd)
			if err := d.unmarshalStruct(val, pval.Message()); err != nil {
				return err
			}
		}

	case reflect.Array, reflect.Slice:
		// Can be either .D or .A
		if _, isD := val.(bson.D); isD {
			fd = m.Descriptor().Fields().ByNumber(genid.Value_StructValue_field_number)
			pval = m.NewField(fd)
			if err := d.unmarshalStruct(val, pval.Message()); err != nil {
				return err
			}
		}
		if _, isA := val.(bson.A); isA {
			fd = m.Descriptor().Fields().ByNumber(genid.Value_ListValue_field_number)
			pval = m.NewField(fd)
			if err := d.unmarshalListValue(val, pval.Message()); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("invalid %v: %v", genid.Value_message_fullname, val)
	}
	m.Set(fd, pval)
	return nil
}

// The JSON representation for a Duration is a JSON string that ends in the
// suffix "s" (indicating seconds) and is preceded by the number of seconds,
// with nanoseconds expressed as fractional seconds.
//
// Durations less than one second are represented with a 0 seconds field and a
// positive or negative nanos field. For durations of one second or more, a
// non-zero value for the nanos field must be of the same sign as the seconds
// field.
//
// Duration.seconds must be from -315,576,000,000 to +315,576,000,000 inclusive.
// Duration.nanos must be from -999,999,999 to +999,999,999 inclusive.

const (
	secondsInNanos       = 999999999
	maxSecondsInDuration = 315576000000
)

func isValidDuration(secs, nanos int64) (bool, error) {
	if secs < -maxSecondsInDuration || secs > maxSecondsInDuration {
		return false, fmt.Errorf("%s: seconds out of range %v", genid.Duration_message_fullname, secs)
	}
	if nanos < -secondsInNanos || nanos > secondsInNanos {
		return false, fmt.Errorf("%s: nanos out of range %v", genid.Duration_message_fullname, nanos)
	}
	if (secs > 0 && nanos < 0) || (secs < 0 && nanos > 0) {
		return false, fmt.Errorf("%s: signs of seconds and nanos do not match", genid.Duration_message_fullname)
	}
	return true, nil
}

func (e encoder) marshalDuration(m pref.Message) (interface{}, error) {
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(genid.Duration_Seconds_field_number)
	fdNanos := fds.ByNumber(genid.Duration_Nanos_field_number)

	secsVal := m.Get(fdSeconds)
	nanosVal := m.Get(fdNanos)
	secs := secsVal.Int()
	nanos := nanosVal.Int()

	if _, err := isValidDuration(secs, nanos); err != nil {
		return bson.D{}, err
	}
	return bson.D{
		{Key: "Seconds", Value: secs},
		{Key: "Nanos", Value: nanos},
	}, nil
}

func (d decoder) unmarshalDuration(val interface{}, m pref.Message) error {
	var err = fmt.Errorf("invalid google.protobuf.Duration value %s", quoted(val))
	dur, ok := val.(bson.D)
	if !ok {
		return err
	}
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(genid.Duration_Seconds_field_number)
	fdNanos := fds.ByNumber(genid.Duration_Nanos_field_number)

	var secs, nanos int64
	if seconds, ok := dur.Map()["Seconds"]; ok {
		switch reflect.TypeOf(seconds).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			secs = reflect.ValueOf(seconds).Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			secs = int64(reflect.ValueOf(seconds).Uint())
		default:
			return fmt.Errorf("invalid google.protobuf.Duration seconds: %v (want int64 but got %T)", quoted(seconds), seconds)
		}
	}
	if nanoseconds, ok := dur.Map()["Nanos"]; ok {
		switch reflect.TypeOf(nanoseconds).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			nanos = int64(reflect.ValueOf(nanoseconds).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			nanos = int64(reflect.ValueOf(nanoseconds).Uint())
		default:
			return fmt.Errorf("invalid google.protobuf.Duration nanoseconds: %v (want int32 but got %T)", quoted(nanoseconds), nanoseconds)
		}
	}

	if _, err := isValidDuration(secs, nanos); err != nil {
		return err
	}

	m.Set(fdSeconds, pref.ValueOfInt64(secs))
	m.Set(fdNanos, pref.ValueOfInt32(int32(nanos)))
	return nil
}

// parseDuration parses the given input string for seconds and nanoseconds value
// for the Duration JSON format. The format is a decimal number with a suffix
// 's'. It can have optional plus/minus sign. There needs to be at least an
// integer or fractional part. Fractional part is limited to 9 digits only for
// nanoseconds precision, regardless of whether there are trailing zero digits.
// Example values are 1s, 0.1s, 1.s, .1s, +1s, -1s, -.1s.
func parseDuration(input string) (int64, int32, bool) {
	b := []byte(input)
	size := len(b)
	if size < 2 {
		return 0, 0, false
	}
	if b[size-1] != 's' {
		return 0, 0, false
	}
	b = b[:size-1]

	// Read optional plus/minus symbol.
	var neg bool
	switch b[0] {
	case '-':
		neg = true
		b = b[1:]
	case '+':
		b = b[1:]
	}
	if len(b) == 0 {
		return 0, 0, false
	}

	// Read the integer part.
	var intp []byte
	switch {
	case b[0] == '0':
		b = b[1:]

	case '1' <= b[0] && b[0] <= '9':
		intp = b[0:]
		b = b[1:]
		n := 1
		for len(b) > 0 && '0' <= b[0] && b[0] <= '9' {
			n++
			b = b[1:]
		}
		intp = intp[:n]

	case b[0] == '.':
		// Continue below.

	default:
		return 0, 0, false
	}

	hasFrac := false
	var frac [9]byte
	if len(b) > 0 {
		if b[0] != '.' {
			return 0, 0, false
		}
		// Read the fractional part.
		b = b[1:]
		n := 0
		for len(b) > 0 && n < 9 && '0' <= b[0] && b[0] <= '9' {
			frac[n] = b[0]
			n++
			b = b[1:]
		}
		// It is not valid if there are more bytes left.
		if len(b) > 0 {
			return 0, 0, false
		}
		// Pad fractional part with 0s.
		for i := n; i < 9; i++ {
			frac[i] = '0'
		}
		hasFrac = true
	}

	var secs int64
	if len(intp) > 0 {
		var err error
		secs, err = strconv.ParseInt(string(intp), 10, 64)
		if err != nil {
			return 0, 0, false
		}
	}

	var nanos int64
	if hasFrac {
		nanob := bytes.TrimLeft(frac[:], "0")
		if len(nanob) > 0 {
			var err error
			nanos, err = strconv.ParseInt(string(nanob), 10, 32)
			if err != nil {
				return 0, 0, false
			}
		}
	}

	if neg {
		if secs > 0 {
			secs = -secs
		}
		if nanos > 0 {
			nanos = -nanos
		}
	}
	return secs, int32(nanos), true
}

// The JSON representation for a Timestamp is a JSON string in the RFC 3339
// format, i.e. "{year}-{month}-{day}T{hour}:{min}:{sec}[.{frac_sec}]Z" where
// {year} is always expressed using four digits while {month}, {day}, {hour},
// {min}, and {sec} are zero-padded to two digits each. The fractional seconds,
// which can go up to 9 digits, up to 1 nanosecond resolution, is optional. The
// "Z" suffix indicates the timezone ("UTC"); the timezone is required. Encoding
// should always use UTC (as indicated by "Z") and a decoder should be able to
// accept both UTC and other timezones (as indicated by an offset).
//
// Timestamp.seconds must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z
// inclusive.
// Timestamp.nanos must be from 0 to 999,999,999 inclusive.

const (
	maxTimestampSeconds = 253402300799
	minTimestampSeconds = -62135596800
)

func isValidTimestamp(secs, nanos int64) (bool, error) {
	if secs < minTimestampSeconds || secs > maxTimestampSeconds {
		return false, fmt.Errorf("%s: seconds out of range %v", genid.Timestamp_message_fullname, secs)
	}
	if nanos < 0 || nanos > secondsInNanos {
		return false, fmt.Errorf("%s: nanos out of range %v", genid.Timestamp_message_fullname, nanos)
	}
	return true, nil
}

func (e encoder) marshalTimestamp(m pref.Message) (interface{}, error) {
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(genid.Timestamp_Seconds_field_number)
	fdNanos := fds.ByNumber(genid.Timestamp_Nanos_field_number)

	secsVal := m.Get(fdSeconds)
	nanosVal := m.Get(fdNanos)
	secs := int64(secsVal.Int())
	nanos := int64(nanosVal.Int())

	if _, err := isValidTimestamp(secs, nanos); err != nil {
		return bson.D{}, err
	}
	return primitive.NewDateTimeFromTime(time.Unix(secs, nanos).UTC()), nil
}

func (d decoder) unmarshalTimestamp(val interface{}, m pref.Message) error {
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(genid.Timestamp_Seconds_field_number)
	fdNanos := fds.ByNumber(genid.Timestamp_Nanos_field_number)

	if ts, ok := val.(primitive.DateTime); ok {
		t := ts.Time()
		m.Set(fdSeconds, pref.ValueOfInt64(t.Unix()))
		m.Set(fdNanos, pref.ValueOfInt32(int32(t.Nanosecond())))
		return nil
	}
	var secs int64
	switch reflect.TypeOf(val).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		secs = reflect.ValueOf(val).Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		secs = int64(reflect.ValueOf(val).Uint())
	default:
		return fmt.Errorf("invalid google.protobuf.Timestamp value %s", quoted(val))
	}
	m.Set(fdSeconds, pref.ValueOfInt64(secs))
	return nil
}

// The JSON representation for a FieldMask is a JSON string where paths are
// separated by a comma. Fields name in each path are converted to/from
// lower-camel naming conventions. Encoding should fail if the path name would
// end up differently after a round-trip.

func (e encoder) marshalFieldMask(m pref.Message) (interface{}, error) {
	fd := m.Descriptor().Fields().ByNumber(genid.FieldMask_Paths_field_number)
	list := m.Get(fd).List()
	paths := bson.A{}

	for i := 0; i < list.Len(); i++ {
		s := list.Get(i).String()
		if !pref.FullName(s).IsValid() {
			return nil, fmt.Errorf("%s contains invalid path: %q", genid.FieldMask_Paths_field_fullname, s)
		}
		// Return error if conversion to camelCase is not reversible.
		cc := JSONCamelCase(s)
		if s != JSONSnakeCase(cc) {
			return nil, fmt.Errorf("%s contains irreversible value %q", genid.FieldMask_Paths_field_fullname, s)
		}
		paths = append(paths, cc)
	}
	return paths, nil
}

/*
// TODO
func (d decoder) unmarshalFieldMask(m pref.Message) error {
	tok, err := d.Read()
	if err != nil {
		return err
	}
	if tok.Kind() != json.String {
		return d.unexpectedTokenError(tok)
	}
	str := strings.TrimSpace(tok.ParsedString())
	if str == "" {
		return nil
	}
	paths := strings.Split(str, ",")

	fd := m.Descriptor().Fields().ByNumber(genid.FieldMask_Paths_field_number)
	list := m.Mutable(fd).List()

	for _, s0 := range paths {
		s := JSONSnakeCase(s0)
		if strings.Contains(s0, "_") || !pref.FullName(s).IsValid() {
			return d.newError(tok.Pos(), "%v contains invalid path: %q", genid.FieldMask_Paths_field_fullname, s0)
		}
		list.Append(pref.ValueOfString(s))
	}
	return nil
}
*/
