package update

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type BrewCaskInfo struct {
	Name       string
	Installed  string
	Latest     string
	Outdated   bool
}

func parseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

func BrewListOutdated() ([]BrewCaskInfo, error) {
	cmd := exec.Command("brew", "outdated", "--cask", "--json")
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, nil
		}
		return nil, err
	}

	return parseBrewOutdated(string(out))
}

func parseBrewOutdated(jsonStr string) ([]BrewCaskInfo, error) {
	type brewCask struct {
		Name      string   `json:"name"`
		Installed []string `json:"installed_versions"`
		Current   string   `json:"current_version"`
	}

	var casks []brewCask
	if err := parseJSON(jsonStr, &casks); err != nil {
		return nil, err
	}

	var result []BrewCaskInfo
	for _, c := range casks {
		installed := ""
		if len(c.Installed) > 0 {
			installed = c.Installed[0]
		}
		result = append(result, BrewCaskInfo{
			Name:      c.Name,
			Installed: installed,
			Latest:    c.Current,
			Outdated:  true,
		})
	}

	return result, nil
}

func BrewUpgrade(name string) (string, error) {
	cmd := exec.Command("brew", "upgrade", "--cask", name)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func BrewListInstalled() (map[string]string, error) {
	cmd := exec.Command("brew", "list", "--cask", "--versions")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			result[parts[0]] = ""
		}
	}

	return result, nil
}
