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
        .target(
            name: "FrostSwift",
            dependencies: [
                .target(name: "FrostSwiftFFI")
            ], path: "FrostSwift/Sources/FrostSwift"
        ),
        .binaryTarget(name: "RustFramework", url: "https://github.com/pacu/frost-uniffi-sdk/releases/download/0.0.4/RustFramework.xcframework.zip", checksum: "aa4f942896c1cd069be859797e19b3891c73e85b6181a2d3b18aeec60cbafd42"),
        .target(
            name: "FrostSwiftFFI",
            dependencies: [
                .target(name: "RustFramework")
            ], path: "FrostSwift/Sources/FrostSwiftFFI"
        ),
        .testTarget(
            name: "NotRedPallasTests",
            dependencies: ["FrostSwiftFFI"],
            path: "FrostSwift/Tests/FrostSwiftFFI"
        ),
        .testTarget(
            name: "FrostTests",
            dependencies: ["FrostSwift"],
            path: "FrostSwift/Tests/FrostSwift"
        )
    ]
)
