package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdTocSchema returns the JSON Schema for the md_toc tool.
func MdTocSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to extract table of contents from",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdToc returns a tool handler that extracts a table of contents from markdown headings.
func MdToc() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		entries := parser.ExtractTOC([]byte(md))

		// Convert to slice of maps for JSON serialization
		tocItems := make([]any, len(entries))
		for i, e := range entries {
			item := map[string]any{
				"level": e.Level,
				"text":  e.Text,
			}
			if e.ID != "" {
				item["id"] = e.ID
			}
			tocItems[i] = item
		}

		result := map[string]any{
			"toc":   tocItems,
			"count": len(entries),
		}

		resp, err := helpers.JSONResult(result)
		if err != nil {
			return helpers.ErrorResult("internal_error", err.Error()), nil
		}
		return resp, nil
	}
}
