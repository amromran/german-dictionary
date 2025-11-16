package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <german-word>\n", os.Args[0])
		os.Exit(1)
	}

	word := strings.ToLower(strings.TrimSpace(os.Args[1]))
	if word == "" {
		fmt.Fprintf(os.Stderr, "Please provide a word to search\n")
		os.Exit(1)
	}

	meanings, err := searchWord(word)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(meanings) == 0 {
		fmt.Printf("No translations found for '%s'\n", word)
		return
	}

	fmt.Printf("\nTranslations for '%s':\n", word)
	for i, meaning := range meanings {
		fmt.Printf("%d. %s\n", i+1, meaning)
	}
}

func searchWord(word string) ([]string, error) {
	dictURL := fmt.Sprintf("https://en.langenscheidt.com/german-english/%s", url.PathEscape(word))

	client := &http.Client{}
	req, err := http.NewRequest("GET", dictURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %v", err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %v", err)
	}

	meanings := extractMeanings(doc)

	if len(meanings) > 5 {
		meanings = meanings[:5]
	}

	return meanings, nil
}

func extractMeanings(n *html.Node) []string {
	var meanings []string
	seen := make(map[string]bool)

	// Find all <a> tags and skip those inside example sections
	var findLinks func(*html.Node, bool)
	findLinks = func(n *html.Node, inExample bool) {

		if n.Type == html.ElementNode {
			class := getAttr(n, "class")

			if strings.Contains(class, "example") ||
				strings.Contains(class, "usage") ||
				strings.Contains(class, "sentence") ||
				n.Data == "example" {
				inExample = true
			}
		}

		if n.Type == html.ElementNode && n.Data == "a" && !inExample {
			href := getAttr(n, "href")

			if strings.HasPrefix(href, "/english-german/") {
				text := getTextContent(n)
				text = strings.TrimSpace(text)
				if text != "" && len(text) > 1 && len(text) < 30 && !seen[text] {
					// Filter out non-word content
					if isValidTranslation(text) {
						meanings = append(meanings, text)
						seen[text] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c, inExample)
		}
	}

	findLinks(n, false)
	return meanings
}

func isValidTranslation(s string) bool {

	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z\s-]*[a-zA-Z]$`, s)
	if !matched {
		return false
	}

	// Filter out common words for now
	lower := strings.ToLower(s)
	blacklist := []string{
		"translation", "overview", "examples", "show", "hide",
		"click", "tap", "more", "less", "see", "view",
		"german", "english", "noun", "verb", "adjective",
	}

	for _, bad := range blacklist {
		if strings.Contains(lower, bad) {
			return false
		}
	}

	return true
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return text
}
