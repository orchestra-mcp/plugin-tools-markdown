package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdRenderPlaintextSchema returns the JSON Schema for the md_render_plaintext tool.
func MdRenderPlaintextSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to convert to plain text",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdRenderPlaintext returns a tool handler that strips markdown formatting and returns plain text.
func MdRenderPlaintext() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		plaintext := parser.RenderPlaintext([]byte(md))

		return helpers.TextResult(plaintext), nil
	}
}
