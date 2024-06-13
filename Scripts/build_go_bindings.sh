#!/bin/sh

cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.2.1+v0.25.0

uniffi-bindgen-go --library './target/debug/libfrost_uniffi_sdk.dylib' --out-dir .