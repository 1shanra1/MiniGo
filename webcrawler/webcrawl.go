package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
)

type CrawlTask struct {
	BaseUrl string
	URL     *url.URL
	Depth   int
}

type CrawlResult struct {
	ParentTask CrawlTask
	FoundLinks []*url.URL
}

type WebCrawler struct {
	tasks   chan CrawlTask
	results chan CrawlResult
	wg      sync.WaitGroup
	mu      sync.Mutex
	visited map[string]bool
}

func (wc *WebCrawler) worker(id int) {
	log.Printf("Worker %d started\n", id)
	for task := range wc.tasks {
		log.Printf("Worker %d processing: %s (depth: %d)\n", id, task.URL.String(), task.Depth)
		httpContent := getHttpContent(task.URL.String())
		if httpContent == nil {
			log.Printf("Worker %d: Failed to get content for %s\n", id, task.URL.String())
			wc.wg.Done()
			continue
		}
		links := extractAbsoluteTags(httpContent)
		log.Printf("Worker %d: Extracted %d links from %s\n", id, len(links), task.URL.String())
		filtered, _ := filterAndResolveTags(task.BaseUrl, links)
		log.Printf("Worker %d: Filtered to %d same-domain links\n", id, len(filtered))

		// Don't call Done yet - wait until results are processed
		result := CrawlResult{
			ParentTask: task,
			FoundLinks: filtered,
		}
		wc.results <- result
		log.Printf("Worker %d: Sent result for %s\n", id, task.URL.String())
	}
	log.Printf("Worker %d exiting\n", id)
}

func getHttpContent(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return nil
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Print(err)
	}

	return body
}

func extractAbsoluteTags(b []byte) []string {
	r := regexp.MustCompile(`<a\s+.*?href\s*=\s*['"]([^'"]+)['"]`)

	matches := r.FindAllSubmatch(b, -1)

	var urls []string

	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, string(match[1]))
		}
	}

	return urls
}

// Here I need to make sure that it is only links related to the same website that I care about
func filterAndResolveTags(baseUrl string, urls []string) ([]*url.URL, error) {
	var resolvedUrls []*url.URL

	base, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("Error with parsing: %w", err)
	}

	for _, urlString := range urls {
		parsed, err := url.Parse(urlString)
		if err != nil {
			return nil, fmt.Errorf("Error with parsing: %w", err)
		}
		resolved := base.ResolveReference(parsed)
		if resolved.Host == base.Host {
			resolvedUrls = append(resolvedUrls, resolved)
		}
	}

	return resolvedUrls, nil
}

func main() {
	wc := WebCrawler{
		tasks:   make(chan CrawlTask, 1000), // Large buffer
		results: make(chan CrawlResult, 1000),
		visited: make(map[string]bool),
	}

	for i := 0; i < 5; i++ {
		go wc.worker(i)
	}

	startURLString := "https://daylightcomputer.com"
	startURL, err := url.Parse(startURLString)
	if err != nil {
		log.Printf("Failed to parse the start URL: %v", err)
	}

	wc.wg.Add(1)
	wc.visited[startURL.String()] = true
	wc.tasks <- CrawlTask{BaseUrl: startURLString, URL: startURL, Depth: 0}

	log.Println("Crawler started. Waiting for results...")

	// Process results in a separate goroutine
	done := make(chan bool)
	go func() {
		for result := range wc.results {
			log.Printf("Crawled: %s, Found %d links.\n", result.ParentTask.URL, len(result.FoundLinks))

			// Only process links if we're not at max depth
			if result.ParentTask.Depth < 2 {
				for _, link := range result.FoundLinks {
					wc.mu.Lock()
					if !wc.visited[link.String()] {
						wc.visited[link.String()] = true
						wc.wg.Add(1)
						log.Printf("Main: Queueing new task: %s (depth: %d)\n", link.String(), result.ParentTask.Depth+1)
						wc.tasks <- CrawlTask{BaseUrl: result.ParentTask.BaseUrl, URL: link, Depth: result.ParentTask.Depth + 1}
					}
					wc.mu.Unlock()
				}
			}

			// Mark the parent task as done after processing its results
			wc.wg.Done()
			log.Printf("Main: Marked task done for %s\n", result.ParentTask.URL.String())
		}
		done <- true
	}()

	// Wait for all tasks to complete, then close channels
	log.Println("Monitor: Waiting for all tasks to complete...")
	wc.wg.Wait()
	log.Println("Monitor: All tasks complete, closing channels...")
	close(wc.tasks)
	close(wc.results)

	// Wait for result processing to finish
	<-done

	log.Printf("Crawl finished. Visited %d pages.", len(wc.visited))
	fmt.Println("\n--- Site Map ---")
	for page := range wc.visited {
		fmt.Println(page)
	}
}
