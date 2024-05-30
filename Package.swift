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
        .binaryTarget(name: "RustFramework", path: "./FrostSwiftFFI/RustFramework.xcframework"),
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
