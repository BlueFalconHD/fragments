package main

import (
	"fmt"
	"regexp"
	"strings"
)

func (s *Site) createFragment(name string, code string) *Fragment {
	f := &Fragment{name: name, code: code, site: s}
	s.fragments[name], _ = f.evaluate()
	return f
}

type fragmentOptions struct {
	renderAsPage bool
	pagePath     string
	scripts      []string
}
type Fragment struct {
	name string
	code string
	meta metaMap
	site *Site

	options fragmentOptions
}

func (f *Fragment) evaluate() (*Fragment, string) {
	localmeta := metaMap{}

	if strings.Contains(f.code, "---") {
		parts := strings.SplitN(f.code, "---", 3)
		metaBlock := parts[1]
		f.code = parts[2]
		// trim leading and trailing whitespace from f.code
		f.code = strings.TrimSpace(f.code)
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
			logWarning(fmt.Sprintf("No meta found for key '%s'", key))
		}
	}

	for _, match := range fragRegex.FindAllStringSubmatch(f.code, -1) {
		fragKey := match[1]
		if fragment, ok := f.site.fragments[fragKey]; ok {
			evaluatedFragment, _ := fragment.evaluate()
			replacements = append(replacements, match[0], evaluatedFragment.code)
		} else {
			logError(fmt.Sprintf("Fragment '%s' not found", fragKey))
		}
	}

	content := strings.NewReplacer(replacements...).Replace(f.code)

	// Run scripts
	for _, scriptName := range f.options.scripts {
		f.runScript(scriptName)
	}

	return &Fragment{code: content, meta: localmeta, site: f.site, name: f.name}, content
}
func (f *Fragment) logMeta() {
	logMap(f.meta, f.name)
}
func (f *Fragment) runScript(scriptName string) {
	if script, exists := scripts[scriptName]; exists {
		script.run(f)
	} else {
		logError(fmt.Sprintf("Script '%s' not found", scriptName))
	}
}
