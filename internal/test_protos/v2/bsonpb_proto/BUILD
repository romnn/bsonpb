load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "test_proto2_proto",
    srcs = ["test_proto2.proto"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_google_protobuf//:any_proto",
        "@com_google_protobuf//:duration_proto",
        "@com_google_protobuf//:struct_proto",
        "@com_google_protobuf//:timestamp_proto",
        "@com_google_protobuf//:wrappers_proto",
    ],
)

proto_library(
    name = "test_proto3_proto",
    srcs = ["test_proto3.proto"],
    visibility = ["//visibility:public"],
    deps = [],
)

go_proto_library(
    name = "test_proto2_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/bsonpb_proto2",
    proto = ":test_proto2_proto",
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_protobuf//types/known/anypb:go_default_library",
        "@org_golang_google_protobuf//types/known/durationpb:go_default_library",
        "@org_golang_google_protobuf//types/known/structpb:go_default_library",
        "@org_golang_google_protobuf//types/known/timestamppb:go_default_library",
        "@org_golang_google_protobuf//types/known/wrapperspb:go_default_library",
    ],
)

go_proto_library(
    name = "test_proto3_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/bsonpb_proto3",
    proto = ":test_proto3_proto",
    visibility = ["//visibility:public"],
    deps = [
        
    ],
)