load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "weak1_proto",
    srcs = ["weak1.proto"],
    visibility = ["//visibility:public"],
    deps = [
    ],
)

proto_library(
    name = "weak2_proto",
    srcs = ["weak2.proto"],
    visibility = ["//visibility:public"],
    deps = [
    ],
)

proto_library(
    name = "test_import_proto",
    srcs = ["test_import.proto"],
    visibility = ["//visibility:public"],
    deps = [
    ],
)

proto_library(
    name = "test_public_proto",
    srcs = ["test_public.proto"],
    visibility = ["//visibility:public"],
    deps = [
    ],
)

proto_library(
    name = "test_proto",
    srcs = ["test.proto"],
    visibility = ["//visibility:public"],
    deps = [
        ":weak1_proto",
        ":weak2_proto",
        ":test_import_proto",
        ":test_public_proto",
    ],
)

go_proto_library(
    name = "weak1_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/weak_proto",
    proto = ":weak1_proto",
    visibility = ["//visibility:public"],
    deps = [],
)

go_proto_library(
    name = "test_go_proto",
    importpath = "github.com/romnnn/bsonpb/internal/testprotos/v2/test_proto",
    proto = ":test_proto",
    visibility = ["//visibility:public"],
    deps = [],
)