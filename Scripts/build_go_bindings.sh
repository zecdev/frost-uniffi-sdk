#!/bin/bash

ls target 

cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.2.1+v0.25.0


if [[ "$OSTYPE" == "darwin"* ]]; then
ARCH=$(arch)

if [[ "$ARCH" == "arm64" ]]; then
uniffi-bindgen-go --library './target/aarch64-apple-darwin/debug/libfrost_uniffi_sdk.dylib' --out-dir .
else
uniffi-bindgen-go --library './target/x86_64-apple-darwin/debug/libfrost_uniffi_sdk.dylib' --out-dir .
fi

else
uniffi-bindgen-go --library './target/debug/libfrost_uniffi_sdk.dylib' --out-dir .
fi
