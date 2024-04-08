package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Site struct {
	pages     map[string]Page
	fragments map[string]Fragment
}
type Page struct {
	markdown string
	meta     map[string]string
}
type Fragment struct {
	code string
}

func getGlobalMeta() map[string]string {
	// Get the current time
	currentTime := time.Now()

	// Create a map to hold the metadata
	meta := map[string]string{
		"timestamp": currentTime.Format(time.RFC3339),      // Full timestamp
		"date":      currentTime.Format("2006-01-02"),      // Date in YYYY-MM-DD format
		"month":     currentTime.Format("01"),              // Month as a two-digit number
		"year":      currentTime.Format("2006"),            // Year in YYYY format
		"unix":      fmt.Sprintf("%d", currentTime.Unix()), // Unix timestamp
	}

	return meta
}

func (p *Page) getMeta() {
	if p.meta == nil {
		p.meta = make(map[string]string)
	}

	// Use a scanner to read the markdown line by line
	scanner := bufio.NewScanner(strings.NewReader(p.markdown))

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`"([^"]+)":\s*"([^"]+)"`)

	for scanner.Scan() {
		line := scanner.Text()

		// Stop reading metadata if we reach a line starting with '---'
		if line == "---" {
			break
		}

		// Extract the key-value pair from the line using regular expression
		matches := re.FindStringSubmatch(line)
		if matches != nil && len(matches) == 3 {
			key := matches[1]
			value := matches[2]
			p.meta[key] = value
		}
	}

	// Add global metadata to the page metadata
	globalMeta := getGlobalMeta()
	for key, value := range globalMeta {
		p.meta[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading markdown:", err)
	}
}
func (p *Page) eval(s Site) string {
	newMarkdown := p.markdown

	matches := regexp.MustCompile(`\$\{(.*?)}`).FindAllStringSubmatch(newMarkdown, -1)

	for _, match := range matches {
		key := match[1]
		if fragment, ok := s.fragments[key]; ok {
			replacement := fragment.eval(*p)
			newMarkdown = strings.Replace(newMarkdown, match[0], replacement, -1)
		} else {
			newMarkdown = strings.Replace(newMarkdown, match[0], "", -1)
		}
	}

	return newMarkdown
}

func (f *Fragment) eval(parent Page) string {
	// Create a new variable to hold the modified code
	newCode := f.code

	// Find any parts of the code matching ${...}
	matches := regexp.MustCompile(`\$\{(.*?)}`).FindAllStringSubmatch(newCode, -1)

	// Check the parent page's meta for the key (...), and
	// replace the matched pattern with the meta value.
	for _, match := range matches {
		// Update parent metadata
		parent.getMeta()

		key := match[1]
		if value, ok := parent.meta[key]; ok {
			newCode = strings.Replace(newCode, match[0], value, -1)
		} else {
			newCode = strings.Replace(newCode, match[0], "", -1)
		}
	}

	return newCode
}
