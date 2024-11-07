package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type FragmentEvaluationState int

const (
	PENDING FragmentEvaluationState = iota
	EVALUATING
	EVALUATED
)

type FragmentType int

const (
	FRAGMENT FragmentType = iota
	PAGE
	TEMPLATE
)

// TODO: refactor with different fragment types, support templated pages, and template fragments.
// idea dump: template fragments have access to shared metadata as well.

type Fragment struct {
	Name       string
	Type       FragmentType
	Code       string
	Depth      int
	Parent     *Fragment
	LocalMeta  CoreTable
	SharedMeta *CoreTable
	EvalState  FragmentEvaluationState
	Builders   CoreTable
}

func (f *Fragment) MakeChild(name string, code string) *Fragment {
	return &Fragment{
		Name:       name,
		Type:       FRAGMENT,
		Code:       code,
		Depth:      f.Depth + 1,
		Parent:     f,
		LocalMeta:  *NewEmptyCoreTable(),
		SharedMeta: f.SharedMeta,
	}
}

func (f *Fragment) RetrieveSharedMetadata() *CoreTable {
	// Recursively call this function on the parent fragment until depth is 0
	if f.Depth == 0 {
		return f.SharedMeta
	}

	return f.Parent.RetrieveSharedMetadata()
}

func (f *Fragment) MakeLFragment() *LFragment {

	if f.Parent == nil {
		return &LFragment{
			Fragment:   f,
			Parent:     nil,
			LocalMeta:  &f.LocalMeta,
			SharedMeta: f.RetrieveSharedMetadata(),
		}
	}

	return &LFragment{
		Fragment:   f,
		Parent:     f.Parent.MakeLFragment(),
		LocalMeta:  &f.LocalMeta,
		SharedMeta: f.RetrieveSharedMetadata(),
	}
}

func (f *Fragment) NewChildFragmentFromName(name string) *Fragment {
	// TODO: determine where a good root for the fragments are, currently just the same directory where run

	rd := "exampleSite/fragment/"

	// get run directory
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// get fragment directory
	dir = dir + "/" + rd

	// look for the name provided + ".frag"
	file, err := os.Open(dir + name + ".frag")
	if err != nil {
		panic(err)
	}

	// read the file
	b := make([]byte, 1024)
	n, err := file.Read(b)
	if err != nil {
		fmt.Println("Error reading file:", err)
		panic(err)
	}

	// create a new fragment
	return f.MakeChild(name, string(b[:n]))
}

/*

Rendering Process:

fragment:
 - Evaluate lua
 - Replace references in fragment code to meta with actual values
 - Run the builder functions to replace builder references with return values
 - Run fragment process on each fragment reference and replace with the result

Render Page:
 - Run fragment process on page fragment
 - Replace ${CONTENT} in the page template with the result of the page fragment process

*/

func (f *Fragment) Evaluate() string {
	parts := strings.Split(f.Code, "---")
	var lua, code string

	if len(parts) == 1 {
		// No "---" found, treat the entire f.Code as "code" part
		lua = ""
		code = parts[0]
	} else {
		// Split into lua and code parts as expected
		lua = parts[0]
		code = parts[1]
	}

	// Evaluate lua if it's present
	if lua != "" {
		f.RunLua(lua)
	}

	// Replace references in fragment code to meta with actual values
	metaReferences := regexp.MustCompile(`\$\{([^}]+)\}`).FindAllStringSubmatch(code, -1)
	builderReferences := regexp.MustCompile(`\*\{([^}]+)\}`).FindAllStringSubmatch(code, -1)
	fragReferences := regexp.MustCompile(`@\{([^}]+)\}`).FindAllStringSubmatch(code, -1)

	for _, ref := range metaReferences {
		key := ref[1]
		value := f.LocalMeta.v[key]
		code = strings.ReplaceAll(code, "${"+key+"}", value.goType().(string))
	}

	for _, ref := range builderReferences {
		builderName := ref[1]
		code = strings.ReplaceAll(code, "*{"+builderName+"}", "<BUILDER TODO>")
	}

	for _, ref := range fragReferences {
		fragmentName := ref[1]
		childFragment := f.NewChildFragmentFromName(fragmentName)
		code = strings.ReplaceAll(code, "@{"+fragmentName+"}", childFragment.Evaluate())
	}

	return code
}
