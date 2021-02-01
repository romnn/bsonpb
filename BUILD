load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

# gazelle:prefix github.com/romnn/bsonpb
gazelle(name = "gazelle")

test_suite(
    name = "go_default_test",
    tests = [
        "//v2:go_default_test",
    ]
)
