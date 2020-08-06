package bsonpb

import (
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	anypb "github.com/golang/protobuf/ptypes/any"
	durpb "github.com/golang/protobuf/ptypes/duration"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/romnnn/bsonpb/internal/testprotos/v1/test_objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	if err := proto.SetExtension(realNumber, pb.E_Name, &realNumberName); err != nil {
		panic(err)
	}
	if err := proto.SetExtension(realNumber, pb.E_Complex_RealExtension, complexNumber); err != nil {
		panic(err)
	}
	registerDynamicMessage()
}

func marshaledNestedProto(msg proto.Message) []byte {
	marshaledProto, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return marshaledProto
}

func protoTimestamp(t time.Time) *tspb.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return ts
}

func MarshalBsonToJson(p interface{}) (string, error) {
	b, err := bson.MarshalExtJSON(p, true, true)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var (
	defaultMarshaler = Marshaler{}

	simpleObject = &pb.Simple{
		OBool:   proto.Bool(true),
		OInt32:  proto.Int32(-32),
		OInt64:  proto.Int64(-6400000000),
		OUint32: proto.Uint32(32),
		OUint64: proto.Uint64(6400000000),
		OSint32: proto.Int32(-13),
		OSint64: proto.Int64(-2600000000),
		OFloat:  proto.Float32(3.14),
		ODouble: proto.Float64(6.02214179e23),
		OString: proto.String("hello \"there\""),
		OBytes:  []byte("beep boop"),
	}

	simpleObjectOutputBSON = bson.D{
		{"oBool", true},
		{"oInt32", int32(-32)},
		{"oInt64", int64(-6400000000)},
		{"oUint32", uint32(32)},
		{"oUint64", uint64(6400000000)},
		{"oSint32", int32(-13)},
		{"oSint64", int64(-2600000000)},
		{"oFloat", float32(3.14)},
		{"oDouble", float64(6.02214179e+23)},
		{"oString", "hello \"there\""},
		{"oBytes", primitive.Binary{Data: []byte("beep boop")}},
	}

	repeatsObject = &pb.Repeats{
		RBool:   []bool{true, false, true},
		RInt32:  []int32{-3, -4, -5},
		RInt64:  []int64{-123456789, -987654321},
		RUint32: []uint32{1, 2, 3},
		RUint64: []uint64{6789012345, 3456789012},
		RSint32: []int32{-1, -2, -3},
		RSint64: []int64{-6789012345, -3456789012},
		RFloat:  []float32{3.14, 6.28},
		RDouble: []float64{299792458 * 1e20, 6.62606957e-34},
		RString: []string{"happy", "days"},
		RBytes:  [][]byte{[]byte("skittles"), []byte("m&m's")},
	}

	repeatsObjectBSON = bson.D{
		{"rBool", bson.A{true, false, true}},
		{"rInt32", bson.A{int32(-3), int32(-4), int32(-5)}},
		{"rInt64", bson.A{int64(-123456789), int64(-987654321)}},
		{"rUint32", bson.A{uint32(1), uint32(2), uint32(3)}},
		{"rUint64", bson.A{uint64(6789012345), uint64(3456789012)}},
		{"rSint32", bson.A{int32(-1), int32(-2), int32(-3)}},
		{"rSint64", bson.A{int64(-6789012345), int64(-3456789012)}},
		{"rFloat", bson.A{float32(3.14), float32(6.28)}},
		{"rDouble", bson.A{float64(2.99792458e+28), float64(6.62606957e-34)}},
		{"rString", bson.A{"happy", "days"}},
		{"rBytes", bson.A{primitive.Binary{Data: []byte("skittles")}, primitive.Binary{Data: []byte("m&m's")}}},
	}

	meaningOfLife = int32(42)

	innerSimple   = &pb.Simple{OInt32: &meaningOfLife}
	innerSimple2  = &pb.Simple{OInt32: &meaningOfLife, OInt64: proto.Int64(25)}
	innerRepeats  = &pb.Repeats{RString: []string{"roses", "red"}}
	innerRepeats2 = &pb.Repeats{RString: []string{"violets", "blue"}}
	complexObject = &pb.Widget{
		Color:    pb.Widget_GREEN.Enum(),
		RColor:   []pb.Widget_Color{pb.Widget_RED, pb.Widget_GREEN, pb.Widget_BLUE},
		Simple:   innerSimple,
		RSimple:  []*pb.Simple{innerSimple, innerSimple2},
		Repeats:  innerRepeats,
		RRepeats: []*pb.Repeats{innerRepeats, innerRepeats2},
	}

	innerSimpleBSON = bson.D{
		{"oBool", primitive.Null{}},
		{"oInt32", int32(42)},
		{"oInt64", primitive.Null{}},
		{"oUint32", primitive.Null{}},
		{"oUint64", primitive.Null{}},
		{"oSint32", primitive.Null{}},
		{"oSint64", primitive.Null{}},
		{"oFloat", primitive.Null{}},
		{"oDouble", primitive.Null{}},
		{"oString", primitive.Null{}},
		{"oBytes", primitive.Binary{}},
	}

	innerSimpleMinimalBSON = bson.D{
		{"oInt32", int32(42)},
	}

	innerSimple2BSON = bson.D{
		{"oBool", primitive.Null{}},
		{"oInt32", int32(42)},
		{"oInt64", int64(25)},
		{"oUint32", primitive.Null{}},
		{"oUint64", primitive.Null{}},
		{"oSint32", primitive.Null{}},
		{"oSint64", primitive.Null{}},
		{"oFloat", primitive.Null{}},
		{"oDouble", primitive.Null{}},
		{"oString", primitive.Null{}},
		{"oBytes", primitive.Binary{}},
	}

	innerSimple2MinimalBSON = bson.D{
		{"oInt32", int32(42)},
		{"oInt64", int64(25)},
	}

	innerRepeatsBSON = bson.D{
		{"rBool", bson.A{}},
		{"rInt32", bson.A{}},
		{"rInt64", bson.A{}},
		{"rUint32", bson.A{}},
		{"rUint64", bson.A{}},
		{"rSint32", bson.A{}},
		{"rSint64", bson.A{}},
		{"rFloat", bson.A{}},
		{"rDouble", bson.A{}},
		{"rString", bson.A{"roses", "red"}},
		{"rBytes", bson.A{}},
	}

	innerRepeatsMinimalBSON = bson.D{
		{"rString", bson.A{"roses", "red"}},
	}

	innerRepeats2BSON = bson.D{
		{"rBool", bson.A{}},
		{"rInt32", bson.A{}},
		{"rInt64", bson.A{}},
		{"rUint32", bson.A{}},
		{"rUint64", bson.A{}},
		{"rSint32", bson.A{}},
		{"rSint64", bson.A{}},
		{"rFloat", bson.A{}},
		{"rDouble", bson.A{}},
		{"rString", bson.A{"violets", "blue"}},
		{"rBytes", bson.A{}},
	}

	innerRepeats2MinimalBSON = bson.D{
		{"rString", bson.A{"violets", "blue"}},
	}

	complexObjectBSON = bson.D{
		{"color", "GREEN"},
		{"rColor", bson.A{"RED", "GREEN", "BLUE"}},
		{"simple", innerSimpleBSON},
		{"rSimple", bson.A{innerSimpleBSON, innerSimple2BSON}},
		{"repeats", innerRepeatsBSON},
		{"rRepeats", bson.A{
			innerRepeatsBSON,
			innerRepeats2BSON,
		}},
	}

	complexObjectMinimalBSON = bson.D{
		{"color", "GREEN"},
		{"rColor", bson.A{"RED", "GREEN", "BLUE"}},
		{"simple", innerSimpleMinimalBSON},
		{"rSimple", bson.A{innerSimpleMinimalBSON, innerSimple2MinimalBSON}},
		{"repeats", innerRepeatsMinimalBSON},
		{"rRepeats", bson.A{
			innerRepeatsMinimalBSON,
			innerRepeats2MinimalBSON,
		}},
	}

	realNumber     = &pb.Real{Value: proto.Float64(3.14159265359)}
	realNumberName = "Pi"
	complexNumber  = &pb.Complex{Imaginary: proto.Float64(0.5772156649)}
	realNumberBSON = bson.D{
		{"value", float64(3.14159265359)},
		{"[test_objects.Complex.real_extension]", bson.D{{"imaginary", float64(0.5772156649)}}},
		{"[test_objects.name]", "Pi"},
	}

	anySimple = &pb.KnownTypes{
		An: &anypb.Any{
			TypeUrl: "something.example.com/test_objects.Simple",
			Value: marshaledNestedProto(&pb.Simple{
				OInt32: proto.Int32(12),
				OBool:  proto.Bool(true),
			}),
		},
	}

	anySimpleBSON = bson.D{
		{"an", bson.D{
			{"@type", "something.example.com/test_objects.Simple"},
			{"oBool", true},
			{"oInt32", int32(12)},
		}},
	}

	anyWellKnown = &pb.KnownTypes{
		An: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Duration",
			Value: marshaledNestedProto(&durpb.Duration{
				Seconds: 1,
				Nanos:   212000000,
			}),
		},
	}
	anyWellKnownBSON = bson.D{
		{"an", bson.D{
			{"@type", "type.googleapis.com/google.protobuf.Duration"},
			{"value", 1.212},
		}},
	}

	nonFinites = &pb.NonFinites{
		FNan:  proto.Float32(float32(math.NaN())),
		FPinf: proto.Float32(float32(math.Inf(1))),
		FNinf: proto.Float32(float32(math.Inf(-1))),
		DNan:  proto.Float64(float64(math.NaN())),
		DPinf: proto.Float64(float64(math.Inf(1))),
		DNinf: proto.Float64(float64(math.Inf(-1))),
	}
	nonFinitesBSON = bson.D{
		{"fNan", float32(math.NaN())},
		{"fPinf", float32(math.Inf(1))},
		{"fNinf", float32(math.Inf(-1))},
		{"dNan", float64(math.NaN())},
		{"dPinf", float64(math.Inf(1))},
		{"dNinf", float64(math.Inf(-1))},
	}
)
