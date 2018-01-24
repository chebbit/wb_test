package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

func init() {
	log.SetFlags(0)
}

// countPatternInURL - get response by url and calculating
// pattern in HTML body
func countPatternInURL(url string, pattern string) int {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return 0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return 0
	}
	r, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return 0
	}
	indexes := r.FindAllStringIndex(string(body), -1)
	count := len(indexes)
	log.Printf("Count for %v: %v", url, count)
	return count
}

// sendToChannel push url to channel
func sendToChannel(ch chan string, url string) {
	ch <- url
}

func main() {
	const PATTERN string = `\bGo\b`
	var n int
	k := 5
	var mu sync.Mutex
	var urls = make(chan string, 100)
	var tokens = make(chan struct{}, k)
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		n++ // counting url
		text := scanner.Text()
		go sendToChannel(urls, text)

	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	total := 0

	// calculate sum of matched "Go" in URLs
	for ; n > 0; n-- {
		url := <-urls
		wg.Add(1)
		go func(url string) {
			tokens <- struct{}{}
			count := countPatternInURL(url, PATTERN)
			mu.Lock()
			total += count
			mu.Unlock()
			<-tokens
			wg.Done()
		}(url)
	}
	wg.Wait()
	log.Printf("Total: %v", total)
}
