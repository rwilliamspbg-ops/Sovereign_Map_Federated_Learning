// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "SovereignNodeApp",
    platforms: [
        .iOS(.v14)
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
