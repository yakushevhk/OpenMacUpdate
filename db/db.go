package db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	remoteURL   = "https://raw.githubusercontent.com/yakushevhk/macupdate/main/db/apps.json"
	cacheMaxAge = 24 * time.Hour
)

type AppEntry struct {
	FeedURL        string `json:"fe,omitempty"`
	CodeSignature  string `json:"cs,omitempty"`
	PurchaseURL    string `json:"pu,omitempty"`
	PackagePattern string `json:"pp,omitempty"`
	IsPaid         int    `json:"pl,omitempty"`
	IsIgnored      int    `json:"ig,omitempty"`
	IgnoreWhy      string `json:"ie,omitempty"`
	LookupURL      string `json:"lu,omitempty"`
	ReleaseNotes   string `json:"rn,omitempty"`
	InstallPath    string `json:"ip,omitempty"`
	OverrideBID    string `json:"or,omitempty"`
	ForceName      string `json:"fn,omitempty"`
	CaskName       string `json:"ca,omitempty"`
	ForcePremium   int    `json:"fp,omitempty"`
	FeedURLARM     string `json:"feas,omitempty"`
	MultiPackage   int    `json:"mp,omitempty"`
}

type Database struct {
	BundleIDMapping map[string]AppEntry `json:"bundleIDMapping"`
	IgnoredPrefixes []string            `json:"ignoredBundleIDPrefixes"`
	IgnoredSuffixes []string            `json:"ignoredBundleIDSuffixes"`
	IgnoredSubstr   []string            `json:"ignoredBundleIDSubstrings"`
	Source          string              `json:"-"`
}

func cacheDir() string {
	if d := os.Getenv("XDG_CACHE_HOME"); d != "" {
		return filepath.Join(d, "macupdate")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "macupdate")
}

func cachePath() string {
	return filepath.Join(cacheDir(), "apps.json")
}

func isCacheFresh() bool {
	info, err := os.Stat(cachePath())
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < cacheMaxAge
}

func fetchRemote() ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(remoteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func saveCache(data []byte) error {
	dir := cacheDir()
	os.MkdirAll(dir, 0o755)
	return os.WriteFile(cachePath(), data, 0o644)
}

func loadCache() ([]byte, error) {
	return os.ReadFile(cachePath())
}

func Load() (*Database, error) {
	var data []byte
	var source string

	if isCacheFresh() {
		if d, err := loadCache(); err == nil {
			data = d
			source = "cache"
		}
	}

	if data == nil {
		if d, err := fetchRemote(); err == nil {
			data = d
			source = "remote"
			saveCache(d)
		} else if d, err := loadCache(); err == nil {
			data = d
			source = "cache (offline)"
		} else {
			return nil, fmt.Errorf("cannot load database: network failed, no cache")
		}
	}

	db, err := parseDB(data)
	if err != nil {
		return nil, err
	}
	db.Source = source
	return db, nil
}

func parseDB(data []byte) (*Database, error) {
	var raw struct {
		Mapping map[string]json.RawMessage `json:"bundleIDMapping"`
		Prefix  []string                   `json:"ignoredBundleIDPrefixes"`
		Suffix  []string                   `json:"ignoredBundleIDSuffixes"`
		Substr  []string                   `json:"ignoredBundleIDSubstrings"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	db := &Database{
		BundleIDMapping: make(map[string]AppEntry, len(raw.Mapping)),
		IgnoredPrefixes: raw.Prefix,
		IgnoredSuffixes: raw.Suffix,
		IgnoredSubstr:   raw.Substr,
	}

	for bid, rawEntry := range raw.Mapping {
		var entry AppEntry
		json.Unmarshal(rawEntry, &entry)
		db.BundleIDMapping[bid] = entry
	}

	return db, nil
}

func (e *AppEntry) GetFeedURL(arch string) string {
	if arch == "arm64" && e.FeedURLARM != "" {
		return e.FeedURLARM
	}
	return e.FeedURL
}

func (db *Database) Lookup(bid string) (AppEntry, bool) {
	entry, ok := db.BundleIDMapping[bid]
	return entry, ok
}

func (db *Database) Search(q string) []struct {
	BID   string
	Entry AppEntry
} {
	var res []struct {
		BID   string
		Entry AppEntry
	}
	for bid, entry := range db.BundleIDMapping {
		if contains(bid, q) || contains(entry.ForceName, q) {
			res = append(res, struct {
				BID   string
				Entry AppEntry
			}{bid, entry})
		}
	}
	return res
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
