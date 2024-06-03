#!/bin/sh

cd frost-uniffi-sdk
cargo swift package --platforms macos ios --name FrostSwiftFFI --release
cd ..

# Rsync the FrostSwiftFFI file
rsync -avr --exclude='*.DS_Store' frost-uniffi-sdk/FrostSwiftFFI/Sources FrostSwift/
cp -rf frost-uniffi-sdk/FrostSwiftFFI/RustFramework.xcframework FrostSwift/RustFramework.xcframework
rm -rf frost-uniffi-sdk/FrostSwiftFFI/

# Zip the xcframework
zip -r FrostSwift/RustFramework.xcframework.zip FrostSwift/RustFramework.xcframework

echo "CHECKSUM:"
shasum -a 256 FrostSwift/RustFramework.xcframework.zip

