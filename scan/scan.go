package scan

import (
	"os"
	"path/filepath"

	"howett.net/plist"
)

type AppInfo struct {
	Name       string
	Path       string
	BundleID   string
	Version    string
	Build      string
	FeedURL    string
}

func readPlist(path string) (map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data map[string]interface{}
	_, err = plist.Unmarshal(f, &data)
	return data, err
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
	return ScanDir("/Applications")
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
