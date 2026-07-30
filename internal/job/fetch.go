package job

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	githubDomains = []string{
		"https://raw.githubusercontent.com",
		"https://github.com",
	}

	fetchClient = &http.Client{
		Timeout: 60 * time.Second,
	}
)

func fetchLines(rawURL, ghProxy string) ([]string, error) {
	url := rewriteURL(rawURL, ghProxy)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ikuai-tools-service/1.0")

	resp, err := fetchClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP %s from %s", resp.Status, url)
	}

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

// rewriteURL prepends ghProxy for GitHub URLs so they are fetched via the proxy.
// Example: "https://ghproxy.com/" + "https://raw.githubusercontent.com/..."
func rewriteURL(rawURL, ghProxy string) string {
	if ghProxy == "" {
		return rawURL
	}
	for _, prefix := range githubDomains {
		if strings.HasPrefix(rawURL, prefix) {
			return strings.TrimRight(ghProxy, "/") + "/" + rawURL
		}
	}
	return rawURL
}
