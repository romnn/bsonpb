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

go_proto_library(
    name = "test_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/proto3_proto",
    proto = ":test_proto",
    visibility = ["//visibility:public"],
    deps = [
        "//internal/test_protos/v2/proto2_proto:test_go_proto",
        "@org_golang_google_protobuf//types/known/anypb:go_default_library",
    ],
)