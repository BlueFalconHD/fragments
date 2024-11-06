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

type Fragment struct {
	Name       string
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

func (f *Fragment) MakeLFragment(parent *LFragment) *LFragment {
	return &LFragment{
		Fragment:   f,
		Parent:     parent,
		LocalMeta:  &f.LocalMeta,
		SharedMeta: f.RetrieveSharedMetadata(),
	}
}

// TODO: new fragment refactor add some evaluation functions and link to LFragment

func replaceMetaReferences(code string, sm *CoreTable, m *CoreTable) string {
	replace := func(meta map[string]interface{}) {
		for k, v := range meta {
			code = strings.ReplaceAll(code, "${"+k+"}", v.(string))
		}
	}

	replace(sm.goType().(map[string]interface{}))
	replace(m.goType().(map[string]interface{}))

	return code
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

	// Retrieve the first part of the file (before "====="), which is the lua

	// lua := f.Code[:strings.Index(f.Code, "=====")]
	// TODO: Evaluate lua

	// Retrieve the second part of the file (after "====="), which is the actual code
	code := f.Code[strings.Index(f.Code, "=====")+5:]

	// Replace references to meta with actual values
	code = replaceMetaReferences(code, f.SharedMeta, &f.LocalMeta)

	// Run the builder functions to replace builder references with return values
	// Builder functions are formatted as *{builderName}
	// TODO: Run builder functions

	// Run fragment process on each fragment reference and replace with the result
	// Fragment references are formatted as @{fragmentName}

	references := regexp.MustCompile(`@\{([^}]+)\}`).FindAllStringSubmatch(code, -1)
	for _, ref := range references {
		fragmentName := ref[1]
		childFragment := f.NewChildFragmentFromName(fragmentName)
		code = strings.ReplaceAll(code, "@{"+fragmentName+"}", childFragment.Evaluate())
	}

	// Return the evaluated code
	return code
}
