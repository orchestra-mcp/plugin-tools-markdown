package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdParseFrontmatterSchema returns the JSON Schema for the md_parse_frontmatter tool.
func MdParseFrontmatterSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string with optional YAML frontmatter",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdParseFrontmatter returns a tool handler that extracts YAML frontmatter from markdown.
func MdParseFrontmatter() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		frontmatter, body, hasFrontmatter := parser.ExtractFrontmatter([]byte(md))

		result := map[string]any{
			"frontmatter":     frontmatter,
			"body":            body,
			"has_frontmatter": hasFrontmatter,
		}

		resp, err := helpers.JSONResult(result)
		if err != nil {
			return helpers.ErrorResult("internal_error", err.Error()), nil
		}
		return resp, nil
	}
}
