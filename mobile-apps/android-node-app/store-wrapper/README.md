# Android Play Store Wrapper Package

This wrapper package provides production release scaffolding for publishing the Sovereign Node Android app to Google Play.

## Included

- `build-release-android.sh`: deterministic release build and optional signing helper.
- `validate-store-assets.sh`: validates expected metadata files before release.
- `config/keystore.properties.template`: signing configuration template.
- `config/gradle.properties.template`: release tuning template.
- `metadata/en-US/*`: Play listing text templates.

## Required Inputs

- Android SDK + Build Tools
- Java 17+
- `gradlew` wrapper in app project
- Play app signing key (`.jks`/`.keystore`)

## Quick Start

1. Copy templates:
   - `cp config/keystore.properties.template ../keystore.properties`
   - `cp config/gradle.properties.template ../gradle.properties.release`
2. Populate metadata files in `metadata/en-US/`.
3. Run validation:
   - `bash store-wrapper/validate-store-assets.sh`
4. Build release artifacts:
   - `bash store-wrapper/build-release-android.sh`

## Outputs

- Unsigned/signed AAB: `build/outputs/bundle/release/*.aab`
- Unsigned/signed APK: `build/outputs/apk/release/*.apk`

## Notes

- Keep keystore files and real passwords out of git.
- Prefer AAB upload for Play Store.
- Use Play Console internal track for smoke validation first.
