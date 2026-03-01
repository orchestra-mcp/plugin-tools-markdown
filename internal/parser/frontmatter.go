package parser

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// ExtractFrontmatter splits a markdown document into frontmatter (YAML) and body.
// Returns frontmatter string, body string, and whether frontmatter was found.
func ExtractFrontmatter(source []byte) (string, string, bool) {
	s := string(source)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return "", s, false
	}

	// Find closing ---
	rest := s[4:] // skip opening ---\n
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", s, false
	}

	frontmatter := rest[:idx]
	body := rest[idx+4:] // skip \n---
	// Strip leading newline from body
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimPrefix(body, "\r\n")

	return frontmatter, body, true
}

// RenderPlaintext strips all markdown formatting and returns plain text.
func RenderPlaintext(source []byte) string {
	_, body, _ := ExtractFrontmatter(source)
	node := Parse([]byte(body))
	var buf bytes.Buffer
	extractText(node, []byte(body), &buf)
	return buf.String()
}

func extractText(node ast.Node, source []byte, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *ast.Text:
		buf.Write(n.Text(source))
		if n.SoftLineBreak() || n.HardLineBreak() {
			buf.WriteByte('\n')
		}
	case *ast.String:
		buf.Write(n.Value)
	case *ast.FencedCodeBlock:
		buf.Write(n.Text(source))
		buf.WriteByte('\n')
	case *ast.CodeSpan:
		buf.Write(n.Text(source))
	default:
		// Block-level nodes need paragraph spacing
		if node.Type() == ast.TypeBlock && node.ChildCount() > 0 {
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				extractText(child, source, buf)
			}
			if _, ok := node.(*ast.Paragraph); ok {
				buf.WriteByte('\n')
			}
			return
		}
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		extractText(child, source, buf)
	}
}
