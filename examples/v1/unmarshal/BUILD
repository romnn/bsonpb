load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "unmarshal",
    srcs = [
        "unmarshal.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//:go_default_library",
        "//test_protos:test_objects_go_proto",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/primitive:go_default_library",
    ],
)
