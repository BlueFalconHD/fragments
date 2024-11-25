package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"strings"
)

// Custom error interfaces and types
type CodeError interface {
	error
	Line() int
	Column() int
}

type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

type EvaluationError struct {
	Line    int
	Column  int
	Message string
}

func (e *EvaluationError) Error() string {
	return fmt.Sprintf("Evaluation error at line %d, column %d: %s", e.Line, e.Column, e.Message)
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

func (n *TextNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
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

func (n *MetaReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	value := f.LocalMeta.v[n.Key]
	if value == nil {
		return "", &EvaluationError{
			Line:    n.Line(),
			Column:  n.Column(),
			Message: fmt.Sprintf("Metadata key not found: `%s`", n.Key),
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
			Line:    n.Line(),
			Column:  n.Column(),
			Message: fmt.Sprintf("Builder not found: %s", n.Name),
		}
	}

	var content string
	if n.Content != "" {
		// Parse and evaluate the content
		contentNodes, err := ParseCode(n.Content)
		if err != nil {
			return "", &EvaluationError{
				Line:    n.line,
				Column:  n.column,
				Message: fmt.Sprintf("Error parsing content in builder %s: %v", n.Name, err),
			}
		}
		var contentBuilder strings.Builder
		for _, node := range contentNodes {
			s, err := node.Evaluate(f, L)
			if err != nil {
				return "", &EvaluationError{
					Line:    node.Line(),
					Column:  node.Column(),
					Message: fmt.Sprintf("Error evaluating content node in builder %s: %v", n.Name, err),
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
			Line:    n.line,
			Column:  n.column,
			Message: fmt.Sprintf("Error calling builder function %s: %v", n.Name, err),
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
		contentNodes, err := ParseCode(n.Content)
		if err != nil {
			return "", &EvaluationError{
				Line:    n.line,
				Column:  n.column,
				Message: fmt.Sprintf("Error parsing content in fragment %s: %v", n.Name, err),
			}
		}
		var contentBuilder strings.Builder
		for _, node := range contentNodes {
			s, err := node.Evaluate(f, L)
			if err != nil {
				return "", &EvaluationError{
					Line:    node.Line(),
					Column:  node.Column(),
					Message: fmt.Sprintf("Error evaluating content node in fragment %s: %v", n.Name, err),
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

func ParseCode(code string) ([]Node, error) {
	lexer := NewLexer(code)
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
				Line:    tok.Line,
				Column:  tok.Column,
				Message: fmt.Sprintf("Unknown token type: %s", tok.Type),
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
				Line:    startLine,
				Column:  startColumn,
				Message: "Unexpected EOF while parsing reference",
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
			// Start of content
			content, err := parseContent(lexer, tok)
			if err != nil {
				return "", "", err
			}
			// Continue parsing until the closing brace
			for {
				tok := lexer.NextToken()
				if tok.Type == TOKEN_CLOSE_BRACE {
					braceCount--
					if braceCount == 0 {
						break
					}
				} else if tok.Type == TOKEN_EOF {
					return "", "", &ParseError{
						Line:    startLine,
						Column:  startColumn,
						Message: "Unexpected EOF while parsing reference",
					}
				} else {
					nameBuilder.WriteString(tok.Literal)
				}
			}
			return strings.TrimSpace(nameBuilder.String()), content, nil
		} else if tok.Type == TOKEN_EOF {
			return "", "", &ParseError{
				Line:    startLine,
				Column:  startColumn,
				Message: "Unexpected EOF while parsing reference",
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
	startLine := startToken.Line
	startColumn := startToken.Column

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
				Line:    startLine,
				Column:  startColumn,
				Message: "Unexpected EOF while parsing content",
			}
		} else {
			contentBuilder.WriteString(tok.Literal)
		}
	}

	return contentBuilder.String(), nil
}
