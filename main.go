package main

import (
	"github.com/charmbracelet/log"
)

type metaMap map[string]string
type fragmentMap map[string]*Fragment

const otherFc = `
this:setTemplate("page")

print(string.sub(this.name, 1, 2))

---

@{../page/index}
`

func testLua() {

	fcache := make(FragmentCache)

	// Print the keys of the fragment cache
	for k := range fcache {
		log.Info("Fragment cache key", "key", k)
	}

	pof := &Fragment{
		Name:          "other",
		Code:          otherFc,
		Depth:         0,
		Parent:        nil,
		LocalMeta:     *NewEmptyCoreTable(),
		SharedMeta:    NewEmptyCoreTable(),
		Builders:      NewEmptyCoreTable(),
		FragmentCache: &fcache,
	}

	log.Info("Output of evaluation of other", "result", pof.Evaluate())

}

func main() {
	testLua()
}
