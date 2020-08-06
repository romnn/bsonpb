load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "8663604808d2738dc615a2c3eb70eba54a9a982089dd09f6ffe5d0e75771bc4f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.6/rules_go-v0.23.6.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.6/rules_go-v0.23.6.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    # sha256 = "c5dc4cacbb303d5d0aa20c5cbb5cb88ef82ac61641c951cdf6b8e054184c5e22",
    strip_prefix = "protobuf-3.11.4",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.11.4.zip"],
)

"""
http_archive(
    name = "com_google_protobuf",
    sha256 = "0075c64cef80524b1d855df5f405845ded9b8d055022cc17b94e1589eb946b90",
    strip_prefix = "protobuf-4.0.0-rc2",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v4.0.0-rc2.zip"],
)
"""

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

# Override the dependencies here!
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

"""
go_repository(
    name = "org_golang_google_protobuf",
    importpath = "google.golang.org/protobuf",
    sum = "h1:Ejskq+SyPohKW+1uil0JJMtmHCgJPJ/qWTxr8qp+R4c=",
    version = "v1.25.0",
)
"""

"""
go_repository(
    name = "com_github_golang_protobuf",
    build_file_proto_mode = "disable_global",
    importpath = "github.com/golang/protobuf",
    patch_args = ["-p1"],
    patches = ["@io_bazel_rules_go//third_party:com_github_golang_protobuf-extras.patch"],
    sum = "h1:87PNWwrRvUSnqS4dlcBU/ftvOIBep4sYuBLlh6rX2wk=",
    version = "v1.3.4",
)
"""

go_repository(
    name = "org_golang_google_genproto",
    importpath = "google.golang.org/genproto",
    sum = "h1:wTk5DQB3+1darAz4Ldomo0r5bUOCKX7gilxQ4sb2kno=",
    version = "v0.0.0-20200731012542-8145dea6a485",
)

protobuf_deps()

go_rules_dependencies()

go_register_toolchains()

gazelle_dependencies()

go_repository(
    name = "com_github_romnnn_deepequal",
    importpath = "github.com/romnnn/deepequal",
    sum = "h1:UXnTzW6gXkguwf/N6M0lxrEKn2VOMclaQGTLBqPE9zI=",
    version = "v0.0.0-20200304130557-0992d7d478a0",
)

go_repository(
    name = "org_mongodb_go_mongo_driver",
    importpath = "go.mongodb.org/mongo-driver",
    sum = "h1:op56IfTQiaY2679w922KVWa3qcHdml2K/Io8ayAOUEQ=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
    version = "v1.8.0",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    sum = "h1:UBcNElsrwanuuMsnGSlYmtmgbb23qDR5dG+6X6Oo89I=",
    version = "v1.6.0",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    sum = "h1:JFrFEBb2xKufg6XkJsJr+WbKb4FQlURi5RUcBveYu9k=",
    version = "v0.5.1",
)

go_repository(
    name = "org_golang_google_protobuf",
    importpath = "google.golang.org/protobuf",
    sum = "h1:UhZDfRO8JRQru4/+LlLE0BRKGF8L+PICnvYZmx/fEGA=",
    version = "v1.24.0",
)

go_repository(
    name = "com_github_lunemec_as",
    importpath = "github.com/lunemec/as",
    sum = "h1:76CLvdcM2GTl7908l53dswTjxj777jul1l/YCwK4eX8=",
    version = "v1.0.0",
)
