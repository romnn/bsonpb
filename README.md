
## bsonpb

This package allows to serialize/deserialize go `protobuf` messages into/from `bson` documents.

It was inspired by the official [golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) package.

```go
import "github.com/romnnn/bsonbp"
```

#### Usage
t.b.a

#### Tests
```bash
bazel test //:go_default_test
bazel test //:marshal_test  # Marshalling tests only
bazel test //:unmarshal_test  # Unmarshalling tests only
```

#### Acknowledgements
- Authors of the [golang/protobuf/jsonpb](https://github.com/golang/protobuf/tree/master/jsonpb) package that was used as a starting point.
