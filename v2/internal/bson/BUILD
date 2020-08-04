load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "encode.go",
        "decode.go",
        "decode_number.go",
        "decode_string.go",
        "decode_token.go",
    ],
    importpath = "github.com/romnnn/bsonpb/v2/internal/bson",
    visibility = ["//visibility:public"],
    deps = [
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
    ],
)