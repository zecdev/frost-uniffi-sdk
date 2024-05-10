// swift-tools-version:5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.
// Swift Package: frost-swift-sdk

import PackageDescription;

let package = Package(
    name: "frost-swift-sdk",
    platforms: [
        .iOS(.v13),
        .macOS(.v10_15)
    ],
    products: [
        .library(
            name: "frost-swift-sdk",
            targets: ["frost-swift-sdk"]
        )
    ],
    dependencies: [ ],
    targets: [
        .binaryTarget(name: "RustFramework", path: "./RustFramework.xcframework"),
        .target(
            name: "frost-swift-sdk",
            dependencies: [
                .target(name: "RustFramework")
            ]
        ),
    ]
)