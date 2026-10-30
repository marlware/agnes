package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	margoURL = env("MARGO_URL", "http://localhost:8081")
	edithURL = env("EDITH_URL", "http://localhost:8082")
)

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
// Only Margo-approved assets are served; an ad is fetched from Edith.
func play(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("assetId")
	viewer := r.URL.Query().Get("viewerId")
	if viewer == "" {
		viewer = "anonymous"
	}

	var a asset
	code, err := getJSON(fmt.Sprintf("%s/assets/%s", margoURL, id), &a)
	if err != nil {
		http.Error(w, "margo unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	if code == http.StatusNotFound {
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /play/{assetId}", play)
	mux.HandleFunc("POST /play/{assetId}/complete", complete)
	mux.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir("hls"))))

	log.Println("agnes listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
