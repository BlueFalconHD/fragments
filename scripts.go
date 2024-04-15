package main

import "github.com/russross/blackfriday/v2"

type Script struct {
	name    string
	handler func(f *Fragment)
}

func (s *Script) run(f *Fragment) {
	s.handler(f)
}

var scripts = map[string]*Script{
	"RenderMarkdown": &Script{
		name: "RenderMarkdown",
		handler: func(f *Fragment) {
			f.code = string(blackfriday.Run([]byte(f.code)))
		},
	},
}
