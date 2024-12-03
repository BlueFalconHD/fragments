package main

import (
	"github.com/charmbracelet/log"
	"os"
	"strings"
)

import (
	"io/fs"
	"path/filepath"
)

func RecursivelyFindPages(op string, cache *FragmentCache) map[string]*Fragment {

	pageMap := make(map[string]*Fragment)

	err := filepath.WalkDir(op, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Create a fragment for each page
			// construct a path like "posts/example" from "{pagePath}/posts/example.frag"
			// and create a fragment for it
			fragmentName := strings.TrimSuffix(strings.TrimPrefix(path, op+"/"), ".frag")
			f := GetFragmentFromName(fragmentName, PAGE, cache)

			// Add the fragment to the page map
			pageMap[fragmentName] = f

		}
		return nil
	})
	if err != nil {
		log.Error("Error walking the path", "path", op, "error", err)
	}

	return pageMap
}

func testLua() {

	config := &Config{
		SiteRoot:      "exampleSite",
		FragmentsPath: "fragment/",
		PagePath:      "page/",
		BuildPath:     "build",
	}
	fcache := NewFragmentCache(config)

	// Look through the page directory and create a fragment for each page (recursively)
	// This is the first step in the build process
	pageMap := RecursivelyFindPages("exampleSite/page", fcache)

	// Make the directory for the build
	os.MkdirAll("exampleSite/build", os.ModePerm)

	// Print the pageMap
	for k, v := range pageMap {
		log.Info("Building page", "name", k)
		res := v.Evaluate()

		// Write to the file in the build directory with the same name as the page (k) + ".html"
		err := os.MkdirAll(filepath.Dir("exampleSite/build/"+k+".html"), os.ModePerm)
		if err != nil {
			log.Error("Error creating directories", "error", err)
		}
		file, err := os.Create("exampleSite/build/" + k + ".html")

		if err != nil {
			log.Error("Error creating file", "error", err)
		}

		_, err = file.Write([]byte(res))

		if err != nil {
			log.Error("Error writing to file", "error", err)
		}

		file.Close()

		log.Info("Page built", "name", k)
	}
}

func main() {
	testLua()
}
