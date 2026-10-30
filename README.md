# Agnes
- HLS playback service
- full name: Adaptive Gateway for Networked Edge Streaming

<img src="/agnes.gif" width="200" alt="Agnes GIF">

## What's going on here

Agnes is one of three small projects. Together, they pretend to be a video streaming service. Everything is simulated, so the "videos" and "ads" are generated test patterns. But the pieces really do talk to each other.

- **[Margo](https://github.com/marlware/margo)** checks new videos before anyone can watch them
- **[Edith](https://github.com/marlware/edith)** picks which ad to show and counts how many times each ad was seen
- **Agnes** (this one) streams the videos to viewers and plays the ads

## What a viewing looks like

Here's the whole story of one person pressing play.

- Margo already checked the video and announced that it's approved
- Agnes heard that announcement and remembers it
- The viewer asks Agnes to play a video
- Agnes makes sure the video was approved, and turns away rejected or unknown ones
- Agnes asks Edith for an ad that suits this viewer
- The viewer's phone plays the ad first, then the video
- When the ad finishes, Agnes tells Edith that this person just saw it
- Edith counts the view

The videos are streamed in a format called HLS. It chops a video into little pieces and sends them one at a time. Real streaming services use the same trick. It lets playback start fast and keep going smoothly.

## The Android app

There's a small Kotlin app in the `android` folder. It does the viewer's part.

- Asks Agnes what to play
- Plays the ad, then the video, using Google's Media3 player
- Reports back when the ad is done

I built it with a Pixel 8 Pro in mind. But I haven't run it on a real phone yet. So treat it as built, but not battle-tested.

## Why these technologies

Every tool here earns its spot. Here's what each one does for Agnes.

- **Go** makes small, fast programs with a web server built in, and it handles lots of viewers at once without much effort
- **HLS** chops a video into little pieces, so playback starts fast and keeps going smoothly
- **ffmpeg** makes the test videos, so you don't need any real footage
- **Kafka** lets Agnes hear about approved videos and report finished ads without being tied to Margo or Edith
- **Kotlin** is the main language for Android apps
- **Media3** is Google's player library, and it handles HLS playback for you
- **Node** runs the stress-test script without any extra setup
- **Docker** runs Kafka for you, so there's nothing to install by hand

## Trying it out

Here's what you'll need.

- Go
- Docker
- ffmpeg, to make the test videos
- The other two projects running

Then do things in this order.

- Start the shared stack from the Margo repo with `docker compose up -d`
- Start Margo, then Edith (their READMEs have the commands)
- Make the test videos with `FFMPEG=/path/to/ffmpeg scripts/make-hls.sh`
- Skip the `FFMPEG=` part if ffmpeg is already on your PATH
- That creates a 30 second demo and three short ads in a `hls` folder
- Start Agnes with `go run .` (it listens on port 8080 and is ready a few seconds after launch)
- Press play, sort of, with the requests below

```
curl "localhost:8080/play/demo?viewerId=alice"       # approved, you get the video plus an ad
curl "localhost:8080/play/broken?viewerId=alice"     # turned away, Margo rejected it
curl -X POST "localhost:8080/play/demo/complete?viewerId=alice&adId=ad-1"   # "alice saw the ad"
```

- Check `localhost:8082/impressions` afterwards to see Edith's note

## Running the Android app

Here's what you'll need.

- The Android SDK (platform 36)
- A Java between 17 and 21

Then do these steps.

- Put the SDK location in `android/local.properties` as `sdk.dir=...`
- Build with `cd android && ./gradlew assembleDebug`
- Plug in your phone with USB debugging on
- Make sure the whole stack is running
- Run `adb reverse tcp:8080 tcp:8080` so the phone can reach Agnes on your computer
- Install with `adb install -r app/build/outputs/apk/debug/app-debug.apk`
- Open Agnes on the phone
- No USB cable? Launch with your computer's address instead: `adb shell am start -n app.marlware.agnes/.MainActivity -e base http://<your-ip>:8080`

## Stress test

Running `node scripts/load-impressions.mjs` fakes 10,000 people finishing an ad. Watch Edith's counts climb at `localhost:8082/stats`.

## Good to know

- Agnes keeps what it knows about videos in memory
- But it re-reads Margo's whole announcement history on every start
- So a restart doesn't make it forget anything
- Ads are always playable
- Videos only play if Margo approved them
- OpenTelemetry and AWS are on the wishlist, but they aren't built yet
- The endpoints are `GET /play/{assetId}`, `POST /play/{assetId}/complete`, and `/hls/...` for the video files
