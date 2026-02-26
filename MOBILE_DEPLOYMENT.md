# Mobile App Deployment Guide

Complete guide for deploying the Sovereign Node app to iOS and Android app stores.

## 1. iOS App Store Deployment

### Prerequisites

- Apple Developer Account ($99/year)
- Xcode 14+
- Certificate & Provisioning Profiles
- TestFlight access

### Steps

#### 1.1 Create App Store Connect Entry

```bash
1. Visit: https://appstoreconnect.apple.com
2. Click "+ New App"
3. Select "iOS"
4. Fill in:
   - Name: "Sovereign Node"
   - Bundle ID: "io.sovereignmap.node"
   - SKU: "io.sovereignmap.node.1"
5. Click "Create"
```

#### 1.2 Configure App Information

```
Price: Free
Age Rating: 4+
Content Rights: Yours
Contact Email: your-email@sovereignmap.io
```

#### 1.3 Build Release Version

```bash
cd mobile-apps/ios-node-app

# Update version
# File: Info.plist
# CFBundleShortVersionString = "1.0.0"
# CFBundleVersion = "1"

# Build archive
xcodebuild \
    -workspace SovereignNodeApp.xcworkspace \
    -scheme SovereignNodeApp \
    -configuration Release \
    -archivePath /tmp/SovereignNode.xcarchive \
    archive

# Export for distribution
xcodebuild \
    -exportArchive \
    -archivePath /tmp/SovereignNode.xcarchive \
    -exportOptionsPlist ExportOptions.plist \
    -exportPath /tmp/export
```

#### 1.4 Submit to TestFlight

```bash
# Upload to App Store Connect
xcrun altool --upload-app \
    -f /tmp/export/SovereignNode.ipa \
    -t ios \
    -u your-apple-id@example.com \
    -p your-app-specific-password
```

#### 1.5 Promote to App Store

```
1. In App Store Connect, click "Prepare for Submission"
2. Add screenshots (5-10 images, 1242x2208 px)
3. Add app preview video (optional)
4. Add description (500 chars max)
5. Click "Submit for Review"
6. Wait 24-48 hours for approval
```

### iOS App Store Screenshots

**Home Screen (1242x2208)**
```
Title: "Join Network with One Tap"
Subtitle: "Participate in Federated Learning"
```

**Training Screen (1242x2208)**
```
Title: "Real-Time Metrics"
Subtitle: "Watch your node improve accuracy"
```

**Settings Screen (1242x2208)**
```
Title: "Customizable Configuration"
Subtitle: "Control training parameters"
```

---

## 2. Android Play Store Deployment

### Prerequisites

- Google Play Developer Account ($25 one-time)
- App signed with release key
- Google Play Console access

### Steps

#### 2.1 Create App Signing Key

```bash
keytool -genkey -v \
    -keystore sovereign-node-key.keystore \
    -keyalg RSA \
    -keysize 4096 \
    -validity 10000 \
    -alias sovereign-node

# Store this key securely!
# Back up: sovereign-node-key.keystore
```

#### 2.2 Configure Gradle Signing

```gradle
# android-node-app/build.gradle

signingConfigs {
    release {
        storeFile file("sovereign-node-key.keystore")
        storePassword = System.getenv("KEYSTORE_PASSWORD")
        keyAlias = "sovereign-node"
        keyPassword = System.getenv("KEY_PASSWORD")
    }
}

buildTypes {
    release {
        signingConfig signingConfigs.release
    }
}
```

#### 2.3 Build Release APK & AAB

```bash
cd mobile-apps/android-node-app

# Set environment variables
export KEYSTORE_PASSWORD="your-keystore-password"
export KEY_PASSWORD="your-key-password"

# Build App Bundle (for Play Store)
./gradlew bundleRelease

# Build APK (for direct distribution)
./gradlew assembleRelease

# Output:
# - app/build/outputs/bundle/release/app-release.aab
# - app/build/outputs/apk/release/app-release.apk
```

#### 2.4 Upload to Play Console

```bash
1. Visit: https://play.google.com/console
2. Click "Create App"
3. Enter name: "Sovereign Node"
4. Accept declaration
5. Click "Create"
```

#### 2.5 Fill In App Details

```
Title: "Sovereign Node"
Short description: "1-tap federated learning"
Full description: "Join decentralized federated learning..."
Category: "Social"
Content rating: PG
Privacy policy: [Your URL]
```

#### 2.6 Upload App Bundle

```
1. Left menu: "Release" → "Production"
2. Click "Create new release"
3. Upload AAB: app/build/outputs/bundle/release/app-release.aab
4. Add release notes:
   v1.0.0: "Initial launch - 1-tap join, real-time metrics"
5. Review and publish
```

#### 2.7 Add Screenshots

Upload 2-8 screenshots (1080x1920 px) for each:
- Phone 5.5" (required)
- 7" tablet (optional)
- 10" tablet (optional)

**Screenshot 1: Home Screen**
```
- Big "Join Network" button
- Status showing "Offline"
```

**Screenshot 2: Training**
```
- Round counter
- Accuracy graph
- Loss display
```

**Screenshot 3: Settings**
```
- Server URL configuration
- Byzantine mode toggle
- Node ID display
```

### Android Gradle Configuration

```gradle
// android-node-app/build.gradle

android {
    compileSdk 33
    
    defaultConfig {
        applicationId "io.sovereignmap.node"
        minSdk 26
        targetSdk 33
        versionCode 1
        versionName "1.0.0"
    }
    
    buildTypes {
        release {
            minifyEnabled true
            shrinkResources true
            proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'), 'proguard-rules.pro'
        }
    }
}
```

---

## 3. Distribution Methods

### Method 1: App Store / Play Store (Recommended for Production)

**Pros**:
- Automatic updates
- Rating & reviews
- Discoverability
- Secure

**Cons**:
- Review time (1-3 days)
- Store policies
- Commission (Google: 15%, Apple: 15%)

### Method 2: TestFlight / Internal Testing (Beta)

**Pros**:
- No commission
- Instant deployment
- Easier updates

**Cons**:
- Limited to 100 testers (TestFlight)
- No public listing

**iOS TestFlight:**
```
1. Build in Xcode
2. Upload to App Store Connect
3. Add beta testers (max 100)
4. Testers can install via TestFlight app
```

**Android Internal Testing:**
```
1. Build APK
2. Upload to Play Console (Internal Testing track)
3. Add internal testers
4. Get link to install
```

### Method 3: Direct APK (Android Only)

**Pros**:
- No approval needed
- Instant deployment

**Cons**:
- Manual updates
- Security concerns
- No ratings

```bash
# Generate signed APK
./gradlew assembleRelease

# Share: app/build/outputs/apk/release/app-release.apk
# Users tap to install
```

---

## 4. Version Management

### Semantic Versioning

```
1.0.0
├── Major (breaking changes)
├── Minor (new features)
└── Patch (bug fixes)
```

### Update Checklist

```
□ Update version in build files
□ Update CHANGELOG
□ Build and test
□ Tag release in Git
□ Build release artifacts
□ Create release notes
□ Submit to stores
□ Monitor ratings
□ Respond to feedback
```

---

## 5. Monitoring & Analytics

### Crash Reports

**iOS (Xcode Organizer)**:
```
Xcode > Organizer > Crashes
```

**Android (Play Console)**:
```
Play Console > Vitals > Crashes
```

### User Engagement

**iOS (App Store Connect)**:
- Downloads
- Crashes
- Ratings

**Android (Play Console)**:
- Installs
- Crashes
- Ratings & Reviews

### Custom Analytics (Optional)

```swift
// iOS - Firebase Analytics
import FirebaseAnalytics

Analytics.logEvent("node_joined", parameters: [
    "node_id": nodeID,
    "timestamp": Date()
])
```

```kotlin
// Android - Firebase Analytics
FirebaseAnalytics.getInstance(this).logEvent("node_joined", bundle)
```

---

## 6. Support & Updates

### Issue Handling

```
User Report → Verify → Fix → Test → Release → Monitor
```

### Update Strategy

**Critical Bug** → Emergency release (same day)
**Feature** → Quarterly release
**UI/UX** → Quarterly release

---

## Checklist for Launch

- [ ] App Store Connect entry created
- [ ] Privacy policy written
- [ ] Terms of service written
- [ ] Screenshots prepared (5-10)
- [ ] Video preview created (optional)
- [ ] App description finalized
- [ ] Support email set up
- [ ] App signed and tested
- [ ] Release notes written
- [ ] Store submission completed
- [ ] Monitoring set up
- [ ] Response team assigned

---

**Status**: Ready for submission  
**Target Timeline**: 2-4 weeks from submission to live  
**Estimated Users (Year 1)**: 1,000-10,000  
**Update Frequency**: Monthly security, Quarterly features

