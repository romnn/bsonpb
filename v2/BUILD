load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "copied.go",
        "well_known_types.go",
        "decode.go",
        "encode.go",
    ],
    importpath = "github.com/romnn/bsonpb/v2",
    visibility = ["//visibility:public"],
    deps = [
        "//v2/internal/genid:go_default_library",
        "@org_golang_google_protobuf//proto:go_default_library",
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//reflect/protoregistry:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        "@org_golang_google_protobuf//types/known/timestamppb:go_default_library",
        "@org_golang_google_protobuf//types/known/durationpb:go_default_library",
        "@org_golang_google_protobuf//types/known/anypb:go_default_library",
        "@org_golang_google_protobuf//types/known/structpb:go_default_library",
        "@org_golang_google_protobuf//types/known/wrapperspb:go_default_library",
        "@org_golang_google_protobuf//types/known/emptypb:go_default_library",
        "@com_github_lunemec_as//:go_default_library",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/bsonrw:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/primitive:go_default_library",
    ],
)

TEST_DEPS = [
    "@com_github_google_go_cmp//cmp:go_default_library",
    "@org_golang_google_protobuf//testing/protopack:go_default_library",
    "@com_github_romnn_deepequal//:go_default_library",
    "//internal/test_protos/v2/textpb2_proto:test_go_proto",
    "//internal/test_protos/v2/textpb3_proto:test_go_proto",
]

go_test(
    name = "encode",
    srcs = [
        "encode_test.go",
    ],
    embed = [":go_default_library"],
    deps = TEST_DEPS,
    visibility = ["//visibility:public"],
)

go_test(
    name = "decode",
    srcs = [
        "decode_test.go",
    ],
    embed = [":go_default_library"],
    deps = TEST_DEPS,
    visibility = ["//visibility:public"],
)

test_suite(
    name = "go_default_test",
    tests = [
        ":encode",
        ":decode",
    ],
    tags = [],
)
