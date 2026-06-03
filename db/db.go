package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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
}

func dbPath() string {
	if f := os.Getenv("MACBEE_DB"); f != "" {
		return f
	}
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "apps.json")
}

func Load() (*Database, error) {
	path := dbPath()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var raw struct {
		Mapping map[string]json.RawMessage `json:"bundleIDMapping"`
		Prefix  []string                   `json:"ignoredBundleIDPrefixes"`
		Suffix  []string                   `json:"ignoredBundleIDSuffixes"`
		Substr  []string                   `json:"ignoredBundleIDSubstrings"`
	}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
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

func (db *Database) Lookup(bid string) (AppEntry, bool) {
	entry, ok := db.BundleIDMapping[bid]
	return entry, ok
}

func (db *Database) Search(q string) []struct{ BID string; Entry AppEntry } {
	var res []struct{ BID string; Entry AppEntry }
	for bid, entry := range db.BundleIDMapping {
		if contains(bid, q) || contains(entry.ForceName, q) {
			res = append(res, struct{ BID string; Entry AppEntry }{bid, entry})
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
