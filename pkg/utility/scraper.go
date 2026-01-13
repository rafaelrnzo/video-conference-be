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
	// User-Agent to look like a real browser (Chrome on Windows)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	// 3. Read body (limit to 1MB)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}
	html := string(bodyBytes)

	meta := &OGMeta{URL: targetURL}

	// 4. Robust parsing
	meta.Title = findMetaProperty(html, "og:title")
	if meta.Title == "" {
		meta.Title = findMetaName(html, "title") // twitter:title etc
	}
	if meta.Title == "" {
		meta.Title = extractTitleTag(html)
	}

	meta.Description = findMetaProperty(html, "og:description")
	if meta.Description == "" {
		meta.Description = findMetaName(html, "description")
	}

	meta.Image = findMetaProperty(html, "og:image")
	if meta.Image == "" {
		meta.Image = findMetaName(html, "twitter:image")
	}

	return meta, nil
}

// Helpers for flexible attribute parsing
func findMetaProperty(html, prop string) string {
	// Matches <meta ... property="prop" ... content="value" ... > OR content first
	// We use a simplified scan to find the tag containing the property, then extract content.
	// This is regex-heavy but stdlib constraint limits us.
	// Allow property="X" or property='X'
	return extractContentFromMeta(html, `property=["']`+regexp.QuoteMeta(prop)+`["']`)
}

func findMetaName(html, name string) string {
	return extractContentFromMeta(html, `name=["']`+regexp.QuoteMeta(name)+`["']`)
}

func extractContentFromMeta(html, identPattern string) string {
	// Look for a <meta> tag containing identPattern
	// <meta (attrs)>
	reTag := regexp.MustCompile(`<meta\s+([^>]+)>`)
	matches := reTag.FindAllStringSubmatch(html, -1)
	
	identRe := regexp.MustCompile(identPattern)

	for _, m := range matches {
		attrs := m[1]
		if identRe.MatchString(attrs) {
			// This tag allows the property/name. Now find content.
			// content="value" or content='value'
			contentRe := regexp.MustCompile(`content=["']([^"']+)["']`)
			cMatch := contentRe.FindStringSubmatch(attrs)
			if len(cMatch) > 1 {
				return htmlUnescape(cMatch[1])
			}
		}
	}
	return ""
}

func extractTitleTag(html string) string {
	re := regexp.MustCompile(`<title>([^<]+)</title>`)
	m := re.FindStringSubmatch(html)
	if len(m) > 1 {
		return htmlUnescape(m[1])
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
