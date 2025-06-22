package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

func getHttpContent(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
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
	url := "https://www.apple.com"
	body := getHttpContent(url)
	absoluteTags := extractAbsoluteTags(body)
	filteredTags, _ := filterAndResolveTags(url, absoluteTags)
	fmt.Printf("%v", filteredTags[1].Host)
}
