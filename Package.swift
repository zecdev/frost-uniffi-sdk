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
        .binaryTarget(name: "RustFramework", url: "https://github.com/pacu/frost-uniffi-sdk/releases/download/0.0.2/RustFramework.xcframework.zip", checksum: "19c88b713a302c188284c8bbc1b8214682b90402bc655874e3fd17b3e196b4a4"),
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
