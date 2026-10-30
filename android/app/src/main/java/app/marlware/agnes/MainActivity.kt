package app.marlware.agnes

import android.app.Activity
import android.os.Bundle
import android.widget.TextView
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread

private const val ASSET_ID = "demo"
private const val VIEWER_ID = "pixel8pro"

class MainActivity : Activity() {

    private var player: ExoPlayer? = null
    private lateinit var status: TextView
    private lateinit var base: String
    private var impressionSent = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        status = findViewById(R.id.status)

        // Reached through `adb reverse tcp:8080 tcp:8080`; override with: am start -e base http://<ip>:8080
        base = intent.getStringExtra("base") ?: "http://localhost:8080"
        say("Asking Agnes for $ASSET_ID...")

        thread {
            try {
                val play = JSONObject(request("GET", "$base/play/$ASSET_ID?viewerId=$VIEWER_ID"))
                val ad = play.getJSONObject("ad")
                val video = play.getJSONObject("asset")
                runOnUiThread { start(ad.getString("id"), ad.getString("hlsPath"), video.getString("hlsPath")) }
            } catch (e: Exception) {
                say("Could not start: ${e.message}")
            }
        }
    }

    // Plays the ad first, then the approved video. The ad finishing counts as an impression.
    private fun start(adId: String, adPath: String, videoPath: String) {
        val exo = ExoPlayer.Builder(this).build()
        player = exo
        findViewById<PlayerView>(R.id.player).player = exo

        exo.addListener(object : Player.Listener {
            override fun onMediaItemTransition(mediaItem: MediaItem?, reason: Int) {
                say(if (exo.currentMediaItemIndex == 0) "Playing ad $adId" else "Playing $ASSET_ID")
                if (reason == Player.MEDIA_ITEM_TRANSITION_REASON_AUTO && exo.currentMediaItemIndex == 1) {
                    reportImpression(adId)
                }
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
                say("Playback error: ${error.errorCodeName}")
            }
        })

        exo.setMediaItems(listOf(MediaItem.fromUri(base + adPath), MediaItem.fromUri(base + videoPath)))
        exo.prepare()
        exo.playWhenReady = true
    }

    private fun reportImpression(adId: String) {
        if (impressionSent) return
        impressionSent = true
        thread {
            try {
                request("POST", "$base/play/$ASSET_ID/complete?viewerId=$VIEWER_ID&adId=$adId")
                say("Playing $ASSET_ID (impression sent for $adId)")
            } catch (e: Exception) {
                say("Impression failed: ${e.message}")
            }
        }
    }

    private fun request(method: String, url: String): String {
        val conn = URL(url).openConnection() as HttpURLConnection
        conn.requestMethod = method
        conn.connectTimeout = 5000
        conn.readTimeout = 5000
        try {
            if (conn.responseCode >= 400) error("HTTP ${conn.responseCode}")
            return conn.inputStream.bufferedReader().readText()
        } finally {
            conn.disconnect()
        }
    }

    private fun say(text: String) = runOnUiThread { status.text = text }

    override fun onDestroy() {
        player?.release()
        super.onDestroy()
    }
}
