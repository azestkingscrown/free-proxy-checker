package main

import (
	"context"
	"flag"
	"log"
	"time"

	"proxyforge-clone/internal/checker"
	"proxyforge-clone/internal/output"
	"proxyforge-clone/internal/scraper"
)

func main() {
	workers := flag.Int("workers", 50, "Number of concurrent workers for checking proxies")
	loop := flag.Bool("loop", false, "Run in an infinite loop for local daemon mode")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	for {
		log.Println("INFO: Starting proxy scraping and checking process...")
		runCycle(*workers)

		if !*loop {
			break
		}

		log.Println("INFO: Cycle complete. Waiting 10 seconds before next cycle...")
		time.Sleep(10 * time.Second)
	}

	log.Println("INFO: Proxy scraping and checking process finished.")
}

func runCycle(workers int) {
	// Create context with a 5-minute timeout for the entire cycle
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. Scrape Proxies
	proxies, err := scraper.ScrapeProxies()
	if err != nil {
		log.Printf("ERROR: Failed to scrape proxies: %v", err)
		return
	}

	if len(proxies) == 0 {
		log.Println("WARN: No proxies scraped. Exiting cycle.")
		return
	}

	// 2. Check Proxies concurrently
	log.Printf("INFO: Starting proxy checks with %d workers", workers)
	workingProxies := checker.CheckProxies(ctx, proxies, workers)
	log.Printf("INFO: Finished checking. Found %d working proxies out of %d", len(workingProxies), len(proxies))

	// 3. Write Output
	if err := output.WriteOutputs(workingProxies); err != nil {
		log.Printf("ERROR: Failed to write outputs: %v", err)
		return
	}

	log.Println("INFO: Successfully saved working proxies to files")
}
