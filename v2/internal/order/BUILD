load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "order.go",
        "range.go",
    ],
    importpath = "github.com/romnnn/bsonpb/v2/internal/order",
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
    ],
)