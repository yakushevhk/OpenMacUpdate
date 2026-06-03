package scan

import (
	"os"
	"path/filepath"
	"runtime"

	"howett.net/plist"
)

type AppInfo struct {
	Name       string
	Path       string
	BundleID   string
	Version    string
	Build      string
	FeedURL    string
	FeedURLARM string
	Arch       string
}

func readPlist(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	_, err = plist.Unmarshal(data, &result)
	return result, err
}

func ScanDir(dir string) ([]AppInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var apps []AppInfo
	for _, e := range entries {
		if !e.IsDir() || filepath.Ext(e.Name()) != ".app" {
			continue
		}
		appPath := filepath.Join(dir, e.Name())
		infoPath := filepath.Join(appPath, "Contents", "Info.plist")

		data, err := readPlist(infoPath)
		if err != nil {
			continue
		}

		app := AppInfo{
			Name:    e.Name()[:len(e.Name())-4],
			Path:    appPath,
			FeedURL: plistString(data, "SUFeedURL"),
			Arch:    runtime.GOARCH,
		}

		if bid, ok := data["CFBundleIdentifier"]; ok {
			app.BundleID = toString(bid)
		}
		if v, ok := data["CFBundleShortVersionString"]; ok {
			app.Version = toString(v)
		}
		if b, ok := data["CFBundleVersion"]; ok {
			app.Build = toString(b)
		}

		apps = append(apps, app)
	}
	return apps, nil
}

func ScanDefault() ([]AppInfo, error) {
	var allApps []AppInfo

	dirs := []string{"/Applications"}

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}

	seen := make(map[string]bool)
	for _, dir := range dirs {
		apps, err := ScanDir(dir)
		if err != nil {
			continue
		}
		for _, app := range apps {
			if !seen[app.BundleID] {
				seen[app.BundleID] = true
				allApps = append(allApps, app)
			}
		}
	}

	return allApps, nil
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func plistString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		return toString(v)
	}
	return ""
}
