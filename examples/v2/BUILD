load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "example",
    srcs = [
        "example.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//v2:go_default_library",
        "//internal/test_protos/v2/proto3_proto:test_go_proto",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@com_github_sirupsen_logrus//:go_default_library",
        "@org_golang_google_protobuf//types/known/timestamppb:go_default_library",
        "@org_golang_google_protobuf//types/known/wrapperspb:go_default_library",
    ],
)
