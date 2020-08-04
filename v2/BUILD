load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

# gazelle:go_proto_compilers //:go_proto_apiv2

go_library(
    name = "go_default_library",
    srcs = [
        "protov2copy.go",
        "copied.go",
        "well_known_types.go",
        "decode.go",
        "encode.go",
    ],
    importpath = "github.com/romnnn/bsonpb/v2",
    visibility = ["//visibility:public"],
    deps = [
        "//internal/test_protos/v2/bsonpb_proto:test_proto2_go_proto",
        "//internal/test_protos/v2/bsonpb_proto:test_proto3_go_proto",
        "//internal/test_protos/v2/proto3_proto:test_go_proto",
        "//v2/internal/order:go_default_library",
        "//v2/internal/json:go_default_library",
        "//v2/internal/genid:go_default_library",
        # Old
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_golang_protobuf//ptypes/wrappers:go_default_library",
        "@com_github_golang_protobuf//protoc-gen-go/descriptor:go_default_library",
        "@com_github_golang_protobuf//ptypes/any:go_default_library",
        "@com_github_golang_protobuf//ptypes/duration:go_default_library",
        "@com_github_golang_protobuf//ptypes/struct:go_default_library",
        "@com_github_golang_protobuf//ptypes/timestamp:go_default_library",
        # New
        "@org_golang_google_protobuf//encoding/protojson:go_default_library",
        "@org_golang_google_protobuf//proto:go_default_library",
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//reflect/protoregistry:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        # Well known types
        "@org_golang_google_protobuf//types/known/timestamppb:go_default_library",
        "@org_golang_google_protobuf//types/known/durationpb:go_default_library",
        "@org_golang_google_protobuf//types/known/anypb:go_default_library",
        "@org_golang_google_protobuf//types/known/structpb:go_default_library",
        "@org_golang_google_protobuf//types/known/wrapperspb:go_default_library",
        "@org_golang_google_protobuf//types/known/emptypb:go_default_library",
        # "@org_golang_google_protobuf//types/known/fieldmaskpb:go_default_library",
        # Internal
        # "@org_golang_google_protobuf//internal/encoding/json:go_default_library",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/bsonrw:go_default_library",
        "@org_mongodb_go_mongo_driver//bson/primitive:go_default_library",
    ],
)

go_test(
    name = "encode",
    srcs = [
        "encode_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "@com_github_google_go_cmp//cmp:go_default_library",
        "@org_golang_google_protobuf//testing/protopack:go_default_library",
        "@com_github_romnnn_deepequal//:go_default_library",
        "//internal/test_protos/v2/textpb2_proto:test_go_proto",
        "//internal/test_protos/v2/textpb3_proto:test_go_proto",
    ],
    visibility = ["//visibility:public"],
)