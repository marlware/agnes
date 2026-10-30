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
