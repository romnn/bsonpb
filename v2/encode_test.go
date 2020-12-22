package bsonpb

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/romnnn/deepequal"

	"google.golang.org/protobuf/proto"
	preg "google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protopack"

	pb2 "github.com/romnnn/bsonpb/internal/testprotos/v2/textpb2_proto"
	pb3 "github.com/romnnn/bsonpb/internal/testprotos/v2/textpb3_proto"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		desc    string
		mo      MarshalOptions
		input   proto.Message
		want    interface{}
		wantErr bool
		skip    bool
	}{{
		desc:  "proto2 optional scalars not set",
		input: &pb2.Scalars{},
		want:  bson.D{},
	}, {
		desc:  "proto3 scalars not set",
		input: &pb3.Scalars{},
		want:  bson.D{},
	},
		/*
			{
				desc:  "proto3 optional not set",
				input: &pb3.Proto3Optional{},
				want:  bson.D{},
			},
		*/
		{
			desc: "proto2 optional scalars set to zero values",
			input: &pb2.Scalars{
				OptBool:     proto.Bool(false),
				OptInt32:    proto.Int32(0),
				OptInt64:    proto.Int64(0),
				OptUint32:   proto.Uint32(0),
				OptUint64:   proto.Uint64(0),
				OptSint32:   proto.Int32(0),
				OptSint64:   proto.Int64(0),
				OptFixed32:  proto.Uint32(0),
				OptFixed64:  proto.Uint64(0),
				OptSfixed32: proto.Int32(0),
				OptSfixed64: proto.Int64(0),
				OptFloat:    proto.Float32(0),
				OptDouble:   proto.Float64(0),
				OptBytes:    []byte{},
				OptString:   proto.String(""),
			},
			want: bson.D{
				{Key: "optBool", Value: false},
				{Key: "optInt32", Value: int32(0)},
				{Key: "optInt64", Value: int64(0)},
				{Key: "optUint32", Value: uint32(0)},
				{Key: "optUint64", Value: uint64(0)},
				{Key: "optSint32", Value: int32(0)},
				{Key: "optSint64", Value: int64(0)},
				{Key: "optFixed32", Value: uint32(0)},
				{Key: "optFixed64", Value: uint64(0)},
				{Key: "optSfixed32", Value: int32(0)},
				{Key: "optSfixed64", Value: int64(0)},
				{Key: "optFloat", Value: float32(0)},
				{Key: "optDouble", Value: float64(0)},
				{Key: "optBytes", Value: primitive.Binary{Data: []byte{}}},
				{Key: "optString", Value: ""},
			},
		},
		/*{
				desc: "proto3 optional set to zero values",
				input: &pb3.Proto3Optional{
					OptBool:    proto.Bool(false),
					OptInt32:   proto.Int32(0),
					OptInt64:   proto.Int64(0),
					OptUint32:  proto.Uint32(0),
					OptUint64:  proto.Uint64(0),
					OptFloat:   proto.Float32(0),
					OptDouble:  proto.Float64(0),
					OptString:  proto.String(""),
					OptBytes:   []byte{},
					OptEnum:    pb3.Enum_ZERO.Enum(),
					OptMessage: &pb3.Nested{},
				},
				want: `{
		  "optBool": false,
		  "optInt32": 0,
		  "optInt64": "0",
		  "optUint32": 0,
		  "optUint64": "0",
		  "optFloat": 0,
		  "optDouble": 0,
		  "optString": "",
		  "optBytes": "",
		  "optEnum": "ZERO",
		  "optMessage": {}
		}`,
			},*/
		{
			desc: "proto2 optional scalars set to some values",
			input: &pb2.Scalars{
				OptBool:     proto.Bool(true),
				OptInt32:    proto.Int32(0xff),
				OptInt64:    proto.Int64(0xdeadbeef),
				OptUint32:   proto.Uint32(47),
				OptUint64:   proto.Uint64(0xdeadbeef),
				OptSint32:   proto.Int32(-1001),
				OptSint64:   proto.Int64(-0xffff),
				OptFixed64:  proto.Uint64(64),
				OptSfixed32: proto.Int32(-32),
				OptFloat:    proto.Float32(1.02),
				OptDouble:   proto.Float64(1.234),
				OptBytes:    []byte("谷歌"),
				OptString:   proto.String("谷歌"),
			},
			want: bson.D{
				{Key: "optBool", Value: true},
				{Key: "optInt32", Value: int32(255)},
				{Key: "optInt64", Value: int64(3735928559)},
				{Key: "optUint32", Value: uint32(47)},
				{Key: "optUint64", Value: uint64(3735928559)},
				{Key: "optSint32", Value: int32(-1001)},
				{Key: "optSint64", Value: int64(-65535)},
				{Key: "optFixed64", Value: uint64(64)},
				{Key: "optSfixed32", Value: int32(-32)},
				{Key: "optFloat", Value: float32(1.02)},
				{Key: "optDouble", Value: float64(1.234)},
				{Key: "optBytes", Value: primitive.Binary{Data: []byte("谷歌")}},
				{Key: "optString", Value: "谷歌"},
			},
		}, {
			desc: "string",
			input: &pb3.Scalars{
				SString: "谷歌",
			},
			want: bson.D{
				{Key: "sString", Value: "谷歌"},
			},
		}, {
			desc: "string with invalid UTF8",
			input: &pb3.Scalars{
				SString: "abc\xff",
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "float nan",
			input: &pb3.Scalars{
				SFloat: float32(math.NaN()),
			},
			want: bson.D{
				{Key: "sFloat", Value: float32(math.NaN())},
			},
		}, {
			desc: "float positive infinity",
			input: &pb3.Scalars{
				SFloat: float32(math.Inf(1)),
			},
			want: bson.D{
				{Key: "sFloat", Value: float32(math.Inf(1))},
			},
		}, {
			desc: "float negative infinity",
			input: &pb3.Scalars{
				SFloat: float32(math.Inf(-1)),
			},
			want: bson.D{
				{Key: "sFloat", Value: float32(math.Inf(-1))},
			},
		}, {
			desc: "double nan",
			input: &pb3.Scalars{
				SDouble: math.NaN(),
			},
			want: bson.D{
				{Key: "sDouble", Value: math.NaN()},
			},
		}, {
			desc: "double positive infinity",
			input: &pb3.Scalars{
				SDouble: math.Inf(1),
			},
			want: bson.D{
				{Key: "sDouble", Value: math.Inf(1)},
			},
		}, {
			desc: "double negative infinity",
			input: &pb3.Scalars{
				SDouble: math.Inf(-1),
			},
			want: bson.D{
				{Key: "sDouble", Value: math.Inf(-1)},
			},
		}, {
			desc:  "proto2 enum not set",
			input: &pb2.Enums{},
			want:  bson.D{},
		}, {
			desc: "proto2 enum set to zero value",
			input: &pb2.Enums{
				OptEnum:       pb2.Enum(0).Enum(),
				OptNestedEnum: pb2.Enums_NestedEnum(0).Enum(),
			},
			want: bson.D{
				{Key: "optEnum", Value: int64(0)},
				{Key: "optNestedEnum", Value: int64(0)},
			},
		}, {
			desc: "proto2 enum",
			input: &pb2.Enums{
				OptEnum:       pb2.Enum_ONE.Enum(),
				OptNestedEnum: pb2.Enums_UNO.Enum(),
			},
			want: bson.D{
				{Key: "optEnum", Value: "ONE"},
				{Key: "optNestedEnum", Value: "UNO"},
			},
		}, {
			desc: "proto2 enum set to numeric values",
			input: &pb2.Enums{
				OptEnum:       pb2.Enum(2).Enum(),
				OptNestedEnum: pb2.Enums_NestedEnum(2).Enum(),
			},
			want: bson.D{
				{Key: "optEnum", Value: "TWO"},
				{Key: "optNestedEnum", Value: "DOS"},
			},
		}, {
			desc: "proto2 enum set to unnamed numeric values",
			input: &pb2.Enums{
				OptEnum:       pb2.Enum(101).Enum(),
				OptNestedEnum: pb2.Enums_NestedEnum(-101).Enum(),
			},
			want: bson.D{
				{Key: "optEnum", Value: int64(101)},
				{Key: "optNestedEnum", Value: int64(-101)},
			},
		}, {
			desc:  "proto3 enum not set",
			input: &pb3.Enums{},
			want:  bson.D{},
		}, {
			desc: "proto3 enum set to zero value",
			input: &pb3.Enums{
				SEnum:       pb3.Enum_ZERO,
				SNestedEnum: pb3.Enums_CERO,
			},
			want: bson.D{},
		}, {
			desc: "proto3 enum",
			input: &pb3.Enums{
				SEnum:       pb3.Enum_ONE,
				SNestedEnum: pb3.Enums_UNO,
			},
			want: bson.D{
				{Key: "sEnum", Value: "ONE"},
				{Key: "sNestedEnum", Value: "UNO"},
			},
		}, {
			desc: "proto3 enum set to numeric values",
			input: &pb3.Enums{
				SEnum:       2,
				SNestedEnum: 2,
			},
			want: bson.D{
				{Key: "sEnum", Value: "TWO"},
				{Key: "sNestedEnum", Value: "DOS"},
			},
		}, {
			desc: "proto3 enum set to unnamed numeric values",
			input: &pb3.Enums{
				SEnum:       -47,
				SNestedEnum: 47,
			},
			want: bson.D{
				{Key: "sEnum", Value: int64(-47)},
				{Key: "sNestedEnum", Value: int64(47)},
			},
		}, {
			desc:  "proto2 nested message not set",
			input: &pb2.Nests{},
			want:  bson.D{},
		}, {
			desc: "proto2 nested message set to empty",
			input: &pb2.Nests{
				OptNested: &pb2.Nested{},
				Optgroup:  &pb2.Nests_OptGroup{},
			},
			want: bson.D{
				{Key: "optNested", Value: bson.D{}},
				{Key: "optgroup", Value: bson.D{}},
			},
		}, {
			desc: "proto2 nested messages",
			input: &pb2.Nests{
				OptNested: &pb2.Nested{
					OptString: proto.String("nested message"),
					OptNested: &pb2.Nested{
						OptString: proto.String("another nested message"),
					},
				},
			},
			want: bson.D{
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: "nested message"},
					{Key: "optNested", Value: bson.D{
						{Key: "optString", Value: "another nested message"},
					}},
				}},
			},
		}, {
			desc: "proto2 groups",
			input: &pb2.Nests{
				Optgroup: &pb2.Nests_OptGroup{
					OptString: proto.String("inside a group"),
					OptNested: &pb2.Nested{
						OptString: proto.String("nested message inside a group"),
					},
					Optnestedgroup: &pb2.Nests_OptGroup_OptNestedGroup{
						OptFixed32: proto.Uint32(47),
					},
				},
			},
			want: bson.D{
				{Key: "optgroup", Value: bson.D{
					{Key: "optString", Value: "inside a group"},
					{Key: "optNested", Value: bson.D{
						{Key: "optString", Value: "nested message inside a group"},
					}},
					{Key: "optnestedgroup", Value: bson.D{
						{Key: "optFixed32", Value: uint32(47)},
					}},
				}},
			},
		}, {
			desc:  "proto3 nested message not set",
			input: &pb3.Nests{},
			want:  bson.D{},
		}, {
			desc: "proto3 nested message set to empty",
			input: &pb3.Nests{
				SNested: &pb3.Nested{},
			},
			want: bson.D{
				{Key: "sNested", Value: bson.D{}},
			},
		}, {
			desc: "proto3 nested message",
			input: &pb3.Nests{
				SNested: &pb3.Nested{
					SString: "nested message",
					SNested: &pb3.Nested{
						SString: "another nested message",
					},
				},
			},
			want: bson.D{
				{Key: "sNested", Value: bson.D{
					{Key: "sString", Value: "nested message"},
					{Key: "sNested", Value: bson.D{
						{Key: "sString", Value: "another nested message"},
					}},
				}},
			},
		}, {
			desc:  "oneof not set",
			input: &pb3.Oneofs{},
			want:  bson.D{},
		}, {
			desc: "oneof set to empty string",
			input: &pb3.Oneofs{
				Union: &pb3.Oneofs_OneofString{},
			},
			want: bson.D{
				{Key: "oneofString", Value: ""},
			},
		}, {
			desc: "oneof set to string",
			input: &pb3.Oneofs{
				Union: &pb3.Oneofs_OneofString{
					OneofString: "hello",
				},
			},
			want: bson.D{
				{Key: "oneofString", Value: "hello"},
			},
		}, {
			desc: "oneof set to enum",
			input: &pb3.Oneofs{
				Union: &pb3.Oneofs_OneofEnum{
					OneofEnum: pb3.Enum_ZERO,
				},
			},
			want: bson.D{
				{Key: "oneofEnum", Value: "ZERO"},
			},
		}, {
			desc: "oneof set to empty message",
			input: &pb3.Oneofs{
				Union: &pb3.Oneofs_OneofNested{
					OneofNested: &pb3.Nested{},
				},
			},
			want: bson.D{
				{Key: "oneofNested", Value: bson.D{}},
			},
		}, {
			desc: "oneof set to message",
			input: &pb3.Oneofs{
				Union: &pb3.Oneofs_OneofNested{
					OneofNested: &pb3.Nested{
						SString: "nested message",
					},
				},
			},
			want: bson.D{
				{Key: "oneofNested", Value: bson.D{
					{Key: "sString", Value: "nested message"},
				}},
			},
		}, {
			desc:  "repeated fields not set",
			input: &pb2.Repeats{},
			want:  bson.D{},
		}, {
			desc: "repeated fields set to empty slices",
			input: &pb2.Repeats{
				RptBool:   []bool{},
				RptInt32:  []int32{},
				RptInt64:  []int64{},
				RptUint32: []uint32{},
				RptUint64: []uint64{},
				RptFloat:  []float32{},
				RptDouble: []float64{},
				RptBytes:  [][]byte{},
			},
			want: bson.D{},
		}, {
			desc: "repeated fields set to some values",
			input: &pb2.Repeats{
				RptBool:   []bool{true, false, true, true},
				RptInt32:  []int32{1, 6, 0, 0},
				RptInt64:  []int64{-64, 47},
				RptUint32: []uint32{0xff, 0xffff},
				RptUint64: []uint64{0xdeadbeef},
				RptFloat:  []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1)), 1.034},
				RptDouble: []float64{math.NaN(), math.Inf(1), math.Inf(-1), 1.23e-308},
				RptString: []string{"hello", "世界"},
				RptBytes: [][]byte{
					[]byte("hello"),
					[]byte("\xe4\xb8\x96\xe7\x95\x8c"),
				},
			},
			want: bson.D{
				{Key: "rptBool", Value: bson.A{
					true,
					false,
					true,
					true,
				}},
				{Key: "rptInt32", Value: bson.A{
					int32(1),
					int32(6),
					int32(0),
					int32(0),
				}},
				{Key: "rptInt64", Value: bson.A{
					int64(-64),
					int64(47),
				}},
				{Key: "rptUint32", Value: bson.A{
					uint32(255),
					uint32(65535),
				}},
				{Key: "rptUint64", Value: bson.A{
					uint64(3735928559),
				}},
				{Key: "rptFloat", Value: bson.A{
					float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1)), float32(1.034),
				}},
				{Key: "rptDouble", Value: bson.A{
					math.NaN(), math.Inf(1), math.Inf(-1), float64(1.23e-308),
				}},
				{Key: "rptString", Value: bson.A{
					"hello",
					"世界",
				}},
				{Key: "rptBytes", Value: bson.A{
					primitive.Binary{Data: []byte("hello")},
					primitive.Binary{Data: []byte("\xe4\xb8\x96\xe7\x95\x8c")},
				}},
			},
		}, {
			desc: "repeated enums",
			input: &pb2.Enums{
				RptEnum:       []pb2.Enum{pb2.Enum_ONE, 2, pb2.Enum_TEN, 42},
				RptNestedEnum: []pb2.Enums_NestedEnum{2, 47, 10},
			},
			want: bson.D{
				{Key: "rptEnum", Value: bson.A{
					"ONE",
					"TWO",
					"TEN",
					int64(42),
				}},
				{Key: "rptNestedEnum", Value: bson.A{
					"DOS",
					int64(47),
					"DIEZ",
				}},
			},
		}, {
			desc: "repeated messages set to empty",
			input: &pb2.Nests{
				RptNested: []*pb2.Nested{},
				Rptgroup:  []*pb2.Nests_RptGroup{},
			},
			want: bson.D{},
		}, {
			desc: "repeated messages",
			input: &pb2.Nests{
				RptNested: []*pb2.Nested{
					{
						OptString: proto.String("repeat nested one"),
					},
					{
						OptString: proto.String("repeat nested two"),
						OptNested: &pb2.Nested{
							OptString: proto.String("inside repeat nested two"),
						},
					},
					{},
				},
			},
			want: bson.D{
				{Key: "rptNested", Value: bson.A{
					bson.D{{Key: "optString", Value: "repeat nested one"}},
					bson.D{
						{Key: "optString", Value: "repeat nested two"},
						{Key: "optNested", Value: bson.D{
							{Key: "optString", Value: "inside repeat nested two"},
						}},
					},
					bson.D{},
				}},
			},
		}, {
			desc: "repeated messages contains nil value",
			input: &pb2.Nests{
				RptNested: []*pb2.Nested{nil, {}},
			},
			want: bson.D{
				{Key: "rptNested", Value: bson.A{
					bson.D{},
					bson.D{},
				}},
			},
		}, {
			desc: "repeated groups",
			input: &pb2.Nests{
				Rptgroup: []*pb2.Nests_RptGroup{
					{
						RptString: []string{"hello", "world"},
					},
					{},
					nil,
				},
			},
			want: bson.D{
				{Key: "rptgroup", Value: bson.A{
					bson.D{
						{Key: "rptString", Value: bson.A{"hello", "world"}},
					},
					bson.D{},
					bson.D{},
				}},
			},
		}, {
			desc:  "map fields not set",
			input: &pb3.Maps{},
			want:  bson.D{},
		}, {
			desc: "map fields set to empty",
			input: &pb3.Maps{
				Int32ToStr:   map[int32]string{},
				BoolToUint32: map[bool]uint32{},
				Uint64ToEnum: map[uint64]pb3.Enum{},
				StrToNested:  map[string]*pb3.Nested{},
				StrToOneofs:  map[string]*pb3.Oneofs{},
			},
			want: bson.D{},
		}, {
			desc: "map fields 1",
			input: &pb3.Maps{
				BoolToUint32: map[bool]uint32{
					true:  42,
					false: 101,
				},
			},
			want: bson.D{
				{Key: "boolToUint32", Value: bson.D{
					{Key: "false", Value: uint32(101)},
					{Key: "true", Value: uint32(42)},
				}},
			},
		}, {
			desc: "map fields 2",
			input: &pb3.Maps{
				Int32ToStr: map[int32]string{
					-101: "-101",
					0xff: "0xff",
					0:    "zero",
				},
			},
			want: bson.D{
				{Key: "int32ToStr", Value: bson.D{
					{Key: "-101", Value: "-101"},
					{Key: "0", Value: "zero"},
					{Key: "255", Value: "0xff"},
				}},
			},
		}, {
			desc: "map fields 3",
			input: &pb3.Maps{
				Uint64ToEnum: map[uint64]pb3.Enum{
					1:  pb3.Enum_ONE,
					2:  pb3.Enum_TWO,
					10: pb3.Enum_TEN,
					47: 47,
				},
			},
			want: bson.D{
				{Key: "uint64ToEnum", Value: bson.D{
					{Key: "1", Value: "ONE"},
					{Key: "2", Value: "TWO"},
					{Key: "10", Value: "TEN"},
					{Key: "47", Value: int64(47)},
				}},
			},
		}, {
			desc: "map fields 4",
			input: &pb3.Maps{
				StrToNested: map[string]*pb3.Nested{
					"nested": &pb3.Nested{
						SString: "nested in a map",
					},
				},
			},
			want: bson.D{
				{Key: "strToNested", Value: bson.D{
					{Key: "nested", Value: bson.D{
						{Key: "sString", Value: "nested in a map"},
					}},
				}},
			},
		}, {
			desc: "map fields 5",
			input: &pb3.Maps{
				StrToOneofs: map[string]*pb3.Oneofs{
					"string": &pb3.Oneofs{
						Union: &pb3.Oneofs_OneofString{
							OneofString: "hello",
						},
					},
					"nested": &pb3.Oneofs{
						Union: &pb3.Oneofs_OneofNested{
							OneofNested: &pb3.Nested{
								SString: "nested oneof in map field value",
							},
						},
					},
				},
			},
			want: bson.D{
				{Key: "strToOneofs", Value: bson.D{
					{Key: "nested", Value: bson.D{
						{Key: "oneofNested", Value: bson.D{
							{Key: "sString", Value: "nested oneof in map field value"},
						}},
					}},
					{Key: "string", Value: bson.D{
						{Key: "oneofString", Value: "hello"},
					}},
				}},
			},
		}, {
			desc: "map field contains nil value",
			input: &pb3.Maps{
				StrToNested: map[string]*pb3.Nested{
					"nil": nil,
				},
			},
			want: bson.D{
				{Key: "strToNested", Value: bson.D{
					{Key: "nil", Value: bson.D{}},
				}},
			},
		}, {
			desc:    "required fields not set",
			input:   &pb2.Requireds{},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "required fields partially set",
			input: &pb2.Requireds{
				ReqBool:     proto.Bool(false),
				ReqSfixed64: proto.Int64(0),
				ReqDouble:   proto.Float64(1.23),
				ReqString:   proto.String("hello"),
				ReqEnum:     pb2.Enum_ONE.Enum(),
			},
			want: bson.D{
				{Key: "reqBool", Value: false},
				{Key: "reqSfixed64", Value: int64(0)},
				{Key: "reqDouble", Value: float64(1.23)},
				{Key: "reqString", Value: "hello"},
				{Key: "reqEnum", Value: "ONE"},
			},
			wantErr: true,
		}, {
			desc: "required fields not set with AllowPartial",
			mo:   MarshalOptions{AllowPartial: true},
			input: &pb2.Requireds{
				ReqBool:     proto.Bool(false),
				ReqSfixed64: proto.Int64(0),
				ReqDouble:   proto.Float64(1.23),
				ReqString:   proto.String("hello"),
				ReqEnum:     pb2.Enum_ONE.Enum(),
			},
			want: bson.D{
				{Key: "reqBool", Value: false},
				{Key: "reqSfixed64", Value: int64(0)},
				{Key: "reqDouble", Value: float64(1.23)},
				{Key: "reqString", Value: "hello"},
				{Key: "reqEnum", Value: "ONE"},
			},
		}, {
			desc: "required fields all set",
			input: &pb2.Requireds{
				ReqBool:     proto.Bool(false),
				ReqSfixed64: proto.Int64(0),
				ReqDouble:   proto.Float64(1.23),
				ReqString:   proto.String("hello"),
				ReqEnum:     pb2.Enum_ONE.Enum(),
				ReqNested:   &pb2.Nested{},
			},
			want: bson.D{
				{Key: "reqBool", Value: false},
				{Key: "reqSfixed64", Value: int64(0)},
				{Key: "reqDouble", Value: float64(1.23)},
				{Key: "reqString", Value: "hello"},
				{Key: "reqEnum", Value: "ONE"},
				{Key: "reqNested", Value: bson.D{}},
			},
		}, {
			desc: "indirect required field",
			input: &pb2.IndirectRequired{
				OptNested: &pb2.NestedWithRequired{},
			},
			want: bson.D{
				{Key: "optNested", Value: bson.D{}},
			},
			wantErr: true,
		}, {
			desc: "indirect required field with AllowPartial",
			mo:   MarshalOptions{AllowPartial: true},
			input: &pb2.IndirectRequired{
				OptNested: &pb2.NestedWithRequired{},
			},
			want: bson.D{
				{Key: "optNested", Value: bson.D{}},
			},
		}, {
			desc: "indirect required field in empty repeated",
			input: &pb2.IndirectRequired{
				RptNested: []*pb2.NestedWithRequired{},
			},
			want: bson.D{},
		}, {
			desc: "indirect required field in repeated",
			input: &pb2.IndirectRequired{
				RptNested: []*pb2.NestedWithRequired{
					&pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "rptNested", Value: bson.A{bson.D{}}},
			},
			wantErr: true,
		}, {
			desc: "indirect required field in repeated with AllowPartial",
			mo:   MarshalOptions{AllowPartial: true},
			input: &pb2.IndirectRequired{
				RptNested: []*pb2.NestedWithRequired{
					&pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "rptNested", Value: bson.A{bson.D{}}},
			},
		}, {
			desc: "indirect required field in empty map",
			input: &pb2.IndirectRequired{
				StrToNested: map[string]*pb2.NestedWithRequired{},
			},
			want: bson.D{},
		}, {
			desc: "indirect required field in map",
			input: &pb2.IndirectRequired{
				StrToNested: map[string]*pb2.NestedWithRequired{
					"fail": &pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "strToNested", Value: bson.D{
					{Key: "fail", Value: bson.D{}},
				}},
			},
			wantErr: true,
		}, {
			desc: "indirect required field in map with AllowPartial",
			mo:   MarshalOptions{AllowPartial: true},
			input: &pb2.IndirectRequired{
				StrToNested: map[string]*pb2.NestedWithRequired{
					"fail": &pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "strToNested", Value: bson.D{
					{Key: "fail", Value: bson.D{}},
				}},
			},
		}, {
			desc: "indirect required field in oneof",
			input: &pb2.IndirectRequired{
				Union: &pb2.IndirectRequired_OneofNested{
					OneofNested: &pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "oneofNested", Value: bson.D{}},
			},
			wantErr: true,
		}, {
			desc: "indirect required field in oneof with AllowPartial",
			mo:   MarshalOptions{AllowPartial: true},
			input: &pb2.IndirectRequired{
				Union: &pb2.IndirectRequired_OneofNested{
					OneofNested: &pb2.NestedWithRequired{},
				},
			},
			want: bson.D{
				{Key: "oneofNested", Value: bson.D{}},
			},
		}, {
			desc: "unknown fields are ignored",
			input: func() proto.Message {
				m := &pb2.Scalars{
					OptString: proto.String("no unknowns"),
				}
				m.ProtoReflect().SetUnknown(protopack.Message{
					protopack.Tag{101, protopack.BytesType}, protopack.String("hello world"),
				}.Marshal())
				return m
			}(),
			want: bson.D{
				{Key: "optString", Value: "no unknowns"},
			},
		}, {
			desc: "json_name",
			input: &pb3.JSONNames{
				SString: "json_name",
			},
			want: bson.D{
				{Key: "foo_bar", Value: "json_name"},
			},
		}, {
			desc: "extensions of non-repeated fields",
			input: func() proto.Message {
				m := &pb2.Extensions{
					OptString: proto.String("non-extension field"),
					OptBool:   proto.Bool(true),
					OptInt32:  proto.Int32(42),
				}
				proto.SetExtension(m, pb2.E_OptExtBool, true)
				proto.SetExtension(m, pb2.E_OptExtString, "extension field")
				proto.SetExtension(m, pb2.E_OptExtEnum, pb2.Enum_TEN)
				proto.SetExtension(m, pb2.E_OptExtNested, &pb2.Nested{
					OptString: proto.String("nested in an extension"),
					OptNested: &pb2.Nested{
						OptString: proto.String("another nested in an extension"),
					},
				})
				return m
			}(),
			want: bson.D{
				{Key: "optString", Value: "non-extension field"},
				{Key: "optBool", Value: true},
				{Key: "optInt32", Value: int32(42)},
				{Key: "[textpb2_proto.opt_ext_bool]", Value: true},
				{Key: "[textpb2_proto.opt_ext_enum]", Value: "TEN"},
				{Key: "[textpb2_proto.opt_ext_nested]", Value: bson.D{
					{Key: "optString", Value: "nested in an extension"},
					{Key: "optNested", Value: bson.D{
						{Key: "optString", Value: "another nested in an extension"},
					}},
				}},
				{Key: "[textpb2_proto.opt_ext_string]", Value: "extension field"},
			},
		}, {
			desc: "extensions of repeated fields",
			input: func() proto.Message {
				m := &pb2.Extensions{}
				proto.SetExtension(m, pb2.E_RptExtEnum, []pb2.Enum{pb2.Enum_TEN, 101, pb2.Enum_ONE})
				proto.SetExtension(m, pb2.E_RptExtFixed32, []uint32{42, 47})
				proto.SetExtension(m, pb2.E_RptExtNested, []*pb2.Nested{
					&pb2.Nested{OptString: proto.String("one")},
					&pb2.Nested{OptString: proto.String("two")},
					&pb2.Nested{OptString: proto.String("three")},
				})
				return m
			}(),
			want: bson.D{
				{Key: "[textpb2_proto.rpt_ext_enum]", Value: bson.A{
					"TEN",
					int64(101),
					"ONE",
				}},
				{Key: "[textpb2_proto.rpt_ext_fixed32]", Value: bson.A{
					uint32(42),
					uint32(47),
				}},
				{Key: "[textpb2_proto.rpt_ext_nested]", Value: bson.A{
					bson.D{{Key: "optString", Value: "one"}},
					bson.D{{Key: "optString", Value: "two"}},
					bson.D{{Key: "optString", Value: "three"}},
				}},
			},
		}, {
			desc: "extensions of non-repeated fields in another message",
			input: func() proto.Message {
				m := &pb2.Extensions{}
				proto.SetExtension(m, pb2.E_ExtensionsContainer_OptExtBool, true)
				proto.SetExtension(m, pb2.E_ExtensionsContainer_OptExtString, "extension field")
				proto.SetExtension(m, pb2.E_ExtensionsContainer_OptExtEnum, pb2.Enum_TEN)
				proto.SetExtension(m, pb2.E_ExtensionsContainer_OptExtNested, &pb2.Nested{
					OptString: proto.String("nested in an extension"),
					OptNested: &pb2.Nested{
						OptString: proto.String("another nested in an extension"),
					},
				})
				return m
			}(),
			want: bson.D{
				{Key: "[textpb2_proto.ExtensionsContainer.opt_ext_bool]", Value: true},
				{Key: "[textpb2_proto.ExtensionsContainer.opt_ext_enum]", Value: "TEN"},
				{Key: "[textpb2_proto.ExtensionsContainer.opt_ext_nested]", Value: bson.D{
					{Key: "optString", Value: "nested in an extension"},
					{Key: "optNested", Value: bson.D{
						{Key: "optString", Value: "another nested in an extension"},
					}},
				}},
				{Key: "[textpb2_proto.ExtensionsContainer.opt_ext_string]", Value: "extension field"},
			},
		}, {
			desc: "extensions of repeated fields in another message",
			input: func() proto.Message {
				m := &pb2.Extensions{
					OptString: proto.String("non-extension field"),
					OptBool:   proto.Bool(true),
					OptInt32:  proto.Int32(42),
				}
				proto.SetExtension(m, pb2.E_ExtensionsContainer_RptExtEnum, []pb2.Enum{pb2.Enum_TEN, 101, pb2.Enum_ONE})
				proto.SetExtension(m, pb2.E_ExtensionsContainer_RptExtString, []string{"hello", "world"})
				proto.SetExtension(m, pb2.E_ExtensionsContainer_RptExtNested, []*pb2.Nested{
					&pb2.Nested{OptString: proto.String("one")},
					&pb2.Nested{OptString: proto.String("two")},
					&pb2.Nested{OptString: proto.String("three")},
				})
				return m
			}(),
			want: bson.D{
				{Key: "optString", Value: "non-extension field"},
				{Key: "optBool", Value: true},
				{Key: "optInt32", Value: int32(42)},
				{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_enum]", Value: bson.A{
					"TEN",
					int64(101),
					"ONE",
				}},
				{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_nested]", Value: bson.A{
					bson.D{{Key: "optString", Value: "one"}},
					bson.D{{Key: "optString", Value: "two"}},
					bson.D{{Key: "optString", Value: "three"}},
				}},
				{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_string]", Value: bson.A{
					"hello",
					"world",
				}},
			},
		}, /* {
			desc: "MessageSet",
			input: func() proto.Message {
				m := &pb2.MessageSet{}
				proto.SetExtension(m, pb2.E_MessageSetExtension_MessageSetExtension, &pb2.MessageSetExtension{
					OptString: proto.String("a messageset extension"),
				})
				proto.SetExtension(m, pb2.E_MessageSetExtension_NotMessageSetExtension, &pb2.MessageSetExtension{
					OptString: proto.String("not a messageset extension"),
				})
				proto.SetExtension(m, pb2.E_MessageSetExtension_ExtNested, &pb2.Nested{
					OptString: proto.String("just a regular extension"),
				})
				return m
			}(),
			want: bson.D{
				{Key: "[textpb2_proto.ExtensionsContainer.ext_nested]", Value: bson.D{
					{Key: "optString", Value: "just a regular extension"},
				}},
				{Key: "[textpb2_proto.MessageSetExtension]", Value: bson.D{
					{Key: "optString", Value: "a messageset extension"},
				}},
				{Key: "[textpb2_proto.MessageSetExtension.not_message_set_extension]", Value: bson.D{
					{Key: "optString", Value: "not a messageset extension"},
				}},
			},
			skip: !protoLegacy,
		}, {
			desc: "not real MessageSet 1",
			input: func() proto.Message {
				m := &pb2.FakeMessageSet{}
				proto.SetExtension(m, pb2.E_FakeMessageSetExtension_MessageSetExtension, &pb2.FakeMessageSetExtension{
					OptString: proto.String("not a messageset extension"),
				})
				return m
			}(),
			want: bson.D{
				{Key: "[textpb2_proto.FakeMessageSetExtension.message_set_extension]", Value: bson.D{
					{Key: "optString", Value: "not a messageset extension"},
				}},
			},
			skip: !protoLegacy,
		}, {
			desc: "not real MessageSet 2",
			input: func() proto.Message {
				m := &pb2.MessageSet{}
				proto.SetExtension(m, pb2.E_MessageSetExtension, &pb2.FakeMessageSetExtension{
					OptString: proto.String("another not a messageset extension"),
				})
				return m
			}(),
			want: bson.D{
				{Key: "[textpb2_proto.message_set_extension]", Value: bson.D{
					{Key: "optString", Value: "another not a messageset extension"},
				}},
			},
			skip: !protoLegacy,
		}, */{
			desc:  "BoolValue empty",
			input: &wrapperspb.BoolValue{},
			want:  false,
		}, {
			desc:  "BoolValue",
			input: &wrapperspb.BoolValue{Value: true},
			want:  true,
		}, {
			desc:  "Int32Value empty",
			input: &wrapperspb.Int32Value{},
			want:  int32(0),
		}, {
			desc:  "Int32Value",
			input: &wrapperspb.Int32Value{Value: 42},
			want:  int32(42),
		}, {
			desc:  "Int64Value",
			input: &wrapperspb.Int64Value{Value: 42},
			want:  int64(42),
		}, {
			desc:  "UInt32Value",
			input: &wrapperspb.UInt32Value{Value: 42},
			want:  uint32(42),
		}, {
			desc:  "UInt64Value",
			input: &wrapperspb.UInt64Value{Value: 42},
			want:  uint64(42),
		}, {
			desc:  "FloatValue",
			input: &wrapperspb.FloatValue{Value: 1.02},
			want:  float32(1.02),
		}, {
			desc:  "FloatValue Infinity",
			input: &wrapperspb.FloatValue{Value: float32(math.Inf(-1))},
			want:  float32(math.Inf(-1)),
		}, {
			desc:  "DoubleValue",
			input: &wrapperspb.DoubleValue{Value: 1.02},
			want:  float64(1.02),
		}, {
			desc:  "DoubleValue NaN",
			input: &wrapperspb.DoubleValue{Value: math.NaN()},
			want:  float64(math.NaN()),
		}, {
			desc:  "StringValue empty",
			input: &wrapperspb.StringValue{},
			want:  "",
		}, {
			desc:  "StringValue",
			input: &wrapperspb.StringValue{Value: "谷歌"},
			want:  "谷歌",
		}, {
			desc:    "StringValue with invalid UTF8 error",
			input:   &wrapperspb.StringValue{Value: "abc\xff"},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "StringValue field with invalid UTF8 error",
			input: &pb2.KnownTypes{
				OptString: &wrapperspb.StringValue{Value: "abc\xff"},
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:  "BytesValue",
			input: &wrapperspb.BytesValue{Value: []byte("hello")},
			want:  primitive.Binary{Data: []byte("hello")},
		}, {
			desc:  "Empty",
			input: &emptypb.Empty{},
			want:  primitive.Null{},
		}, {
			desc:  "NullValue field",
			input: &pb2.KnownTypes{OptNull: new(structpb.NullValue)},
			want:  bson.D{{Key: "optNull", Value: primitive.Null{}}},
		}, {
			desc:    "Value empty",
			input:   &structpb.Value{},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Value empty field",
			input: &pb2.KnownTypes{
				OptValue: &structpb.Value{},
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:  "Value contains NullValue",
			input: &structpb.Value{Kind: &structpb.Value_NullValue{}},
			want:  primitive.Null{},
		}, {
			desc:  "Value contains BoolValue",
			input: &structpb.Value{Kind: &structpb.Value_BoolValue{}},
			want:  false,
		}, {
			desc:  "Value contains NumberValue",
			input: &structpb.Value{Kind: &structpb.Value_NumberValue{1.02}},
			want:  1.02,
		}, {
			desc:  "Value contains StringValue",
			input: &structpb.Value{Kind: &structpb.Value_StringValue{"hello"}},
			want:  "hello",
		}, {
			desc:    "Value contains StringValue with invalid UTF8",
			input:   &structpb.Value{Kind: &structpb.Value_StringValue{"\xff"}},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Value contains Struct",
			input: &structpb.Value{
				Kind: &structpb.Value_StructValue{
					&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"null":   {Kind: &structpb.Value_NullValue{}},
							"number": {Kind: &structpb.Value_NumberValue{}},
							"string": {Kind: &structpb.Value_StringValue{}},
							"struct": {Kind: &structpb.Value_StructValue{}},
							"list":   {Kind: &structpb.Value_ListValue{}},
							"bool":   {Kind: &structpb.Value_BoolValue{}},
						},
					},
				},
			},
			want: bson.D{
				{Key: "bool", Value: false},
				{Key: "list", Value: bson.A{}},
				{Key: "null", Value: primitive.Null{}},
				{Key: "number", Value: float64(0)},
				{Key: "string", Value: ""},
				{Key: "struct", Value: bson.D{}},
			},
		}, {
			desc: "Value contains ListValue",
			input: &structpb.Value{
				Kind: &structpb.Value_ListValue{
					&structpb.ListValue{
						Values: []*structpb.Value{
							{Kind: &structpb.Value_BoolValue{}},
							{Kind: &structpb.Value_NullValue{}},
							{Kind: &structpb.Value_NumberValue{}},
							{Kind: &structpb.Value_StringValue{}},
							{Kind: &structpb.Value_StructValue{}},
							{Kind: &structpb.Value_ListValue{}},
						},
					},
				},
			},
			want: bson.A{
				false,
				primitive.Null{},
				float64(0),
				"",
				bson.D{},
				bson.A{},
			},
		}, {
			desc:  "Struct with nil map",
			input: &structpb.Struct{},
			want:  bson.D{},
		}, {
			desc: "Struct with empty map",
			input: &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			},
			want: bson.D{},
		}, {
			desc: "Struct",
			input: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"bool":   {Kind: &structpb.Value_BoolValue{true}},
					"null":   {Kind: &structpb.Value_NullValue{}},
					"number": {Kind: &structpb.Value_NumberValue{3.1415}},
					"string": {Kind: &structpb.Value_StringValue{"hello"}},
					"struct": {
						Kind: &structpb.Value_StructValue{
							&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"string": {Kind: &structpb.Value_StringValue{"world"}},
								},
							},
						},
					},
					"list": {
						Kind: &structpb.Value_ListValue{
							&structpb.ListValue{
								Values: []*structpb.Value{
									{Kind: &structpb.Value_BoolValue{}},
									{Kind: &structpb.Value_NullValue{}},
									{Kind: &structpb.Value_NumberValue{}},
								},
							},
						},
					},
				},
			},
			want: bson.D{
				{Key: "bool", Value: true},
				{Key: "list", Value: bson.A{
					false,
					primitive.Null{},
					float64(0),
				}},
				{Key: "null", Value: primitive.Null{}},
				{Key: "number", Value: float64(3.1415)},
				{Key: "string", Value: "hello"},
				{Key: "struct", Value: bson.D{
					{Key: "string", Value: "world"},
				}},
			},
		}, {
			desc: "Struct message with invalid UTF8 string",
			input: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"string": {Kind: &structpb.Value_StringValue{"\xff"}},
				},
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:  "ListValue with nil values",
			input: &structpb.ListValue{},
			want:  bson.A{},
		}, {
			desc: "ListValue with empty values",
			input: &structpb.ListValue{
				Values: []*structpb.Value{},
			},
			want: bson.A{},
		}, {
			desc: "ListValue",
			input: &structpb.ListValue{
				Values: []*structpb.Value{
					{Kind: &structpb.Value_BoolValue{true}},
					{Kind: &structpb.Value_NullValue{}},
					{Kind: &structpb.Value_NumberValue{3.1415}},
					{Kind: &structpb.Value_StringValue{"hello"}},
					{
						Kind: &structpb.Value_ListValue{
							&structpb.ListValue{
								Values: []*structpb.Value{
									{Kind: &structpb.Value_BoolValue{}},
									{Kind: &structpb.Value_NullValue{}},
									{Kind: &structpb.Value_NumberValue{}},
								},
							},
						},
					},
					{
						Kind: &structpb.Value_StructValue{
							&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"string": {Kind: &structpb.Value_StringValue{"world"}},
								},
							},
						},
					},
				},
			},
			want: bson.A{
				true,
				primitive.Null{},
				float64(3.1415),
				"hello",
				bson.A{
					false,
					primitive.Null{},
					float64(0),
				},
				bson.D{
					{Key: "string", Value: "world"},
				},
			},
		}, {
			desc: "ListValue with invalid UTF8 string",
			input: &structpb.ListValue{
				Values: []*structpb.Value{
					{Kind: &structpb.Value_StringValue{"\xff"}},
				},
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:  "Duration empty",
			input: &durationpb.Duration{},
			want:  bson.D{{Key: "Seconds", Value: int64(0)}, {Key: "Nanos", Value: int64(0)}},
		}, {
			desc:  "Duration with secs",
			input: &durationpb.Duration{Seconds: 3},
			want:  bson.D{{Key: "Seconds", Value: int64(3)}, {Key: "Nanos", Value: int64(0)}},
		}, {
			desc:  "Duration with -secs",
			input: &durationpb.Duration{Seconds: -3},
			want:  bson.D{{Key: "Seconds", Value: int64(-3)}, {Key: "Nanos", Value: int64(0)}},
		}, {
			desc:  "Duration with nanos",
			input: &durationpb.Duration{Nanos: 1e6},
			want:  bson.D{{Key: "Seconds", Value: int64(0)}, {Key: "Nanos", Value: int64(1e6)}},
		}, {
			desc:  "Duration with -nanos",
			input: &durationpb.Duration{Nanos: -1e6},
			want:  bson.D{{Key: "Seconds", Value: int64(0)}, {Key: "Nanos", Value: int64(-1e6)}},
		}, {
			desc:  "Duration with large secs",
			input: &durationpb.Duration{Seconds: 1e10, Nanos: 1},
			want:  bson.D{{Key: "Seconds", Value: int64(1e10)}, {Key: "Nanos", Value: int64(1)}},
		}, {
			desc:  "Duration with 6-digit nanos",
			input: &durationpb.Duration{Nanos: 1e4},
			want:  bson.D{{Key: "Seconds", Value: int64(0)}, {Key: "Nanos", Value: int64(1e4)}},
		}, {
			desc:  "Duration with 3-digit nanos",
			input: &durationpb.Duration{Nanos: 1e6},
			want:  bson.D{{Key: "Seconds", Value: int64(0)}, {Key: "Nanos", Value: int64(1e6)}},
		}, {
			desc:  "Duration with -secs -nanos",
			input: &durationpb.Duration{Seconds: -123, Nanos: -450},
			want:  bson.D{{Key: "Seconds", Value: int64(-123)}, {Key: "Nanos", Value: int64(-450)}},
		}, {
			desc:  "Duration max value",
			input: &durationpb.Duration{Seconds: 315576000000, Nanos: 999999999},
			want:  bson.D{{Key: "Seconds", Value: int64(315576000000)}, {Key: "Nanos", Value: int64(999999999)}},
		}, {
			desc:  "Duration min value",
			input: &durationpb.Duration{Seconds: -315576000000, Nanos: -999999999},
			want:  bson.D{{Key: "Seconds", Value: int64(-315576000000)}, {Key: "Nanos", Value: int64(-999999999)}},
		}, {
			desc:    "Duration with +secs -nanos",
			input:   &durationpb.Duration{Seconds: 1, Nanos: -1},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Duration with -secs +nanos",
			input:   &durationpb.Duration{Seconds: -1, Nanos: 1},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Duration with +secs out of range",
			input:   &durationpb.Duration{Seconds: 315576000001},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Duration with -secs out of range",
			input:   &durationpb.Duration{Seconds: -315576000001},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Duration with +nanos out of range",
			input:   &durationpb.Duration{Seconds: 0, Nanos: 1e9},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Duration with -nanos out of range",
			input:   &durationpb.Duration{Seconds: 0, Nanos: -1e9},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:  "Timestamp zero",
			input: &timestamppb.Timestamp{},
			want:  primitive.NewDateTimeFromTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)),
		}, {
			desc:  "Timestamp",
			input: &timestamppb.Timestamp{Seconds: 1553036601},
			want:  primitive.NewDateTimeFromTime(time.Unix(1553036601, 0)),
		}, {
			desc:  "Timestamp with nanos",
			input: &timestamppb.Timestamp{Seconds: 1553036601, Nanos: 1},
			want:  primitive.NewDateTimeFromTime(time.Unix(1553036601, 1)),
		}, {
			desc:  "Timestamp with 6-digit nanos",
			input: &timestamppb.Timestamp{Nanos: 1e3},
			want:  primitive.NewDateTimeFromTime(time.Unix(0, 1e3)),
		}, {
			desc:  "Timestamp with 3-digit nanos",
			input: &timestamppb.Timestamp{Nanos: 1e7},
			want:  primitive.NewDateTimeFromTime(time.Unix(0, 1e7)),
		}, {
			desc:  "Timestamp max value",
			input: &timestamppb.Timestamp{Seconds: 253402300799, Nanos: 999999999},
			want:  primitive.NewDateTimeFromTime(time.Unix(253402300799, 999999999)),
		}, {
			desc:  "Timestamp min value",
			input: &timestamppb.Timestamp{Seconds: -62135596800},
			want:  primitive.NewDateTimeFromTime(time.Unix(-62135596800, 0)),
		}, {
			desc:    "Timestamp with +secs out of range",
			input:   &timestamppb.Timestamp{Seconds: 253402300800},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Timestamp with -secs out of range",
			input:   &timestamppb.Timestamp{Seconds: -62135596801},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Timestamp with -nanos",
			input:   &timestamppb.Timestamp{Nanos: -1},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc:    "Timestamp with +nanos out of range",
			input:   &timestamppb.Timestamp{Nanos: 1e9},
			want:    bson.D{},
			wantErr: true,
		}, /* {
			desc:  "FieldMask empty",
			input: &fieldmaskpb.FieldMask{},
			want:  "",
		}, {
			desc: "FieldMask",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{
					"foo",
					"foo_bar",
					"foo.bar_qux",
					"_foo",
				},
			},
			want: `"foo,fooBar,foo.barQux,Foo"`,
		}, {
			desc: "FieldMask empty string path",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{""},
			},
			wantErr: true,
		}, {
			desc: "FieldMask path contains spaces only",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{"  "},
			},
			wantErr: true,
		}, {
			desc: "FieldMask irreversible error 1",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{"foo_"},
			},
			wantErr: true,
		}, {
			desc: "FieldMask irreversible error 2",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{"foo__bar"},
			},
			wantErr: true,
		}, {
			desc: "FieldMask invalid char",
			input: &fieldmaskpb.FieldMask{
				Paths: []string{"foo@bar"},
			},
			wantErr: true,
		}, */{
			desc:  "Any empty",
			input: &anypb.Any{},
			want:  bson.D{},
		}, {
			desc: "Any with non-custom message",
			input: func() proto.Message {
				m := &pb2.Nested{
					OptString: proto.String("embedded inside Any"),
					OptNested: &pb2.Nested{
						OptString: proto.String("inception"),
					},
				}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "foo/textpb2_proto.Nested",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "foo/textpb2_proto.Nested"},
				{Key: "optString", Value: "embedded inside Any"},
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: "inception"},
				}},
			},
		}, {
			desc:  "Any with empty embedded message",
			input: &anypb.Any{TypeUrl: "foo/textpb2_proto.Nested"},
			want: bson.D{
				{Key: "@type", Value: "foo/textpb2_proto.Nested"},
			},
		}, {
			desc:    "Any without registered type",
			mo:      MarshalOptions{Resolver: new(preg.Types)},
			input:   &anypb.Any{TypeUrl: "foo/textpb2_proto.Nested"},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with missing required",
			input: func() proto.Message {
				m := &pb2.PartialRequired{
					OptString: proto.String("embedded inside Any"),
				}
				b, err := proto.MarshalOptions{
					AllowPartial:  true,
					Deterministic: true,
				}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: string(m.ProtoReflect().Descriptor().FullName()),
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "textpb2_proto.PartialRequired"},
				{Key: "optString", Value: "embedded inside Any"},
			},
		}, {
			desc: "Any with partial required and AllowPartial",
			mo: MarshalOptions{
				AllowPartial: true,
			},
			input: func() proto.Message {
				m := &pb2.PartialRequired{
					OptString: proto.String("embedded inside Any"),
				}
				b, err := proto.MarshalOptions{
					AllowPartial:  true,
					Deterministic: true,
				}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: string(m.ProtoReflect().Descriptor().FullName()),
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "textpb2_proto.PartialRequired"},
				{Key: "optString", Value: "embedded inside Any"},
			},
		}, {
			desc: "Any with EmitUnpopulated",
			mo: MarshalOptions{
				EmitUnpopulated: true,
			},
			input: func() proto.Message {
				return &anypb.Any{
					TypeUrl: string(new(pb3.Scalars).ProtoReflect().Descriptor().FullName()),
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "textpb3_proto.Scalars"},
				{Key: "sBool", Value: false},
				{Key: "sInt32", Value: int32(0)},
				{Key: "sInt64", Value: int64(0)},
				{Key: "sUint32", Value: uint32(0)},
				{Key: "sUint64", Value: uint64(0)},
				{Key: "sSint32", Value: int32(0)},
				{Key: "sSint64", Value: int64(0)},
				{Key: "sFixed32", Value: uint32(0)},
				{Key: "sFixed64", Value: uint64(0)},
				{Key: "sSfixed32", Value: int32(0)},
				{Key: "sSfixed64", Value: int64(0)},
				{Key: "sFloat", Value: float32(0)},
				{Key: "sDouble", Value: float64(0)},
				{Key: "sBytes", Value: primitive.Binary{}},
				{Key: "sString", Value: ""},
			},
		}, {
			desc: "Any with invalid UTF8",
			input: func() proto.Message {
				m := &pb2.Nested{
					OptString: proto.String("abc\xff"),
				}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "foo/pb2.Nested",
					Value:   b,
				}
			}(),
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with invalid value",
			input: &anypb.Any{
				TypeUrl: "foo/pb2.Nested",
				Value:   []byte("\x80"),
			},
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with BoolValue",
			input: func() proto.Message {
				m := &wrapperspb.BoolValue{Value: true}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.BoolValue",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "type.googleapis.com/google.protobuf.BoolValue"},
				{Key: "value", Value: true},
			},
		}, {
			desc: "Any with Empty",
			input: func() proto.Message {
				m := &emptypb.Empty{}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.Empty",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "type.googleapis.com/google.protobuf.Empty"},
				{Key: "value", Value: primitive.Null{}},
			},
		}, {
			desc: "Any with StringValue containing invalid UTF8",
			input: func() proto.Message {
				m := &wrapperspb.StringValue{Value: "abcd"}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "google.protobuf.StringValue",
					Value:   bytes.Replace(b, []byte("abcd"), []byte("abc\xff"), -1),
				}
			}(),
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with Int64Value",
			input: func() proto.Message {
				m := &wrapperspb.Int64Value{Value: 42}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "google.protobuf.Int64Value",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "google.protobuf.Int64Value"},
				{Key: "value", Value: int64(42)},
			},
		}, {
			desc: "Any with Duration",
			input: func() proto.Message {
				m := &durationpb.Duration{}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.Duration",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "type.googleapis.com/google.protobuf.Duration"},
				{Key: "value", Value: bson.D{
					{Key: "Seconds", Value: int64(0)},
					{Key: "Nanos", Value: int64(0)},
				}},
			},
		}, {
			desc: "Any with empty Value",
			input: func() proto.Message {
				m := &structpb.Value{}
				b, err := proto.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.Value",
					Value:   b,
				}
			}(),
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with Value of StringValue",
			input: func() proto.Message {
				m := &structpb.Value{Kind: &structpb.Value_StringValue{"abcd"}}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.Value",
					Value:   bytes.Replace(b, []byte("abcd"), []byte("abc\xff"), -1),
				}
			}(),
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "Any with Value of NullValue",
			input: func() proto.Message {
				m := &structpb.Value{Kind: &structpb.Value_NullValue{}}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.Value",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "type.googleapis.com/google.protobuf.Value"},
				{Key: "value", Value: primitive.Null{}},
			},
		}, {
			desc: "Any with Struct",
			input: func() proto.Message {
				m := &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"bool":   {Kind: &structpb.Value_BoolValue{true}},
						"null":   {Kind: &structpb.Value_NullValue{}},
						"string": {Kind: &structpb.Value_StringValue{"hello"}},
						"struct": {
							Kind: &structpb.Value_StructValue{
								&structpb.Struct{
									Fields: map[string]*structpb.Value{
										"string": {Kind: &structpb.Value_StringValue{"world"}},
									},
								},
							},
						},
					},
				}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					TypeUrl: "google.protobuf.Struct",
					Value:   b,
				}
			}(),
			want: bson.D{
				{Key: "@type", Value: "google.protobuf.Struct"},
				{Key: "value", Value: bson.D{
					{Key: "bool", Value: true},
					{Key: "null", Value: primitive.Null{}},
					{Key: "string", Value: "hello"},
					{Key: "struct", Value: bson.D{
						{Key: "string", Value: "world"},
					}},
				}},
			},
		}, {
			desc: "Any with missing type_url",
			input: func() proto.Message {
				m := &wrapperspb.BoolValue{Value: true}
				b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
				if err != nil {
					t.Fatalf("error in binary marshaling message for Any.value: %v", err)
				}
				return &anypb.Any{
					Value: b,
				}
			}(),
			want:    bson.D{},
			wantErr: true,
		}, {
			desc: "well known types as field values",
			input: &pb2.KnownTypes{
				OptBool:      &wrapperspb.BoolValue{Value: false},
				OptInt32:     &wrapperspb.Int32Value{Value: 42},
				OptInt64:     &wrapperspb.Int64Value{Value: 42},
				OptUint32:    &wrapperspb.UInt32Value{Value: 42},
				OptUint64:    &wrapperspb.UInt64Value{Value: 42},
				OptFloat:     &wrapperspb.FloatValue{Value: 1.23},
				OptDouble:    &wrapperspb.DoubleValue{Value: 3.1415},
				OptString:    &wrapperspb.StringValue{Value: "hello"},
				OptBytes:     &wrapperspb.BytesValue{Value: []byte("hello")},
				OptDuration:  &durationpb.Duration{Seconds: 123},
				OptTimestamp: &timestamppb.Timestamp{Seconds: 1553036601},
				OptStruct: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"string": {Kind: &structpb.Value_StringValue{"hello"}},
					},
				},
				OptList: &structpb.ListValue{
					Values: []*structpb.Value{
						{Kind: &structpb.Value_NullValue{}},
						{Kind: &structpb.Value_StringValue{}},
						{Kind: &structpb.Value_StructValue{}},
						{Kind: &structpb.Value_ListValue{}},
					},
				},
				OptValue: &structpb.Value{
					Kind: &structpb.Value_StringValue{"world"},
				},
				OptEmpty: &emptypb.Empty{},
				OptAny: &anypb.Any{
					TypeUrl: "google.protobuf.Empty",
				},
				/*
					OptFieldmask: &fieldmaskpb.FieldMask{
						Paths: []string{"foo_bar", "bar_foo"},
					},
				*/
			},
			want: bson.D{
				{Key: "optBool", Value: false},
				{Key: "optInt32", Value: int32(42)},
				{Key: "optInt64", Value: int64(42)},
				{Key: "optUint32", Value: uint32(42)},
				{Key: "optUint64", Value: uint64(42)},
				{Key: "optFloat", Value: float32(1.23)},
				{Key: "optDouble", Value: float64(3.1415)},
				{Key: "optString", Value: "hello"},
				{Key: "optBytes", Value: primitive.Binary{Data: []byte("hello")}},
				{Key: "optDuration", Value: bson.D{
					{Key: "Seconds", Value: int64(123)},
					{Key: "Nanos", Value: int64(0)},
				}},
				{Key: "optTimestamp", Value: primitive.NewDateTimeFromTime(time.Unix(1553036601, 0))},
				{Key: "optStruct", Value: bson.D{
					{Key: "string", Value: "hello"},
				}},
				{Key: "optList", Value: bson.A{
					primitive.Null{},
					"",
					bson.D{},
					bson.A{},
				}},
				{Key: "optValue", Value: "world"},
				{Key: "optEmpty", Value: primitive.Null{}},
				{Key: "optAny", Value: bson.D{
					{Key: "@type", Value: "google.protobuf.Empty"},
					{Key: "value", Value: primitive.Null{}},
				}},
				// {Key: "optFieldmask", Value: "fooBar,barFoo"},
			},
		}, {
			desc:  "EmitUnpopulated: proto2 optional scalars",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Scalars{},
			want: bson.D{
				{Key: "optBool", Value: primitive.Null{}},
				{Key: "optInt32", Value: primitive.Null{}},
				{Key: "optInt64", Value: primitive.Null{}},
				{Key: "optUint32", Value: primitive.Null{}},
				{Key: "optUint64", Value: primitive.Null{}},
				{Key: "optSint32", Value: primitive.Null{}},
				{Key: "optSint64", Value: primitive.Null{}},
				{Key: "optFixed32", Value: primitive.Null{}},
				{Key: "optFixed64", Value: primitive.Null{}},
				{Key: "optSfixed32", Value: primitive.Null{}},
				{Key: "optSfixed64", Value: primitive.Null{}},
				{Key: "optFloat", Value: primitive.Null{}},
				{Key: "optDouble", Value: primitive.Null{}},
				{Key: "optBytes", Value: primitive.Null{}},
				{Key: "optString", Value: primitive.Null{}},
			},
		}, {
			desc:  "EmitUnpopulated: proto3 scalars",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Scalars{},
			want: bson.D{
				{Key: "sBool", Value: false},
				{Key: "sInt32", Value: int32(0)},
				{Key: "sInt64", Value: int64(0)},
				{Key: "sUint32", Value: uint32(0)},
				{Key: "sUint64", Value: uint64(0)},
				{Key: "sSint32", Value: int32(0)},
				{Key: "sSint64", Value: int64(0)},
				{Key: "sFixed32", Value: uint32(0)},
				{Key: "sFixed64", Value: uint64(0)},
				{Key: "sSfixed32", Value: int32(0)},
				{Key: "sSfixed64", Value: int64(0)},
				{Key: "sFloat", Value: float32(0)},
				{Key: "sDouble", Value: float64(0)},
				{Key: "sBytes", Value: primitive.Binary{}},
				{Key: "sString", Value: ""},
			},
		}, {
			desc:  "EmitUnpopulated: proto2 enum",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Enums{},
			want: bson.D{
				{Key: "optEnum", Value: primitive.Null{}},
				{Key: "rptEnum", Value: bson.A{}},
				{Key: "optNestedEnum", Value: primitive.Null{}},
				{Key: "rptNestedEnum", Value: bson.A{}},
			},
		}, {
			desc:  "EmitUnpopulated: proto3 enum",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Enums{},
			want: bson.D{
				{Key: "sEnum", Value: "ZERO"},
				{Key: "sNestedEnum", Value: "CERO"},
			},
		}, {
			desc:  "EmitUnpopulated: proto2 message and group fields",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Nests{},
			want: bson.D{
				{Key: "optNested", Value: primitive.Null{}},
				{Key: "optgroup", Value: primitive.Null{}},
				{Key: "rptNested", Value: bson.A{}},
				{Key: "rptgroup", Value: bson.A{}},
			},
		}, {
			desc:  "EmitUnpopulated: proto3 message field",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Nests{},
			want: bson.D{
				{Key: "sNested", Value: primitive.Null{}},
			},
		}, {
			desc: "EmitUnpopulated: proto2 empty message and group fields",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Nests{
				OptNested: &pb2.Nested{},
				Optgroup:  &pb2.Nests_OptGroup{},
			},
			want: bson.D{
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: primitive.Null{}},
					{Key: "optNested", Value: primitive.Null{}},
				}},
				{Key: "optgroup", Value: bson.D{
					{Key: "optString", Value: primitive.Null{}},
					{Key: "optNested", Value: primitive.Null{}},
					{Key: "optnestedgroup", Value: primitive.Null{}},
				}},
				{Key: "rptNested", Value: bson.A{}},
				{Key: "rptgroup", Value: bson.A{}},
			},
		}, {
			desc: "EmitUnpopulated: proto3 empty message field",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Nests{
				SNested: &pb3.Nested{},
			},
			want: bson.D{
				{Key: "sNested", Value: bson.D{
					{Key: "sString", Value: ""},
					{Key: "sNested", Value: primitive.Null{}},
				}},
			},
		}, {
			desc: "EmitUnpopulated: proto2 required fields",
			mo: MarshalOptions{
				AllowPartial:    true,
				EmitUnpopulated: true,
			},
			input: &pb2.Requireds{},
			want: bson.D{
				{Key: "reqBool", Value: primitive.Null{}},
				{Key: "reqSfixed64", Value: primitive.Null{}},
				{Key: "reqDouble", Value: primitive.Null{}},
				{Key: "reqString", Value: primitive.Null{}},
				{Key: "reqEnum", Value: primitive.Null{}},
				{Key: "reqNested", Value: primitive.Null{}},
			},
		}, {
			desc:  "EmitUnpopulated: repeated fields",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Repeats{},
			want: bson.D{
				{Key: "rptBool", Value: bson.A{}},
				{Key: "rptInt32", Value: bson.A{}},
				{Key: "rptInt64", Value: bson.A{}},
				{Key: "rptUint32", Value: bson.A{}},
				{Key: "rptUint64", Value: bson.A{}},
				{Key: "rptFloat", Value: bson.A{}},
				{Key: "rptDouble", Value: bson.A{}},
				{Key: "rptString", Value: bson.A{}},
				{Key: "rptBytes", Value: bson.A{}},
			},
		}, {
			desc: "EmitUnpopulated: repeated containing empty message",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Nests{
				RptNested: []*pb2.Nested{nil, {}},
			},
			want: bson.D{
				{Key: "optNested", Value: primitive.Null{}},
				{Key: "optgroup", Value: primitive.Null{}},
				{Key: "rptNested", Value: bson.A{
					bson.D{
						{Key: "optString", Value: primitive.Null{}},
						{Key: "optNested", Value: primitive.Null{}},
					},
					bson.D{
						{Key: "optString", Value: primitive.Null{}},
						{Key: "optNested", Value: primitive.Null{}},
					},
				}},
				{Key: "rptgroup", Value: bson.A{}},
			},
		}, {
			desc:  "EmitUnpopulated: map fields",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Maps{},
			want: bson.D{
				{Key: "int32ToStr", Value: bson.D{}},
				{Key: "boolToUint32", Value: bson.D{}},
				{Key: "uint64ToEnum", Value: bson.D{}},
				{Key: "strToNested", Value: bson.D{}},
				{Key: "strToOneofs", Value: bson.D{}},
			},
		}, {
			desc: "EmitUnpopulated: map containing empty message",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Maps{
				StrToNested: map[string]*pb3.Nested{
					"nested": &pb3.Nested{},
				},
				StrToOneofs: map[string]*pb3.Oneofs{
					"nested": &pb3.Oneofs{},
				},
			},
			want: bson.D{
				{Key: "int32ToStr", Value: bson.D{}},
				{Key: "boolToUint32", Value: bson.D{}},
				{Key: "uint64ToEnum", Value: bson.D{}},
				{Key: "strToNested", Value: bson.D{
					{Key: "nested", Value: bson.D{
						{Key: "sString", Value: ""},
						{Key: "sNested", Value: primitive.Null{}},
					}},
				}},
				{Key: "strToOneofs", Value: bson.D{
					{Key: "nested", Value: bson.D{}},
				}},
			},
		}, {
			desc:  "EmitUnpopulated: oneof fields",
			mo:    MarshalOptions{EmitUnpopulated: true},
			input: &pb3.Oneofs{},
			want:  bson.D{},
		}, {
			desc: "EmitUnpopulated: extensions",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: func() proto.Message {
				m := &pb2.Extensions{}
				proto.SetExtension(m, pb2.E_OptExtNested, &pb2.Nested{})
				proto.SetExtension(m, pb2.E_RptExtNested, []*pb2.Nested{
					nil,
					{},
				})
				return m
			}(),
			want: bson.D{
				{Key: "optString", Value: primitive.Null{}},
				{Key: "optBool", Value: primitive.Null{}},
				{Key: "optInt32", Value: primitive.Null{}},
				{Key: "[textpb2_proto.opt_ext_nested]", Value: bson.D{
					{Key: "optString", Value: primitive.Null{}},
					{Key: "optNested", Value: primitive.Null{}},
				}},
				{Key: "[textpb2_proto.rpt_ext_nested]", Value: bson.A{
					bson.D{
						{Key: "optString", Value: primitive.Null{}},
						{Key: "optNested", Value: primitive.Null{}},
					},
					bson.D{
						{Key: "optString", Value: primitive.Null{}},
						{Key: "optNested", Value: primitive.Null{}},
					},
				}},
			},
		}, {
			desc: "EmitUnpopulated: with populated fields",
			mo:   MarshalOptions{EmitUnpopulated: true},
			input: &pb2.Scalars{
				OptInt32:    proto.Int32(0xff),
				OptUint32:   proto.Uint32(47),
				OptSint32:   proto.Int32(-1001),
				OptFixed32:  proto.Uint32(32),
				OptSfixed32: proto.Int32(-32),
				OptFloat:    proto.Float32(1.02),
				OptBytes:    []byte("谷歌"),
			},
			want: bson.D{
				{Key: "optBool", Value: primitive.Null{}},
				{Key: "optInt32", Value: int32(255)},
				{Key: "optInt64", Value: primitive.Null{}},
				{Key: "optUint32", Value: uint32(47)},
				{Key: "optUint64", Value: primitive.Null{}},
				{Key: "optSint32", Value: int32(-1001)},
				{Key: "optSint64", Value: primitive.Null{}},
				{Key: "optFixed32", Value: uint32(32)},
				{Key: "optFixed64", Value: primitive.Null{}},
				{Key: "optSfixed32", Value: int32(-32)},
				{Key: "optSfixed64", Value: primitive.Null{}},
				{Key: "optFloat", Value: float32(1.02)},
				{Key: "optDouble", Value: primitive.Null{}},
				{Key: "optBytes", Value: primitive.Binary{Data: []byte("谷歌")}},
				{Key: "optString", Value: primitive.Null{}},
			},
		}, {
			desc: "UseEnumNumbers in singular field",
			mo:   MarshalOptions{UseEnumNumbers: true},
			input: &pb2.Enums{
				OptEnum:       pb2.Enum_ONE.Enum(),
				OptNestedEnum: pb2.Enums_UNO.Enum(),
			},
			want: bson.D{
				{Key: "optEnum", Value: int64(1)},
				{Key: "optNestedEnum", Value: int64(1)},
			},
		}, {
			desc: "UseEnumNumbers in repeated field",
			mo:   MarshalOptions{UseEnumNumbers: true},
			input: &pb2.Enums{
				RptEnum:       []pb2.Enum{pb2.Enum_ONE, 2, pb2.Enum_TEN, 42},
				RptNestedEnum: []pb2.Enums_NestedEnum{pb2.Enums_UNO, pb2.Enums_DOS, 47},
			},
			want: bson.D{
				{Key: "rptEnum", Value: bson.A{
					int64(1),
					int64(2),
					int64(10),
					int64(42),
				}},
				{Key: "rptNestedEnum", Value: bson.A{
					int64(1),
					int64(2),
					int64(47),
				}},
			},
		}, {
			desc: "UseEnumNumbers in map field",
			mo:   MarshalOptions{UseEnumNumbers: true},
			input: &pb3.Maps{
				Uint64ToEnum: map[uint64]pb3.Enum{
					1:  pb3.Enum_ONE,
					2:  pb3.Enum_TWO,
					10: pb3.Enum_TEN,
					47: 47,
				},
			},
			want: bson.D{
				{Key: "uint64ToEnum", Value: bson.D{
					{Key: "1", Value: int64(1)},
					{Key: "2", Value: int64(2)},
					{Key: "10", Value: int64(10)},
					{Key: "47", Value: int64(47)},
				}},
			},
		}, {
			desc: "UseProtoNames",
			mo:   MarshalOptions{UseProtoNames: true},
			input: &pb2.Nests{
				OptNested: &pb2.Nested{},
				Optgroup: &pb2.Nests_OptGroup{
					OptString: proto.String("inside a group"),
					OptNested: &pb2.Nested{
						OptString: proto.String("nested message inside a group"),
					},
					Optnestedgroup: &pb2.Nests_OptGroup_OptNestedGroup{
						OptFixed32: proto.Uint32(47),
					},
				},
				Rptgroup: []*pb2.Nests_RptGroup{
					{
						RptString: []string{"hello", "world"},
					},
				},
			},
			want: bson.D{
				{Key: "opt_nested", Value: bson.D{}},
				{Key: "OptGroup", Value: bson.D{
					{Key: "opt_string", Value: "inside a group"},
					{Key: "opt_nested", Value: bson.D{
						{Key: "opt_string", Value: "nested message inside a group"},
					}},
					{Key: "OptNestedGroup", Value: bson.D{
						{Key: "opt_fixed32", Value: uint32(47)},
					}},
				}},
				{Key: "RptGroup", Value: bson.A{
					bson.D{{Key: "rpt_string", Value: bson.A{
						"hello",
						"world",
					}}},
				}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		if tt.skip {
			continue
		}
		t.Run(tt.desc, func(t *testing.T) {
			result, err := tt.mo.Marshal(tt.input)
			if err != nil && !tt.wantErr {
				t.Errorf("Marshal() returned error: %v\n", err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("Marshal() got nil error, want error\n")
			}
			if equal, err := deepequal.DeepEqual(result, tt.want); !equal {
				if diff := cmp.Diff(tt.want, result); diff != "" {
					t.Errorf("Marshal() diff -want +got\n%v\n", diff)
				} else {
					t.Errorf("Marshal() wrong: %v\n", err)
				}
			}
		})
	}
}
