## bsonpb

[![Build Status](https://travis-ci.com/romnnn/bsonpb.svg?branch=master)](https://travis-ci.com/romnnn/bsonpb)
[![GitHub](https://img.shields.io/github/license/romnnn/bsonpb)](https://github.com/romnnn/bsonpb)
[![GoDoc](https://godoc.org/github.com/romnnn/bsonpb?status.svg)](https://godoc.org/github.com/romnnn/bsonpb)
[![Test Coverage](https://codecov.io/gh/romnnn/bsonpb/branch/master/graph/badge.svg)](https://codecov.io/gh/romnnn/bsonpb)
[![Release](https://img.shields.io/github/v/release/romnnn/bsonpb)](https://github.com/romnnn/bsonpb/releases/latest)

This package allows to serialize/deserialize go `protobuf` messages into/from `bson` documents.

It was inspired by the official [golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) package.

```go
import "github.com/romnnn/bsonbp"
```

#### Marshaling

```golang
marshaler := bsonbp.Marshaler{}
myProto := &pb.Widget{RColor: []pb.Widget_Color{pb.Widget_RED}}
bson, err := marshaler.Marshal(myProto)
if err != nil {
    log.Fatalf("Failed to marshal with error: %s\n", err.Error())
}
log.Printf("Marshaled bson: %s\n", bson)
```

#### Unmarshaling

```golang
unmarshaler := bsonpb.Unmarshaler{}
widgetBson := bson.D{{"rColor", bson.A{"RED"}}}

// Marshal bson to bytes
rawBson, err := bson.Marshal(widgetBson)
if err != nil {
    fmt.Printf("marshaling bson to bytes failed: %s", err.Error())
}

var result pb.Widget
err = unmarshaler.Unmarshal(rawBson, &result)
if err != nil {
    log.Fatalf("unmarshaling failed: %s\n", err.Error())
}
log.Printf("Unmarshaled proto: %s\n", result)
```

You can also run the examples if you wish:

```bash
bazel run //examples/marshal
bazel run //examples/unmarshal
```

#### Tests

```bash
bazel test //:go_default_test
bazel test //:marshal_test  # Marshalling tests only
bazel test //:unmarshal_test  # Unmarshalling tests only
```

#### Acknowledgements

- Authors of the [golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) package that was used as a starting point.
