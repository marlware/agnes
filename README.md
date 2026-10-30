# Agnes

- Adaptive Gateway for Networked Edge Streaming
- Go, Kotlin, HLS, Media3, AWS, OpenTelemetry, Docker

## Running

- Start Kafka (`docker compose up -d` in `margo`), Margo, and Edith
- `go run .` - serves on :8080
- Only assets seen as APPROVED on the Kafka `asset-updates` topic can be played
- `GET /play/{assetId}?viewerId=` - asset plus an ad from Edith
- `POST /play/{assetId}/complete?viewerId=&adId=` - sends the impression to Edith
