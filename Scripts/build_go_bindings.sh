#!/bin/sh
set -euxo pipefail
cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.4.0+v0.28.3

uniffi-bindgen-go --library './target/debug/libfrost_uniffi_sdk.dylib' --out-dir .