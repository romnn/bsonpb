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
    name = "proto2_proto",
    srcs = ["proto2_proto.proto"],
    visibility = ["//visibility:public"],
    deps = [],
)

proto_library(
    name = "proto3_proto",
    srcs = ["proto3_proto.proto"],
    visibility = ["//visibility:public"],
    deps = [
        ":proto2_proto",
        "@com_google_protobuf//:any_proto",
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
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v1/test_objects",
    proto = ":test_objects_proto",
    visibility = ["//visibility:public"],
    deps = [
        ":bson_go_proto",
    ]
)

go_proto_library(
    name = "proto2_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v1/proto2_proto",
    proto = ":proto2_proto",
    visibility = ["//visibility:public"],
    deps = [
        ":bson_go_proto",
    ]
)

go_proto_library(
    name = "proto3_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v1/proto3_proto",
    proto = ":proto3_proto",
    visibility = ["//visibility:public"],
    deps = [
        ":proto2_go_proto",
        "@com_github_golang_protobuf//ptypes/any:go_default_library",
    ]
)