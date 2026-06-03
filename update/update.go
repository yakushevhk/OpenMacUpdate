package update

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Appcast struct {
	XMLName xml.Name       `xml:"rss"`
	Channel AppcastChannel `xml:"channel"`
}

type AppcastChannel struct {
	Items []AppcastItem `xml:"item"`
}

type AppcastItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	Enclosure   *struct {
		URL    string `xml:"url,attr"`
		Version string `xml:"version,attr"`
		ShortVersionString string `xml:"shortVersionString,attr"`
	} `xml:"enclosure"`
	ReleaseNotesLink string `xml:"http://www.andymatuschak.org/xml-namespaces/sparkle releaseNotesLink"`
}

type UpdateInfo struct {
	LatestVersion string
	DownloadURL   string
	ReleaseNotes  string
	PubDate       time.Time
}

func FetchFeed(url string) (*UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	var appcast Appcast
	if err := xml.NewDecoder(resp.Body).Decode(&appcast); err != nil {
		return nil, fmt.Errorf("parse %s: %w", url, err)
	}

	if len(appcast.Channel.Items) == 0 {
		return nil, nil
	}

	latest := appcast.Channel.Items[0]
	info := &UpdateInfo{
		LatestVersion: latest.Enclosure.ShortVersionString,
		DownloadURL:   latest.Enclosure.URL,
		ReleaseNotes:  latest.ReleaseNotesLink,
	}

	if info.LatestVersion == "" {
		info.LatestVersion = latest.Enclosure.Version
	}

	if info.LatestVersion == "" {
		info.LatestVersion = latest.Title
	}

	if t, err := time.Parse(time.RFC1123, latest.PubDate); err == nil {
		info.PubDate = t
	}

	return info, nil
}

func CompareVersions(installed, latest string) int {
	a, b := normalize(installed), normalize(latest)
	if a == b {
		return 0
	}

	pa := parseVersion(a)
	pb := parseVersion(b)

	minLen := len(pa)
	if len(pb) < minLen {
		minLen = len(pb)
	}

	for i := range minLen {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}

	if len(pa) < len(pb) {
		return -1
	}
	if len(pa) > len(pb) {
		return 1
	}
	return 0
}

func normalize(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

func parseVersion(v string) []int {
	var parts []int
	for _, p := range strings.Split(v, ".") {
		p = strings.TrimSpace(p)
		numStr := ""
		for _, c := range p {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if numStr == "" {
			parts = append(parts, 0)
		} else {
			n, _ := strconv.Atoi(numStr)
			parts = append(parts, n)
		}
	}
	return parts
}
