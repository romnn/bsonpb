## bsonpb

[![Build Status](https://travis-ci.com/romnnn/bsonpb.svg?branch=master)](https://travis-ci.com/romnnn/bsonpb)
[![GitHub](https://img.shields.io/github/license/romnnn/bsonpb)](https://github.com/romnnn/bsonpb)
[![GoDoc](https://godoc.org/github.com/romnnn/bsonpb?status.svg)](https://godoc.org/github.com/romnnn/bsonpb)
[![Test Coverage](https://codecov.io/gh/romnnn/bsonpb/branch/master/graph/badge.svg)](https://codecov.io/gh/romnnn/bsonpb)
[![Release](https://img.shields.io/github/v/release/romnnn/bsonpb)](https://github.com/romnnn/bsonpb/releases/latest)

This package allows to serialize/deserialize go `protobuf` messages into/from `bson` documents.

**Important notes**: 
- As of the time of writing, the golang protobuf implementation is actively transitioning from the old [github.com/golang/protobuf](https://github.com/golang/protobuf) to the new [google.golang.org/protobuf](https://github.com/protocolbuffers/protobuf-go) API.
- Because of numerous version instabilities and inter-dependencies of both packages - especially with the bazel build system, **both v1 and v2** are in an **experimental state**.

```go
import "github.com/romnnn/bsonbp/v1" // v1 (github.com/golang/protobuf)
import "github.com/romnnn/bsonbp/v2" // v2 (google.golang.org/protobuf)
```

#### Usage (v2)

###### Marshaling

```golang
import "github.com/romnnn/bsonbp/v2"

myProto := &pb.Message{Name: "Test", Hilarity: pb.Message_SLAPSTICK}
opts := bsonpb.MarshalOptions{}
marshaled, err := opts.Marshal(someProto)
if err != nil {
    log.Fatal(err)
}
log.Infof("Marshaled: %v", marshaled)
```

###### Unmarshaling

```golang
import "github.com/romnnn/bsonbp/v2"

var myProto pb.Message
inputBson := bson.D{{Key: "Name", Value: "Test"}}
if err := bsonpb.Unmarshal(inputBson, &myProto); err != nil {
    log.Fatal(err)
}
log.Infof("Unmarshaled: %v", myProto)
```

If you want to try it, you can run the provided example with
```bash
bazel run //examples/v2:example
```

#### Usage (v1)

###### Marshaling

```golang
marshaler := bsonbp.Marshaler{}
myProto := &pb.Widget{RColor: []pb.Widget_Color{pb.Widget_RED}}
marshaledBson, err := marshaler.Marshal(myProto)
if err != nil {
    log.Fatalf("Failed to marshal with error: %s\n", err.Error())
}
log.Printf("Marshaled bson: %s\n", marshaledBson)
```

###### Unmarshaling

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

If you want to try it, you can run the provided example with
```bash
bazel run //examples/v1:example
```

#### Tests

```bash
bazel test //:go_default_test
bazel test //v1:go_default_test # v1 only
bazel test //v2:go_default_test # v2 only
```

#### Acknowledgements

- The v1 implementation was inspired by the official [github.com/golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) implementation.
- The v2 implementation was inspired by the official [google.golang.org/protobuf/encoding/protojson](https://github.com/protocolbuffers/protobuf-go/blob/master/encoding/protojson) implementation.
