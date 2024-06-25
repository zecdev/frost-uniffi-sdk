#!/bin/sh

cargo build --package frost-uniffi-sdk --no-default-features
cargo build --package uniffi-bindgen 
