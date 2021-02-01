load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "any_gen.go",
        "api_gen.go",
        "descriptor_gen.go",
        "doc.go",
        "duration_gen.go",
        "empty_gen.go",
        "field_mask_gen.go",
        "goname.go",
        "map_entry.go",
        "source_context_gen.go",
        "struct_gen.go",
        "timestamp_gen.go",
        "type_gen.go",
        "wrappers.go",
        "wrappers_gen.go",
    ],
    importpath = "github.com/romnn/bsonpb/v2/internal/genid",
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
    ],
)
