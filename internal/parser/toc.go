package parser

import (
	"github.com/yuin/goldmark/ast"
)

// TOCEntry represents a single heading in the table of contents.
type TOCEntry struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id,omitempty"`
}

// ExtractTOC walks the AST and extracts all headings.
func ExtractTOC(source []byte) []TOCEntry {
	node := Parse(source)
	var entries []TOCEntry

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok {
			text := extractNodeText(heading, source)
			id := ""
			if idAttr, ok := heading.AttributeString("id"); ok {
				if idStr, ok := idAttr.([]byte); ok {
					id = string(idStr)
				}
			}
			entries = append(entries, TOCEntry{
				Level: heading.Level,
				Text:  text,
				ID:    id,
			})
		}
		return ast.WalkContinue, nil
	})

	return entries
}

func extractNodeText(node ast.Node, source []byte) string {
	var text []byte
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			text = append(text, t.Text(source)...)
		} else if cs, ok := child.(*ast.CodeSpan); ok {
			text = append(text, cs.Text(source)...)
		} else {
			// Recurse into inline elements
			text = append(text, []byte(extractNodeText(child, source))...)
		}
	}
	return string(text)
}
