package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdRenderHtmlSchema returns the JSON Schema for the md_render_html tool.
func MdRenderHtmlSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to render as HTML",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdRenderHtml returns a tool handler that renders markdown to HTML.
func MdRenderHtml() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		html, err := parser.RenderHTML([]byte(md))
		if err != nil {
			return helpers.ErrorResult("render_error", err.Error()), nil
		}

		return helpers.TextResult(html), nil
	}
}
