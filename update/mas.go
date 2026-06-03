package update

import (
	"os/exec"
	"strings"
)

type MasAppInfo struct {
	ID      string
	Name    string
	Version string
}

func MasListOutdated() ([]MasAppInfo, error) {
	cmd := exec.Command("mas", "outdated")
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, nil
		}
		return nil, err
	}

	return parseMasOutdated(string(out))
}

func parseMasOutdated(output string) ([]MasAppInfo, error) {
	var result []MasAppInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		id := strings.TrimSpace(parts[0])
		nameVer := strings.TrimSpace(parts[1])

		name := nameVer
		ver := ""

		if idx := strings.LastIndex(nameVer, " "); idx > 0 {
			name = nameVer[:idx]
			ver = nameVer[idx+1:]
		}

		result = append(result, MasAppInfo{
			ID:      id,
			Name:    name,
			Version: ver,
		})
	}

	return result, nil
}

func MasUpgrade(id string) (string, error) {
	cmd := exec.Command("mas", "upgrade", id)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func MasUpgradeAll() (string, error) {
	cmd := exec.Command("mas", "upgrade")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
