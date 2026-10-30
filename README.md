# Agnes

Adaptive Gateway for Networked Edge Streaming

## What's going on here

Agnes is one of three small projects that together pretend to be a video streaming service, kind of like a baby version of what sits behind Netflix or YouTube. Everything is simulated, so the "videos" and "ads" are generated test patterns, but the way the pieces talk to each other is the real deal.

- **[Margo](https://github.com/marlware/margo)** checks new videos before anyone can watch them
- **Agnes** (this one) is the part viewers actually talk to, since it streams the videos and shows ads
- **[Edith](https://github.com/marlware/edith)** picks which ad to show and counts how many times each one was seen

## What a viewing looks like

Here is the whole story of one person pressing play:

- Margo has already checked the video and announced that it is approved, and Agnes heard that announcement and remembers it
- The viewer asks Agnes to play a video, and Agnes makes sure it was approved (rejected or unknown videos get turned away)
- Agnes asks Edith for an ad that suits this viewer
- The viewer's phone plays the ad first, then the video
- When the ad finishes, Agnes tells Edith that this person just saw that ad, so it gets counted

The videos are streamed in the HLS format, which chops a video into little pieces that get sent one at a time. It is the same trick real streaming services use so that playback can start fast and keep going smoothly.

## The Android app

There is a small Kotlin app in the `android` folder that does the viewer's part. It asks Agnes what to play, plays the ad and then the video using Google's Media3 player, and reports back when the ad is done. I built it with a Pixel 8 Pro in mind, but I haven't had it run on a real phone yet, so treat it as built but not battle-tested.

## Trying it out

You will need Go, Docker, ffmpeg (to make the test videos), and the other two projects running. Here is the order to do things in:

- Start the shared stack from the Margo repo with `docker compose up -d`
- Start Margo, then Edith, since their READMEs have the commands
- Make the test videos with `FFMPEG=/path/to/ffmpeg scripts/make-hls.sh` (or just `scripts/make-hls.sh` if ffmpeg is on your PATH), which creates a 30 second demo and three short ads in a `hls` folder
- Start Agnes with `go run .`, which listens on port 8080 and is ready a few seconds after launch
- Press play, sort of, with these requests:

```
curl "localhost:8080/play/demo?viewerId=alice"       # approved, you get the video plus an ad
curl "localhost:8080/play/broken?viewerId=alice"     # turned away, Margo rejected it
curl -X POST "localhost:8080/play/demo/complete?viewerId=alice&adId=ad-1"   # "alice saw the ad"
```

- Check `localhost:8082/impressions` afterwards, and Edith will have the note

## Running the Android app

You will need the Android SDK (platform 36) and a Java between 17 and 21. Then:

- Put the SDK location in `android/local.properties` as `sdk.dir=...`
- Build it with `cd android && ./gradlew assembleDebug`
- Plug in your phone with USB debugging on, with the whole stack running
- Run `adb reverse tcp:8080 tcp:8080` so the phone can reach Agnes on your computer
- Install with `adb install -r app/build/outputs/apk/debug/app-debug.apk`, then open Agnes on the phone
- If you would rather skip the USB cable, launch it with your computer's address instead: `adb shell am start -n app.marlware.agnes/.MainActivity -e base http://<your-ip>:8080`

## Stress test

Running `node scripts/load-impressions.mjs` fakes 10,000 people finishing an ad. You can watch Edith's counts climb at `localhost:8082/stats`.

## Good to know

- Agnes keeps what it knows about videos in memory, but it re-reads Margo's whole announcement history every time it starts, so a restart doesn't make it forget anything
- Ads are always playable, while videos are only playable if Margo approved them
- OpenTelemetry and AWS are on the wishlist but aren't built yet
- The endpoints, if you want them, are `GET /play/{assetId}`, `POST /play/{assetId}/complete`, and `/hls/...` for the video files themselves
