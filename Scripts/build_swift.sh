#!/bin/sh

cd frost-uniffi-sdk
cargo swift package --platforms macos ios --name FrostSwiftFFI --release
cd ..

# Rsync the FrostSwiftFFI file
rsync -avr --exclude='*.DS_Store' frost-uniffi-sdk/FrostSwiftFFI/ FrostSwiftFFI/

rm -rf frost-uniffi-sdk/FrostSwiftFFI/

# move the Package file to project root 
mv -f FrostSwiftFFI/Package.swift Package.swift