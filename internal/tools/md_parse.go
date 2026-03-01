package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdParseSchema returns the JSON Schema for the md_parse tool.
func MdParseSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to parse into AST",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdParse returns a tool handler that parses markdown into a typed JSON AST.
func MdParse() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		source := []byte(md)
		node := parser.Parse(source)
		astMap := parser.NodeToMap(node, source)

		resp, err := helpers.JSONResult(astMap)
		if err != nil {
			return helpers.ErrorResult("parse_error", err.Error()), nil
		}
		return resp, nil
	}
}
