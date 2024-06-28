package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

    "github.com/fatih/color"
	"github.com/corpix/uarand"
)

type FetchConfig struct {
	threads     int
	directory   string
	retries     int
	timeout     time.Duration
	headers     map[string]string
	randomAgent bool
	proxy       string
	silent      bool
}

// Defining colors
var green = color.New(color.FgGreen)
var red = color.New(color.FgRed)
var boldCyan = color.New(color.FgCyan).Add(color.Bold)


// Driver code 
func main() {

	// configuring cli args
	filePath := flag.String("f", "", "File containing URLs (one per line)")
	threads := flag.Int("t", 60, "Number of threads")
	directory := flag.String("dir", "fetched", "Directory to save output")
	retries := flag.Int("r", 3, "Number of retries")
	timeout := flag.Int("x", 12, "Timeout in seconds")
	headers := flag.String("H", "", "Headers to send with request (comma-separated key:value pairs)")
	randomAgent := flag.Bool("ra", false, "Use random user agent")
	proxy := flag.String("proxy", "", "Proxy (format: IP:PORT)")
	silent := flag.Bool("silent", false, "Silent mode, only output URLs that are fetched successfully")

	flag.Parse()

	if *filePath == "" {
		red.Println("[x] File path is required")
		flag.Usage()
		return
	}

	config := FetchConfig{
		threads:     *threads,
		directory:   *directory,
		retries:     *retries,
		timeout:     time.Duration(*timeout) * time.Second,
		headers:     parseHeaders(*headers),
		randomAgent: *randomAgent,
		proxy:       *proxy,
		silent:      *silent,
	}


	err := os.MkdirAll(config.directory, os.ModePerm)
	if err != nil {
		if !config.silent {
			red.Printf("[x] Failed to create directory: [%v]\n", err)
		}
		return
	}

	urls, err := readLines(*filePath)
	if err != nil {
		if !config.silent {
			red.Printf("[x] Failed to read file: [%v]\n", err)
		}
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.threads)
	successCount := 0
	var mu sync.Mutex

	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			success := fetchURL(url, config)
			if success {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(url)
	}


	boldCyan.Printf("---\n")
	boldCyan.Printf("[>] URLs provided [%d]\n", len(urls))
	boldCyan.Printf("---\n")


	wg.Wait()
	if !config.silent {
		boldCyan.Printf("---\n[>] Successfully fetched [%d/%d] URLs\n---\n", successCount, len(urls))
	}
}

// Requesting and saving the response 
func fetchURL(url string, config FetchConfig) bool {
	client := createHTTPClient(config)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if !config.silent {
			red.Printf("[x] Failed to create request for  [%s]: [%v]\n", url, err)
		}
		return false
	}

	for key, value := range config.headers {
		req.Header.Set(key, value)
	}

	if config.randomAgent {
		req.Header.Set("User-Agent", uarand.GetRandom())
	}

	for i := 0; i <= config.retries; i++ {
		resp, err := client.Do(req)
		if err != nil {
			if i == config.retries {
				if !config.silent {
					red.Printf("[x] Failed [%s] after [%d] attempts\n", url, config.retries)
				}
				return false
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if !config.silent {
				red.Printf("[x] Status [%d] [%s]\n", resp.StatusCode, url)
			}
			return false
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if i == config.retries {
				if !config.silent {
					red.Printf("[x] Failed to read response body for [%s] after [%d] attempts\n", url, config.retries+1,)
				}
				return false
			}
			continue
		}

		filename := filepath.Join(config.directory, sanitizeFilename(url))
		err = ioutil.WriteFile(filename, body, 0644)
		if err != nil {
			if !config.silent {
				red.Printf("[x] Failed to write to file [%s]: [%v]\n", filename, err)
			}
			return false
		}

		if !config.silent {
			green.Printf("[>] Fetched: [%s]\n", url)
		} else if config.silent {
			fmt.Printf("%s\n", url)
		}

		return true
	}

	return false
}

// Making custom http client to disable SSL certificate verification
func createHTTPClient(config FetchConfig) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if config.proxy != "" {
		proxyURL, _ := url.Parse("http://" + config.proxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{
		Transport: transport,
		Timeout:   config.timeout,
	}
}

// Parsing headers so it can pe passed with request 
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// Reading Urls from provided file 
func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Replacing '/' with '_' & stripping 'http://' & 'https://'
func sanitizeFilename(urlStr string) string {
	urlStr = strings.ReplaceAll(urlStr, "http://", "")
	urlStr = strings.ReplaceAll(urlStr, "https://", "")
	urlStr = strings.ReplaceAll(urlStr, "/", "_")
	return urlStr
}
