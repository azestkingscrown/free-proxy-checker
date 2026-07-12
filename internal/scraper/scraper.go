package scraper

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"proxyforge-clone/internal/models"
)

const proxySourceURL = "https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/all/data.txt"

// ScrapeProxies fetches the proxy list from the designated source using the provided context.
func ScrapeProxies(ctx context.Context) ([]models.Proxy, error) {
	client := &http.Client{}

	log.Printf("INFO: Fetching proxies from %s", proxySourceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, proxySourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proxies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var proxies []models.Proxy
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Expected format: protocol://ip:port
		parts := strings.SplitN(line, "://", 2)
		if len(parts) != 2 {
			continue
		}

		protocol := parts[0]
		addr := parts[1]

		addrParts := strings.SplitN(addr, ":", 2)
		if len(addrParts) != 2 {
			continue
		}

		ip := addrParts[0]
		port := addrParts[1]

		proxies = append(proxies, models.Proxy{
			Protocol: protocol,
			IP:       ip,
			Port:     port,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	log.Printf("INFO: Successfully scraped %d proxies", len(proxies))
	return proxies, nil
}
