package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	caskAPIURL = "https://formulae.brew.sh/api/cask.json"
	rawGitHub  = "https://raw.githubusercontent.com/Homebrew/homebrew-cask/HEAD/"
	dbPath     = "db/apps.json"
	workers    = 16
)

type Cask struct {
	Token       string          `json:"token"`
	Name        []string        `json:"name"`
	Version     string          `json:"version"`
	Homepage    string          `json:"homepage"`
	Artifacts   json.RawMessage `json:"artifacts"`
	RubySrcPath string          `json:"ruby_source_path"`
}

type LivecheckInfo struct {
	URL      string
	Strategy string
}

type AppEntry struct {
	FeedURL       string `json:"fe,omitempty"`
	CodeSignature string `json:"cs,omitempty"`
	PurchaseURL   string `json:"pu,omitempty"`
	IsIgnored     int    `json:"ig,omitempty"`
	OverrideBID   string `json:"or,omitempty"`
	ForceName     string `json:"fn,omitempty"`
	CaskName      string `json:"ca,omitempty"`
	FeedURLARM    string `json:"feas,omitempty"`
	LookupURL     string `json:"lu,omitempty"`
}

type Database struct {
	BundleIDMapping map[string]json.RawMessage `json:"bundleIDMapping"`
	IgnoredPrefixes []string                   `json:"ignoredBundleIDPrefixes"`
	IgnoredSuffixes []string                   `json:"ignoredBundleIDSuffixes"`
	IgnoredSubstr   []string                   `json:"ignoredBundleIDSubstrings"`
}

func main() {
	fmt.Println("=== MacUpdate Cask Sync ===")
	fmt.Println()

	// Load local database
	dbData, err := os.ReadFile(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading database: %v\n", err)
		os.Exit(1)
	}

	var db Database
	if err := json.Unmarshal(dbData, &db); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing database: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded database: %d entries\n", len(db.BundleIDMapping))

	// Fetch cask API
	fmt.Println("Fetching Homebrew Cask API...")
	casks, err := fetchCaskAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching cask API: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched %d casks\n", len(casks))

	// Extract bundle IDs from casks
	fmt.Println("Extracting bundle IDs from casks...")
	caskBundleIDs := make(map[string]*Cask) // bundleID -> cask
	for i := range casks {
		bids := extractBundleIDs(&casks[i])
		for _, bid := range bids {
			caskBundleIDs[bid] = &casks[i]
		}
	}
	fmt.Printf("Extracted bundle IDs from %d casks\n", len(caskBundleIDs))

	// Match and update
	fmt.Println("Matching against database...")
	stats := syncStats{}

	for bid, cask := range caskBundleIDs {
		rawEntry, exists := db.BundleIDMapping[bid]
		if !exists {
			continue
		}

		var entry AppEntry
		json.Unmarshal(rawEntry, &entry)

		changed := false

		// Add cask name if missing
		if entry.CaskName == "" {
			entry.CaskName = cask.Token
			changed = true
			stats.caskAdded++
		}

		if changed {
			updated, _ := json.Marshal(entry)
			db.BundleIDMapping[bid] = updated
		}
	}

	// Fetch livecheck URLs for matched casks in parallel
	fmt.Println("Fetching livecheck URLs from cask source files...")
	lcResults := fetchLivecheckURLs(caskBundleIDs)

	for bid, lc := range lcResults {
		rawEntry, exists := db.BundleIDMapping[bid]
		if !exists {
			continue
		}

		var entry AppEntry
		json.Unmarshal(rawEntry, &entry)

		changed := false

		if lc.Strategy == "sparkle" && entry.FeedURL == "" && lc.URL != "" {
			entry.FeedURL = lc.URL
			changed = true
			stats.feedAdded++
		} else if lc.URL != "" && entry.LookupURL == "" && entry.FeedURL == "" {
			entry.LookupURL = lc.URL
			changed = true
			stats.lookupAdded++
		}

		if changed {
			updated, _ := json.Marshal(entry)
			db.BundleIDMapping[bid] = updated
		}
	}

	// Save updated database
	output, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling database: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(dbPath, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("=== Sync Complete ===")
	fmt.Printf("Cask names added:  %d\n", stats.caskAdded)
	fmt.Printf("Sparkle feeds:     %d\n", stats.feedAdded)
	fmt.Printf("Lookup URLs:       %d\n", stats.lookupAdded)
	fmt.Printf("Total matched:     %d\n", stats.caskAdded+stats.feedAdded+stats.lookupAdded)
}

type syncStats struct {
	caskAdded   int
	feedAdded   int
	lookupAdded int
}

func fetchCaskAPI() ([]Cask, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(caskAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var casks []Cask
	if err := json.Unmarshal(data, &casks); err != nil {
		return nil, err
	}

	return casks, nil
}

func extractBundleIDs(cask *Cask) []string {
	var bids []string

	// Try to parse artifacts for uninstall -> quit
	var artifacts []struct {
		Uninstall json.RawMessage `json:"uninstall,omitempty"`
		Zap       json.RawMessage `json:"zap,omitempty"`
	}
	json.Unmarshal(cask.Artifacts, &artifacts)

	for _, a := range artifacts {
		if a.Uninstall != nil {
			bids = append(bids, extractQuitBIDs(a.Uninstall)...)
		}
		if a.Zap != nil {
			bids = append(bids, extractTrashBIDs(a.Zap)...)
		}
	}

	return unique(bids)
}

func extractQuitBIDs(data json.RawMessage) []string {
	var bids []string

	// Try as array
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		for _, item := range arr {
			if quit, ok := item["quit"]; ok {
				switch v := quit.(type) {
				case string:
					bids = append(bids, v)
				case []interface{}:
					for _, s := range v {
						if str, ok := s.(string); ok {
							bids = append(bids, str)
						}
					}
				}
			}
		}
		return bids
	}

	// Try as object
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		if quit, ok := obj["quit"]; ok {
			switch v := quit.(type) {
			case string:
				bids = append(bids, v)
			case []interface{}:
				for _, s := range v {
					if str, ok := s.(string); ok {
						bids = append(bids, str)
					}
				}
			}
		}
	}

	return bids
}

func extractTrashBIDs(data json.RawMessage) []string {
	var bids []string

	var items []struct {
		Trash interface{} `json:"trash"`
	}
	if err := json.Unmarshal(data, &items); err == nil {
		for _, item := range items {
			switch v := item.Trash.(type) {
			case string:
				bids = append(bids, extractBIDFromPath(v)...)
			case []interface{}:
				for _, s := range v {
					if str, ok := s.(string); ok {
						bids = append(bids, extractBIDFromPath(str)...)
					}
				}
			}
		}
	}

	return bids
}

var bidPattern = regexp.MustCompile(`com\.[a-zA-Z0-9._-]+|[a-zA-Z0-9]+(\.[a-zA-Z0-9_-]+){2,}`)

func extractBIDFromPath(path string) []string {
	var bids []string
	matches := bidPattern.FindAllString(path, -1)
	for _, m := range matches {
		if strings.Count(m, ".") >= 2 {
			bids = append(bids, m)
		}
	}
	return bids
}

func fetchLivecheckURLs(caskBIDs map[string]*Cask) map[string]*LivecheckInfo {
	results := make(map[string]*LivecheckInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Deduplicate by ruby source path
	pathToBIDs := make(map[string][]string)
	for bid, cask := range caskBIDs {
		pathToBIDs[cask.RubySrcPath] = append(pathToBIDs[cask.RubySrcPath], bid)
	}

	sem := make(chan struct{}, workers)

	for path, bids := range pathToBIDs {
		wg.Add(1)
		go func(p string, b []string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			lc, err := fetchLivecheck(p)
			if err != nil || lc == nil {
				return
			}

			mu.Lock()
			for _, bid := range b {
				results[bid] = lc
			}
			mu.Unlock()
		}(path, bids)
	}

	wg.Wait()
	return results
}

var (
	livecheckURLRe    = regexp.MustCompile(`livecheck\s+do\s*\n\s*url\s+"([^"]+)"`)
	sparkleStrategyRe = regexp.MustCompile(`strategy\s+:sparkle`)
)

func fetchLivecheck(rubyPath string) (*LivecheckInfo, error) {
	url := rawGitHub + rubyPath

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)

	// Find livecheck block
	lcMatch := livecheckURLRe.FindStringSubmatch(content)
	if lcMatch == nil {
		return nil, nil
	}

	lc := &LivecheckInfo{
		URL: lcMatch[1],
	}

	if sparkleStrategyRe.MatchString(content) {
		lc.Strategy = "sparkle"
	} else if strings.Contains(content, "strategy :header_match") {
		lc.Strategy = "header_match"
	} else if strings.Contains(content, "strategy :json") {
		lc.Strategy = "json"
	} else if strings.Contains(content, "strategy :electron_builder") {
		lc.Strategy = "electron_builder"
	} else if strings.Contains(content, "strategy :github_latest") {
		lc.Strategy = "github_latest"
	} else if strings.Contains(content, "strategy :github_releases") {
		lc.Strategy = "github_releases"
	} else {
		lc.Strategy = "other"
	}

	return lc, nil
}

func unique(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
