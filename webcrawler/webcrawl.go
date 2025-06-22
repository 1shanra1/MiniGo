package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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

func main() {
	url := "http://www.example.com"
	body := getHttpContent(url)
	fmt.Printf("%s", body)
}
