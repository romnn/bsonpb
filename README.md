## bsonpb

[![Build Status](https://travis-ci.com/romnnn/bsonpb.svg?branch=master)](https://travis-ci.com/romnnn/bsonpb)
[![GitHub](https://img.shields.io/github/license/romnnn/bsonpb)](https://github.com/romnnn/bsonpb)
[![GoDoc](https://godoc.org/github.com/romnnn/bsonpb?status.svg)](https://godoc.org/github.com/romnnn/bsonpb)
[![Test Coverage](https://codecov.io/gh/romnnn/bsonpb/branch/master/graph/badge.svg)](https://codecov.io/gh/romnnn/bsonpb)
[![Release](https://img.shields.io/github/v/release/romnnn/bsonpb)](https://github.com/romnnn/bsonpb/releases/latest)

This package allows to serialize/deserialize golang `protobuf` messages into/from `bson` documents.

**Important notes**: 
- This implementation has transitioned from the old [github.com/golang/protobuf](https://github.com/golang/protobuf) to the new [google.golang.org/protobuf](https://github.com/protocolbuffers/protobuf-go) API. The v1 implementation had various bazel related conflicts with the protobug dependency and is now abandoned under the `v1` branch.

```go
import "github.com/romnnn/bsonbp/v2" // only works with google.golang.org/protobuf, NOT github.com/golang/protobuf
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

#### Tests

```bash
bazel test //:go_default_test
bazel test //v2:go_default_test # v2 only
```

#### Acknowledgements

- The v1 implementation was inspired by the official [github.com/golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) implementation.
- The v2 implementation was inspired by the official [google.golang.org/protobuf/encoding/protojson](https://github.com/protocolbuffers/protobuf-go/blob/master/encoding/protojson) implementation.
