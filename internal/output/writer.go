package output

import (
	"encoding/json"
	"fmt"
	"os"

	"proxyforge-clone/internal/models"
)

const txtFileName = "working_proxies.txt"
const jsonFileName = "proxy_stats.json"

// WriteOutputs writes the list of working proxies to txt and json files.
func WriteOutputs(proxies []models.Proxy) error {
	if err := writeTXT(proxies); err != nil {
		return fmt.Errorf("failed to write txt output: %w", err)
	}

	if err := writeJSON(proxies); err != nil {
		return fmt.Errorf("failed to write json output: %w", err)
	}

	return nil
}

func writeTXT(proxies []models.Proxy) error {
	file, err := os.Create(txtFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, p := range proxies {
		line := fmt.Sprintf("%s://%s:%s\n", p.Protocol, p.IP, p.Port)
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(proxies []models.Proxy) error {
	file, err := os.Create(jsonFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(proxies); err != nil {
		return err
	}

	return nil
}
