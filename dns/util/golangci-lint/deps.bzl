load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

_DOWNLOAD_URI = (
    "https://github.com/golangci/golangci-lint/releases/download/v{version}/" +
    "golangci-lint-{version}-{arch}.{archive}"
)
_PREFIX = (
    "golangci-lint-{version}-{arch}"
)

_ARCHIVE_TYPE = ["zip", "tar.gz"]

_VERSION = "1.50.1"
_CHECKSUMS = {
    "windows-386": "6d11fb6ed91ba3aecbf2ea8e1a95dce16cf0449d54aa77c607ac4e75cc43213a",
    "windows-amd64": "8c2da214884db02fb7f3d929672c515ae3b9d10defad4dd661c4ab365a316d68",
    "linux-amd64": "4ba1dc9dbdf05b7bdc6f0e04bdfe6f63aa70576f51817be1b2540bbce017b69a",
    "linux-386": "8965c4327e742444bd4359bd952d42ebc645eba929d3826dd12adb8a14cc6248",
    "darwin-amd64": "0f615fb8c364f6e4a213f2ed2ff7aa1fc2b208addf29511e89c03534067bbf57",
}

def _golangcilint_download_impl(ctx):
    if ctx.os.name == "linux":
        arch = "linux-amd64"
    elif ctx.os.name == "mac os x":
        arch = "darwin-amd64"
    else:
        fail("Unsupported operating system: {}".format(ctx.os.name))

    if arch not in _CHECKSUMS:
        fail("Unsupported arch {}".format(arch))

    if arch.startswith("windows"):
        archive = _ARCHIVE_TYPE[0]
    else:
        archive = _ARCHIVE_TYPE[1]

    url = _DOWNLOAD_URI.format(version = _VERSION, arch = arch, archive = archive)
    prefix = _PREFIX.format(version = _VERSION, arch = arch)
    sha256 = _CHECKSUMS[arch]

    ctx.template(
        "BUILD.bazel",
        Label("@com_github_atlassian_bazel_tools//golangcilint:golangcilint.build.bazel"),
        executable = False,
    )
    ctx.download_and_extract(
        stripPrefix = prefix,
        url = url,
        sha256 = sha256,
    )

_golangcilint_download = repository_rule(
    implementation = _golangcilint_download_impl,
)

def golangcilint_dependencies():
    maybe(
        http_archive,
        name = "bazel_skylib",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        ],
        sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
    )
    _golangcilint_download(
        name = "com_github_atlassian_bazel_tools_golangcilint",
    )