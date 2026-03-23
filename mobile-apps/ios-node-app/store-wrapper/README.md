# iOS App Store Wrapper Package

This wrapper package provides production release scaffolding for publishing the Sovereign Node iOS app to Apple App Store Connect.

## Included

- `build-release-ios.sh`: archive + export script for App Store IPA builds.
- `validate-store-assets.sh`: validates metadata/config files before packaging.
- `APP_STORE_CONNECT_SUBMISSION_CHECKLIST.md`: end-to-end App Store Connect submission checklist.
- `config/ExportOptions-AppStore.plist`: xcodebuild export options template.
- `config/SovereignNodeApp.entitlements`: hardened runtime and keychain baseline.
- `config/PrivacyInfo.xcprivacy`: privacy manifest starter.
- `config/AppStoreConnect.env.template`: upload/signing environment template.
- `metadata/en-US/*`: App Store listing text templates.

## Required Inputs

- Xcode 15+
- Apple Developer Team + Distribution cert/profile
- An `.xcodeproj` or `.xcworkspace` containing SovereignNodeApp target

## Quick Start

1. Copy environment template:
   - `cp config/AppStoreConnect.env.template .appstore.env`
2. Populate metadata files in `metadata/en-US/`.
3. Validate package files:
   - `bash store-wrapper/validate-store-assets.sh`
4. Archive + export:
   - `bash store-wrapper/build-release-ios.sh`

## Outputs

- Archive: `build/SovereignNodeApp.xcarchive`
- IPA: `build/export/SovereignNodeApp.ipa`

## Notes

- Keep real API keys and signing assets out of git.
- Submit first to TestFlight for device and crash validation.
- Follow `APP_STORE_CONNECT_SUBMISSION_CHECKLIST.md` before production submission.
