package parser_test

import (
	"strings"
	"testing"

	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
)

func TestParse_ReturnsDocument(t *testing.T) {
	node := parser.Parse([]byte("# Hello\n\nWorld\n"))
	if node == nil {
		t.Fatal("expected non-nil AST node")
	}
	m := parser.NodeToMap(node, []byte("# Hello\n\nWorld\n"))
	if m["type"] != "document" {
		t.Fatalf("expected document, got %s", m["type"])
	}
	children, ok := m["children"].([]map[string]any)
	if !ok || len(children) == 0 {
		t.Fatal("expected children in document")
	}
}

func TestParse_Heading(t *testing.T) {
	src := []byte("# My Heading\n")
	node := parser.Parse(src)
	m := parser.NodeToMap(node, src)
	children := m["children"].([]map[string]any)
	if children[0]["type"] != "heading" {
		t.Fatalf("expected heading, got %s", children[0]["type"])
	}
	if children[0]["level"] != 1 {
		t.Fatalf("expected level 1, got %v", children[0]["level"])
	}
}

func TestParse_FencedCodeBlock(t *testing.T) {
	src := []byte("```go\nfmt.Println()\n```\n")
	node := parser.Parse(src)
	m := parser.NodeToMap(node, src)
	children := m["children"].([]map[string]any)
	if children[0]["type"] != "fenced_code_block" {
		t.Fatalf("expected fenced_code_block, got %s", children[0]["type"])
	}
	if children[0]["language"] != "go" {
		t.Fatalf("expected language go, got %v", children[0]["language"])
	}
}

func TestParse_Link(t *testing.T) {
	src := []byte("[click](https://example.com)\n")
	node := parser.Parse(src)
	m := parser.NodeToMap(node, src)

	// walk to find the link
	var findLink func(n map[string]any) map[string]any
	findLink = func(n map[string]any) map[string]any {
		if n["type"] == "link" {
			return n
		}
		if ch, ok := n["children"].([]map[string]any); ok {
			for _, c := range ch {
				if r := findLink(c); r != nil {
					return r
				}
			}
		}
		return nil
	}
	link := findLink(m)
	if link == nil {
		t.Fatal("expected to find a link node")
	}
	if link["url"] != "https://example.com" {
		t.Fatalf("expected url https://example.com, got %v", link["url"])
	}
}

func TestRenderHTML_Basic(t *testing.T) {
	html, err := parser.RenderHTML([]byte("# Title\n\nParagraph.\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<h1") {
		t.Fatalf("expected <h1 in output, got: %s", html)
	}
	if !strings.Contains(html, "Paragraph") {
		t.Fatalf("expected Paragraph in output")
	}
}

func TestRenderHTML_GFMTable(t *testing.T) {
	src := []byte("| A | B |\n|---|---|\n| 1 | 2 |\n")
	html, err := parser.RenderHTML(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<table") {
		t.Fatalf("expected <table in output, got: %s", html)
	}
}

func TestExtractFrontmatter_Present(t *testing.T) {
	src := []byte("---\ntitle: Hello\n---\nBody text\n")
	fm, body, ok := parser.ExtractFrontmatter(src)
	if !ok {
		t.Fatal("expected frontmatter to be found")
	}
	if !strings.Contains(fm, "title: Hello") {
		t.Fatalf("expected title in frontmatter, got: %s", fm)
	}
	if !strings.Contains(body, "Body text") {
		t.Fatalf("expected body text, got: %s", body)
	}
}

func TestExtractFrontmatter_Absent(t *testing.T) {
	src := []byte("No frontmatter here.\n")
	_, body, ok := parser.ExtractFrontmatter(src)
	if ok {
		t.Fatal("expected no frontmatter")
	}
	if !strings.Contains(body, "No frontmatter") {
		t.Fatalf("expected body unchanged, got: %s", body)
	}
}

func TestRenderPlaintext(t *testing.T) {
	src := []byte("# Heading\n\n**Bold** and _italic_ text.\n")
	plain := parser.RenderPlaintext(src)
	if !strings.Contains(plain, "Bold") {
		t.Fatalf("expected Bold in plaintext, got: %s", plain)
	}
	if strings.Contains(plain, "**") {
		t.Fatalf("expected ** stripped from plaintext, got: %s", plain)
	}
}

func TestExtractTOC(t *testing.T) {
	src := []byte("# Title\n\n## Section One\n\n### Subsection\n\n## Section Two\n")
	entries := parser.ExtractTOC(src)
	if len(entries) != 4 {
		t.Fatalf("expected 4 TOC entries, got %d", len(entries))
	}
	if entries[0].Level != 1 || entries[0].Text != "Title" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Level != 2 || entries[1].Text != "Section One" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
	if entries[2].Level != 3 {
		t.Fatalf("expected level 3, got %d", entries[2].Level)
	}
}
