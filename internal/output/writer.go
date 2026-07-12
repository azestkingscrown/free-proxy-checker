package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"proxyforge-clone/internal/models"
)

const txtFileName = "working_proxies.txt"
const jsonFileName = "proxy_stats.json"

// WriteOutputs writes the list of working proxies to txt and json files in the specified outdir.
func WriteOutputs(proxies []models.Proxy, outdir string) error {
	if err := os.MkdirAll(outdir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	txtPath := filepath.Join(outdir, txtFileName)
	jsonPath := filepath.Join(outdir, jsonFileName)

	if err := writeTXT(proxies, txtPath); err != nil {
		return fmt.Errorf("failed to write txt output: %w", err)
	}

	if err := writeJSON(proxies, jsonPath); err != nil {
		return fmt.Errorf("failed to write json output: %w", err)
	}

	return nil
}

func writeTXT(proxies []models.Proxy, filePath string) error {
	file, err := os.Create(filePath)
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

func writeJSON(proxies []models.Proxy, filePath string) error {
	file, err := os.Create(filePath)
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
