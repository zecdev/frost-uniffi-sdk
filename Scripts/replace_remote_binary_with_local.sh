#!/bin/sh
set -euxo pipefail
if [[ "$OSTYPE" == "darwin"* ]]; then
sed -i '' 's|^[[:space:]]*\.binaryTarget(name: "RustFramework", url: "[^"]*", checksum: "[^"]*")\,|        .binaryTarget(name: "RustFramework", path: "FrostSwift/RustFramework.xcframework.zip"),|' Package.swift

else
sed -i 's|^[[:space:]]*\.binaryTarget(name: "RustFramework", url: "[^"]*", checksum: "[^"]*")\,|        .binaryTarget(name: "RustFramework", path: "FrostSwift/RustFramework.xcframework.zip"),|' Package.swift

fi


