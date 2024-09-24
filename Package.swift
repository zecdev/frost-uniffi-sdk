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
    dependencies: [
        .package(url: "https://github.com/pacu/frost-uniffi-sdk/releases/download/0.0.7/RustFramework.xcframework.zip", from: "1.8.3")
    ],
    targets: [
        .target(
            name: "FrostSwift",
            dependencies: [
                .target(name: "FrostSwiftFFI")
            ], path: "FrostSwift/Sources/FrostSwift"
        ),
        .binaryTarget(name: "RustFramework", url: "https://github.com/pacu/frost-uniffi-sdk/releases/download/0.0.7/RustFramework.xcframework.zip", checksum: "4a9d0051a83e5c363dea06107972b82c6c72a28e464a5f4e70d3edbcb419d228"),
        .target(
            name: "FrostSwiftFFI",
            dependencies: [
                .target(name: "RustFramework")
            ], path: "FrostSwift/Sources/FrostSwiftFFI"
        ),
        .testTarget(
            name: "NotRedPallasTests",
            dependencies: [
                "FrostSwiftFFI",
            ],
            path: "FrostSwift/Tests/FrostSwiftFFI"
        ),
        .testTarget(
            name: "FrostTests",
            dependencies: ["FrostSwift"],
            path: "FrostSwift/Tests/FrostSwift"
        ),
        .testTarget(
            name: "OrchardSwiftFFITests",
            dependencies: [
                "FrostSwift",
                "FrostSwiftFFI",
                "CryptoSwift"
            ],
            path: "FrostSwift/Tests/OrchardSwiftFFI"
        )
    ]
)
