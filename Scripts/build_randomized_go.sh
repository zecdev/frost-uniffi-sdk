#!/bin/sh
set -euxo pipefail
cargo build --package frost-uniffi-sdk --package uniffi-bindgen --features redpallas
