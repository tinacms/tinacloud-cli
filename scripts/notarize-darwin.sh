#!/usr/bin/env bash
set -euo pipefail

for secret_name in APPLE_CERTIFICATE_BASE64 APPLE_CERTIFICATE_PASSWORD APPLE_ID APPLE_TEAM_ID APPLE_APP_SPECIFIC_PASSWORD; do
  if [[ -z "${!secret_name:-}" ]]; then
    echo "Missing required secret: $secret_name"
    exit 1
  fi
done

if [[ "$#" -eq 0 ]]; then
  echo "No binaries supplied for signing"
  exit 1
fi

certificate_path="$RUNNER_TEMP/developer-id.p12"
keychain_path="$RUNNER_TEMP/tinacloud-signing.keychain-db"
notarization_archive="$RUNNER_TEMP/tinacloud-notarization.zip"
notarization_result="$RUNNER_TEMP/tinacloud-notarization.json"
keychain_password="$(openssl rand -base64 32)"

cleanup() {
  security delete-keychain "$keychain_path" 2>/dev/null || true
  rm -f "$certificate_path" "$notarization_archive" "$notarization_result"
}
trap cleanup EXIT

printf '%s' "$APPLE_CERTIFICATE_BASE64" | base64 -D > "$certificate_path"
security create-keychain -p "$keychain_password" "$keychain_path"
security set-keychain-settings -lut 21600 "$keychain_path"
security unlock-keychain -p "$keychain_password" "$keychain_path"
security import "$certificate_path" -P "$APPLE_CERTIFICATE_PASSWORD" -A -t cert -f pkcs12 -k "$keychain_path"
security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$keychain_password" "$keychain_path"
security list-keychains -d user -s "$keychain_path"

signing_identity="$(security find-identity -v -p codesigning "$keychain_path" | awk '/Developer ID Application/ { print $2; exit }')"
if [[ -z "$signing_identity" ]]; then
  echo "No Developer ID Application identity found in certificate"
  exit 1
fi

for binary in "$@"; do
  codesign --force --options runtime --timestamp --identifier io.tina.tinacloud --sign "$signing_identity" "$binary"
  codesign --verify --strict --verbose=2 "$binary"
done

ditto -c -k --keepParent build/dist "$notarization_archive"
xcrun notarytool submit "$notarization_archive" \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$APPLE_APP_SPECIFIC_PASSWORD" \
  --wait \
  --output-format json | tee "$notarization_result"

notarization_status="$(plutil -extract status raw -o - "$notarization_result")"
if [[ "$notarization_status" != "Accepted" ]]; then
  submission_id="$(plutil -extract id raw -o - "$notarization_result")"
  xcrun notarytool log "$submission_id" \
    --apple-id "$APPLE_ID" \
    --team-id "$APPLE_TEAM_ID" \
    --password "$APPLE_APP_SPECIFIC_PASSWORD" || true
  exit 1
fi
