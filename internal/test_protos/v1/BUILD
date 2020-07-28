load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "bson_proto",
    srcs = ["bson.proto"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_google_protobuf//:descriptor_proto",
    ],
)

proto_library(
    name = "test_objects_proto",
    srcs = [
        "test_objects.proto",
        "more_test_objects.proto",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":bson_proto",
        "@com_google_protobuf//:any_proto",
        "@com_google_protobuf//:duration_proto",
        "@com_google_protobuf//:struct_proto",
        "@com_google_protobuf//:timestamp_proto",
        "@com_google_protobuf//:wrappers_proto",
    ],
)

go_proto_library(
    name = "bson_go_proto",
    importpath = "github.com/romnnn/bsonpb/bson",
    proto = ":bson_proto",
    visibility = ["//visibility:public"],
    deps = [],
)

go_proto_library(
    name = "test_objects_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/test_protos/v1/test_objects",
    proto = ":test_objects_proto",
    visibility = ["//visibility:public"],
    deps = [
        ":bson_go_proto",
    ]
)