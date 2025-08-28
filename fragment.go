package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
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

//type FragmentCache map[string]*Fragment

type FragmentCache struct {
	Cache  map[string]*Fragment
	Config *Config
}

func NewFragmentCache(c *Config) *FragmentCache {
	return &FragmentCache{
		Cache:  make(map[string]*Fragment),
		Config: c,
	}
}

type Fragment struct {
	Name          string
	Type          FragmentType
	Code          string
	Depth         int
	Parent        *Fragment
	LocalMeta     CoreTable
	SharedMeta    *CoreTable
	EvalState     FragmentEvaluationState
	Builders      *CoreTable
	Template      *Fragment
	FragmentCache *FragmentCache
	Config        *Config
}

func (f *Fragment) MakeChild(name string, code string) *Fragment {
	return &Fragment{
		Name:          name,
		Type:          FRAGMENT,
		Code:          code,
		Depth:         f.Depth + 1,
		Parent:        f,
		LocalMeta:     *NewEmptyCoreTable(),
		SharedMeta:    f.SharedMeta,
		Builders:      NewEmptyCoreTable(),
		FragmentCache: f.FragmentCache,
		Config:        f.Config,
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

func GetFragmentFromName(name string, fragType FragmentType, cache *FragmentCache) *Fragment {
	// TODO: determine where a good root for the fragments are, currently just the same directory where run

	// Depending on the type of fragment, get the directory specified in the config
	var rd string
	switch fragType {
	case FRAGMENT:
		rd = cache.Config.FragmentsPath
	case PAGE:
		rd = cache.Config.PagePath
	case TEMPLATE:
		rd = cache.Config.FragmentsPath
	}

	// Build full path to fragment file using config site root and paths
	fullPath := filepath.Join(cache.Config.SiteRoot, rd, name+".frag")

	b, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	// create a new fragment
	return &Fragment{
		Name:          name,
		Type:          fragType,
		Code:          string(b),
		Depth:         0,
		Parent:        nil,
		LocalMeta:     *NewEmptyCoreTable(),
		SharedMeta:    NewEmptyCoreTable(),
		Builders:      NewEmptyCoreTable(),
		FragmentCache: cache,
		Config:        cache.Config,
	}
}

func (c *FragmentCache) GetAll(fragType FragmentType) map[string]*Fragment {
	result := make(map[string]*Fragment)
	for name, f := range c.Cache {
		if f.Type == fragType {
			result[name] = f
		}
	}
	return result
}

func (c *FragmentCache) Get(name string, fragType FragmentType) *Fragment {
	if f, ok := (c.Cache)[name]; ok {
		return f
	}
	f := GetFragmentFromName(name, fragType, c)
	if f == nil {
		log.Error("Fragment not found", "name", name)
		return nil
	}

	f.Evaluate()
	return f
}

func (c *FragmentCache) Add(name string, f *Fragment) {
	if c == nil {
		// cannot initialize a nil receiver; just return to avoid panic
		return
	}
	if c.Cache == nil {
		c.Cache = make(map[string]*Fragment)
	}
	if c.Config == nil && f != nil {
		c.Config = f.Config
	}
	c.Cache[name] = f
}

func (f *Fragment) NewChildFragmentFromName(name string) *Fragment {
	// TODO: determine where a good root for the fragments are, currently just the same directory where run

	nf := GetFragmentFromName(name, FRAGMENT, f.FragmentCache)

	// Set the parent of the new fragment to this fragment
	nf.Parent = f

	return nf
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
	f.EvalState = PENDING

	// Setup lua state
	L := f.CreateState()
	defer L.Close()

	cleaned := strings.TrimSpace(f.Code)

	// Split fragment into lua and code parts, separated by '~~~'
	parts := strings.SplitN(cleaned, "~~~", 2)
	var luaCode, code string
	if len(parts) >= 2 {
		luaCode = parts[0]
		code = parts[1]
	} else {
		luaCode = ""
		code = parts[0]
	}

	f.EvalState = EVALUATING

	// Strip leading and trailing whitespace from code
	code = strings.TrimSpace(code)

	// Evaluate lua if it's present
	if luaCode != "" {
		err := L.DoString(luaCode)
		if err != nil {
			log.Error(err)
		}
	}

	// Parse code into AST
	nodes, err := ParseCode(code, f)
	if err != nil {
		log.Error(err)
		return ""
	}

	// Evaluate nodes
	var result strings.Builder
	for _, node := range nodes {
		s, err := node.Evaluate(f, L)
		if err != nil {
			log.Error(err)
			continue
		}
		result.WriteString(s)
	}

	f.EvalState = EVALUATED

	if f.Template != nil {
		if f.Depth == 0 {

			// Set the template's shared metadata to the fragment's shared metadata
			f.Template.SharedMeta = f.SharedMeta

			// Set ${CONTENT} in the template to the result of this fragment
			if f.Template.LocalMeta.v == nil {
				f.Template.LocalMeta.v = make(map[string]CoreType)
			}
			f.Template.LocalMeta.v["CONTENT"] = NewCoreString(result.String())
			// Add the fragment to the cache before returning so listings can discover it
			f.FragmentCache.Add(f.Name, f)
			// Evaluate the template
			return f.Template.Evaluate()
		} else {
			log.Error("Template fragments are only allowed at the root of the fragment tree.")
		}
	}

	// Add the fragment to the cache
	f.FragmentCache.Add(f.Name, f)

	return result.String()
}

func (f *Fragment) WithContent(content string, of *Fragment) string {

	// Merge this fragment's shared metadata with the provided fragment's shared metadata
	f.SharedMeta.mergeMut(of.SharedMeta)

	// Replace ${CONTENT} in the fragment code with the content provided
	if f.LocalMeta.v == nil {
		f.LocalMeta.v = make(map[string]CoreType)
	}
	f.LocalMeta.v["CONTENT"] = NewCoreString(content)

	c := f.Evaluate()
	return c
}

func (f *Fragment) CreateState() *lua.LState {
	lf := f.MakeLFragment()
	L := lua.NewState()

	// Register fragment and fragments module types
	registerFragmentType(L)
	registerFragmentsModuleType(L)
	registerCoreTableType(L)

	// Register the markdown rendering function
	L.SetGlobal("renderMarkdown", L.NewFunction(renderMarkdown))

	// Preload standard libraries
	libs.Preload(L)

	// Register 'this' fragment
	lf.registerThisFragmentAs(L, "this")

	// Create and register the fragments module
	fragPath := filepath.Join(f.Config.SiteRoot, f.Config.FragmentsPath)
	pagePath := filepath.Join(f.Config.SiteRoot, f.Config.PagePath)
	fragmentsModule := newFragmentsModule(f.FragmentCache, fragPath, pagePath)
	ud := L.NewUserData()
	ud.Value = fragmentsModule
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentModuleTypeName))
	L.SetGlobal("fragments", ud)

	return L
}
