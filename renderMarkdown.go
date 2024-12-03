package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	lua "github.com/yuin/gopher-lua"
)

// RenderMarkdownToHTML renders a GFM markdown string to HTML
func RenderMarkdownToHTML(markdown string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // Enable GitHub Flavored Markdown
			extension.Footnote,      // Enable footnotes
			extension.Table,         // Enable tables
			extension.Strikethrough, // Enable strikethrough
			extension.TaskList,      // Enable task lists
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // Wrap lines as <br> tags
			html.WithUnsafe(),    // Allow unsafe HTML
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID()),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderMarkdown(L *lua.LState) int {
	if L.GetTop() < 1 {
		L.ArgError(1, "string expected")
	}

	if L.Get(1).Type() != lua.LTString {
		L.ArgError(1, "string expected")
	}

	content := L.CheckString(1)
	html, err := RenderMarkdownToHTML(content)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(NewCoreString(html).luaType(L))
	return 1
}
