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

func main() {
	// site := site()

	testLua()
}
