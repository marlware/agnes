// Completes N plays through Agnes; each one publishes an impression to Kafka.
// Usage: node scripts/load-impressions.mjs [count] [agnesUrl]
const count = Number(process.argv[2] ?? 10000);
const base = process.argv[3] ?? "http://localhost:8080";
const concurrency = 50;
const viewers = ["alice", "bob", "carol", "dave"];

let next = 0;
let ok = 0;
let failed = 0;

async function worker() {
  while (next < count) {
    const i = next++;
    const viewer = `${viewers[i % viewers.length]}${i % 250}`;
    const adId = `ad-${(i % 3) + 1}`;
    try {
      const res = await fetch(
        `${base}/play/demo/complete?viewerId=${viewer}&adId=${adId}`,
        { method: "POST" },
      );
      if (res.status === 202) ok++;
      else failed++;
    } catch {
      failed++;
    }
  }
}

const start = Date.now();
await Promise.all(Array.from({ length: concurrency }, worker));
const secs = (Date.now() - start) / 1000;
console.log(`sent ${count} in ${secs.toFixed(1)}s (${Math.round(count / secs)}/s)`);
console.log(`accepted=${ok} failed=${failed}`);
