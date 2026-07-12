package checker

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"h12.io/socks"
	"proxyforge-clone/internal/models"
)

const checkURL = "http://httpbin.org/ip"
const timeout = 10 * time.Second

// CheckProxies takes a list of proxies, checks them concurrently using a worker pool,
// and returns a list of working proxies.
func CheckProxies(ctx context.Context, proxies []models.Proxy, maxWorkers int) []models.Proxy {
	proxyCh := make(chan models.Proxy, len(proxies))
	resultCh := make(chan models.Proxy, len(proxies))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, proxyCh, resultCh)
	}

	// Feed proxies to the pool
	go func() {
		for _, p := range proxies {
			proxyCh <- p
		}
		close(proxyCh)
	}()

	// Wait for workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var workingProxies []models.Proxy
	for result := range resultCh {
		workingProxies = append(workingProxies, result)
	}

	return workingProxies
}

func worker(ctx context.Context, wg *sync.WaitGroup, proxyCh <-chan models.Proxy, resultCh chan<- models.Proxy) {
	defer wg.Done()
	for p := range proxyCh {
		select {
		case <-ctx.Done():
			return
		default:
			// Proceed with check
		}

		if checkProxy(ctx, &p) {
			resultCh <- p
		}
	}
}

func checkProxy(ctx context.Context, p *models.Proxy) bool {
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var transport *http.Transport

	switch p.Protocol {
	case "http", "https":
		proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%s", p.IP, p.Port)) // using http:// for both proxy types usually
		if err != nil {
			return false
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	case "socks4", "socks4a", "socks5":
		proxyAddr := fmt.Sprintf("%s:%s", p.IP, p.Port)

		// Use h12.io/socks for both SOCKS4 and SOCKS5 as it handles them elegantly.
		// Added "?timeout=10s" to prevent the dialing goroutine from hanging indefinitely.
		dialSocksProxy := socks.Dial(fmt.Sprintf("%s://%s?timeout=10s", p.Protocol, proxyAddr))

		transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				type connRes struct {
					conn net.Conn
					err  error
				}
				ch := make(chan connRes, 1)
				go func() {
					c, e := dialSocksProxy(network, addr)
					ch <- connRes{c, e}
				}()

				select {
				case <-ctx.Done():
					// To prevent connection leak, we spawn a goroutine that waits for the connection
					// and closes it if it eventually succeeds after context timeout
					go func() {
						res := <-ch
						if res.conn != nil {
							res.conn.Close()
						}
					}()
					return nil, ctx.Err()
				case res := <-ch:
					return res.conn, res.err
				}
			},
		}
	default:
		// Unsupported protocol
		return false
	}

	client := &http.Client{
		Transport: transport,
		// Timeout removed to rely solely on context timeout
	}

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, checkURL, nil)
	if err != nil {
		return false
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// Verify that the response contains our proxy IP or looks like httpbin response
	// The httpbin /ip endpoint returns {"origin": "IP1, IP2..."}
	if !strings.Contains(string(body), "origin") {
		return false
	}

	latency := time.Since(start).Milliseconds()
	p.Latency = latency
	p.Timestamp = time.Now().Format(time.RFC3339)

	return true
}
