package tools_test

import (
	"context"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/tools"
	"google.golang.org/protobuf/types/known/structpb"
)

// makeArgs builds a structpb.Struct from a plain map.
func makeArgs(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("makeArgs: %v", err)
	}
	return s
}

// resultStr returns the string value at Result[key].
func resultStr(t *testing.T, resp *pluginv1.ToolResponse, key string) string {
	t.Helper()
	if resp.Result == nil {
		t.Fatalf("result is nil (ErrorCode=%s ErrorMessage=%s)", resp.ErrorCode, resp.ErrorMessage)
	}
	v, ok := resp.Result.Fields[key]
	if !ok {
		t.Fatalf("result missing key %q (keys: %v)", key, resp.Result.Fields)
	}
	return v.GetStringValue()
}

// resultFloat returns a float64 from Result[key].
func resultFloat(t *testing.T, resp *pluginv1.ToolResponse, key string) float64 {
	t.Helper()
	if resp.Result == nil {
		t.Fatalf("result is nil")
	}
	v, ok := resp.Result.Fields[key]
	if !ok {
		t.Fatalf("result missing key %q", key)
	}
	return v.GetNumberValue()
}

// resultBool returns a bool from Result[key].
func resultBool(t *testing.T, resp *pluginv1.ToolResponse, key string) bool {
	t.Helper()
	if resp.Result == nil {
		t.Fatalf("result is nil")
	}
	v, ok := resp.Result.Fields[key]
	if !ok {
		t.Fatalf("result missing key %q", key)
	}
	return v.GetBoolValue()
}

// ---- md_parse ----

func TestMdParse_BasicDocument(t *testing.T) {
	fn := tools.MdParse()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Hello\n\nWorld\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	nodeType := resultStr(t, resp, "type")
	if nodeType != "document" {
		t.Fatalf("expected document, got %s", nodeType)
	}
}

func TestMdParse_MissingField(t *testing.T) {
	fn := tools.MdParse()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected error response for missing markdown field")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected validation_error, got %s", resp.ErrorCode)
	}
}

// ---- md_render_html ----

func TestMdRenderHTML_Basic(t *testing.T) {
	fn := tools.MdRenderHtml()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Hi\n\nParagraph.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	html := resultStr(t, resp, "text")
	if !strings.Contains(html, "<h1") {
		t.Fatalf("expected <h1 in html, got: %s", html)
	}
	if !strings.Contains(html, "Paragraph") {
		t.Fatalf("expected Paragraph in html")
	}
}

func TestMdRenderHTML_GFMTable(t *testing.T) {
	fn := tools.MdRenderHtml()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "| A | B |\n|---|---|\n| 1 | 2 |\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	html := resultStr(t, resp, "text")
	if !strings.Contains(html, "<table") {
		t.Fatalf("expected <table in output, got: %s", html)
	}
}

// ---- md_render_plaintext ----

func TestMdRenderPlaintext(t *testing.T) {
	fn := tools.MdRenderPlaintext()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "**Bold** and _italic_ text.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	text := resultStr(t, resp, "text")
	if strings.Contains(text, "**") {
		t.Fatalf("expected ** stripped, got: %s", text)
	}
	if !strings.Contains(text, "Bold") {
		t.Fatalf("expected Bold in plaintext, got: %s", text)
	}
}

// ---- md_parse_frontmatter ----

func TestMdParseFrontmatter_Present(t *testing.T) {
	fn := tools.MdParseFrontmatter()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "---\ntitle: Test\nauthor: Alice\n---\nBody content\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if !resultBool(t, resp, "has_frontmatter") {
		t.Fatal("expected has_frontmatter=true")
	}
	fm := resultStr(t, resp, "frontmatter")
	if !strings.Contains(fm, "title: Test") {
		t.Fatalf("expected title in frontmatter, got: %s", fm)
	}
}

func TestMdParseFrontmatter_Absent(t *testing.T) {
	fn := tools.MdParseFrontmatter()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "Just content, no frontmatter.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if resultBool(t, resp, "has_frontmatter") {
		t.Fatal("expected has_frontmatter=false")
	}
}

// ---- md_toc ----

func TestMdToc_ExtractsHeadings(t *testing.T) {
	fn := tools.MdToc()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Title\n\n## Section A\n\n## Section B\n\n### Sub\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	count := resultFloat(t, resp, "count")
	if count != 4 {
		t.Fatalf("expected 4 TOC entries, got %v", count)
	}
}

func TestMdToc_NoHeadings(t *testing.T) {
	fn := tools.MdToc()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "No headings here.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	count := resultFloat(t, resp, "count")
	if count != 0 {
		t.Fatalf("expected 0 TOC entries, got %v", count)
	}
}

// ---- md_lint ----

func TestMdLint_CleanDoc(t *testing.T) {
	fn := tools.MdLint()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Title\n\nParagraph.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if !resultBool(t, resp, "valid") {
		t.Fatal("expected valid=true for clean doc")
	}
}

func TestMdLint_TrailingWhitespace(t *testing.T) {
	fn := tools.MdLint()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Title  \n\nParagraph.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if resultBool(t, resp, "valid") {
		t.Fatal("expected valid=false for trailing whitespace")
	}
}

func TestMdLint_MissingFinalNewline(t *testing.T) {
	fn := tools.MdLint()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "# Title\n\nNo trailing newline",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if resultBool(t, resp, "valid") {
		t.Fatal("expected valid=false for missing final newline")
	}
}

func TestMdLint_ConsecutiveBlankLines(t *testing.T) {
	fn := tools.MdLint()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown": "Para one.\n\n\nPara two.\n",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	if resultBool(t, resp, "valid") {
		t.Fatal("expected valid=false for consecutive blank lines")
	}
}

// ---- md_transform ----

func TestMdTransform_NormalizeWhitespace(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "Line with trailing space   \n\n\n\nAnother line\n",
		"transforms": []any{"normalize_whitespace"},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	result := resultStr(t, resp, "markdown")
	if strings.Contains(result, "   \n") {
		t.Fatalf("expected trailing spaces stripped, got: %q", result)
	}
}

func TestMdTransform_StripHTML(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "Text with <b>bold</b> tag.\n",
		"transforms": []any{"strip_html"},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	result := resultStr(t, resp, "markdown")
	if strings.Contains(result, "<b>") {
		t.Fatalf("expected <b> stripped, got: %s", result)
	}
}

func TestMdTransform_ResolveRelativeLinks(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "[link](docs/page.md)\n",
		"transforms": []any{"resolve_relative_links"},
		"base_url":   "https://example.com/repo/",
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	result := resultStr(t, resp, "markdown")
	if !strings.Contains(result, "https://example.com") {
		t.Fatalf("expected absolute URL in result, got: %s", result)
	}
}

func TestMdTransform_ResolveRelativeLinks_MissingBaseURL(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "[link](docs/page.md)\n",
		"transforms": []any{"resolve_relative_links"},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected error when base_url missing for resolve_relative_links")
	}
}

func TestMdTransform_NoTransforms(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "Hello\n",
		"transforms": []any{},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected error when no transforms provided")
	}
}

func TestMdTransform_UnknownTransform(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "Hello\n",
		"transforms": []any{"bogus_transform"},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected error for unknown transform")
	}
}

func TestMdTransform_AddHeadingIDs(t *testing.T) {
	fn := tools.MdTransform()
	req := &pluginv1.ToolRequest{Arguments: makeArgs(t, map[string]any{
		"markdown":   "# My Heading\n\n## Sub Section\n",
		"transforms": []any{"add_heading_ids"},
	})}
	resp, err := fn(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.ErrorMessage)
	}
	result := resultStr(t, resp, "markdown")
	if !strings.Contains(result, "{#") {
		t.Fatalf("expected heading IDs in result, got: %s", result)
	}
}
