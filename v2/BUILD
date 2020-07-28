load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "copied.go",
        "well_known_types.go",
        "decode.go",
        "encode.go",
    ],
    importpath = "github.com/romnnn/bsonpb",
    visibility = ["//visibility:public"],
    deps = [
        "//internal/test_protos/v2/bsonpb_proto:test_proto2_go_proto",
        "//internal/test_protos/v2/bsonpb_proto:test_proto3_go_proto",
        "//internal/test_protos/v2/proto3_proto:test_go_proto",
        "//v2/internal/order:go_default_library",
        "//v2/internal/json:go_default_library",
        "//v2/internal/genid:go_default_library",
        # Old
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_golang_protobuf//ptypes/wrappers:go_default_library",
        "@com_github_golang_protobuf//protoc-gen-go/descriptor:go_default_library",
        "@com_github_golang_protobuf//ptypes/any:go_default_library",
        "@com_github_golang_protobuf//ptypes/duration:go_default_library",
        "@com_github_golang_protobuf//ptypes/struct:go_default_library",
        "@com_github_golang_protobuf//ptypes/timestamp:go_default_library",
        # New
        "@org_golang_google_protobuf//encoding/protojson:go_default_library",
        "@org_golang_google_protobuf//proto:go_default_library",
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//reflect/protoregistry:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        # Internal
        # "@org_golang_google_protobuf//internal/encoding/json:go_default_library",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/bsonrw:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/primitive:go_default_library",
    ],
)

go_test(
    name = "marshal_test",
    srcs = [
        "bson_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "unmarshal_test",
    srcs = [
        "bson_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bson_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
