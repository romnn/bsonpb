package bsonpb

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	// proto3pb "github.com/golang/protobuf/internal/testprotos/proto3_proto"
	proto3pb "github.com/romnnn/bsonpb/internal/testprotos/v1/proto3_proto"
	anypb "github.com/golang/protobuf/ptypes/any"
	durpb "github.com/golang/protobuf/ptypes/duration"
	stpb "github.com/golang/protobuf/ptypes/struct"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	wpb "github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/romnnn/bsonpb/internal/testprotos/v1/test_objects"
	"github.com/romnnn/deepequal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var unmarshalingTests = []struct {
	desc        string
	unmarshaler Unmarshaler
	bson        bson.D
	pb          proto.Message
}{
	{"simple flat object", Unmarshaler{}, simpleObjectOutputBSON, simpleObject},
	{"repeated fields flat object", Unmarshaler{}, repeatsObjectBSON, repeatsObject},
	{"nested message/enum flat object", Unmarshaler{}, complexObjectMinimalBSON, complexObject},
	{"enum-string object", Unmarshaler{},
		bson.D{{"color", "BLUE"}},
		&pb.Widget{Color: pb.Widget_BLUE.Enum()},
	},
	{"enum-value object", Unmarshaler{},
		bson.D{{"color", 2}},
		&pb.Widget{Color: pb.Widget_BLUE.Enum()},
	},
	{"unknown field with allowed option", Unmarshaler{AllowUnknownFields: true},
		bson.D{{"unknown", "foo"}, {"oInt32", int32(42)}},
		innerSimple,
	},
	{"proto3 enum string", Unmarshaler{},
		bson.D{{"hilarity", "PUNS"}},
		&proto3pb.Message{Hilarity: proto3pb.Message_PUNS},
	},
	{"proto3 enum value", Unmarshaler{},
		bson.D{{"hilarity", 1}},
		&proto3pb.Message{Hilarity: proto3pb.Message_PUNS},
	},
	{"unknown enum value object",
		Unmarshaler{},
		bson.D{{"color", 1000}, {"r_color", bson.A{"RED"}}},
		&pb.Widget{Color: pb.Widget_Color(1000).Enum(), RColor: []pb.Widget_Color{pb.Widget_RED}},
	},
	{"repeated proto3 enum", Unmarshaler{},
		bson.D{{"rFunny", bson.A{"PUNS", "SLAPSTICK"}}},
		&proto3pb.Message{RFunny: []proto3pb.Message_Humour{
			proto3pb.Message_PUNS,
			proto3pb.Message_SLAPSTICK,
		}},
	},
	{"repeated proto3 enum as int", Unmarshaler{},
		bson.D{{"rFunny", bson.A{1, 2}}},
		&proto3pb.Message{RFunny: []proto3pb.Message_Humour{
			proto3pb.Message_PUNS,
			proto3pb.Message_SLAPSTICK,
		}},
	},
	{"repeated proto3 enum as mix of strings and ints", Unmarshaler{},
		bson.D{{"rFunny", bson.A{"PUNS", 2}}},
		&proto3pb.Message{RFunny: []proto3pb.Message_Humour{
			proto3pb.Message_PUNS,
			proto3pb.Message_SLAPSTICK,
		}},
	},
	{"unquoted int64 object", Unmarshaler{},
		bson.D{{"oInt64", -314}, {"oInt32", int32(12)}},
		&pb.Simple{OInt64: proto.Int64(-314), OInt32: proto.Int32(12)},
	},
	{"unquoted uint64 object", Unmarshaler{},
		bson.D{{"oUint64", 123}, {"oInt32", int32(12)}},
		&pb.Simple{OUint64: proto.Uint64(123), OInt32: proto.Int32(12)},
	},
	{"NaN", Unmarshaler{},
		bson.D{{"oDouble", float64(math.NaN())}, {"oInt32", int32(12)}},
		&pb.Simple{ODouble: proto.Float64(math.NaN()), OInt32: proto.Int32(12)},
	},
	{"Inf", Unmarshaler{},
		bson.D{{"oFloat", proto.Float32(float32(math.Inf(1)))}, {"oInt32", int32(12)}},
		&pb.Simple{OFloat: proto.Float32(float32(math.Inf(1))), OInt32: proto.Int32(12)},
	},
	{"-Inf", Unmarshaler{},
		bson.D{{"oDouble", proto.Float64(math.Inf(-1))}, {"oInt32", int32(12)}},
		&pb.Simple{ODouble: proto.Float64(math.Inf(-1)), OInt32: proto.Int32(12)},
	},
	{"map<int64, int32>", Unmarshaler{},
		bson.D{{"nummy", bson.D{{"1", 2}, {"3", 4}}}},
		&pb.Mappy{Nummy: map[int64]int32{1: 2, 3: 4}},
	},
	{"map<string, string>", Unmarshaler{},
		bson.D{{"strry", bson.D{{"one", "two"}, {"three", "four"}}}},
		&pb.Mappy{Strry: map[string]string{"one": "two", "three": "four"}},
	},
	{"map<int32, Object>", Unmarshaler{},
		bson.D{{"objjy", bson.D{{"1", bson.D{{"dub", 1}}}}}},
		&pb.Mappy{Objjy: map[int32]*pb.Simple3{1: {Dub: 1}}},
	},
	{"proto2 extension", Unmarshaler{}, realNumberBSON, realNumber},
	{"Any with message", Unmarshaler{}, anySimpleBSON, anySimple},
	{"Any with WKT", Unmarshaler{}, anyWellKnownBSON, anyWellKnown},
	{"map<string, enum>", Unmarshaler{},
		bson.D{{"enumy", bson.D{{"XIV", "ROMAN"}}}},
		&pb.Mappy{Enumy: map[string]pb.Numeral{"XIV": pb.Numeral_ROMAN}},
	},
	{"map<string, enum as int>", Unmarshaler{},
		bson.D{{"enumy", bson.D{{"XIV", 2}}}},
		&pb.Mappy{Enumy: map[string]pb.Numeral{"XIV": pb.Numeral_ROMAN}},
	},
	{"oneof", Unmarshaler{},
		bson.D{{"salary", 31000}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_Salary{31000}},
	},
	{"oneof spec name", Unmarshaler{},
		bson.D{{"Country", "Australia"}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_Country{"Australia"}},
	},
	{"oneof orig_name", Unmarshaler{},
		bson.D{{"Country", "Australia"}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_Country{"Australia"}},
	},
	{"oneof spec name2", Unmarshaler{},
		bson.D{{"homeAddress", "Australia"}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_HomeAddress{"Australia"}},
	},
	{"oneof orig_name2", Unmarshaler{},
		bson.D{{"home_address", "Australia"}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_HomeAddress{"Australia"}},
	},
	{"orig_name input", Unmarshaler{},
		bson.D{{"o_bool", true}, {"o_int32", int32(12)}},
		&pb.Simple{OBool: proto.Bool(true), OInt32: proto.Int32(12)},
	},
	{"camelName input", Unmarshaler{},
		bson.D{{"oBool", true}, {"o_int32", int32(12)}},
		&pb.Simple{OBool: proto.Bool(true), OInt32: proto.Int32(12)},
	},
	{"Duration", Unmarshaler{},
		bson.D{{"dur", float64(3)}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 3}},
	},
	{"Duration", Unmarshaler{},
		bson.D{{"dur", float64(4)}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 4}},
	},
	{"Duration with unicode", Unmarshaler{},
		bson.D{{"dur", float64(3)}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 3}},
	},
	{"null Duration", Unmarshaler{},
		bson.D{{"dur", primitive.Null{}}},
		&pb.KnownTypes{Dur: nil},
	},
	{"Timestamp", Unmarshaler{},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(14e8, 21e6))}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 14e8, Nanos: 21e6}},
	},
	{"Timestamp", Unmarshaler{},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(14e8, 0))}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 14e8, Nanos: 0}},
	},
	{"Timestamp with unicode", Unmarshaler{},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(14e8, 0))}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 14e8, Nanos: 0}},
	},
	{"PreEpochTimestamp", Unmarshaler{},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(-2, 999999995))}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: -1}},
	},
	{"ZeroTimeTimestamp", Unmarshaler{},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(-62135596800, 0))}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: -6795364579, Nanos: 129000000}},
	},
	{"null Timestamp", Unmarshaler{},
		bson.D{{"ts", primitive.Null{}}},
		&pb.KnownTypes{Ts: nil},
	},
	{"null Struct", Unmarshaler{},
		bson.D{},
		&pb.KnownTypes{St: nil},
	},
	{"empty Struct", Unmarshaler{},
		bson.D{{"st", bson.D{}}},
		&pb.KnownTypes{St: &stpb.Struct{}},
	},
	{"basic Struct", Unmarshaler{},
		bson.D{{"st", bson.D{{"a", "x"}, {"b", primitive.Null{}}, {"c", 3}, {"d", true}}}},
		&pb.KnownTypes{St: &stpb.Struct{Fields: map[string]*stpb.Value{
			"a": {Kind: &stpb.Value_StringValue{"x"}},
			"b": {Kind: &stpb.Value_NullValue{}},
			"c": {Kind: &stpb.Value_NumberValue{3}},
			"d": {Kind: &stpb.Value_BoolValue{true}},
		}}},
	},
	{"nested Struct", Unmarshaler{},
		bson.D{{"st", bson.D{{"a", bson.D{{"b", 1}, {"c", bson.A{bson.D{{"d", true}}, "f"}}}}}}},
		&pb.KnownTypes{St: &stpb.Struct{Fields: map[string]*stpb.Value{
			"a": {Kind: &stpb.Value_StructValue{&stpb.Struct{Fields: map[string]*stpb.Value{
				"b": {Kind: &stpb.Value_NumberValue{1}},
				"c": {Kind: &stpb.Value_ListValue{&stpb.ListValue{Values: []*stpb.Value{
					{Kind: &stpb.Value_StructValue{&stpb.Struct{Fields: map[string]*stpb.Value{"d": {Kind: &stpb.Value_BoolValue{true}}}}}},
					{Kind: &stpb.Value_StringValue{"f"}},
				}}}},
			}}}},
		}}},
	},
	{"null ListValue", Unmarshaler{},
		bson.D{{"lv", primitive.Null{}}},
		&pb.KnownTypes{Lv: nil},
	},
	{"empty ListValue", Unmarshaler{},
		bson.D{{"lv", bson.A{}}},
		&pb.KnownTypes{Lv: &stpb.ListValue{}},
	},
	{"basic ListValue", Unmarshaler{},
		bson.D{{"lv", bson.A{"x", primitive.Null{}, 3, true}}},
		&pb.KnownTypes{Lv: &stpb.ListValue{Values: []*stpb.Value{
			{Kind: &stpb.Value_StringValue{"x"}},
			{Kind: &stpb.Value_NullValue{}},
			{Kind: &stpb.Value_NumberValue{3}},
			{Kind: &stpb.Value_BoolValue{true}},
		}}},
	},
	{"number Value", Unmarshaler{},
		bson.D{{"val", 1}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_NumberValue{1}}},
	},
	{"null Value", Unmarshaler{},
		bson.D{{"val", primitive.Null{}}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_NullValue{stpb.NullValue_NULL_VALUE}}},
	},
	{"bool Value", Unmarshaler{},
		bson.D{{"val", true}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_BoolValue{true}}},
	},
	{"string Value", Unmarshaler{},
		bson.D{{"val", "x"}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_StringValue{"x"}}},
	},
	{"string number value", Unmarshaler{},
		bson.D{{"val", "9223372036854775807"}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_StringValue{"9223372036854775807"}}},
	},
	{"list of lists Value", Unmarshaler{},
		bson.D{{"val", bson.A{"x", bson.A{bson.A{"y"}, "z"}}}},
		&pb.KnownTypes{Val: &stpb.Value{
			Kind: &stpb.Value_ListValue{&stpb.ListValue{
				Values: []*stpb.Value{
					{Kind: &stpb.Value_StringValue{"x"}},
					{Kind: &stpb.Value_ListValue{&stpb.ListValue{
						Values: []*stpb.Value{
							{Kind: &stpb.Value_ListValue{&stpb.ListValue{
								Values: []*stpb.Value{{Kind: &stpb.Value_StringValue{"y"}}},
							}}},
							{Kind: &stpb.Value_StringValue{"z"}},
						},
					}}},
				},
			}}},
		},
	},
	{"DoubleValue", Unmarshaler{},
		bson.D{{"dbl", 1.2}},
		&pb.KnownTypes{Dbl: &wpb.DoubleValue{Value: 1.2}},
	},
	{"FloatValue", Unmarshaler{},
		bson.D{{"flt", 1.2}},
		&pb.KnownTypes{Flt: &wpb.FloatValue{Value: 1.2}},
	},
	{"Int64Value", Unmarshaler{},
		bson.D{{"i64", -3}},
		&pb.KnownTypes{I64: &wpb.Int64Value{Value: -3}},
	},
	{"UInt64Value", Unmarshaler{},
		bson.D{{"u64", 3}},
		&pb.KnownTypes{U64: &wpb.UInt64Value{Value: 3}},
	},
	{"Int32Value", Unmarshaler{},
		bson.D{{"i32", -4}},
		&pb.KnownTypes{I32: &wpb.Int32Value{Value: -4}},
	},
	{"UInt32Value", Unmarshaler{},
		bson.D{{"u32", 4}},
		&pb.KnownTypes{U32: &wpb.UInt32Value{Value: 4}},
	},
	{"BoolValue", Unmarshaler{},
		bson.D{{"bool", true}},
		&pb.KnownTypes{Bool: &wpb.BoolValue{Value: true}},
	},
	{"StringValue", Unmarshaler{},
		bson.D{{"str", "plush"}},
		&pb.KnownTypes{Str: &wpb.StringValue{Value: "plush"}},
	},
	{"StringValue containing escaped character", Unmarshaler{},
		bson.D{{"str", "a/b"}},
		&pb.KnownTypes{Str: &wpb.StringValue{Value: "a/b"}},
	},
	{"StructValue containing StringValue's", Unmarshaler{},
		bson.D{{"escaped", "a/b"}, {"unicode", "\u00004E16\u0000754C"}},
		&stpb.Struct{
			Fields: map[string]*stpb.Value{
				"escaped": {Kind: &stpb.Value_StringValue{"a/b"}},
				"unicode": {Kind: &stpb.Value_StringValue{"\u00004E16\u0000754C"}},
			},
		},
	},
	{"BytesValue", Unmarshaler{},
		bson.D{{"bytes", primitive.Binary{Data: []byte("wow")}}},
		&pb.KnownTypes{Bytes: &wpb.BytesValue{Value: []byte("wow")}},
	},

	// Ensure that `null` as a value ends up with a nil pointer instead of a [type]Value struct.
	{"null DoubleValue", Unmarshaler{}, bson.D{{"dbl", primitive.Null{}}}, &pb.KnownTypes{Dbl: nil}},
	{"null FloatValue", Unmarshaler{}, bson.D{{"flt", primitive.Null{}}}, &pb.KnownTypes{Flt: nil}},
	{"null Int64Value", Unmarshaler{}, bson.D{{"i64", primitive.Null{}}}, &pb.KnownTypes{I64: nil}},
	{"null UInt64Value", Unmarshaler{}, bson.D{{"u64", primitive.Null{}}}, &pb.KnownTypes{U64: nil}},
	{"null Int32Value", Unmarshaler{}, bson.D{{"i32", primitive.Null{}}}, &pb.KnownTypes{I32: nil}},
	{"null UInt32Value", Unmarshaler{}, bson.D{{"u32", primitive.Null{}}}, &pb.KnownTypes{U32: nil}},
	{"null BoolValue", Unmarshaler{}, bson.D{{"bool", primitive.Null{}}}, &pb.KnownTypes{Bool: nil}},
	{"null StringValue", Unmarshaler{}, bson.D{{"str", primitive.Null{}}}, &pb.KnownTypes{Str: nil}},
	{"null BytesValue", Unmarshaler{}, bson.D{{"bytes", primitive.Null{}}}, &pb.KnownTypes{Bytes: nil}},

	{"required", Unmarshaler{}, bson.D{{"str", "hello"}}, &pb.MsgWithRequired{Str: proto.String("hello")}},
	{"required bytes", Unmarshaler{}, bson.D{{"byts", primitive.Binary{}}}, &pb.MsgWithRequiredBytes{Byts: []byte{}}},
}

func TestUnmarshaling(t *testing.T) {
	for _, tt := range unmarshalingTests {
		if equal, err := compareUnmarshaled(tt.unmarshaler, tt.bson, tt.pb); err != nil || !equal {
			t.Errorf("\n%s: %s\n", tt.desc, err.Error())
		}
	}
}

func compareUnmarshaled(um Unmarshaler, b bson.D, pb proto.Message) (bool, error) {
	observed := reflect.New(reflect.TypeOf(pb).Elem()).Interface().(proto.Message)

	// Marshal to bson bytes first
	rawBson, mErr := bson.Marshal(b)
	if mErr != nil {
		return false, fmt.Errorf("marshaling bson to bytes failed: %v", mErr)
	}

	// Now unmarshal to proto message
	umErr := um.Unmarshal(rawBson, observed)
	if umErr != nil {
		return false, fmt.Errorf("unmarshaling failed: %v", umErr)
	}
	equal, err := deepequal.DeepEqual(observed, pb)
	if err != nil {
		return false, fmt.Errorf("\n Got %v\n\n Want %v\n\nError: %s", observed, pb, err.Error())
	}
	return equal, nil
}

func TestUnmarshalNullArray(t *testing.T) {
	var repeats pb.Repeats
	if equal, err := compareUnmarshaled(Unmarshaler{}, bson.D{{"rBool", primitive.Null{}}}, &repeats); err != nil || !equal {
		t.Errorf("\n%s: %s\n", t.Name, err.Error())
	}
}

func TestUnmarshalNullObject(t *testing.T) {
	var maps pb.Maps
	if equal, err := compareUnmarshaled(Unmarshaler{}, bson.D{{"mInt64Str", primitive.Null{}}}, &maps); err != nil || !equal {
		t.Errorf("\n%s: %s\n", t.Name, err.Error())
	}
}

/*
// TODO: Fix this test!
func TestUnmarshalNext(t *testing.T) {
	// We only need to check against a few, not all of them.
	tests := unmarshalingTests[:5]

	// Create a buffer with many concatenated BSON objects.
	var b []byte // bytes.Buffer
	for _, tt := range tests {
		// b.WriteString(tt.bson)
		bb, err := bson.MarshalAppend(b, tt.bson)
		if err != nil {
			t.Errorf("Failed to marshal into BSON stream: %s", err.Error())
		}
		b = bb
	}

	// dec := bson.NewDecoder(&b)
	reader := bsonrw.NewBSONDocumentReader(b)
	dec, err := bson.NewDecoder(reader)
	if err != nil {
		t.Errorf("Failed to create decoder for BSON stream: %s", err.Error())
	}
	for _, tt := range tests {
		// Make a new instance of the type of our expected object.
		p := reflect.New(reflect.TypeOf(tt.pb).Elem()).Interface().(proto.Message)
		umErr := UnmarshalNext(dec, p)
		if err != nil {
			t.Errorf("%s: %v", tt.desc, umErr)
			return
			// continue
		}

		equal, err := deepequal.DeepEqual(p, tt.pb)
		if err != nil {
			t.Errorf(err.Error())
			// return
		}
		if !equal {
			t.Errorf("\n Got %v\n\n Want %v\n\nError: %s", p, tt.pb, err.Error())
		}
		if err != nil && !equal {
			return
		}

		// For easier diffs, compare text strings of the protos.
		exp := proto.MarshalTextString(tt.pb)
		act := proto.MarshalTextString(p)
		if string(exp) != string(act) {
			t.Errorf("%s: got [%s] want [%s]", tt.desc, act, exp)
		}

	}

	p := &pb.Simple{}
	umErr2 := UnmarshalNext(dec, p)
	if umErr2 != io.EOF {
		t.Errorf("eof: got %v, expected io.EOF", umErr2)
	}
}
*/

var unmarshalingShouldError = []struct {
	desc string
	in   bson.D
	pb   proto.Message
}{
	{"a value", bson.D{{"666", 1}}, new(pb.Simple)},
	{"gibberish", bson.D{{"adskja123;", "l23"}, {"=-", "="}}, new(pb.Simple)},
	{"unknown field", bson.D{{"unknown", "foo"}}, new(pb.Simple)},
	{"unknown enum name", bson.D{{"hilarity", "DAVE"}}, new(proto3pb.Message)},
	{"Duration containing invalid character", bson.D{{"dur", "3\\U0073"}}, &pb.KnownTypes{}},
	{"Timestamp containing invalid character", bson.D{{"ts", "2014-05-13T16:53:20\\U005a"}}, &pb.KnownTypes{}},
	{"StringValue containing invalid character", bson.D{{"str", "\U00004E16\U0000754C"}}, &pb.KnownTypes{}},
	{"StructValue containing invalid character", bson.D{{"str", "\U00004E16\U0000754C"}}, &stpb.Struct{}},
	{"repeated proto3 enum with non array input", bson.D{{"rFunny", "PUNS"}}, &proto3pb.Message{RFunny: []proto3pb.Message_Humour{}}},
}

func TestUnmarshalingBadInput(t *testing.T) {
	for _, tt := range unmarshalingShouldError {
		if _, err := compareUnmarshaled(Unmarshaler{}, tt.in, tt.pb); err == nil {
			t.Errorf("an error was expected when parsing %q instead of an object", tt.desc)
		}
	}
}

type funcResolver func(turl string) (proto.Message, error)

func (fn funcResolver) Resolve(turl string) (proto.Message, error) {
	return fn(turl)
}

func TestAnyWithCustomResolver(t *testing.T) {
	var resolvedTypeUrls []string
	resolver := funcResolver(func(turl string) (proto.Message, error) {
		resolvedTypeUrls = append(resolvedTypeUrls, turl)
		return new(pb.Simple), nil
	})
	msg := &pb.Simple{
		OBytes:  []byte{1, 2, 3, 4},
		OBool:   proto.Bool(true),
		OString: proto.String("foobar"),
		OInt32:  proto.Int32(12),
		OInt64:  proto.Int64(1020304),
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling message: %v", err)
	}
	// make an Any with a type URL that won't resolve w/out custom resolver
	any := &anypb.Any{
		TypeUrl: "https://foobar.com/some.random.MessageKind",
		Value:   msgBytes,
	}

	m := Marshaler{AnyResolver: resolver, Omit: OmitOptions{All: true}}
	marshaled, err := m.Marshal(any)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling any to BSON: %v", err)
	}
	if len(resolvedTypeUrls) != 1 {
		t.Errorf("custom resolver was not invoked during marshaling")
	} else if resolvedTypeUrls[0] != "https://foobar.com/some.random.MessageKind" {
		t.Errorf("custom resolver was invoked with wrong URL: got %q, wanted %q", resolvedTypeUrls[0], "https://foobar.com/some.random.MessageKind")
	}

	wanted := bson.D{
		{"@type", "https://foobar.com/some.random.MessageKind"},
		{"oBool", true},
		{"oInt32", int32(12)},
		{"oInt64", int64(1020304)},
		{"oString", "foobar"},
		{"oBytes", primitive.Binary{Data: []byte{1, 2, 3, 4}}},
	}
	equal, err := deepequal.DeepEqual(marshaled, wanted)
	if err != nil {
		t.Errorf("\n Got %v\n\n Want %v\n\nError: %s", marshaled, wanted, err.Error())
	}
	if !equal {
		t.Errorf("marshaling BSON produced incorrect output: \ngot %v\nwanted %v", marshaled, wanted)
	}

	u := Unmarshaler{AnyResolver: resolver}
	equal, umErr := compareUnmarshaled(u, marshaled.(bson.D), any)
	if umErr != nil {
		t.Errorf(umErr.Error())
		return
	}
	if len(resolvedTypeUrls) != 2 {
		t.Errorf("custom resolver was not invoked during marshaling")
	} else if resolvedTypeUrls[1] != "https://foobar.com/some.random.MessageKind" {
		t.Errorf("custom resolver was invoked with wrong URL: got %q, wanted %q", resolvedTypeUrls[1], "https://foobar.com/some.random.MessageKind")
	}
	if !equal {
		t.Errorf("message contents not set correctly after unmarshaling BSON")
	}
}

/*
// TODO: FIX FLAKY
func TestUnmarshalBSONPBUnmarshaler(t *testing.T) {
	b := bson.D{{"foo", "bar"}, {"baz", bson.A{0, 1, 2, 3}}}
	var msg dynamicMessage
	// Marshal to bson bytes first
	rawBson, mErr := bson.Marshal(b)
	if mErr != nil {
		t.Errorf("marshaling bson to bytes failed: %v", mErr)
		return
	}

	// Now unmarshal to proto message
	umErr := (&Unmarshaler{}).Unmarshal(rawBson, &msg)
	if umErr != nil {
		t.Errorf("unmarshaling failed: %v", umErr)
		return
	}
	equal, err := deepequal.DeepEqual(rawBson, msg.RawBson)
	if err != nil {
		t.Errorf("\n Got %v\n\n Want %v\n\nError: %s", rawBson, msg.RawBson, err.Error())
		return
	}
	if !equal {
		t.Errorf("Not equal")
	}
}
*/

/*
// TODO: FIX FLAKY
func TestUnmarshalNullWithBSONPBUnmarshaler(t *testing.T) {
	b := bson.D{{"stringField", primitive.Null{}}}
	var ptrFieldMsg ptrFieldMessage
	// Marshal to bson bytes first
	rawBson, mErr := bson.Marshal(b)
	if mErr != nil {
		t.Errorf("marshaling bson to bytes failed: %v", mErr)
		return
	}

	umErr := (&Unmarshaler{}).Unmarshal(rawBson, &ptrFieldMsg)
	if umErr != nil {
		t.Errorf("unmarshaling failed: %v", umErr)
		return
	}

	want := ptrFieldMessage{StringField: nil}
	equal, err := deepequal.DeepEqual(ptrFieldMsg, want)
	if err != nil {
		t.Errorf("\n Got %v\n\n Want %v\n\nError: %s", ptrFieldMsg, want, err.Error())
		return
	}
	if !equal {
		t.Errorf("Not equal")
	}
}
*/

/*
// TODO: FIX FLAKY
func TestUnmarshalAnyBSONPBUnmarshaler(t *testing.T) {
	bRaw := bson.D{{"foo", "bar"}, {"baz", bson.A{0, 1, 2, 3}}}
	bAny := append(bRaw, bson.E{"@type", "blah.com/" + dynamicMessageName})
	var got anypb.Any
	rawAny, mAnyErr := bson.Marshal(bAny)
	if mAnyErr != nil {
		t.Errorf("marshaling bson to bytes failed: %v", mAnyErr)
		return
	}
	rawRaw, mRawErr := bson.Marshal(bRaw)
	if mRawErr != nil {
		t.Errorf("marshaling bson to bytes failed: %v", mRawErr)
		return
	}

	umErr := (&Unmarshaler{}).Unmarshal(rawAny, &got)
	if umErr != nil {
		t.Errorf("unmarshaling failed: %v", umErr)
		return
	}

	dm := &dynamicMessage{RawBson: rawRaw}
	var want anypb.Any
	if b, err := proto.Marshal(dm); err != nil {
		t.Errorf("an unexpected error occurred when marshaling message: %v", err)
	} else {
		want.TypeUrl = "blah.com/" + dynamicMessageName
		want.Value = b
	}

	equal, err := deepequal.DeepEqual(got, want)
	if err != nil {
		t.Errorf("\n Got %v\n\n Want %v\n\nError: %s", got, want, err.Error())
	}
	if !equal {
		t.Errorf("marshaling BSON produced incorrect output: \ngot %v\nwanted %v", got, want)
	}
}
*/

// Test unmarshaling message containing unset required fields should produce error.
func TestUnmarshalUnsetRequiredFields(t *testing.T) {
	tests := []struct {
		desc string
		pb   proto.Message
		bson bson.D
	}{
		{
			desc: "direct required field missing",
			pb:   &pb.MsgWithRequired{},
			bson: bson.D{},
		},
		{
			desc: "direct required field set to null",
			pb:   &pb.MsgWithRequired{},
			bson: bson.D{{"str", primitive.Null{}}},
		},
		{
			desc: "indirect required field missing",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"subm", bson.A{}}},
		},
		{
			desc: "indirect required field set to null",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"subm", bson.D{{"str", primitive.Null{}}}}},
		},
		{
			desc: "direct required bytes field missing",
			pb:   &pb.MsgWithRequiredBytes{},
			bson: bson.D{},
		},
		{
			desc: "direct required bytes field set to null",
			pb:   &pb.MsgWithRequiredBytes{},
			bson: bson.D{{"byts", primitive.Null{}}},
		},
		{
			desc: "direct required wkt field missing",
			pb:   &pb.MsgWithRequiredWKT{},
			bson: bson.D{},
		},
		{
			desc: "direct required wkt field set to null",
			pb:   &pb.MsgWithRequiredWKT{},
			bson: bson.D{{"str", primitive.Null{}}},
		},
		{
			desc: "any containing message with required field set to null",
			pb:   &pb.KnownTypes{},
			bson: bson.D{{"an", bson.D{{"@type", "example.com/bsonpb.MsgWithRequired"}, {"str", primitive.Null{}}}}},
		},
		{
			desc: "any containing message with missing required field",
			pb:   &pb.KnownTypes{},
			bson: bson.D{{"an", bson.D{{"@type", "example.com/bsonpb.MsgWithRequired"}}}},
		},
		{
			desc: "missing required in map value",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"map_field", bson.D{{"a", bson.D{}}, {"b", bson.D{{"str", "hi"}}}}}},
		},
		{
			desc: "required in map value set to null",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"map_field", bson.D{{"a", bson.D{{"str", "hello"}}}, {"b", bson.D{{"str", primitive.Null{}}}}}}},
		},
		{
			desc: "missing required in slice item",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"slice_field", bson.A{bson.D{}, bson.D{{"str", "hi"}}}}},
		},
		{
			desc: "required in slice item set to null",
			pb:   &pb.MsgWithIndirectRequired{},
			bson: bson.D{{"slice_field", bson.A{bson.D{{"str", "hello"}}, bson.D{{"str", primitive.Null{}}}}}},
		},
		{
			desc: "required inside oneof missing",
			pb:   &pb.MsgWithOneof{},
			bson: bson.D{{"msgWithRequired", bson.D{}}},
		},
		{
			desc: "required inside oneof set to null",
			pb:   &pb.MsgWithOneof{},
			bson: bson.D{{"msgWithRequired", bson.D{{"str", primitive.Null{}}}}},
		},
		{
			desc: "required field in extension missing",
			pb:   &pb.Real{},
			bson: bson.D{{"[bsonpb.extm]", bson.D{}}},
		},
		{
			desc: "required field in extension set to null",
			pb:   &pb.Real{},
			bson: bson.D{{"[bsonpb.extm]", bson.D{{"str", primitive.Null{}}}}},
		},
	}

	um := Unmarshaler{}
	for _, tc := range tests {
		b, mErr := bson.Marshal(tc.bson)
		if mErr != nil {
			t.Fatalf("marshaling bson to bytes failed: %v", mErr)
			return
		}
		if err := um.Unmarshal(b, tc.pb); err == nil {
			t.Errorf("%s: expecting error in unmarshaling with unset required fields %s", tc.desc, tc.bson)
		}
	}
}
