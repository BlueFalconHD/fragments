package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/russross/blackfriday/v2"
)

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

	"ShellCmd": &Script{
		name: "ShellCmd",
		handler: func(f *Fragment) {
			// check if f.meta["shellCmd"] exists
			// save the code of the fragment to /tmp/<random UUID>
			// if it does, run the command with the path to the file as an argument
			// if it doesn't, return

			if cmd, exists := f.meta["shellCmd"]; exists {

				tmpfile, err := os.CreateTemp("/tmp", "shellcmd-*")
				if err != nil {
					return
				}
				defer os.Remove(tmpfile.Name())

				tmpfile.Write([]byte(f.code))
				tmpfile.Close()

				// replace all instances of ${@} with the path to the temporary file
				cmd = strings.ReplaceAll(cmd, "${@}", tmpfile.Name())

				// run the command
				out, err := exec.Command("sh", "-c", cmd).Output()

				if err != nil {
					return
				}

				// set the code of the fragment to the output of the command
				// trim leading and trailing whitespace
				f.code = strings.TrimSpace(string(out))
			}
		},
	},
}
