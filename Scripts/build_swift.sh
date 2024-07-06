#!/bin/sh
set -euxo pipefail
# this builds randomized frost by default because cargo-swift 0.5 does not have the 
cd frost-uniffi-sdk
cargo install cargo-swift@=0.5 -f  
cargo swift package --platforms macos ios --name FrostSwiftFFI --release
cd ..

# Rsync the FrostSwiftFFI file
rsync -avr --exclude='*.DS_Store' frost-uniffi-sdk/FrostSwiftFFI/Sources FrostSwift/
cp -rf frost-uniffi-sdk/FrostSwiftFFI/RustFramework.xcframework FrostSwift/
rm -rf frost-uniffi-sdk/FrostSwiftFFI/

# Zip the xcframework
zip -r FrostSwift/RustFramework.xcframework.zip FrostSwift/RustFramework.xcframework

echo "CHECKSUM:"
shasum -a 256 FrostSwift/RustFramework.xcframework.zip

