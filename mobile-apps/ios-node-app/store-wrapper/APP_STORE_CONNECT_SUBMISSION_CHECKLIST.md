# App Store Connect Submission Checklist

This checklist is tailored to the current Sovereign Node mobile feature set (hardware-backed signing, node telemetry, and network communication).

## 1) App Identity and Build

1. Confirm bundle identifier is final and consistent across signing profile and target.
2. Verify version and build numbers are incremented before each upload.
3. Archive Release build and export IPA using `build-release-ios.sh`.
4. Upload build to TestFlight and confirm processing completed.

## 2) Required Listing Metadata

1. App Name: Sovereign Node
2. Subtitle: Hardware-signed FL participant
3. Description: use `metadata/en-US/description.txt`
4. Keywords: use `metadata/en-US/keywords.txt`
5. Support URL: use `metadata/en-US/support-url.txt`
6. Release Notes: use `metadata/en-US/release-notes.txt`

## 3) Screenshot Matrix

Provide at minimum one full set for each required iPhone display class.

1. iPhone 6.7" (1290 x 2796 or 1284 x 2778)
2. iPhone 6.5" (1242 x 2688)
3. iPhone 5.5" (1242 x 2208)

Recommended shot sequence:

1. Node dashboard (trust + status)
2. Signed round in progress (status message visible)
3. Metrics panel (round/accuracy/loss)
4. Settings screen (server/training toggles)
5. Security posture callout (Secure Enclave hardware-backed signing)

## 4) Privacy Manifest and Nutrition Labels

Current app behavior basis:

1. No ad tracking logic present
2. No third-party ad SDKs present in current source
3. Network traffic used for node/aggregator communication
4. Device-generated node telemetry displayed in app UI

Suggested App Store privacy answers (confirm with legal/compliance):

1. Data Used to Track You: No
2. Data Linked to User: No, unless account system is added later
3. Data Not Linked to User: Diagnostics/Performance optional (if remote telemetry retained)
4. Precise Location: No
5. Contacts/Photos/Health/Financial data: No
6. Identifier for advertisers (IDFA): No

Use and review:

1. `config/PrivacyInfo.xcprivacy`
2. App Privacy section in App Store Connect before submission

## 5) Export Compliance and Encryption

Because TLS and cryptography are used:

1. Answer export compliance questions truthfully in App Store Connect.
2. Mark app as using standard encryption for transport/security features.
3. Provide any required exemption text if requested by your legal team.

## 6) Age Rating and Content Rights

1. Set age rating based on app operational content (typically low risk / 4+ unless policy changes).
2. Confirm all logos, names, and screenshots are owned or licensed.
3. Confirm support contact and policy URLs are live and public.

## 7) TestFlight Gate

1. Install TestFlight build on at least one modern iPhone and one older supported iPhone.
2. Validate join/leave flow and signed-round status.
3. Validate no startup crash when Secure Enclave fallback conditions occur.
4. Validate backend compatibility for signed payload envelope.

## 8) Submission Gate

1. No critical crashes in TestFlight.
2. Privacy answers finalized.
3. Screenshots and metadata complete.
4. Build assigned to submission and app review notes filled.

## 9) Review Notes Template

Use this in App Review Notes:

"Sovereign Node is a federated learning participant app for secure distributed model updates. The app performs on-device training and signs update payloads using hardware-backed keys (Secure Enclave when available). No user account or ad tracking is required for core functionality."
