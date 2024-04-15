package main

import (
	"fmt"
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

func mainConfigStuff() {
	// TODO --------------------------------------------------------------------------------------------------------
	// 1. read files recursively 1 level deep from the current directory
	//     - Check if the file has the .frag, .fragment extension
	//     - Read the file content
	//     - Create a fragment with the file name (without extension) and the file content
	// 2. Read the config file
	//     - Read the global meta
	//     - Read the pages
	//     - For each page, read the file name and scripts, and add them to the fragment specified by the file name
	// TODO --------------------------------------------------------------------------------------------------------
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
	site.createFragment("about", `# About Us
This is the about page.

@{footer}
`)

	logBreak()

	logMap(site.meta, "Global Meta")

	// log meta of all fragments
	for _, fragment := range site.fragments {
		fragment.logMeta()
	}

	logBreak()

	site.fragments["about"].runScript("RenderMarkdown")
	fmt.Println(site.fragments["about"].code)
}
