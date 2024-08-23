#!/bin/bash
set -euxo pipefail
ROOT_DIR=$(pwd)
SCRIPT_DIR="${SCRIPT_DIR:-$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )}"

BINARIES_DIR="$ROOT_DIR/target/debug"

BINDINGS_DIR="$ROOT_DIR/frost_go_ffi"

pushd $BINDINGS_DIR
LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$BINARIES_DIR" \
	CGO_LDFLAGS="-lfrost_uniffi_sdk -L$BINARIES_DIR -lm -ldl" \
	CGO_ENABLED=1 \
	go test -v $BINDINGS_DIR/frost_go_ffi_randomized_test.go $BINDINGS_DIR/frost_uniffi_sdk.go 
LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$BINARIES_DIR" \
	CGO_LDFLAGS="-lfrost_uniffi_sdk -L$BINARIES_DIR -lm -ldl" \
	CGO_ENABLED=1 \
	go test -v $BINDINGS_DIR/frost_go_ffi_orchard_keys_test.go $BINDINGS_DIR/frost_uniffi_sdk.go 