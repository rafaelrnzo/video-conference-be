package utility

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type OGMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
}

// ScrapeOG fetches the URL and extracts Open Graph tags
func ScrapeOG(targetURL string) (*OGMeta, error) {
	// 1. Validate URL (basic)
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	// 2. Create HTTP Client with timeout and skip SSL verify (optional but safer for scraping variety)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	// User-Agent to look like a browser/bot so sites allow us
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MeetBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	// 3. Read body (limit to 500KB to prevent abuse/dos)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}
	html := string(bodyBytes)

	meta := &OGMeta{URL: targetURL}

	// 4. Regex parsing for OG tags
	// Note: html.Parse is more robust but regex is fine for simple OG extraction
	// Title
	meta.Title = extractTag(html, `property="og:title"\s+content="([^"]+)"`)
	if meta.Title == "" {
		meta.Title = extractTag(html, `<title>([^<]+)</title>`)
	}

	// Description
	meta.Description = extractTag(html, `property="og:description"\s+content="([^"]+)"`)
	if meta.Description == "" {
		meta.Description = extractTag(html, `name="description"\s+content="([^"]+)"`)
	}

	// Image
	meta.Image = extractTag(html, `property="og:image"\s+content="([^"]+)"`)

	return meta, nil
}

func extractTag(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return htmlUnescape(matches[1])
	}
	return ""
}

func htmlUnescape(s string) string {
	// simplistic unescape, mainly for &amp; etc
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return s
}
