package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/segmentio/kafka-go"
)

var (
	edithURL = env("EDITH_URL", "http://localhost:8082")
	kafkaURL = env("KAFKA_URL", "localhost:9092")

	// Assets learned from Margo's asset-updates topic; only these can be served.
	assetsMu sync.RWMutex
	assets   = map[string]asset{}
)

func consumeAssets() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   "asset-updates",
		GroupID: "agnes",
		// Start from the beginning on first run so we see everything Margo published.
		StartOffset: kafka.FirstOffset,
	})
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("kafka read error:", err)
			continue
		}
		var a asset
		if err := json.Unmarshal(m.Value, &a); err != nil {
			log.Println("bad asset message:", err)
			continue
		}
		assetsMu.Lock()
		assets[a.ID] = a
		assetsMu.Unlock()
		log.Printf("asset %s -> %s", a.ID, a.Status)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type asset struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	HLSPath string   `json:"hlsPath"`
	Status  string   `json:"status"`
	Errors  []string `json:"errors"`
}

type ad struct {
	ID         string `json:"id"`
	Advertiser string `json:"advertiser"`
	HLSPath    string `json:"hlsPath"`
}

func getJSON(url string, out any) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
}

// GET /play/{assetId}?viewerId=x
// Only assets Margo approved (learned via Kafka) are served; an ad is fetched from Edith.
func play(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("assetId")
	viewer := r.URL.Query().Get("viewerId")
	if viewer == "" {
		viewer = "anonymous"
	}

	assetsMu.RLock()
	a, ok := assets[id]
	assetsMu.RUnlock()
	if !ok {
		http.Error(w, "unknown asset", http.StatusNotFound)
		return
	}
	if a.Status != "APPROVED" {
		http.Error(w, "asset not approved by margo", http.StatusForbidden)
		return
	}

	var ad ad
	if _, err := getJSON(fmt.Sprintf("%s/ads?viewerId=%s", edithURL, viewer), &ad); err != nil {
		http.Error(w, "edith unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"asset": a, "ad": ad})
}

// POST /play/{assetId}/complete?viewerId=x&adId=y
// Called by the player after the ad plays; forwards the impression to Edith.
func complete(w http.ResponseWriter, r *http.Request) {
	body, _ := json.Marshal(map[string]string{
		"adId":     r.URL.Query().Get("adId"),
		"viewerId": r.URL.Query().Get("viewerId"),
		"assetId":  r.PathValue("assetId"),
	})
	resp, err := http.Post(edithURL+"/impressions", "application/json", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "edith unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
}

func main() {
	go consumeAssets()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /play/{assetId}", play)
	mux.HandleFunc("POST /play/{assetId}/complete", complete)
	mux.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir("hls"))))

	log.Println("agnes listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
