# Agnes

- Adaptive Gateway for Networked Edge Streaming
- Go, Kotlin, HLS, Media3, AWS, OpenTelemetry, Docker

## Running

- Start Kafka (`docker compose up -d` in `margo`), Margo, and Edith
- `go run .` - serves on :8080
- Only assets seen as APPROVED on the Kafka `asset-updates` topic can be played
- `GET /play/{assetId}?viewerId=` - asset plus an ad from Edith
- `POST /play/{assetId}/complete?viewerId=&adId=` - publishes the impression to the Kafka `ad-impressions` topic (Edith consumes it)

## Load test

`node scripts/load-impressions.mjs [count]` - completes 10,000 plays by default, each publishing an impression. Needs the whole stack running.

## HLS video

- `FFMPEG=/path/to/ffmpeg scripts/make-hls.sh` - generates the demo video and three ads under `hls/` (gitignored, needs ffmpeg)
- `/hls/{assetId}/...` is only served if Margo approved that asset (403 if rejected, 404 if unknown); `/hls/ads/...` is always served
- Agnes replays the whole `asset-updates` topic on every start (no consumer group), so it is ready a few seconds after launch

## Android app (Kotlin + Media3)

`android/` plays an ad and then the approved video from Agnes, and reports the impression when the ad finishes.

- Build (JDK 17-21, Android SDK with platform 36; set `sdk.dir` in `android/local.properties`): `cd android && ./gradlew assembleDebug`
- APK: `android/app/build/outputs/apk/debug/app-debug.apk`
- On a phone with USB debugging on, with the whole stack running:
  - `adb reverse tcp:8080 tcp:8080` so the phone reaches Agnes on localhost
  - `adb install -r app/build/outputs/apk/debug/app-debug.apk`
  - open the Agnes app
- Without USB, launch with your computer's address: `adb shell am start -n app.marlware.agnes/.MainActivity -e base http://<ip>:8080`
