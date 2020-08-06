package bsonpb

import (
	"math"
	"strings"
	"time"
	"testing"

	"google.golang.org/protobuf/proto"
	preg "google.golang.org/protobuf/reflect/protoregistry"

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

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		desc         string
		umo          UnmarshalOptions
		inputMessage proto.Message
		inputBson    interface{}
		wantMessage  proto.Message
		wantErr      string
		skip         bool
	}{{
		desc:         "proto2 empty message",
		inputMessage: &pb2.Scalars{},
		inputBson:    bson.D{},
		wantMessage:  &pb2.Scalars{},
	}, {
		desc:         "proto2 optional scalars set to zero values",
		inputMessage: &pb2.Scalars{},
		inputBson: bson.D{
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
		wantMessage: &pb2.Scalars{
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
	}, {
		desc:         "proto3 scalars set to zero values",
		inputMessage: &pb3.Scalars{},
		inputBson: bson.D{
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
			{Key: "sBytes", Value: primitive.Binary{Data: []byte{}}},
			{Key: "sString", Value: ""},
		},
		wantMessage: &pb3.Scalars{},
	}, /* {
		desc:         "proto3 optional set to zero values",
		inputMessage: &pb3.Proto3Optional{},
		inputBson: bson.D{
			{Key: "optBool", Value: false},
			{Key: "optInt32", Value: int32(0)},
			{Key: "optInt64", Value: int64(0)},
			{Key: "optUint32", Value: uint32(0)},
			{Key: "optUint64", Value: uint64(0)},
			{Key: "optFloat", Value: float32(0)},
			{Key: "optDouble", Value: float64(0)},
			{Key: "optString", Value: ""},
			{Key: "optBytes", Value: primitive.Binary{Data: []byte{}}},
			{Key: "optEnum", Value: "ZERO"},
			{Key: "optMessage", Value: bson.D{}},
		},
		wantMessage: &pb3.Proto3Optional{
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
	},*/ {
		desc:         "proto2 optional scalars set to null",
		inputMessage: &pb2.Scalars{},
		inputBson: bson.D{
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
		wantMessage: &pb2.Scalars{},
	}, {
		desc:         "proto3 scalars set to null",
		inputMessage: &pb3.Scalars{},
		inputBson: bson.D{
			{Key: "sBool", Value: primitive.Null{}},
			{Key: "sInt32", Value: primitive.Null{}},
			{Key: "sInt64", Value: primitive.Null{}},
			{Key: "sUint32", Value: primitive.Null{}},
			{Key: "sUint64", Value: primitive.Null{}},
			{Key: "sSint32", Value: primitive.Null{}},
			{Key: "sSint64", Value: primitive.Null{}},
			{Key: "sFixed32", Value: primitive.Null{}},
			{Key: "sFixed64", Value: primitive.Null{}},
			{Key: "sSfixed32", Value: primitive.Null{}},
			{Key: "sSfixed64", Value: primitive.Null{}},
			{Key: "sFloat", Value: primitive.Null{}},
			{Key: "sDouble", Value: primitive.Null{}},
			{Key: "sBytes", Value: primitive.Null{}},
			{Key: "sString", Value: primitive.Null{}},
		},
		wantMessage: &pb3.Scalars{},
	}, {
		desc:         "boolean",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sBool", Value: true},
		},
		wantMessage: &pb3.Scalars{
			SBool: true,
		},
	}, {
		desc:         "not boolean",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sBool", Value: "true"},
		},
		wantErr:      `invalid value for bool type: "true" (has type string)`,
	}, {
		desc:         "float and double",
		inputMessage: &pb3.Scalars{},
		inputBson: bson.D{
			{Key: "sFloat", Value: 1.234},
			{Key: "sDouble", Value: 5.678},
		},
		wantMessage: &pb3.Scalars{
			SFloat:  1.234,
			SDouble: 5.678,
		},
	}, {
		desc:         "float exceeds positive limit",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sFloat", Value: 3.4e+39},
		},
		wantErr:      `invalid value for float type: 3.4e+39 (has type float64)`,
	}, {
		desc:         "float exceeds negative limit",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sFloat", Value: -3.4e+39},
		},
		wantErr:      `invalid value for float type: -3.4e+39 (has type float64)`,
	}, /* {
		desc:         "double exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sDouble", Value: -1.79e+309},
		},
		wantErr:      `invalid value for double type: -1.79e+309`,
	}, {
		desc:         "double in string exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sDouble", Value: 1.79e+309},
		},
		wantErr:      `invalid value for double type: "1.79e+309"`,
	}, */ {
		desc:         "infinites",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sFloat", Value: float32(math.Inf(+1))},
			{Key: "sDouble", Value: math.Inf(-1)},
		},
		wantMessage: &pb3.Scalars{
			SFloat:  float32(math.Inf(+1)),
			SDouble: math.Inf(-1),
		},
	}, {
		desc:         "not float",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sFloat", Value: true},
		},
		wantErr:      `invalid value for float type: true (has type bool)`,
	}, {
		desc:         "not double",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sDouble", Value: "not a number"},
		},
		wantErr:      `invalid value for double type: "not a number" (has type string)`,
	}, {
		desc:         "integers",
		inputMessage: &pb3.Scalars{},
		inputBson: bson.D{
			{Key: "sInt32", Value: 1234},
			{Key: "sInt64", Value: -1234},
			{Key: "sUint32", Value: int(1e2)},
			{Key: "sUint64", Value: int(100E-2)},
			{Key: "sSint32", Value: 1},
			{Key: "sSint64", Value: -1},
			{Key: "sFixed32", Value: int(1.234e+5)},
			{Key: "sFixed64", Value: int(1200E-2)},
			{Key: "sSfixed32", Value: int(-1.234e+05)},
			{Key: "sSfixed64", Value: int(-1200e-02)},
		},
		wantMessage: &pb3.Scalars{
			SInt32:    1234,
			SInt64:    -1234,
			SUint32:   100,
			SUint64:   1,
			SSint32:   1,
			SSint64:   -1,
			SFixed32:  123400,
			SFixed64:  12,
			SSfixed32: -123400,
			SSfixed64: -12,
		},
	}, {
		desc:         "number is not an integer",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sInt32", Value: 1.001},
		},
		wantErr:      `invalid value for int32 type: 1.001 (has type float64)`,
	}, {
		desc:         "32-bit int exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sInt32", Value: int(2e10)},
		},
		wantErr:      `invalid value for int32 type: 20000000000 (has type int)`,
	}, {
		desc:         "not integer",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sInt32", Value: "not a number"},
		},
		wantErr:      `invalid value for int32 type: "not a number"`,
	}, {
		desc:         "not unsigned integer",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sUint32", Value: "not a number"},
		},
		wantErr:      `invalid value for uint32 type: "not a number"`,
	}, {
		desc:         "number is not an unsigned integer",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sUint32", Value: -1},
		},
		wantErr:      `invalid value for uint32 type: -1`,
	}, {
		desc:         "string",
		inputMessage: &pb2.Scalars{},
		inputBson:    bson.D{
			{Key: "optString", Value: "谷歌"},
		},
		wantMessage: &pb2.Scalars{
			OptString: proto.String("谷歌"),
		},
	}, {
		desc:         "string with invalid UTF-8",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sString", Value: "\xff\""},
		},
		wantErr:      "invalid UTF-8: \xff\"",
	}, {
		desc:         "not string",
		inputMessage: &pb2.Scalars{},
		inputBson:    bson.D{
			{Key: "optString", Value: 42},
		},
		wantErr:      `invalid value for string type: 42`,
	}, {
		desc:         "bytes",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sBytes", Value: primitive.Binary{Data: []byte("hello world")}},
		},
		wantMessage: &pb3.Scalars{
			SBytes: []byte("hello world"),
		},
	}, {
		desc:         "not bytes",
		inputMessage: &pb3.Scalars{},
		inputBson:    bson.D{
			{Key: "sBytes", Value: true},
		},
		wantErr:      `invalid value for bytes type: true (has type bool)`,
	}, {
		desc:         "proto2 enum",
		inputMessage: &pb2.Enums{},
		inputBson: bson.D{
			{Key: "optEnum", Value: "ONE"},
			{Key: "optNestedEnum", Value: "UNO"},
		},
		wantMessage: &pb2.Enums{
			OptEnum:       pb2.Enum_ONE.Enum(),
			OptNestedEnum: pb2.Enums_UNO.Enum(),
		},
	}, {
		desc:         "proto3 enum",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: "ONE"},
			{Key: "sNestedEnum", Value: "DIEZ"},
		},
		wantMessage: &pb3.Enums{
			SEnum:       pb3.Enum_ONE,
			SNestedEnum: pb3.Enums_DIEZ,
		},
	}, {
		desc:         "enum numeric value",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: 2},
			{Key: "sNestedEnum", Value: 2},
		},
		wantMessage: &pb3.Enums{
			SEnum:       pb3.Enum_TWO,
			SNestedEnum: pb3.Enums_DOS,
		},
	}, {
		desc:         "enum unnamed numeric value",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: 101},
			{Key: "sNestedEnum", Value: -101},
		},
		wantMessage: &pb3.Enums{
			SEnum:       101,
			SNestedEnum: -101,
		},
	}, {
		desc:         "enum set to number string",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: "1"},
		},
		wantErr: `invalid value for enum type: "1"`,
	}, {
		desc:         "enum set to invalid named",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: "UNNAMED"},
		},
		wantErr: `invalid value for enum type: "UNNAMED"`,
	}, {
		desc:         "enum set to not enum",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: true},
		},
		wantErr: `invalid value for enum type: true`,
	}, {
		desc:         "enum set to JSON null",
		inputMessage: &pb3.Enums{},
		inputBson: bson.D{
			{Key: "sEnum", Value: primitive.Null{}},
		},
		wantMessage: &pb3.Enums{},
	}, {
		desc:         "proto name",
		inputMessage: &pb3.JSONNames{},
		inputBson: bson.D{
			{Key: "s_string", Value: "proto name used"},
		},
		wantMessage: &pb3.JSONNames{
			SString: "proto name used",
		},
	}, {
		desc:         "proto group name",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "OptGroup", Value: bson.D{
				{Key: "optString", Value: "hello"},
			}},
			{Key: "RptGroup", Value: bson.A{
				bson.D{{Key: "rptString", Value: bson.A{"goodbye"}}},
			}},
		},
		wantMessage: &pb2.Nests{
			Optgroup: &pb2.Nests_OptGroup{OptString: proto.String("hello")},
			Rptgroup: []*pb2.Nests_RptGroup{{RptString: []string{"goodbye"}}},
		},
	}, {
		desc:         "json_name",
		inputMessage: &pb3.JSONNames{},
		inputBson: bson.D{
			{Key: "foo_bar", Value: "json_name used"},
		},
		wantMessage: &pb3.JSONNames{
			SString: "json_name used",
		},
	}, {
		desc:         "camelCase name",
		inputMessage: &pb3.JSONNames{},
		inputBson: bson.D{
			{Key: "sString", Value: "camelcase used"},
		},
		wantErr: `unknown field "sString"`,
	}, {
		desc:         "proto name and json_name",
		inputMessage: &pb3.JSONNames{},
		inputBson: bson.D{
			{Key: "foo_bar", Value: "json_name used"},
			{Key: "s_string", Value: "proto name used"},
		},
		wantErr: `duplicate field "s_string"`,
	}, {
		desc:         "duplicate field names",
		inputMessage: &pb3.JSONNames{},
		inputBson: bson.D{
			{Key: "foo_bar", Value: "one"},
			{Key: "foo_bar", Value: "two"},
		},
		wantErr: `duplicate field "foo_bar"`,
	}, /* {
		desc:         "null message",
		inputMessage: &pb2.Nests{},
		inputBson:    primitive.Null{},
		wantErr:      `unexpected token null`,
	}, */ {
		desc:         "proto2 nested message not set",
		inputMessage: &pb2.Nests{},
		inputBson:    bson.D{},
		wantMessage:  &pb2.Nests{},
	}, {
		desc:         "proto2 nested message set to null",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "optNested", Value: primitive.Null{}},
			{Key: "optgroup", Value: primitive.Null{}},
		},
		wantMessage: &pb2.Nests{},
	}, {
		desc:         "proto2 nested message set to empty",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "optNested", Value: bson.D{}},
			{Key: "optgroup", Value: bson.D{}},
		},
		wantMessage: &pb2.Nests{
			OptNested: &pb2.Nested{},
			Optgroup:  &pb2.Nests_OptGroup{},
		},
	}, {
		desc:         "proto2 nested messages",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "optNested", Value: bson.D{
				{Key: "optString", Value: "nested message"},
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: "another nested message"},
				}},
			}},
		},
		wantMessage: &pb2.Nests{
			OptNested: &pb2.Nested{
				OptString: proto.String("nested message"),
				OptNested: &pb2.Nested{
					OptString: proto.String("another nested message"),
				},
			},
		},
	}, {
		desc:         "proto2 groups",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "optgroup", Value: bson.D{
				{Key: "optString", Value: "inside a group"},
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: "nested message inside a group"},
				}},
				{Key: "optnestedgroup", Value: bson.D{
					{Key: "optFixed32", Value: 47},
				}},
			}},
		},
		wantMessage: &pb2.Nests{
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
	}, {
		desc:         "proto3 nested message not set",
		inputMessage: &pb3.Nests{},
		inputBson:    bson.D{},
		wantMessage:  &pb3.Nests{},
	}, {
		desc:         "proto3 nested message set to null",
		inputMessage: &pb3.Nests{},
		inputBson:    bson.D{
			{Key: "sNested", Value: primitive.Null{}},
		},
		wantMessage:  &pb3.Nests{},
	}, {
		desc:         "proto3 nested message set to empty",
		inputMessage: &pb3.Nests{},
		inputBson:    bson.D{
			{Key: "sNested", Value: bson.D{}},
		},
		wantMessage: &pb3.Nests{
			SNested: &pb3.Nested{},
		},
	}, {
		desc:         "proto3 nested message",
		inputMessage: &pb3.Nests{},
		inputBson: bson.D{
			{Key: "sNested", Value: bson.D{
				{Key: "sString", Value: "nested message"},
				{Key: "sNested", Value: bson.D{
					{Key: "sString", Value: "another nested message"},
				}},
			}},
		},
		wantMessage: &pb3.Nests{
			SNested: &pb3.Nested{
				SString: "nested message",
				SNested: &pb3.Nested{
					SString: "another nested message",
				},
			},
		},
	}, {
		desc:         "nested message set to non-message",
		inputMessage: &pb3.Nests{},
		inputBson:    bson.D{
			{Key: "sNested", Value: true},
		},
		wantErr:      `unexpected message value: true`,
	}, {
		desc:         "oneof not set",
		inputMessage: &pb3.Oneofs{},
		inputBson:    bson.D{},
		wantMessage:  &pb3.Oneofs{},
	}, {
		desc:         "oneof set to empty string",
		inputMessage: &pb3.Oneofs{},
		inputBson:    bson.D{
			{Key: "oneofString", Value: ""},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{},
		},
	}, {
		desc:         "oneof set to string",
		inputMessage: &pb3.Oneofs{},
		inputBson:    bson.D{
			{Key: "oneofString", Value: "hello"},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{
				OneofString: "hello",
			},
		},
	}, {
		desc:         "oneof set to enum",
		inputMessage: &pb3.Oneofs{},
		inputBson:    bson.D{
			{Key: "oneofEnum", Value: "ZERO"},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofEnum{
				OneofEnum: pb3.Enum_ZERO,
			},
		},
	}, {
		desc:         "oneof set to empty message",
		inputMessage: &pb3.Oneofs{},
		inputBson:    bson.D{
			{Key: "oneofNested", Value: bson.D{}},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{},
			},
		},
	}, {
		desc:         "oneof set to message",
		inputMessage: &pb3.Oneofs{},
		inputBson: bson.D{
			{Key: "oneofNested", Value: bson.D{
				{Key: "sString", Value: "nested message"},
			}},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{
					SString: "nested message",
				},
			},
		},
	}, {
		desc:         "oneof set to more than one field",
		inputMessage: &pb3.Oneofs{},
		inputBson: bson.D{
			{Key: "oneofEnum", Value: "ZERO"},
			{Key: "oneofString", Value: "hello"},
		},
		wantErr: `error parsing "oneofString", oneof textpb3_proto.Oneofs.union is already set`,
	}, {
		desc:         "oneof set to null and value",
		inputMessage: &pb3.Oneofs{},
		inputBson: bson.D{
			{Key: "oneofEnum", Value: "ZERO"},
			{Key: "oneofString", Value: primitive.Null{}},
		},
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofEnum{
				OneofEnum: pb3.Enum_ZERO,
			},
		},
	}, {
		desc:         "repeated null fields",
		inputMessage: &pb2.Repeats{},
		inputBson: bson.D{
			{Key: "rptString", Value: primitive.Null{}},
			{Key: "rptInt32", Value: primitive.Null{}},
			{Key: "rptFloat", Value: primitive.Null{}},
			{Key: "rptBytes", Value: primitive.Null{}},
		},
		wantMessage: &pb2.Repeats{},
	}, {
		desc:         "repeated scalars",
		inputMessage: &pb2.Repeats{},
		inputBson: bson.D{
			{Key: "rptString", Value: bson.A{"hello", "world"}},
			{Key: "rptInt32", Value: bson.A{-1, 0, 1}},
			{Key: "rptBool", Value: bson.A{false, true}},
		},
		wantMessage: &pb2.Repeats{
			RptString: []string{"hello", "world"},
			RptInt32:  []int32{-1, 0, 1},
			RptBool:   []bool{false, true},
		},
	}, {
		desc:         "repeated enums",
		inputMessage: &pb2.Enums{},
		inputBson: bson.D{
			{Key: "rptEnum", Value: bson.A{"TEN", 1, 42}},
			{Key: "rptNestedEnum", Value: bson.A{"DOS", 2, -47}},
		},
		wantMessage: &pb2.Enums{
			RptEnum:       []pb2.Enum{pb2.Enum_TEN, pb2.Enum_ONE, 42},
			RptNestedEnum: []pb2.Enums_NestedEnum{pb2.Enums_DOS, pb2.Enums_DOS, -47},
		},
	}, {
		desc:         "repeated messages",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{
					{Key: "optString", Value: "repeat nested one"},
				},
				bson.D{
					{Key: "optString", Value: "repeat nested two"},
					{Key: "optNested", Value: bson.D{
						{Key: "optString", Value: "inside repeat nested two"},
					}},
				},
				bson.D{},
			}},
		},
		wantMessage: &pb2.Nests{
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
	}, {
		desc:         "repeated groups",
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "rptgroup", Value: bson.A{
				bson.D{
					{Key: "rptString", Value: bson.A{"hello", "world"}},
				},
				bson.D{},
			}},
		},
		wantMessage: &pb2.Nests{
			Rptgroup: []*pb2.Nests_RptGroup{
				{
					RptString: []string{"hello", "world"},
				},
				{},
			},
		},
	}, {
		desc:         "repeated string contains invalid UTF-8",
		inputMessage: &pb2.Repeats{},
		inputBson:    bson.D{
			{Key: "rptString", Value: bson.A{"abc\xff"}},
		},
		wantErr:      "invalid UTF-8: abc\xff",
	}, {
		desc:         "repeated messages contain invalid UTF-8",
		inputMessage: &pb2.Nests{},
		inputBson:    bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{
					{Key: "optString", Value: "abc\xff"},
				},
			}},
		},
		wantErr:      "invalid UTF-8: abc\xff",
	}, {
		desc:         "repeated scalars contain invalid type",
		inputMessage: &pb2.Repeats{},
		inputBson:    bson.D{
			{Key: "rptString", Value: bson.A{"hello", primitive.Null{}, "world"}},
		},
		wantErr:      `invalid value for string type: {} (has type primitive.Null)`,
	}, {
		desc:         "repeated messages contain invalid type",
		inputMessage: &pb2.Nests{},
		inputBson:    bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{},
				primitive.Null{},
			}},
		},
		wantErr:      `unexpected message value: {}`,
	}, {
		desc:         "map fields 1",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "-101", Value: "-101"},
				{Key: "0", Value: "zero"},
				{Key: "255", Value: "0xff"},
			}},
			{Key: "boolToUint32", Value: bson.D{
				{Key: "false", Value: 101},
				{Key: "true", Value: 42},
			}},
		},
		wantMessage: &pb3.Maps{
			Int32ToStr: map[int32]string{
				-101: "-101",
				0xff: "0xff",
				0:    "zero",
			},
			BoolToUint32: map[bool]uint32{
				true:  42,
				false: 101,
			},
		},
	}, {
		desc:         "map fields 2",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "uint64ToEnum", Value: bson.D{
				{Key: "1", Value: "ONE"},
				{Key: "2", Value: 2},
				{Key: "10", Value: 101},
			}},
		},
		wantMessage: &pb3.Maps{
			Uint64ToEnum: map[uint64]pb3.Enum{
				1:  pb3.Enum_ONE,
				2:  pb3.Enum_TWO,
				10: 101,
			},
		},
	}, {
		desc:         "map fields 3",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "nested_one", Value: bson.D{
					{Key: "sString", Value: "nested in a map"},
				}},
				{Key: "nested_two", Value: bson.D{}},
			}},
		},
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested_one": {
					SString: "nested in a map",
				},
				"nested_two": {},
			},
		},
	}, {
		desc:         "map fields 4",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
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
		wantMessage: &pb3.Maps{
			StrToOneofs: map[string]*pb3.Oneofs{
				"string": {
					Union: &pb3.Oneofs_OneofString{
						OneofString: "hello",
					},
				},
				"nested": {
					Union: &pb3.Oneofs_OneofNested{
						OneofNested: &pb3.Nested{
							SString: "nested oneof in map field value",
						},
					},
				},
			},
		},
	}, {
		desc:         "map contains duplicate keys",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "0", Value: "cero"},
				{Key: "0", Value: "zero"},
			}},
		},
		wantErr: `duplicate map key 0`,
	}, {
		desc:         "map key empty string",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "", Value: bson.D{}},
			}},
		},
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"": {},
			},
		},
	}, {
		desc:         "map contains invalid key 1",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "invalid", Value: "cero"},
			}},
		},
		wantErr: `invalid value for int32 key: "invalid"`,
	}, {
		desc:         "map contains invalid key 2",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "1.02", Value: "float"},
			}},
		},
		wantErr: `invalid value for int32 key: "1.02"`,
	}, {
		desc:         "map contains invalid key 3",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "2147483648", Value: "exceeds 32-bit integer max limit"},
			}},
		},
		wantErr: `invalid value for int32 key: "2147483648"`,
	}, {
		desc:         "map contains invalid key 4",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "uint64ToEnum", Value: bson.D{
				{Key: "-1", Value: 0},
			}},
		},
		wantErr: `invalid value for uint64 key: "-1"`,
	}, {
		desc:         "map contains invalid value",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "101", Value: true},
			}},
		},
		wantErr: `invalid value for string type: true`,
	}, {
		desc:         "map contains null for scalar value",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "int32ToStr", Value: bson.D{
				{Key: "101", Value: primitive.Null{}},
			}},
		},
		wantErr: `invalid value for string type: {} (has type primitive.Null)`,
	}, {
		desc:         "map contains null for message value",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "hello", Value: primitive.Null{}},
			}},
		},
		wantErr: `unexpected message value: {}`,
	}, {
		desc:         "map contains contains message value with invalid UTF-8",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "hello", Value: bson.D{
					{Key: "sString", Value: "abc\xff"},
				}},
			}},
		},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "map key contains invalid UTF-8",
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "hello", Value: bson.D{
					{Key: "abc\xff", Value: bson.D{}},
				}},
			}},
		},
		wantErr: "unknown field \"abc\\xff\"",
	}, {
		desc:         "required fields not set",
		inputMessage: &pb2.Requireds{},
		inputBson:    bson.D{},
		wantErr:      "required field textpb2_proto.Requireds.req_bool not set",
	}, {
		desc:         "required field set",
		inputMessage: &pb2.PartialRequired{},
		inputBson: bson.D{
			{Key: "reqString", Value: "this is required"},
		},
		wantMessage: &pb2.PartialRequired{
			ReqString: proto.String("this is required"),
		},
	}, {
		desc:         "required fields partially set",
		inputMessage: &pb2.Requireds{},
		inputBson: bson.D{
			{Key: "reqBool", Value: false},
			{Key: "reqSfixed64", Value: 42},
			{Key: "reqString", Value: "hello"},
			{Key: "reqEnum", Value: "ONE"},
		},
		wantMessage: &pb2.Requireds{
			ReqBool:     proto.Bool(false),
			ReqSfixed64: proto.Int64(42),
			ReqString:   proto.String("hello"),
			ReqEnum:     pb2.Enum_ONE.Enum(),
		},
		wantErr: "required field textpb2_proto.Requireds.req_double not set",
	}, {
		desc:         "required fields partially set with AllowPartial",
		umo:          UnmarshalOptions{AllowPartial: true},
		inputMessage: &pb2.Requireds{},
		inputBson: bson.D{
			{Key: "reqBool", Value: false},
			{Key: "reqSfixed64", Value: 42},
			{Key: "reqString", Value: "hello"},
			{Key: "reqEnum", Value: "ONE"},
		},
		wantMessage: &pb2.Requireds{
			ReqBool:     proto.Bool(false),
			ReqSfixed64: proto.Int64(42),
			ReqString:   proto.String("hello"),
			ReqEnum:     pb2.Enum_ONE.Enum(),
		},
	}, {
		desc:         "required fields all set",
		inputMessage: &pb2.Requireds{},
		inputBson: bson.D{
			{Key: "reqBool", Value: false},
			{Key: "reqSfixed64", Value: 42},
			{Key: "reqDouble", Value: 1.23},
			{Key: "reqString", Value: "hello"},
			{Key: "reqEnum", Value: "ONE"},
			{Key: "reqNested", Value: bson.D{}},
		},
		wantMessage: &pb2.Requireds{
			ReqBool:     proto.Bool(false),
			ReqSfixed64: proto.Int64(42),
			ReqDouble:   proto.Float64(1.23),
			ReqString:   proto.String("hello"),
			ReqEnum:     pb2.Enum_ONE.Enum(),
			ReqNested:   &pb2.Nested{},
		},
	}, {
		desc:         "indirect required field",
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "optNested", Value: bson.D{}},
		},
		wantMessage: &pb2.IndirectRequired{
			OptNested: &pb2.NestedWithRequired{},
		},
		wantErr: "required field textpb2_proto.NestedWithRequired.req_string not set",
	}, {
		desc:         "indirect required field with AllowPartial",
		umo:          UnmarshalOptions{AllowPartial: true},
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "optNested", Value: bson.D{}},
		},
		wantMessage: &pb2.IndirectRequired{
			OptNested: &pb2.NestedWithRequired{},
		},
	}, {
		desc:         "indirect required field in repeated",
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{{Key: "reqString", Value: "one"}},
				bson.D{},
			}},
		},
		wantMessage: &pb2.IndirectRequired{
			RptNested: []*pb2.NestedWithRequired{
				{
					ReqString: proto.String("one"),
				},
				{},
			},
		},
		wantErr: "required field textpb2_proto.NestedWithRequired.req_string not set",
	}, {
		desc:         "indirect required field in repeated with AllowPartial",
		umo:          UnmarshalOptions{AllowPartial: true},
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{{Key: "reqString", Value: "one"}},
				bson.D{},
			}},
		},
		wantMessage: &pb2.IndirectRequired{
			RptNested: []*pb2.NestedWithRequired{
				{
					ReqString: proto.String("one"),
				},
				{},
			},
		},
	}, {
		desc:         "indirect required field in map",
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "missing", Value: bson.D{}},
				{Key: "contains", Value: bson.D{
					{Key: "reqString", Value: "here"},
				}},
			}},
		},
		wantMessage: &pb2.IndirectRequired{
			StrToNested: map[string]*pb2.NestedWithRequired{
				"missing": &pb2.NestedWithRequired{},
				"contains": &pb2.NestedWithRequired{
					ReqString: proto.String("here"),
				},
			},
		},
		wantErr: "required field textpb2_proto.NestedWithRequired.req_string not set",
	}, {
		desc:         "indirect required field in map with AllowPartial",
		umo:          UnmarshalOptions{AllowPartial: true},
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "missing", Value: bson.D{}},
				{Key: "contains", Value: bson.D{
					{Key: "reqString", Value: "here"},
				}},
			}},
		},
		wantMessage: &pb2.IndirectRequired{
			StrToNested: map[string]*pb2.NestedWithRequired{
				"missing": &pb2.NestedWithRequired{},
				"contains": &pb2.NestedWithRequired{
					ReqString: proto.String("here"),
				},
			},
		},
	}, {
		desc:         "indirect required field in oneof",
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "oneofNested", Value: bson.D{}},
		},
		wantMessage: &pb2.IndirectRequired{
			Union: &pb2.IndirectRequired_OneofNested{
				OneofNested: &pb2.NestedWithRequired{},
			},
		},
		wantErr: "required field textpb2_proto.NestedWithRequired.req_string not set",
	}, {
		desc:         "indirect required field in oneof with AllowPartial",
		umo:          UnmarshalOptions{AllowPartial: true},
		inputMessage: &pb2.IndirectRequired{},
		inputBson: bson.D{
			{Key: "oneofNested", Value: bson.D{}},
		},
		wantMessage: &pb2.IndirectRequired{
			Union: &pb2.IndirectRequired_OneofNested{
				OneofNested: &pb2.NestedWithRequired{},
			},
		},
	}, {
		desc:         "extensions of non-repeated fields",
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
			{Key: "optString", Value: "non-extension field"},
			{Key: "optInt32", Value: 42},
			{Key: "optBool", Value: true},
			{Key: "[textpb2_proto.opt_ext_bool]", Value: true},
			{Key: "[textpb2_proto.opt_ext_nested]", Value: bson.D{
				{Key: "optString", Value: "nested in an extension"},
				{Key: "optNested", Value: bson.D{
					{Key: "optString", Value: "another nested in an extension"},
				}},
			}},
			{Key: "[textpb2_proto.opt_ext_string]", Value: "extension field"},
			{Key: "[textpb2_proto.opt_ext_enum]", Value: "TEN"},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "extensions of repeated fields",
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.rpt_ext_enum]", Value: bson.A{"TEN", 101, "ONE"}},
			{Key: "[textpb2_proto.rpt_ext_fixed32]", Value: bson.A{42, 47}},
			{Key: "[textpb2_proto.rpt_ext_nested]", Value: bson.A{
				bson.D{{Key: "optString", Value: "one"}},
				bson.D{{Key: "optString", Value: "two"}},
				bson.D{{Key: "optString", Value: "three"}},
			}},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "extensions of non-repeated fields in another message",
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
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
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "extensions of repeated fields in another message",
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
			{Key: "optString", Value: "non-extension field"},
			{Key: "optBool", Value: true},
			{Key: "optInt32", Value: 42},
			{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_nested]", Value: bson.A{
				bson.D{{Key: "optString", Value: "one"}},
				bson.D{{Key: "optString", Value: "two"}},
				bson.D{{Key: "optString", Value: "three"}},
			}},
			{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_enum]", Value: bson.A{"TEN", 101, "ONE"}},
			{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_string]", Value: bson.A{"hello", "world"}},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "invalid extension field name",
		inputMessage: &pb2.Extensions{},
		inputBson:    bson.D{
			{Key: "[textpb2_proto.invalid_message_field]", Value: true},
		},
		wantErr:      `unknown field "[textpb2_proto.invalid_message_field]"`,
	}, {
		desc:         "extensions of repeated field contains null",
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.ExtensionsContainer.rpt_ext_nested]", Value: bson.A{
				bson.D{{Key: "optString", Value: "one"}},
				primitive.Null{},
				bson.D{{Key: "optString", Value: "three"}},
			}},
		},
		wantErr: `unexpected message value: {}`,
	}, /* {
		desc:         "MessageSet",
		inputMessage: &pb2.MessageSet{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.MessageSetExtension]", Value: bson.D{
				{Key: "optString", Value: "a messageset extension"},
			}},
			{Key: "[textpb2_proto.MessageSetExtension.ext_nested]", Value: bson.D{
				{Key: "optString", Value: "just a regular extension"},
			}},
			{Key: "[textpb2_proto.MessageSetExtension.not_message_set_extension]", Value: bson.D{
				{Key: "optString", Value: "not a messageset extension"},
			}},
		},
		wantMessage: func() proto.Message {
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
		skip: !protoLegacy,
	}, {
		desc:         "not real MessageSet 1",
		inputMessage: &pb2.FakeMessageSet{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.FakeMessageSetExtension.message_set_extension]", Value: bson.D{
				{Key: "optString", Value: "not a messageset extension"},
			}},
		},
		wantMessage: func() proto.Message {
			m := &pb2.FakeMessageSet{}
			proto.SetExtension(m, pb2.E_FakeMessageSetExtension_MessageSetExtension, &pb2.FakeMessageSetExtension{
				OptString: proto.String("not a messageset extension"),
			})
			return m
		}(),
		skip: !protoLegacy,
	}, {
		desc:         "not real MessageSet 2",
		inputMessage: &pb2.FakeMessageSet{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.FakeMessageSetExtension]", Value: bson.D{
				{Key: "optString", Value: "not a messageset extension"},
			}},
		},
		wantErr: `unknown field "[pb2.FakeMessageSetExtension]"`,
		skip:    !protoLegacy,
	}, {
		desc:         "not real MessageSet 3",
		inputMessage: &pb2.MessageSet{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.message_set_extension]", Value: bson.D{
				{Key: "optString", Value: "another not a messageset extension"},
			}},
		},
		wantMessage: func() proto.Message {
			m := &pb2.MessageSet{}
			proto.SetExtension(m, pb2.E_MessageSetExtension, &pb2.FakeMessageSetExtension{
				OptString: proto.String("another not a messageset extension"),
			})
			return m
		}(),
		skip: !protoLegacy,
	}, */ {
		desc:         "Empty",
		inputMessage: &emptypb.Empty{},
		inputBson:    bson.D{},
		wantMessage:  &emptypb.Empty{},
	}, {
		desc:         "Empty contains unknown",
		inputMessage: &emptypb.Empty{},
		inputBson:    bson.D{{Key: "unknown", Value: primitive.Null{}}},
		wantErr:      `unknown field "unknown"`,
	}, {
		desc:         "BoolValue false",
		inputMessage: &wrapperspb.BoolValue{},
		inputBson:    false,
		wantMessage:  &wrapperspb.BoolValue{},
	}, {
		desc:         "BoolValue true",
		inputMessage: &wrapperspb.BoolValue{},
		inputBson:    true,
		wantMessage:  &wrapperspb.BoolValue{Value: true},
	}, {
		desc:         "BoolValue invalid value",
		inputMessage: &wrapperspb.BoolValue{},
		inputBson:    bson.D{},
		wantErr:      `invalid value for bool type: [] (has type primitive.D)`,
	}, {
		desc:         "Int32Value",
		inputMessage: &wrapperspb.Int32Value{},
		inputBson:    42,
		wantMessage:  &wrapperspb.Int32Value{Value: 42},
	}, {
		desc:         "Int64Value",
		inputMessage: &wrapperspb.Int64Value{},
		inputBson:    42,
		wantMessage:  &wrapperspb.Int64Value{Value: 42},
	}, {
		desc:         "UInt32Value",
		inputMessage: &wrapperspb.UInt32Value{},
		inputBson:    42,
		wantMessage:  &wrapperspb.UInt32Value{Value: 42},
	}, {
		desc:         "UInt64Value",
		inputMessage: &wrapperspb.UInt64Value{},
		inputBson:    42,
		wantMessage:  &wrapperspb.UInt64Value{Value: 42},
	}, {
		desc:         "FloatValue",
		inputMessage: &wrapperspb.FloatValue{},
		inputBson:    1.02,
		wantMessage:  &wrapperspb.FloatValue{Value: 1.02},
	}, {
		desc:         "FloatValue exceeds max limit",
		inputMessage: &wrapperspb.FloatValue{},
		inputBson:    1.23e+40,
		wantErr:      `invalid value for float type: 1.23e+40`,
	}, {
		desc:         "FloatValue Infinity",
		inputMessage: &wrapperspb.FloatValue{},
		inputBson:    float32(math.Inf(-1)),
		wantMessage:  &wrapperspb.FloatValue{Value: float32(math.Inf(-1))},
	}, {
		desc:         "DoubleValue",
		inputMessage: &wrapperspb.DoubleValue{},
		inputBson:    1.02,
		wantMessage:  &wrapperspb.DoubleValue{Value: 1.02},
	}, {
		desc:         "DoubleValue Infinity",
		inputMessage: &wrapperspb.DoubleValue{},
		inputBson:    math.Inf(+1),
		wantMessage:  &wrapperspb.DoubleValue{Value: math.Inf(+1)},
	}, {
		desc:         "StringValue empty",
		inputMessage: &wrapperspb.StringValue{},
		inputBson:    "",
		wantMessage:  &wrapperspb.StringValue{},
	}, {
		desc:         "StringValue",
		inputMessage: &wrapperspb.StringValue{},
		inputBson:    "谷歌",
		wantMessage:  &wrapperspb.StringValue{Value: "谷歌"},
	}, {
		desc:         "StringValue with invalid UTF-8 error",
		inputMessage: &wrapperspb.StringValue{},
		inputBson:    "abc\xff",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "StringValue field with invalid UTF-8 error",
		inputMessage: &pb2.KnownTypes{},
		inputBson:    bson.D{{Key: "optString", Value: "abc\xff"}},
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "NullValue field with JSON null",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optNull", Value: primitive.Null{}}},
		wantMessage: &pb2.KnownTypes{OptNull: new(structpb.NullValue)},
	}, {
		desc:         "NullValue field with string",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optNull", Value: "NULL_VALUE"}},
		wantMessage: &pb2.KnownTypes{OptNull: new(structpb.NullValue)},
	}, {
		desc:         "BytesValue",
		inputMessage: &wrapperspb.BytesValue{},
		inputBson:    primitive.Binary{Data: []byte("hello")},
		wantMessage:  &wrapperspb.BytesValue{Value: []byte("hello")},
	}, {
		desc:         "Value null",
		inputMessage: &structpb.Value{},
		inputBson:    primitive.Null{},
		wantMessage:  &structpb.Value{Kind: &structpb.Value_NullValue{}},
	}, {
		desc:         "Value field null",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: primitive.Null{}}},
		wantMessage: &pb2.KnownTypes{
			OptValue: &structpb.Value{Kind: &structpb.Value_NullValue{}},
		},
	}, {
		desc:         "Value bool",
		inputMessage: &structpb.Value{},
		inputBson:    false,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_BoolValue{}},
	}, {
		desc:         "Value field bool",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: true}},
		wantMessage: &pb2.KnownTypes{
			OptValue: &structpb.Value{Kind: &structpb.Value_BoolValue{true}},
		},
	}, {
		desc:         "Value number",
		inputMessage: &structpb.Value{},
		inputBson:    1.02,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_NumberValue{1.02}},
	}, {
		desc:         "Value field number",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: 1.02}},
		wantMessage: &pb2.KnownTypes{
			OptValue: &structpb.Value{Kind: &structpb.Value_NumberValue{1.02}},
		},
	}, {
		desc:         "Value string",
		inputMessage: &structpb.Value{},
		inputBson:    "hello",
		wantMessage:  &structpb.Value{Kind: &structpb.Value_StringValue{"hello"}},
	}, {
		desc:         "Value string with invalid UTF-8",
		inputMessage: &structpb.Value{},
		inputBson:    "\xff",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Value field string",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: "NaN"}},
		wantMessage: &pb2.KnownTypes{
			OptValue: &structpb.Value{Kind: &structpb.Value_StringValue{"NaN"}},
		},
	}, {
		desc:         "Value field string with invalid UTF-8",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: "\xff"}},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Value empty struct",
		inputMessage: &structpb.Value{},
		inputBson:    bson.D{},
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				&structpb.Struct{Fields: map[string]*structpb.Value{}},
			},
		},
	}, {
		desc:         "Value struct",
		inputMessage: &structpb.Value{},
		inputBson: bson.D{
			{Key: "string", Value: "hello"},
			{Key: "number", Value: 123},
			{Key: "null", Value: primitive.Null{}},
			{Key: "bool", Value: false},
			{Key: "struct", Value: bson.D{
				{Key: "string", Value: "world"},
			}},
			{Key: "list", Value: bson.A{}},
		},
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"string": {Kind: &structpb.Value_StringValue{"hello"}},
						"number": {Kind: &structpb.Value_NumberValue{123}},
						"null":   {Kind: &structpb.Value_NullValue{}},
						"bool":   {Kind: &structpb.Value_BoolValue{false}},
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
							Kind: &structpb.Value_ListValue{&structpb.ListValue{}},
						},
					},
				},
			},
		},
	}, {
		desc:         "Value struct with invalid UTF-8 string",
		inputMessage: &structpb.Value{},
		inputBson:    "{\"string\": \"abc\xff\"}",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Value field struct",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{
			{Key: "optValue", Value: bson.D{
				{Key: "string", Value: "hello"},
			}},
		},
		wantMessage: &pb2.KnownTypes{
			OptValue: &structpb.Value{
				Kind: &structpb.Value_StructValue{
					&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"string": {Kind: &structpb.Value_StringValue{"hello"}},
						},
					},
				},
			},
		},
	}, {
		desc:         "Value empty list",
		inputMessage: &structpb.Value{},
		inputBson:    bson.A{},
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_ListValue{
				&structpb.ListValue{Values: []*structpb.Value{}},
			},
		},
	}, {
		desc:         "Value list",
		inputMessage: &structpb.Value{},
		inputBson: bson.A{
			"string",
			123,
			primitive.Null{},
			true,
			bson.D{},
			bson.A{
				"string",
				1.23,
				primitive.Null{},
				false,
			},
		},
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_ListValue{
				&structpb.ListValue{
					Values: []*structpb.Value{
						{Kind: &structpb.Value_StringValue{"string"}},
						{Kind: &structpb.Value_NumberValue{123}},
						{Kind: &structpb.Value_NullValue{}},
						{Kind: &structpb.Value_BoolValue{true}},
						{Kind: &structpb.Value_StructValue{&structpb.Struct{}}},
						{
							Kind: &structpb.Value_ListValue{
								&structpb.ListValue{
									Values: []*structpb.Value{
										{Kind: &structpb.Value_StringValue{"string"}},
										{Kind: &structpb.Value_NumberValue{1.23}},
										{Kind: &structpb.Value_NullValue{}},
										{Kind: &structpb.Value_BoolValue{false}},
									},
								},
							},
						},
					},
				},
			},
		},
	}, {
		desc:         "Value list with invalid UTF-8 string",
		inputMessage: &structpb.Value{},
		inputBson:    bson.A{"abc\xff"},
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Value field list with invalid UTF-8 string",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{{Key: "optValue", Value: bson.A{"abc\xff"}}},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Duration empty string",
		inputMessage: &durationpb.Duration{},
		inputBson:    "",
		wantErr:      `invalid google.protobuf.Duration value ""`,
	}, {
		desc:         "Duration with secs",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 3},
			{Key: "Nanos", Value: 0},
		},
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with missing nanos",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 3},
		},
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with -secs",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: -3},
			{Key: "Nanos", Value: 0},
		},
		wantMessage:  &durationpb.Duration{Seconds: -3},
	}, {
		desc:         "Duration with nanos",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Nanos", Value: int(1e6)},
		},
		wantMessage:  &durationpb.Duration{Nanos: 1e6},
	}, {
		desc:         "Duration with -nanos",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 0},
			{Key: "Nanos", Value: int(-1e6)},
		},
		wantMessage:  &durationpb.Duration{Nanos: -1e6},
	}, {
		desc:         "Duration with small ints",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: int16(10)},
			{Key: "Nanos", Value: int16(10)},
		},
		wantMessage:  &durationpb.Duration{Seconds: 10, Nanos: 10},
	}, {
		desc:         "Duration with +nanos",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Nanos", Value: int(1e6)},
		},
		wantMessage:  &durationpb.Duration{Nanos: 1e6},
	}, {
		desc:         "Duration with -secs -nanos",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: -123},
			{Key: "Nanos", Value: -450},
		},
		wantMessage:  &durationpb.Duration{Seconds: -123, Nanos: -450},
	}, {
		desc:         "Duration with large secs",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: int(1e10)},
			{Key: "Nanos", Value: 1},
		},
		wantMessage:  &durationpb.Duration{Seconds: 1e10, Nanos: 1},
	}, {
		desc:         "Duration max value",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 315576000000},
			{Key: "Nanos", Value: 999999999},
		},
		wantMessage:  &durationpb.Duration{Seconds: 315576000000, Nanos: 999999999},
	}, {
		desc:         "Duration min value",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: -315576000000},
			{Key: "Nanos", Value: -999999999},
		},
		wantMessage:  &durationpb.Duration{Seconds: -315576000000, Nanos: -999999999},
	}, {
		desc:         "Duration with +secs out of range",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 315576000001},
		},
		wantErr:      `google.protobuf.Duration: seconds out of range 315576000001`,
	}, {
		desc:         "Duration with -secs out of range",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: -315576000001},
		},
		wantErr:      `google.protobuf.Duration: seconds out of range -315576000001`,
	}, {
		desc:         "Duration invalid integer",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Seconds", Value: 1.6},
		},
		wantErr:      `invalid google.protobuf.Duration seconds: 1.6 (want int64 but got float64)`,
	}, {
		desc:         "Duration invalid type",
		inputMessage: &durationpb.Duration{},
		inputBson:    bson.D{
			{Key: "Nanos", Value: true},
		},
		wantErr:      `invalid google.protobuf.Duration nanoseconds: true (want int32 but got bool)`,
	}, {
		desc:         "Timestamp zero",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    0,
		wantMessage:  &timestamppb.Timestamp{},
	}, {
		desc:         "Timestamp with tz adjustment",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    primitive.NewDateTimeFromTime(time.Unix(-3600, 0)),
		wantMessage:  &timestamppb.Timestamp{Seconds: -3600},
	},  {
		desc:         "Timestamp UTC",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    primitive.NewDateTimeFromTime(time.Unix(1553036601, 0)),
		wantMessage:  &timestamppb.Timestamp{Seconds: 1553036601},
	},  {
		desc:         "Timestamp with nanos",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    primitive.NewDateTimeFromTime(time.Unix(1234, 12345)),
		wantMessage:  &timestamppb.Timestamp{Seconds: 1234},
	}, {
		desc:         "Timestamp max value",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    primitive.NewDateTimeFromTime(time.Unix(1234, 12345)),
		wantMessage:  &timestamppb.Timestamp{Seconds: 1234},
	}, /* {
		desc:         "Timestamp above max value",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    primitive.NewDateTimeFromTime(time.Unix(253402300799, 999999999).UTC()),
		wantErr:      `google.protobuf.Timestamp value out of range: "9999-12-31T23:59:59-01:00"`,
	}, {
		desc:         "Timestamp min value",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    `"0001-01-01T00:00:00Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: -62135596800},
	}, {
		desc:         "Timestamp below min value",
		inputMessage: &timestamppb.Timestamp{},
		inputBson:    `"0001-01-01T00:00:00+01:00"`,
		wantErr:      `google.protobuf.Timestamp value out of range: "0001-01-01T00:00:00+01:00"`,
	}, */ /* {
		desc:         "FieldMask empty",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `""`,
		wantMessage:  &fieldmaskpb.FieldMask{Paths: []string{}},
	}, {
		desc:         "FieldMask",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `"foo,fooBar,foo.barQux,Foo"`,
		wantMessage: &fieldmaskpb.FieldMask{
			Paths: []string{
				"foo",
				"foo_bar",
				"foo.bar_qux",
				"_foo",
			},
		},
	}, {
		desc:         "FieldMask empty path 1",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `"foo,"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: ""`,
	}, {
		desc:         "FieldMask empty path 2",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `"foo,  ,bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "  "`,
	}, {
		desc:         "FieldMask invalid char 1",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `"foo_bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "foo_bar"`,
	}, {
		desc:         "FieldMask invalid char 2",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputBson:    `"foo@bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "foo@bar"`,
	}, {
		desc:         "FieldMask field",
		inputMessage: &pb2.KnownTypes{},
		inputBson: `{
  "optFieldmask": "foo,qux.fooBar"
}`,
		wantMessage: &pb2.KnownTypes{
			OptFieldmask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"foo",
					"qux.foo_bar",
				},
			},
		},
	}, */ {
		desc:         "Any empty",
		inputMessage: &anypb.Any{},
		inputBson:    bson.D{},
		wantMessage:  &anypb.Any{},
	}, {
		desc:         "Any with non-custom message",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "foo/textpb2_proto.Nested"},
			{Key: "optString", Value: "embedded inside Any"},
			{Key: "optNested", Value: bson.D{
				{Key: "optString", Value: "inception"},
			}},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with empty embedded message",
		inputMessage: &anypb.Any{},
		inputBson:    bson.D{
			{Key: "@type", Value: "foo/textpb2_proto.Nested"},
		},
		wantMessage:  &anypb.Any{TypeUrl: "foo/textpb2_proto.Nested"},
	}, {
		desc:         "Any without registered type",
		umo:          UnmarshalOptions{Resolver: new(preg.Types)},
		inputMessage: &anypb.Any{},
		inputBson:    bson.D{
			{Key: "@type", Value: "foo/textpb2_proto.Nested"},
		},
		wantErr:      `unable to resolve "foo/textpb2_proto.Nested": proto: not found`,
	}, {
		desc:         "Any with missing required",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "textpb2_proto.PartialRequired"},
			{Key: "optString", Value: "embedded inside Any"},
		},
		wantMessage: func() proto.Message {
			m := &pb2.PartialRequired{
				OptString: proto.String("embedded inside Any"),
			}
			b, err := proto.MarshalOptions{
				Deterministic: true,
				AllowPartial:  true,
			}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: string(m.ProtoReflect().Descriptor().FullName()),
				Value:   b,
			}
		}(),
	}, {
		desc: "Any with partial required and AllowPartial",
		umo: UnmarshalOptions{
			AllowPartial: true,
		},
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "textpb2_proto.PartialRequired"},
			{Key: "optString", Value: "embedded inside Any"},
		},
		wantMessage: func() proto.Message {
			m := &pb2.PartialRequired{
				OptString: proto.String("embedded inside Any"),
			}
			b, err := proto.MarshalOptions{
				Deterministic: true,
				AllowPartial:  true,
			}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: string(m.ProtoReflect().Descriptor().FullName()),
				Value:   b,
			}
		}(),
	}, {
		desc:         "Any with invalid UTF-8",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "foo/textpb2_proto.Nested"},
			{Key: "optString", Value: "abc\xff"},
		},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Any with BoolValue",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "type.googleapis.com/google.protobuf.BoolValue"},
			{Key: "value", Value: true},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with Empty",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "type.googleapis.com/google.protobuf.Empty"},
			{Key: "value", Value: bson.D{}},
		},
		wantMessage: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Empty",
		},
	}, {
		desc:         "Any with missing Empty",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "type.googleapis.com/google.protobuf.Empty"},
		},
		wantErr: `missing "value" field`,
	}, {
		desc:         "Any with StringValue containing invalid UTF-8",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.StringValue"},
			{Key: "value", Value: "abc\xff"},
		},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Any with Int64Value",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.Int64Value"},
			{Key: "value", Value: 42},
		},
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with invalid Int64Value",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.Int64Value"},
			{Key: "value", Value: "forty-two"},
		},
		wantErr: `invalid value for int64 type: "forty-two"`,
	}, {
		desc:         "Any with invalid UInt64Value",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.UInt64Value"},
			{Key: "value", Value: -42},
		},
		wantErr: `invalid value for uint64 type: -42`,
	}, {
		desc:         "Any with Duration",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "type.googleapis.com/google.protobuf.Duration"},
			{Key: "value", Value: bson.D{
				{Key: "Seconds", Value: 3},
			}},
		},
		wantMessage: func() proto.Message {
			m := &durationpb.Duration{Seconds: 3}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Duration",
				Value:   b,
			}
		}(),
	}, {
		desc:         "Any with Value of StringValue",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.Value"},
			{Key: "value", Value: "abc\xff"},
		},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Any with Value of NullValue",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.Value"},
			{Key: "value", Value: primitive.Null{}},
		},
		wantMessage: func() proto.Message {
			m := &structpb.Value{Kind: &structpb.Value_NullValue{}}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "google.protobuf.Value",
				Value:   b,
			}
		}(),
	}, {
		desc:         "Any with Struct",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
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
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with missing @type",
		umo:          UnmarshalOptions{},
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "value", Value: bson.D{}},
		},
		wantErr: `missing @type field in non-empty message`,
	}, {
		desc:         "Any with empty @type",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: ""},
		},
		wantErr: `@type field contains empty or invalid value`,
	}, {
		desc:         "Any with duplicate @type",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.StringValue"},
			{Key: "value", Value: "hello"},
			{Key: "@type", Value: "textpb2_proto.Nested"},
		},
		wantErr: `duplicate @type field`,
	}, {
		desc:         "Any with duplicate value",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "google.protobuf.StringValue"},
			{Key: "value", Value: "hello"},
			{Key: "value", Value: "world"},
		},
		wantErr: `duplicate "value" field`,
	}, {
		desc:         "Any with unknown field",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "textpb2_proto.Nested"},
			{Key: "optString", Value: "hello"},
			{Key: "unknown", Value: "world"},
		},
		wantErr: `unknown field "unknown"`,
	}, {
		desc:         "Any with embedded type containing Any",
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "textpb2_proto.KnownTypes"},
			{Key: "optAny", Value: bson.D{
				{Key: "@type", Value: "google.protobuf.StringValue"},
				{Key: "value", Value: "abc\xff"},
			}},
		},
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "well known types as field values",
		inputMessage: &pb2.KnownTypes{},
		inputBson: bson.D{
			{Key: "optBool", Value: false},
			{Key: "optInt32", Value: 42},
			{Key: "optInt64", Value: 42},
			{Key: "optUint32", Value: 42},
			{Key: "optUint64", Value: 42},
			{Key: "optFloat", Value: 1.23},
			{Key: "optDouble", Value: 3.1415},
			{Key: "optString", Value: "hello"},
			{Key: "optBytes", Value: primitive.Binary{Data: []byte("hello")}},
			{Key: "optDuration", Value: bson.D{
				{Key: "Seconds", Value: 123},
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
			{Key: "optEmpty", Value: bson.D{}},
			{Key: "optAny", Value: bson.D{
				{Key: "@type", Value: "google.protobuf.Empty"},
				{Key: "value", Value: bson.D{}},
			}},
			// {Key: "optFieldmask", Value: "fooBar,barFoo"},
		},
		wantMessage: &pb2.KnownTypes{
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
					{
						Kind: &structpb.Value_StructValue{
							&structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
					},
					{
						Kind: &structpb.Value_ListValue{
							&structpb.ListValue{Values: []*structpb.Value{}},
						},
					},
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
	}, {
		desc:         "DiscardUnknown: regular messages",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb3.Nests{},
		inputBson: bson.D{
			{Key: "sNested", Value: bson.D{
				{Key: "unknown", Value: bson.D{
					{Key: "foo", Value: 1},
					{Key: "bar", Value: bson.A{1, 2, 3}},
				}},
			}},
			{Key: "unknown", Value: "not known"},
		},
		wantMessage: &pb3.Nests{SNested: &pb3.Nested{}},
	}, {
		desc:         "DiscardUnknown: repeated",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb2.Nests{},
		inputBson: bson.D{
			{Key: "rptNested", Value: bson.A{
				bson.D{{Key: "unknown", Value: "blah"}},
				bson.D{{Key: "optString", Value: "hello"}},
			}},
		},
		wantMessage: &pb2.Nests{
			RptNested: []*pb2.Nested{
				{},
				{OptString: proto.String("hello")},
			},
		},
	}, {
		desc:         "DiscardUnknown: map",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb3.Maps{},
		inputBson: bson.D{
			{Key: "strToNested", Value: bson.D{
				{Key: "nested_one", Value: bson.D{
					{Key: "unknown", Value: "what you see is not"},
				}},
			}},
		},
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested_one": {},
			},
		},
	}, {
		desc:         "DiscardUnknown: extension",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb2.Extensions{},
		inputBson: bson.D{
			{Key: "[textpb2_proto.opt_ext_nested]", Value: bson.D{
				{Key: "unknown", Value: bson.A{}},
			}},
		},
		wantMessage: func() proto.Message {
			m := &pb2.Extensions{}
			proto.SetExtension(m, pb2.E_OptExtNested, &pb2.Nested{})
			return m
		}(),
	}, {
		desc:         "DiscardUnknown: Empty",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &emptypb.Empty{},
		inputBson:    bson.D{
			{Key: "unknown", Value: "something"},
		},
		wantMessage:  &emptypb.Empty{},
	}, {
		desc:         "DiscardUnknown: Any without type",
		umo:          UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "value", Value: bson.D{
				{Key: "foo", Value: "bar"},
			}},
			{Key: "unknown", Value: true},
		},
		wantMessage: &anypb.Any{},
	}, {
		desc: "DiscardUnknown: Any",
		umo: UnmarshalOptions{
			DiscardUnknown: true,
		},
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "foo/textpb2_proto.Nested"},
			{Key: "unknown", Value: "none"},
		},
		wantMessage: &anypb.Any{
			TypeUrl: "foo/textpb2_proto.Nested",
		},
	}, {
		desc: "DiscardUnknown: Any with Empty",
		umo: UnmarshalOptions{
			DiscardUnknown: true,
		},
		inputMessage: &anypb.Any{},
		inputBson: bson.D{
			{Key: "@type", Value: "type.googleapis.com/google.protobuf.Empty"},
			{Key: "value", Value: bson.D{
				{Key: "unknown", Value: 47},
			}},
		},
		wantMessage: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Empty",
		},
	}, /* {
		desc:         "weak fields",
		inputMessage: &testpb.TestWeak{},
		inputBson:    bson.D{
			{Key: "weak_message1", Value: bson.D{
				{Key: "a", Value: 1},
			}},
		},
		wantMessage: func() *testpb.TestWeak {
			m := new(testpb.TestWeak)
			m.SetWeakMessage1(&weakpb.WeakImportMessage1{A: proto.Int32(1)})
			return m
		}(),
		skip: !protoLegacy,
	}, {
		desc:         "weak fields; unknown field",
		inputMessage: &testpb.TestWeak{},
		inputBson:    bson.D{
			{Key: "weak_message1", Value: bson.D{
				{Key: "a", Value: 1},
			}},
			{Key: "weak_message2", Value: bson.D{
				{Key: "a", Value: 1},
			}},
		},
		wantErr:      `unknown field "weak_message2"`, // weak_message2 is unknown since the package containing it is not imported
		skip:         !protoLegacy,
	}*/
}

	for _, tt := range tests {
		tt := tt
		if tt.skip {
			continue
		}
		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.umo.Unmarshal(tt.inputBson, tt.inputMessage); err != nil {
				if tt.wantErr == "" {
					t.Errorf("Unmarshal() got unexpected error: %v", err)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Unmarshal() error got %q, want %q", err, tt.wantErr)
				}
				return
			}
			if tt.wantErr != "" {
				t.Errorf("Unmarshal() got nil error, want error %q", tt.wantErr)
			}
			if tt.wantMessage != nil && !proto.Equal(tt.inputMessage, tt.wantMessage) {
				t.Errorf("Unmarshal()\n<got>\n%v\n<want>\n%v\n", tt.inputMessage, tt.wantMessage)
			}
		})
	}
}
