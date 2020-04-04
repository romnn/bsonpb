load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

# gazelle:prefix github.com/romnnn/bsonpb
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = [
        "bsonpb_common.go",
        "bsonpb_marshal.go",
        "bsonpb_unmarshal.go",
    ],
    importpath = "github.com/romnnn/bsonpb",
    visibility = ["//visibility:public"],
    deps = [
        "//test_protos:test_objects_go_proto",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_golang_protobuf//proto/proto3_proto:go_default_library",
        "@com_github_golang_protobuf//ptypes:go_default_library",
        "@com_github_golang_protobuf//ptypes/any:go_default_library",
        "@com_github_golang_protobuf//ptypes/duration:go_default_library",
        "@com_github_golang_protobuf//ptypes/struct:go_default_library",
        "@com_github_golang_protobuf//ptypes/timestamp:go_default_library",
        "@com_github_golang_protobuf//ptypes/wrappers:go_default_library",
        "@com_github_romnnn_deepequal//:go_default_library",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/bsonrw:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/primitive:go_default_library",
    ],
)

go_test(
    name = "marshal_test",
    srcs = [
        "bsonpb_marshal_test.go",
        "bsonpb_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "unmarshal_test",
    srcs = [
        "bsonpb_test.go",
        "bsonpb_unmarshal_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bsonpb_marshal_test.go",
        "bsonpb_test.go",
        "bsonpb_unmarshal_test.go",
    ],
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
