package bsonpb

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	proto3pb "github.com/golang/protobuf/proto/proto3_proto"
	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
	stpb "github.com/golang/protobuf/ptypes/struct"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	wpb "github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/romnnn/bsonpb/test_protos/test_objects"
	"github.com/romnnn/deepequal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var marshalingTests = []struct {
	desc      string
	marshaler Marshaler
	pb        proto.Message
	bson      bson.D
}{
	{"simple flat object", defaultMarshaler, simpleObject, simpleObjectOutputBSON},
	{"non-finite floats fields object", defaultMarshaler, nonFinites, nonFinitesBSON},
	{"repeated fields flat object", defaultMarshaler, repeatsObject, repeatsObjectBSON},
	{"nested message/enum flat object", defaultMarshaler, complexObject, complexObjectBSON},
	{"enum-string flat object", Marshaler{},
		&pb.Widget{Color: pb.Widget_BLUE.Enum()},
		bson.D{
			{"color", "BLUE"},
			{"rColor", bson.A{}},
			{"simple", primitive.Null{}},
			{"rSimple", bson.A{}},
			{"repeats", primitive.Null{}},
			{"rRepeats", bson.A{}},
		},
	},
	{"enum-value pretty object", Marshaler{EnumsAsInts: true, Omit: OmitOptions{All: true}},
		&pb.Widget{Color: pb.Widget_BLUE.Enum()},
		bson.D{{"color", pb.Widget_Color(2)}},
	},
	{"unknown enum value object", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Widget{
			Color:  pb.Widget_Color(1000).Enum(),
			RColor: []pb.Widget_Color{pb.Widget_RED},
		},
		bson.D{{"color", pb.Widget_Color(1000)}, {"rColor", bson.A{"RED"}}},
	},
	{"repeated proto3 enum", Marshaler{},
		&proto3pb.Message{RFunny: []proto3pb.Message_Humour{
			proto3pb.Message_PUNS,
			proto3pb.Message_SLAPSTICK,
		}},
		bson.D{
			{"name", ""},
			{"hilarity", "UNKNOWN"},
			{"heightInCm", uint32(0)},
			{"data", primitive.Binary{}},
			{"resultCount", int64(0)},
			{"trueScotsman", false},
			{"score", float32(0)},
			{"key", bson.A{}},
			{"shortKey", bson.A{}},
			{"nested", primitive.Null{}},
			{"rFunny", bson.A{"PUNS", "SLAPSTICK"}},
			{"terrain", bson.D{}},
			{"proto2Field", primitive.Null{}},
			{"proto2Value", bson.D{}},
			{"anything", primitive.Null{}},
			{"manyThings", bson.A{}},
			{"submessage", primitive.Null{}},
			{"children", bson.A{}},
			{"stringMap", bson.D{}},
		},
	},
	{"repeated proto3 enum as int", Marshaler{EnumsAsInts: true, Omit: OmitOptions{All: true}},
		&proto3pb.Message{RFunny: []proto3pb.Message_Humour{
			proto3pb.Message_PUNS,
			proto3pb.Message_SLAPSTICK,
		}},
		bson.D{
			{"rFunny", bson.A{proto3pb.Message_Humour(1), proto3pb.Message_Humour(2)}},
		},
	},
	{"empty value", Marshaler{Omit: OmitOptions{All: true}}, &pb.Simple3{}, bson.D{}},
	{"empty value emitted", defaultMarshaler, &pb.Simple3{}, bson.D{{"dub", float64(0)}}},
	{"empty repeated emitted", defaultMarshaler, &pb.SimpleSlice3{}, bson.D{{"slices", bson.A{}}}},
	{"empty map emitted", defaultMarshaler, &pb.SimpleMap3{}, bson.D{{"stringy", bson.D{}}}},
	{"nested struct null", defaultMarshaler, &pb.SimpleNull3{}, bson.D{{"simple", primitive.Null{}}}},
	{"map<int64, int32>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Nummy: map[int64]int32{1: 2, 3: 4}},
		bson.D{{"nummy", bson.D{{"1", int32(2)}, {"3", int32(4)}}}},
	},
	{"map<string, string>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Strry: map[string]string{"one": "two", "three": "four"}},
		bson.D{{"strry", bson.D{{"one", "two"}, {"three", "four"}}}},
	},
	{"map<int32, Object>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Objjy: map[int32]*pb.Simple3{1: {Dub: 1}}},
		bson.D{{"objjy", bson.D{{"1", bson.D{{"dub", float64(1)}}}}}},
	},
	{"map<int64, string>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Buggy: map[int64]string{1234: "yup"}},
		bson.D{{"buggy", bson.D{{"1234", "yup"}}}},
	},
	{"map<bool, bool>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Booly: map[bool]bool{false: true}},
		bson.D{{"booly", bson.D{{"false", true}}}},
	},
	{"map<string, enum>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{Enumy: map[string]pb.Numeral{"XIV": pb.Numeral_ROMAN}},
		bson.D{{"enumy", bson.D{{"XIV", "ROMAN"}}}},
	},
	{"map<string, enum as int>", Marshaler{Omit: OmitOptions{All: true}, EnumsAsInts: true},
		&pb.Mappy{Enumy: map[string]pb.Numeral{"XIV": pb.Numeral_ROMAN}},
		bson.D{{"enumy", bson.D{{"XIV", pb.Numeral(2)}}}},
	},
	{"map<int32, bool>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{S32Booly: map[int32]bool{1: true, 3: false, 10: true, 12: false}},
		bson.D{{"s32booly", bson.D{{"1", true}, {"3", false}, {"10", true}, {"12", false}}}},
	},
	{"map<int64, bool>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{S64Booly: map[int64]bool{1: true, 3: false, 10: true, 12: false}},
		bson.D{{"s64booly", bson.D{{"1", true}, {"3", false}, {"10", true}, {"12", false}}}},
	},
	{"map<uint32, bool>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{U32Booly: map[uint32]bool{1: true, 3: false, 10: true, 12: false}},
		bson.D{{"u32booly", bson.D{{"1", true}, {"3", false}, {"10", true}, {"12", false}}}},
	},
	{"map<uint64, bool>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Mappy{U64Booly: map[uint64]bool{1: true, 3: false, 10: true, 12: false}},
		bson.D{{"u64booly", bson.D{{"1", true}, {"3", false}, {"10", true}, {"12", false}}}},
	},
	{"proto2 map<int64, string>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Maps{MInt64Str: map[int64]string{213: "cat"}},
		bson.D{{"mInt64Str", bson.D{{"213", "cat"}}}},
	},
	{"proto2 map<bool, Object>", Marshaler{Omit: OmitOptions{All: true}},
		&pb.Maps{MBoolSimple: map[bool]*pb.Simple{true: {OInt32: proto.Int32(1)}}},
		bson.D{{"mBoolSimple", bson.D{{"true", bson.D{{"oInt32", int32(1)}}}}}},
	},
	{"oneof, not set", Marshaler{Omit: OmitOptions{All: true}},
		&pb.MsgWithOneof{},
		bson.D{},
	},
	{"oneof, set", Marshaler{Omit: OmitOptions{All: true}},
		&pb.MsgWithOneof{Union: &pb.MsgWithOneof_Title{"Grand Poobah"}},
		bson.D{{"title", "Grand Poobah"}},
	},
	{"force orig_name", Marshaler{Omit: OmitOptions{All: true}, OrigName: true},
		&pb.Simple{OInt32: proto.Int32(4)},
		bson.D{{"o_int32", int32(4)}},
	},
	{"proto2 extension", Marshaler{Omit: OmitOptions{All: true}}, realNumber, realNumberBSON},
	{"Any with message", Marshaler{Omit: OmitOptions{All: true}}, anySimple, anySimpleBSON},
	{"Any with WKT", Marshaler{Omit: OmitOptions{All: true}}, anyWellKnown, anyWellKnownBSON},
	{"Duration empty", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{}},
		bson.D{{"dur", float64(0)}},
	},
	{"Duration with secs", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 3}},
		bson.D{{"dur", float64(3)}},
	},
	{"Duration with -secs", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -3}},
		bson.D{{"dur", float64(-3)}},
	},
	{"Duration with nanos", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Nanos: 1e6}},
		bson.D{{"dur", float64(0.001)}},
	},
	{"Duration with -nanos", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Nanos: -1e6}},
		bson.D{{"dur", float64(-0.001)}},
	},
	{"Duration with large secs", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 1e8, Nanos: 1}},
		bson.D{{"dur", float64(1e+8)}},
	},
	{"Duration with 6-digit nanos", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Nanos: 1e4}},
		bson.D{{"dur", float64(1e-05)}},
	},
	{"Duration with 3-digit nanos", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Nanos: 1e6}},
		bson.D{{"dur", float64(0.001)}},
	},
	{"Duration with -secs -nanos", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -123, Nanos: -450}},
		bson.D{{"dur", float64(-123.00000045)}},
	},
	{"Duration max value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 3155760000, Nanos: 999999999}},
		bson.D{{"dur", float64(3.155760001e+09)}},
	},
	{"Duration min value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -3155760000, Nanos: -999999999}},
		bson.D{{"dur", float64(-3.155760001e+09)}},
	},
	{"Struct", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{St: &stpb.Struct{
			Fields: map[string]*stpb.Value{
				"one": {Kind: &stpb.Value_StringValue{"loneliest number"}},
				"two": {Kind: &stpb.Value_NullValue{stpb.NullValue_NULL_VALUE}},
			},
		}},
		bson.D{{"st", bson.D{{"one", "loneliest number"}, {"two", primitive.Null{}}}}},
	},
	{"empty ListValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Lv: &stpb.ListValue{}},
		bson.D{{"lv", bson.A{}}},
	},
	{"basic ListValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Lv: &stpb.ListValue{Values: []*stpb.Value{
			{Kind: &stpb.Value_StringValue{"x"}},
			{Kind: &stpb.Value_NullValue{}},
			{Kind: &stpb.Value_NumberValue{3}},
			{Kind: &stpb.Value_BoolValue{true}},
		}}},
		bson.D{{"lv", bson.A{"x", primitive.Null{}, float64(3), true}}},
	},
	{"Timestamp", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Ts: protoTimestamp(time.Unix(14e8, 21e6))},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(14e8, 21e6))}},
	},
	{"Timestamp", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 14e8, Nanos: 0}},
		bson.D{{"ts", primitive.NewDateTimeFromTime(time.Unix(14e8, 0))}},
	},
	{"number Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_NumberValue{1}}},
		bson.D{{"val", float64(1)}},
	},
	{"null Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_NullValue{stpb.NullValue_NULL_VALUE}}},
		bson.D{{"val", primitive.Null{}}},
	},
	{"string number value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Val: &stpb.Value{Kind: &stpb.Value_StringValue{"9223372036854775807"}}},
		bson.D{{"val", "9223372036854775807"}},
	},
	{"list of lists Value", Marshaler{Omit: OmitOptions{All: true}},
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
			}},
		}},
		bson.D{{"val", bson.A{"x", bson.A{bson.A{"y"}, "z"}}}},
	},
	{"DoubleValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Dbl: &wpb.DoubleValue{Value: 1.2}},
		bson.D{{"dbl", float64(1.2)}},
	},
	{"FloatValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Flt: &wpb.FloatValue{Value: 1.2}},
		bson.D{{"flt", float32(1.2)}},
	},
	{"Int64Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{I64: &wpb.Int64Value{Value: -3}},
		bson.D{{"i64", int64(-3)}},
	},
	{"UInt64Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{U64: &wpb.UInt64Value{Value: 3}},
		bson.D{{"u64", uint64(3)}},
	},
	{"Int32Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{I32: &wpb.Int32Value{Value: -4}},
		bson.D{{"i32", int32(-4)}},
	},
	{"UInt32Value", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{U32: &wpb.UInt32Value{Value: 4}},
		bson.D{{"u32", uint32(4)}},
	},
	{"BoolValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Bool: &wpb.BoolValue{Value: true}},
		bson.D{{"bool", true}},
	},
	{"StringValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Str: &wpb.StringValue{Value: "plush"}},
		bson.D{{"str", "plush"}},
	},
	{"BytesValue", Marshaler{Omit: OmitOptions{All: true}},
		&pb.KnownTypes{Bytes: &wpb.BytesValue{Value: []byte("wow")}},
		bson.D{{"bytes", primitive.Binary{Data: []byte("wow")}}},
	},

	{"required", Marshaler{Omit: OmitOptions{All: true}},
		&pb.MsgWithRequired{Str: proto.String("hello")},
		bson.D{{"str", "hello"}},
	},
	{"required bytes", Marshaler{Omit: OmitOptions{All: true}},
		&pb.MsgWithRequiredBytes{Byts: []byte{}},
		bson.D{{"byts", primitive.Binary{Data: []byte{}}}},
	},
}

func TestMarshaling(t *testing.T) {
	for _, tt := range marshalingTests {
		bson, err := tt.marshaler.Marshal(tt.pb)
		if err != nil {
			t.Errorf("%s: marshaling error: %v", tt.desc, err)
		}

		expected := tt.bson
		observed := bson
		if equal, err := deepequal.DeepEqual(observed, expected); !equal {
			t.Errorf("\n\n%s:\n Got [%v]\n\n Want [%v]\n\nError: %s", tt.desc, observed, expected, err.Error())
		}
	}
}

func TestMarshalingNil(t *testing.T) {
	var msg *pb.Simple
	m := Marshaler{}
	if _, err := m.Marshal(msg); err == nil {
		t.Errorf("mashaling nil returned no error")
	}
}

func TestMarshalIllegalTime(t *testing.T) {
	tests := []struct {
		pb   proto.Message
		fail bool
	}{
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 1, Nanos: 0}}, false},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -1, Nanos: 0}}, false},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 1, Nanos: -1}}, true},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -1, Nanos: 1}}, true},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 315576000001}}, true},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -315576000001}}, true},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: 1, Nanos: 1000000000}}, true},
		{&pb.KnownTypes{Dur: &durpb.Duration{Seconds: -1, Nanos: -1000000000}}, true},
		{&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 1, Nanos: 1}}, false},
		{&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 1, Nanos: -1}}, true},
		{&pb.KnownTypes{Ts: &tspb.Timestamp{Seconds: 1, Nanos: 1000000000}}, true},
	}
	for _, tt := range tests {
		m := Marshaler{}
		_, err := m.Marshal(tt.pb)
		if err == nil && tt.fail {
			t.Errorf("marshaler.Marshal(%v) = _, <nil>; want _, <non-nil>", tt.pb)
		}
		if err != nil && !tt.fail {
			t.Errorf("marshaler.Marshal(%v) = _, %v; want _, <nil>", tt.pb, err)
		}
	}
}

func TestMarshalJSONPBMarshaler(t *testing.T) {
	rawBson := bson.D{{"foo", "bar"}, {"baz", bson.A{int32(0), int32(1), int32(2), int32(3)}}}
	msg := newDynamicMessage(rawBson)
	marshaled, err := new(Marshaler).Marshal(msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling via BSONPBMarshaler: %v", err)
	}
	if equal, err := deepequal.DeepEqual(marshaled, rawBson); !equal {
		t.Errorf("marshaling BSON produced incorrect output: got %v, wanted %v\nError: %v", marshaled, rawBson, err)
	}
}

func TestMarshalAnyJSONPBMarshaler(t *testing.T) {
	rawBson := bson.D{{"foo", "bar"}, {"baz", bson.A{int32(0), int32(1), int32(2), int32(3)}}}
	msg := newDynamicMessage(rawBson)
	anyMsg, err := ptypes.MarshalAny(msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling to Any: %v", err)
	}
	marshaled, err := new(Marshaler).Marshal(anyMsg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling Any to BSON: %v", err)
	}
	// after custom marshaling, it's round-tripped through BSON decoding/encoding already,
	// so the keys are sorted, whitespace is compacted, and "@type" key has been added
	expected := bson.D{
		{"foo", "bar"},
		{"baz", bson.A{int32(0), int32(1), int32(2), int32(3)}},
		// @type always at last as of now
		{"@type", fmt.Sprintf("type.googleapis.com/%s", dynamicMessageName)},
	}
	if equal, err := deepequal.DeepEqual(marshaled, expected); !equal {
		t.Errorf("marshaling BSON produced incorrect output: got %v, wanted %v\nError: %v", marshaled, expected, err)
	}
}

func TestMarshalWithCustomValidation(t *testing.T) {
	rawBson := bson.D{{"foo", "bar"}, {"baz", bson.A{int32(0), int32(1), int32(2), int32(3)}}}
	msg := newDynamicMessage(rawBson)
	msg.Dummy = &dynamicMessage{}
	_, err := new(Marshaler).Marshal(msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling to BSON: %v", err)
	}
}

// Test marshaling message containing unset required fields should produce error.
func TestMarshalUnsetRequiredFields(t *testing.T) {
	msgExt := &pb.Real{}
	proto.SetExtension(msgExt, pb.E_Extm, &pb.MsgWithRequired{})

	tests := []struct {
		desc      string
		marshaler *Marshaler
		pb        proto.Message
	}{
		{
			desc:      "direct required field",
			marshaler: &Marshaler{},
			pb:        &pb.MsgWithRequired{},
		},
		{
			desc:      "direct required field + emit defaults",
			marshaler: &Marshaler{Omit: OmitOptions{All: true}},
			pb:        &pb.MsgWithRequired{},
		},
		{
			desc:      "indirect required field",
			marshaler: &Marshaler{},
			pb:        &pb.MsgWithIndirectRequired{Subm: &pb.MsgWithRequired{}},
		},
		{
			desc:      "indirect required field + emit defaults",
			marshaler: &Marshaler{Omit: OmitOptions{All: true}},
			pb:        &pb.MsgWithIndirectRequired{Subm: &pb.MsgWithRequired{}},
		},
		{
			desc:      "direct required wkt field",
			marshaler: &Marshaler{},
			pb:        &pb.MsgWithRequiredWKT{},
		},
		{
			desc:      "direct required wkt field + emit defaults",
			marshaler: &Marshaler{Omit: OmitOptions{All: true}},
			pb:        &pb.MsgWithRequiredWKT{},
		},
		{
			desc:      "direct required bytes field",
			marshaler: &Marshaler{},
			pb:        &pb.MsgWithRequiredBytes{},
		},
		{
			desc:      "required in map value",
			marshaler: &Marshaler{},
			pb: &pb.MsgWithIndirectRequired{
				MapField: map[string]*pb.MsgWithRequired{
					"key": {},
				},
			},
		},
		{
			desc:      "required in repeated item",
			marshaler: &Marshaler{},
			pb: &pb.MsgWithIndirectRequired{
				SliceField: []*pb.MsgWithRequired{
					{Str: proto.String("hello")},
					{},
				},
			},
		},
		{
			desc:      "required inside oneof",
			marshaler: &Marshaler{},
			pb: &pb.MsgWithOneof{
				Union: &pb.MsgWithOneof_MsgWithRequired{&pb.MsgWithRequired{}},
			},
		},
		{
			desc:      "required inside extension",
			marshaler: &Marshaler{},
			pb:        msgExt,
		},
	}

	for _, tc := range tests {
		if _, err := tc.marshaler.Marshal(tc.pb); err == nil {
			t.Errorf("%s: expecting error in marshaling with unset required fields %+v", tc.desc, tc.pb)
		}
	}
}
