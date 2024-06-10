#!/bin/bash


if [[ "$OSTYPE" == "darwin"* ]]; then
ARCH=$(arch)

if [[ "$ARCH" == "arm64" ]]; then
cargo build --package frost-uniffi-sdk --package uniffi-bindgen --target aarch64-apple-darwin
else
cargo build --package frost-uniffi-sdk --package uniffi-bindgen --target x86_64-apple-darwin
fi

else
cargo build --package frost-uniffi-sdk --package uniffi-bindgen
fi
