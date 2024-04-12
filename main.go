package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type metaMap map[string]string
type fragmentMap map[string]*Fragment

func getGlobalMeta() metaMap {
	currentTime := time.Now()
	return metaMap{
		"timestamp": currentTime.Format(time.RFC3339),
		"date":      currentTime.Format("2006-01-02"),
		"month":     currentTime.Format("01"),
		"year":      currentTime.Format("2006"),
		"unix":      fmt.Sprintf("%d", currentTime.Unix()),
	}
}

func site() *Site {
	return &Site{
		fragments: make(fragmentMap),
		meta:      getGlobalMeta(),
	}
}

type Site struct {
	fragments fragmentMap
	meta      metaMap
}

func (s *Site) createFragment(name string, code string) {
	f := &Fragment{code: code, site: s}
	s.fragments[name] = f.evaluate()
}

type Fragment struct {
	code string
	meta metaMap
	site *Site
}

func (f *Fragment) evaluate() *Fragment {
	localmeta := metaMap{}

	if strings.Contains(f.code, "---") {
		parts := strings.SplitN(f.code, "---", 3)
		metaBlock := parts[1]
		f.code = parts[2]
		metaLines := strings.Split(metaBlock, "\n")
		for _, line := range metaLines {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				localmeta[key] = value
			}
		}
	}

	metaRegex := regexp.MustCompile(`\$\{(.*?)}`)
	fragRegex := regexp.MustCompile(`@\{(.*?)}`)

	var replacements []string
	for _, match := range metaRegex.FindAllStringSubmatch(f.code, -1) {
		key := match[1]
		if value, exists := localmeta[key]; exists {
			replacements = append(replacements, match[0], value)
		} else if value, exists := f.site.meta[key]; exists {
			replacements = append(replacements, match[0], value)
		} else {
			fmt.Printf("Warning: No replacement found for meta key '%s'\n", key)
		}
	}

	for _, match := range fragRegex.FindAllStringSubmatch(f.code, -1) {
		fragKey := match[1]
		if fragment, ok := f.site.fragments[fragKey]; ok {
			evaluatedFragment := fragment.evaluate()
			replacements = append(replacements, match[0], evaluatedFragment.code)
		} else {
			fmt.Printf("Warning: No fragment found for key '%s'\n", fragKey)
		}
	}

	content := strings.NewReplacer(replacements...).Replace(f.code)
	return &Fragment{code: content, meta: localmeta, site: f.site}
}

func main() {
	site := site()
	site.createFragment("footer", `
---
dateUpdated: ${date}
---
<footer>Last updated on ${date}.</footer>
`)

	site.createFragment("home", `
---
title: Home/Welcome
siteName: Example Site
---
Welcome to ${siteName}.
Today's date is ${date}.
Test undefined meta: ${undefined}

@{footer}
`)

	fmt.Println("Home Page Output:")
	fmt.Println(site.fragments["home"].code)
}
