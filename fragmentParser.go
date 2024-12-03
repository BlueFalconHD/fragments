package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	lua "github.com/yuin/gopher-lua"
	"strings"
)

// Custom error interfaces and types
type CodeError interface {
	error
	Line() int
	Column() int
	Fragment() *Fragment
}

type ParseError struct {
	Line     int
	Column   int
	Message  string
	Fragment *Fragment
	Code     string // The code where the error occurred
}

type EvaluationError struct {
	Line     int
	Column   int
	Message  string
	Fragment *Fragment
	Code     string // The code where the error occurred
}

func (e *ParseError) Error() string {
	return formatError("Parse Error", e.Line, e.Column, e.Message, e.Code, e.Fragment)
}

func (e *EvaluationError) Error() string {
	return formatError("Evaluation Error", e.Line, e.Column, e.Message, e.Code, e.Fragment)
}

func formatError(errorType string, line, column int, message, code string, fragment *Fragment) string {
	var sb strings.Builder

	// Define styles using standard ANSI color names
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightRed")).
		Bold(true)
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightYellow"))
	lineNumberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightBlue"))
	errorLineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightMagenta")).
		Bold(true)
	pointerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightRed")).
		Bold(true)
	fragmentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("BrightCyan"))
	codeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("White"))

	// Error header
	header := fmt.Sprintf("%s at line %d, column %d:", errorType, line, column)
	sb.WriteString(headerStyle.Render(header) + "\n")
	sb.WriteString(messageStyle.Render("  "+message) + "\n\n")

	// Include code snippet
	lines := strings.Split(code, "\n")
	if line-1 < len(lines) && line-1 >= 0 {
		// Show lines around the error line
		startLine := line - 2
		if startLine < 0 {
			startLine = 0
		}
		endLine := line + 1
		if endLine > len(lines) {
			endLine = len(lines)
		}

		// Calculate the width needed for line numbers
		lineNumberWidth := len(fmt.Sprintf("%d", endLine))

		for i := startLine; i < endLine; i++ {
			currentLineNumber := i + 1
			linePrefix := "    "
			lineNumber := fmt.Sprintf("%*d | ", lineNumberWidth, currentLineNumber)
			if currentLineNumber == line {
				linePrefix = pointerStyle.Render(" -->")
				lineNumber = lineNumberStyle.Bold(true).Render(lineNumber)
				codeLine := errorLineStyle.Render(lines[i])
				sb.WriteString(fmt.Sprintf("%s %s%s\n", linePrefix, lineNumber, codeLine))

				// Generate pointer to column
				pointerPadding := len(fmt.Sprintf("%s %s", linePrefix, lineNumber))
				pointer := pointerStyle.Render("^")
				sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", pointerPadding+column-1), pointer))
			} else {
				lineNumber = lineNumberStyle.Render(lineNumber)
				codeLine := codeStyle.Render(lines[i])
				sb.WriteString(fmt.Sprintf("%s %s%s\n", linePrefix, lineNumber, codeLine))
			}
		}
	}

	// Include fragment stack
	sb.WriteString(fragmentStyle.Render("\nFragment stack:\n"))
	fragmentStack := getFragmentStack(fragment)
	for _, frag := range fragmentStack {
		sb.WriteString(fragmentStyle.Render(fmt.Sprintf("  In %s\n", frag.Name)))
	}

	return sb.String()
}

func getFragmentStack(f *Fragment) []*Fragment {
	var stack []*Fragment
	for f != nil {

		stack = append([]*Fragment{f}, stack...)
		f = f.Parent
	}
	return stack
}

type Node interface {
	Evaluate(f *Fragment, L *lua.LState) (string, error)
	Line() int
	Column() int
}

type TextNode struct {
	Text   string
	line   int
	column int
}

func (n *TextNode) Evaluate(_ *Fragment, _ *lua.LState) (string, error) {
	return n.Text, nil
}

func (n *TextNode) Line() int {
	return n.line
}

func (n *TextNode) Column() int {
	return n.column
}

type MetaReferenceNode struct {
	Key    string
	line   int
	column int
}

func (n *MetaReferenceNode) Evaluate(f *Fragment, _ *lua.LState) (string, error) {
	// Get the value from the fragment's shared metadata first
	value := f.SharedMeta.v[n.Key]
	if value == nil {
		// If the key is not found in the shared metadata, get it from the local metadata
		value = f.LocalMeta.v[n.Key]
	}

	if value == nil {
		return "", &EvaluationError{
			Line:     n.Line(),
			Column:   n.Column(),
			Message:  fmt.Sprintf("Metadata key not found: `%s`", n.Key),
			Fragment: f,
			Code:     f.Code,
		}
	}
	return value.stringRepresentation(), nil
}

func (n *MetaReferenceNode) Line() int {
	return n.line
}

func (n *MetaReferenceNode) Column() int {
	return n.column
}

type BuilderReferenceNode struct {
	Name    string
	Content string // Parsed and evaluated content
	line    int
	column  int
}

func (n *BuilderReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	builder := f.Builders.v[n.Name]
	if builder == nil {
		return "", &EvaluationError{
			Line:     n.Line(),
			Column:   n.Column(),
			Message:  fmt.Sprintf("Builder not found: %s", n.Name),
			Fragment: f,
			Code:     f.Code,
		}
	}

	var content string
	if n.Content != "" {
		// Parse and evaluate the content
		contentNodes, err := ParseCode(n.Content, f)
		if err != nil {
			return "", &EvaluationError{
				Line:     n.line,
				Column:   n.column,
				Message:  fmt.Sprintf("Error parsing content in builder %s: %v", n.Name, err),
				Fragment: f,
				Code:     f.Code,
			}
		}
		var contentBuilder strings.Builder
		for _, node := range contentNodes {
			s, err := node.Evaluate(f, L)
			if err != nil {

				return "", &EvaluationError{
					Line:     node.Line(),
					Column:   node.Column(),
					Message:  fmt.Sprintf("Error evaluating content node in builder %s: %v", n.Name, err),
					Fragment: f,
					Code:     f.Code,
				}
			}
			contentBuilder.WriteString(s)
		}
		content = contentBuilder.String()
	}

	// Prepare arguments for Lua function
	args := []lua.LValue{}
	if content != "" {
		args = append(args, lua.LString(content))
	}

	err := L.CallByParam(lua.P{
		Fn:      builder.luaType(L),
		NRet:    1,
		Protect: true,
	}, args...)

	if err != nil {
		return "", &EvaluationError{
			Line:     n.line,
			Column:   n.column,
			Message:  fmt.Sprintf("Error calling builder function %s: %v", n.Name, err),
			Fragment: f,
			Code:     f.Code,
		}
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)
	gret := luaToCoreType(ret)
	return gret.stringRepresentation(), nil
}

func (n *BuilderReferenceNode) Line() int {
	return n.line
}

func (n *BuilderReferenceNode) Column() int {
	return n.column
}

type FragmentReferenceNode struct {
	Name    string
	Content string // Parsed and evaluated content
	line    int
	column  int
}

func (n *FragmentReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	childFragment := f.NewChildFragmentFromName(n.Name)
	if n.Content != "" {
		// Parse and evaluate the content
		contentNodes, err := ParseCode(n.Content, f)
		if err != nil {
			return "", &EvaluationError{
				Line:     n.line,
				Column:   n.column,
				Message:  fmt.Sprintf("Error parsing content in fragment %s: %v", n.Name, err),
				Fragment: f,
				Code:     f.Code,
			}
		}
		var contentBuilder strings.Builder
		for _, node := range contentNodes {

			s, err := node.Evaluate(f, L)
			if err != nil {
				return "", &EvaluationError{
					Line:     node.Line(),
					Column:   node.Column(),
					Message:  fmt.Sprintf("Error evaluating content node in fragment %s: %v", n.Name, err),
					Fragment: f,
					Code:     f.Code,
				}
			}
			contentBuilder.WriteString(s)
		}
		content := contentBuilder.String()
		return childFragment.WithContent(content, f), nil
	}
	return childFragment.Evaluate(), nil
}

func (n *FragmentReferenceNode) Line() int {
	return n.line
}

func (n *FragmentReferenceNode) Column() int {
	return n.column
}

func ParseCode(code string, f *Fragment) ([]Node, error) {
	lexer := NewLexer(code, f)
	var nodes []Node

	for tok := lexer.NextToken(); tok.Type != TOKEN_EOF; tok = lexer.NextToken() {

		switch tok.Type {
		case TOKEN_TEXT:

			nodes = append(nodes, &TextNode{Text: tok.Literal, line: tok.Line, column: tok.Column})
		case TOKEN_ESCAPED_CHAR:

			nodes = append(nodes, &TextNode{Text: tok.Literal, line: tok.Line, column: tok.Column})
		case TOKEN_META_REF:
			key, err := parseReference(lexer, tok)

			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &MetaReferenceNode{Key: key, line: tok.Line, column: tok.Column})
		case TOKEN_BUILDER_REF:
			name, content, err := parseReferenceWithContent(lexer, tok)

			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &BuilderReferenceNode{Name: name, Content: content, line: tok.Line, column: tok.Column})
		case TOKEN_FRAGMENT_REF:
			name, content, err := parseReferenceWithContent(lexer, tok)

			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &FragmentReferenceNode{Name: name, Content: content, line: tok.Line, column: tok.Column})
		default:

			return nil, &ParseError{
				Line:     tok.Line,
				Column:   tok.Column,
				Message:  fmt.Sprintf("Unknown token type: %s", tok.Type),
				Fragment: f,
				Code:     code,
			}
		}
	}

	return nodes, nil
}

func parseReference(lexer *Lexer, startToken Token) (string, error) {
	var key strings.Builder
	braceCount := 1
	startLine := startToken.Line
	startColumn := startToken.Column

	for {

		tok := lexer.NextToken()
		if tok.Type == TOKEN_OPEN_BRACE {
			braceCount++
			key.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_CLOSE_BRACE {
			braceCount--
			if braceCount == 0 {
				break
			}
			key.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_EOF {
			return "", &ParseError{
				Line:     startLine,
				Column:   startColumn,
				Message:  "Unexpected EOF while parsing reference",
				Fragment: lexer.fragment,
				Code:     lexer.input,
			}
		} else {
			key.WriteString(tok.Literal)
		}
	}

	return key.String(), nil
}

func parseReferenceWithContent(lexer *Lexer, startToken Token) (string, string, error) {

	var nameBuilder strings.Builder
	braceCount := 1
	startLine := startToken.Line
	startColumn := startToken.Column

	for {
		tok := lexer.NextToken()

		if tok.Type == TOKEN_OPEN_BRACE {
			braceCount++

			nameBuilder.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_CLOSE_BRACE {
			braceCount--

			if braceCount == 0 {

				break
			}
			nameBuilder.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_DOUBLE_OPEN_BRACKET {

			content, err := parseContent(lexer, tok)
			if err != nil {

				return "", "", err
			}

			// Consume any remaining tokens until the closing brace
			for {
				tok := lexer.NextToken()

				if tok.Type == TOKEN_CLOSE_BRACE {
					braceCount--

					if braceCount == 0 {
						break
					}
				} else if tok.Type == TOKEN_EOF {
					return "", "", &ParseError{
						Line:     startLine,
						Column:   startColumn,
						Message:  "Unexpected EOF while parsing reference",
						Fragment: lexer.fragment,
						Code:     lexer.input,
					}
				} else {
					nameBuilder.WriteString(tok.Literal)
				}
			}
			return strings.TrimSpace(nameBuilder.String()), content, nil
		} else if tok.Type == TOKEN_EOF {

			return "", "", &ParseError{
				Line:     startLine,
				Column:   startColumn,
				Message:  "Unexpected EOF while parsing reference",
				Fragment: lexer.fragment,
				Code:     lexer.input,
			}
		} else {

			nameBuilder.WriteString(tok.Literal)
		}
	}

	return strings.TrimSpace(nameBuilder.String()), "", nil
}

func parseContent(lexer *Lexer, startToken Token) (string, error) {

	var contentBuilder strings.Builder
	bracketCount := 1

	for {
		tok := lexer.NextToken()

		if tok.Type == TOKEN_DOUBLE_OPEN_BRACKET {
			bracketCount++

			contentBuilder.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_DOUBLE_CLOSE_BRACKET {
			bracketCount--

			if bracketCount == 0 {

				break
			}
			contentBuilder.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_EOF {

			return "", &ParseError{
				Line:     startToken.Line,
				Column:   startToken.Column,
				Message:  "Unexpected EOF while parsing content",
				Fragment: lexer.fragment,
				Code:     lexer.input,
			}
		} else {

			contentBuilder.WriteString(tok.Literal)
		}
	}

	parsedContent := contentBuilder.String()

	return parsedContent, nil
}
