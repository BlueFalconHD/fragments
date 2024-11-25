package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// RenderMarkdownToHTML renders a GFM markdown string to HTML
func RenderMarkdownToHTML(markdown string) (string, error) {
	// Create a Goldmark instance with GFM extensions
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

	// Prepare a buffer to store the generated HTML
	var buf bytes.Buffer

	// Convert Markdown to HTML
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
