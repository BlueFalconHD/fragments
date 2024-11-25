package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"strings"
)

type Node interface {
	Evaluate(f *Fragment, L *lua.LState) (string, error)
}

type TextNode struct {
	Text string
}

func (n *TextNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	return n.Text, nil
}

type MetaReferenceNode struct {
	Key string
}

func (n *MetaReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	value := f.LocalMeta.v[n.Key]
	if value == nil {
		return "", fmt.Errorf("Meta key not found: %s", n.Key)
	}
	return value.stringRepresentation(), nil
}

type BuilderReferenceNode struct {
	Name    string
	Content string // Empty if no content
}

func (n *BuilderReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	builder := f.Builders.v[n.Name]
	if builder == nil {
		return "", fmt.Errorf("Builder not found: %s", n.Name)
	}

	// If content is provided, push it onto the Lua stack
	if n.Content != "" {
		L.Push(lua.LString(n.Content))
	}

	err := L.CallByParam(lua.P{
		Fn:      builder.luaType(L),
		NRet:    1,
		Protect: true,
	}, lua.LString(n.Content))

	if err != nil {
		return "", fmt.Errorf("Error calling builder function: %s, error: %v", n.Name, err)
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)
	gret := luaToCoreType(ret)
	return gret.stringRepresentation(), nil
}

type FragmentReferenceNode struct {
	Name    string
	Content string // Empty if no content
}

func (n *FragmentReferenceNode) Evaluate(f *Fragment, L *lua.LState) (string, error) {
	childFragment := f.NewChildFragmentFromName(n.Name)
	if n.Content != "" {
		return childFragment.WithContent(n.Content, f), nil
	}
	return childFragment.Evaluate(), nil
}

func ParseCode(code string) ([]Node, error) {
	lexer := NewLexer(code)
	var nodes []Node

	for tok := lexer.NextToken(); tok.Type != TOKEN_EOF; tok = lexer.NextToken() {
		switch tok.Type {
		case TOKEN_TEXT:
			nodes = append(nodes, &TextNode{Text: tok.Literal})
		case TOKEN_ESCAPED_CHAR:
			nodes = append(nodes, &TextNode{Text: tok.Literal})
		case TOKEN_META_REF:
			key, err := parseReference(lexer)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &MetaReferenceNode{Key: key})
		case TOKEN_BUILDER_REF:
			name, content, err := parseReferenceWithContent(lexer)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &BuilderReferenceNode{Name: name, Content: content})
		case TOKEN_FRAGMENT_REF:
			name, content, err := parseReferenceWithContent(lexer)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &FragmentReferenceNode{Name: name, Content: content})
		default:
			return nil, fmt.Errorf("Unknown token type: %s", tok.Type)
		}
	}

	return nodes, nil
}

func parseReference(lexer *Lexer) (string, error) {
	var key strings.Builder
	braceCount := 1

	for {
		tok := lexer.NextToken()
		if tok.Type == TOKEN_OPEN_BRACE {
			braceCount++
		} else if tok.Type == TOKEN_CLOSE_BRACE {
			braceCount--
			if braceCount == 0 {
				break
			}
		} else if tok.Type == TOKEN_EOF {
			return "", fmt.Errorf("Unexpected EOF while parsing reference")
		}
		key.WriteString(tok.Literal)
	}

	return key.String(), nil
}

func parseReferenceWithContent(lexer *Lexer) (string, string, error) {
	var nameBuilder strings.Builder
	var contentBuilder strings.Builder
	braceCount := 1
	inContent := false

	for {
		tok := lexer.NextToken()
		if tok.Type == TOKEN_OPEN_BRACE && !inContent {
			braceCount++
			nameBuilder.WriteString(tok.Literal)
		} else if tok.Type == TOKEN_CLOSE_BRACE {
			braceCount--
			if braceCount == 0 {
				break
			}
			if !inContent {
				nameBuilder.WriteString(tok.Literal)
			} else {
				contentBuilder.WriteString(tok.Literal)
			}
		} else if tok.Type == TOKEN_DOUBLE_OPEN_BRACKET {
			inContent = true
		} else if tok.Type == TOKEN_DOUBLE_CLOSE_BRACKET {
			inContent = false
		} else if tok.Type == TOKEN_EOF {
			return "", "", fmt.Errorf("Unexpected EOF while parsing reference")
		} else {
			if inContent {
				contentBuilder.WriteString(tok.Literal)
			} else {
				nameBuilder.WriteString(tok.Literal)
			}
		}
	}

	return strings.TrimSpace(nameBuilder.String()), contentBuilder.String(), nil
}
