package parser

import (
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// NodeToMap converts a goldmark AST node into a map suitable for JSON.
func NodeToMap(node ast.Node, source []byte) map[string]any {
	result := map[string]any{
		"type": nodeTypeName(node),
	}

	// Add node-specific attributes
	switch n := node.(type) {
	case *ast.Heading:
		result["level"] = n.Level
	case *ast.FencedCodeBlock:
		if n.Language(source) != nil {
			result["language"] = string(n.Language(source))
		}
		result["value"] = string(n.Text(source))
	case *ast.CodeSpan:
		result["value"] = string(n.Text(source))
	case *ast.Link:
		result["url"] = string(n.Destination)
		if n.Title != nil {
			result["title"] = string(n.Title)
		}
	case *ast.Image:
		result["url"] = string(n.Destination)
		if n.Title != nil {
			result["title"] = string(n.Title)
		}
	case *ast.AutoLink:
		result["url"] = string(n.URL(source))
	case *ast.Text:
		result["value"] = string(n.Text(source))
		if n.SoftLineBreak() {
			result["soft_break"] = true
		}
		if n.HardLineBreak() {
			result["hard_break"] = true
		}
	case *ast.String:
		result["value"] = string(n.Value)
	case *ast.HTMLBlock:
		// Collect all lines from the HTML block
		var raw []byte
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			raw = append(raw, seg.Value(source)...)
		}
		result["value"] = string(raw)
	case *ast.RawHTML:
		// Inline HTML
		segs := n.Segments
		var raw []byte
		for i := 0; i < segs.Len(); i++ {
			seg := segs.At(i)
			raw = append(raw, seg.Value(source)...)
		}
		result["value"] = string(raw)
	case *ast.List:
		result["ordered"] = n.IsOrdered()
		if n.IsOrdered() {
			result["start"] = n.Start
		}
	case *east.TaskCheckBox:
		result["checked"] = n.IsChecked
	case *east.Strikethrough:
		// children contain the text
	case *east.Table:
		// children are TableHeader + TableBody
	case *east.TableCell:
		result["alignment"] = n.Alignment.String()
	}

	// Add text content for leaf nodes
	if node.Type() == ast.TypeInline && node.ChildCount() == 0 {
		if _, ok := result["value"]; !ok {
			segs := node.Lines()
			var txt []byte
			for i := 0; i < segs.Len(); i++ {
				seg := segs.At(i)
				txt = append(txt, seg.Value(source)...)
			}
			if len(txt) > 0 {
				result["value"] = string(txt)
			}
		}
	}

	// Recursively add children
	if node.ChildCount() > 0 {
		children := make([]map[string]any, 0, node.ChildCount())
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, NodeToMap(child, source))
		}
		result["children"] = children
	}

	return result
}

func nodeTypeName(node ast.Node) string {
	switch node.(type) {
	case *ast.Document:
		return "document"
	case *ast.Heading:
		return "heading"
	case *ast.Paragraph:
		return "paragraph"
	case *ast.TextBlock:
		return "text_block"
	case *ast.ThematicBreak:
		return "thematic_break"
	case *ast.CodeBlock:
		return "code_block"
	case *ast.FencedCodeBlock:
		return "fenced_code_block"
	case *ast.HTMLBlock:
		return "html_block"
	case *ast.List:
		return "list"
	case *ast.ListItem:
		return "list_item"
	case *ast.Blockquote:
		return "blockquote"
	case *ast.Text:
		return "text"
	case *ast.String:
		return "string"
	case *ast.CodeSpan:
		return "code_span"
	case *ast.Emphasis:
		n := node.(*ast.Emphasis)
		if n.Level == 2 {
			return "strong"
		}
		return "emphasis"
	case *ast.Link:
		return "link"
	case *ast.Image:
		return "image"
	case *ast.AutoLink:
		return "auto_link"
	case *ast.RawHTML:
		return "raw_html"
	case *east.TaskCheckBox:
		return "task_checkbox"
	case *east.Strikethrough:
		return "strikethrough"
	case *east.Table:
		return "table"
	case *east.TableHeader:
		return "table_header"
	case *east.TableRow:
		return "table_row"
	case *east.TableCell:
		return "table_cell"
	case *east.FootnoteLink:
		return "footnote_link"
	case *east.FootnoteBacklink:
		return "footnote_backlink"
	case *east.FootnoteList:
		return "footnote_list"
	case *east.Footnote:
		return "footnote"
	default:
		return node.Kind().String()
	}
}
