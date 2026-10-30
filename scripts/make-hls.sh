#!/bin/sh
# Generates the demo video and three ads as HLS (m3u8 + .ts) under hls/.
# Paths match what Margo and Edith hand out: /hls/demo/ and /hls/ads/{acme,globex,initech}/.
# Usage: FFMPEG=/path/to/ffmpeg scripts/make-hls.sh
set -e
cd "$(dirname "$0")/.."
FFMPEG="${FFMPEG:-ffmpeg}"

make() { # name source seconds segment-seconds tone-hz
  out="hls/$1"
  mkdir -p "$out"
  "$FFMPEG" -y -loglevel error \
    -f lavfi -i "$2=size=640x360:rate=30" \
    -f lavfi -i "sine=frequency=$5:sample_rate=44100" \
    -t "$3" -c:v libx264 -pix_fmt yuv420p -preset veryfast -g 60 -c:a aac -b:a 64k \
    -f hls -hls_time "$4" -hls_playlist_type vod \
    -hls_segment_filename "$out/seg%03d.ts" "$out/index.m3u8"
  echo "made $out"
}

make demo testsrc2 30 6 440
make ads/acme smptebars 6 2 660
make ads/globex rgbtestsrc 6 2 880
make ads/initech mandelbrot 6 2 550
