load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "test_proto",
    srcs = ["test.proto"],
    visibility = ["//visibility:public"],
    deps = [
        
    ],
)

go_proto_library(
    name = "test_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/proto2_proto",
    proto = ":test_proto",
    visibility = ["//visibility:public"],
    deps = [],
)