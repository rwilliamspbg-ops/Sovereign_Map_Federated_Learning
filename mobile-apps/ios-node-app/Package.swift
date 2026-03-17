// swift-tools-version:6.0
import PackageDescription

let package = Package(
    name: "SovereignNodeApp",
    platforms: [
        .iOS(.v16)
    ],
    products: [
        .library(name: "SovereignNodeApp", targets: ["SovereignNodeApp"])
    ],
    dependencies: [],
    targets: [
        .target(
            name: "SovereignNodeApp",
            dependencies: [],
            path: ".",
            sources: ["ContentView.swift", "SovereignNodeApp.swift"]
        )
    ]
)
