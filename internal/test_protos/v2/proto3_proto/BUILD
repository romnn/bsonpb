load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "test_proto",
    srcs = ["test.proto"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_google_protobuf//:any_proto",
        "//internal/test_protos/v2/proto2_proto:test_proto",
    ],
)
"""
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
"""

go_proto_library(
    name = "test_go_proto",
    importpath = "github.com/romnn/bsonpb/internal/testprotos/v2/proto3_proto",
    proto = ":test_proto",
    visibility = ["//visibility:public"],
    deps = [
        "//internal/test_protos/v2/proto2_proto:test_go_proto",
        "@com_github_golang_protobuf//ptypes/any:go_default_library",
    ],
)