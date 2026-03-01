package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdParseFileSchema returns the JSON Schema for the md_parse_file tool.
func MdParseFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the markdown file to parse",
			},
		},
		"required": []any{"path"},
	})
	return s
}

// MdParseFile returns a tool handler that reads a markdown file and parses it into AST.
func MdParseFile() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")

		// Clean and validate path
		path = filepath.Clean(path)
		if !filepath.IsAbs(path) {
			return helpers.ErrorResult("validation_error", "path must be absolute"), nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read file: %v", err)), nil
		}

		node := parser.Parse(source)
		astMap := parser.NodeToMap(node, source)

		result := map[string]any{
			"path": path,
			"ast":  astMap,
		}

		resp, err := helpers.JSONResult(result)
		if err != nil {
			return helpers.ErrorResult("parse_error", err.Error()), nil
		}
		return resp, nil
	}
}
