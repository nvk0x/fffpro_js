package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	workers   = flag.Int("w", 50, "number of workers")
	timeout   = flag.Int("timeout", 25, "request timeout (seconds)")
	retries   = flag.Int("retries", 3, "retry count")
	outputDir = flag.String("o", "out_js", "output directory")
)

var client *http.Client

func main() {
	flag.Parse()
	client = createHTTPClient(time.Duration(*timeout) * time.Second)
	os.MkdirAll(*outputDir, 0755)

	jobs := make(chan string, *workers*2)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		u := strings.TrimSpace(scanner.Text())
		if u != "" && strings.Contains(u, ".js") {
			jobs <- u
		}
	}

	close(jobs)
	wg.Wait()
}

func worker(jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for rawURL := range jobs {
		processURL(rawURL)
	}
}

func processURL(rawURL string) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return
	}

	if !isHostAlive(req.URL.Host) {
		return
	}

	var resp *http.Response

	for i := 0; i < *retries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] %s\n", rawURL)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return
	}

	fmt.Printf("[JS] %d %s\n", resp.StatusCode, rawURL)

	saveBeautifiedJS(req, body)
}

func saveBeautifiedJS(req *http.Request, body []byte) {
	hash := sha1.Sum([]byte(req.URL.String()))

	safePath := sanitizePath(req.URL.Path)
	if safePath == "" {
		safePath = "root"
	}

	name := fmt.Sprintf("%s_%x.js", safePath, hash)

	dir := filepath.Join(*outputDir, req.URL.Hostname())
	os.MkdirAll(dir, 0755)

	filePath := filepath.Join(dir, name)

	cmd := exec.Command("js-beautify")
	cmd.Stdin = bytes.NewReader(body)

	out, err := cmd.Output()
	if err != nil {
		out = body
	}

	f, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString("// URL: " + req.URL.String() + "\n\n")
	f.Write(out)

	logIndex(filePath, req.URL.String())
}

func logIndex(path, url string) {
	f, err := os.OpenFile("index_js.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%s -> %s\n", path, url))
}

func createHTTPClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        300,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func isHostAlive(host string) bool {
	conn, err := net.DialTimeout("tcp", host+":80", 3*time.Second)
	if err != nil {
		conn, err = net.DialTimeout("tcp", host+":443", 3*time.Second)
		if err != nil {
			return false
		}
	}
	conn.Close()
	return true
}

func sanitizePath(p string) string {
	p = strings.Trim(p, "/")
	p = strings.ReplaceAll(p, "/", "_")
	p = strings.ReplaceAll(p, ".", "_")

	if len(p) > 100 {
		p = p[:100]
	}

	return p
}
