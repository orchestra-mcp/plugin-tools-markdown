package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/yuin/goldmark/ast"
	"google.golang.org/protobuf/types/known/structpb"
)

// LintIssue represents a single lint warning or error.
type LintIssue struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// MdLintSchema returns the JSON Schema for the md_lint tool.
func MdLintSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to validate against style rules",
			},
		},
		"required": []any{"markdown"},
	})
	return s
}

// MdLint returns a tool handler that validates markdown against style rules.
func MdLint() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		source := []byte(md)
		issues := runLintRules(source)

		// Convert to serializable format
		issueList := make([]any, len(issues))
		for i, issue := range issues {
			issueList[i] = map[string]any{
				"line":    issue.Line,
				"message": issue.Message,
				"rule":    issue.Rule,
			}
		}

		result := map[string]any{
			"issues": issueList,
			"count":  len(issues),
			"valid":  len(issues) == 0,
		}

		resp, err := helpers.JSONResult(result)
		if err != nil {
			return helpers.ErrorResult("internal_error", err.Error()), nil
		}
		return resp, nil
	}
}

func runLintRules(source []byte) []LintIssue {
	var issues []LintIssue
	lines := strings.Split(string(source), "\n")

	// Rule: trailing whitespace
	for i, line := range lines {
		if strings.TrimRight(line, " \t") != line {
			issues = append(issues, LintIssue{
				Line:    i + 1,
				Message: "Trailing whitespace",
				Rule:    "no-trailing-whitespace",
			})
		}
	}

	// Rule: missing newline at end of file
	if len(source) > 0 && source[len(source)-1] != '\n' {
		issues = append(issues, LintIssue{
			Line:    len(lines),
			Message: "Missing newline at end of file",
			Rule:    "final-newline",
		})
	}

	// Rule: consecutive blank lines (more than 2 newlines in a row = 1+ blank line is OK, 3+ newlines = bad)
	consecutiveBlank := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveBlank++
			if consecutiveBlank > 1 {
				issues = append(issues, LintIssue{
					Line:    i + 1,
					Message: "Multiple consecutive blank lines",
					Rule:    "no-consecutive-blank-lines",
				})
			}
		} else {
			consecutiveBlank = 0
		}
	}

	// Rule: heading level jumps (parse AST for this)
	node := parser.Parse(source)
	var lastHeadingLevel int
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok {
			if lastHeadingLevel > 0 && heading.Level > lastHeadingLevel+1 {
				// Find line number from AST
				line := 0
				if heading.Lines().Len() > 0 {
					seg := heading.Lines().At(0)
					line = countLines(source, seg.Start)
				}
				issues = append(issues, LintIssue{
					Line:    line,
					Message: fmt.Sprintf("Heading level jumped from h%d to h%d", lastHeadingLevel, heading.Level),
					Rule:    "no-heading-level-skip",
				})
			}
			lastHeadingLevel = heading.Level
		}
		return ast.WalkContinue, nil
	})

	// Rule: missing alt text on images
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if img, ok := n.(*ast.Image); ok {
			// Check if image has alt text (children contain the alt text)
			hasAlt := false
			for child := img.FirstChild(); child != nil; child = child.NextSibling() {
				if t, ok := child.(*ast.Text); ok {
					if len(strings.TrimSpace(string(t.Text(source)))) > 0 {
						hasAlt = true
						break
					}
				}
			}
			if !hasAlt {
				line := 0
				if img.Parent() != nil && img.Parent().Lines().Len() > 0 {
					seg := img.Parent().Lines().At(0)
					line = countLines(source, seg.Start)
				}
				issues = append(issues, LintIssue{
					Line:    line,
					Message: fmt.Sprintf("Image missing alt text: %s", string(img.Destination)),
					Rule:    "no-missing-alt-text",
				})
			}
		}
		return ast.WalkContinue, nil
	})

	return issues
}

// countLines counts the number of newlines before the given byte offset.
func countLines(source []byte, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(source); i++ {
		if source[i] == '\n' {
			line++
		}
	}
	return line
}
