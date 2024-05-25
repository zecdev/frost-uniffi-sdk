// swift-tools-version:5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.
// Swift Package: FrostSwiftFFI

import PackageDescription;

let package = Package(
    name: "FrostSwiftFFI",
    platforms: [
        .iOS(.v13),
        .macOS(.v10_15)
    ],
    products: [
        .library(
            name: "FrostSwiftFFI",
            targets: ["FrostSwiftFFI"]
        )
    ],
    dependencies: [ ],
    targets: [
        .binaryTarget(name: "RustFramework", url: "https://github.com/pacu/frost-uniffi-sdk/releases/download/0.0.0/RustFramework.xcframework.zip", checksum: "249ecd6a27c69ff6b8ec408d036ad291516a791e8be5b3d3b069ec74e13f87dd"),
        .target(
            name: "FrostSwiftFFI",
            dependencies: [
                .target(name: "RustFramework")
            ], path: "FrostSwiftFFI/Sources"
        ),
        .testTarget(
            name: "NotRedPallasTests",
            dependencies: ["FrostSwiftFFI"],
            path: "FrostSwiftFFI/Tests"
        )
    ]
)
