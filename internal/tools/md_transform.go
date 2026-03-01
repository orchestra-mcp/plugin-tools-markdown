package tools

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/parser"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MdTransformSchema returns the JSON Schema for the md_transform tool.
func MdTransformSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"markdown": map[string]any{
				"type":        "string",
				"description": "Markdown string to transform",
			},
			"transforms": map[string]any{
				"type":        "array",
				"description": "List of transforms to apply: add_heading_ids, resolve_relative_links, strip_html, normalize_whitespace",
				"items": map[string]any{
					"type": "string",
					"enum": []any{"add_heading_ids", "resolve_relative_links", "strip_html", "normalize_whitespace"},
				},
			},
			"base_url": map[string]any{
				"type":        "string",
				"description": "Base URL for resolve_relative_links transform",
			},
		},
		"required": []any{"markdown", "transforms"},
	})
	return s
}

// MdTransform returns a tool handler that applies transforms to markdown.
func MdTransform() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "markdown"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		md := helpers.GetString(req.Arguments, "markdown")
		transforms := helpers.GetStringSlice(req.Arguments, "transforms")
		baseURL := helpers.GetString(req.Arguments, "base_url")

		if len(transforms) == 0 {
			return helpers.ErrorResult("validation_error", "at least one transform is required"), nil
		}

		result := md
		applied := make([]string, 0, len(transforms))

		for _, t := range transforms {
			switch t {
			case "add_heading_ids":
				result = transformAddHeadingIDs(result)
				applied = append(applied, t)
			case "resolve_relative_links":
				if baseURL == "" {
					return helpers.ErrorResult("validation_error", "base_url is required for resolve_relative_links transform"), nil
				}
				var err error
				result, err = transformResolveRelativeLinks(result, baseURL)
				if err != nil {
					return helpers.ErrorResult("transform_error", fmt.Sprintf("resolve_relative_links: %v", err)), nil
				}
				applied = append(applied, t)
			case "strip_html":
				result = transformStripHTML(result)
				applied = append(applied, t)
			case "normalize_whitespace":
				result = transformNormalizeWhitespace(result)
				applied = append(applied, t)
			default:
				return helpers.ErrorResult("validation_error", fmt.Sprintf("unknown transform: %s", t)), nil
			}
		}

		output := map[string]any{
			"markdown":           result,
			"transforms_applied": applied,
		}

		resp, err := helpers.JSONResult(output)
		if err != nil {
			return helpers.ErrorResult("internal_error", err.Error()), nil
		}
		return resp, nil
	}
}

// transformAddHeadingIDs adds HTML anchor IDs to all headings based on their text.
func transformAddHeadingIDs(md string) string {
	source := []byte(md)
	entries := parser.ExtractTOC(source)

	lines := strings.Split(md, "\n")
	var result []string

	headingIdx := 0
	headingPattern := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	for _, line := range lines {
		matches := headingPattern.FindStringSubmatch(line)
		if matches != nil && headingIdx < len(entries) {
			id := slugifyHeading(entries[headingIdx].Text)
			// Replace the heading line with one that includes an anchor
			result = append(result, fmt.Sprintf("%s %s {#%s}", matches[1], matches[2], id))
			headingIdx++
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// transformResolveRelativeLinks resolves relative URLs against a base URL.
func transformResolveRelativeLinks(md string, baseURLStr string) (string, error) {
	base, err := url.Parse(baseURLStr)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Match markdown links: [text](url) and images: ![alt](url)
	linkPattern := regexp.MustCompile(`(!?\[[^\]]*\])\(([^)]+)\)`)

	result := linkPattern.ReplaceAllStringFunc(md, func(match string) string {
		parts := linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		linkText := parts[1]
		linkURL := parts[2]

		// Split off any title from the URL (e.g., `url "title"`)
		title := ""
		if idx := strings.Index(linkURL, " \""); idx >= 0 {
			title = linkURL[idx:]
			linkURL = linkURL[:idx]
		}

		parsed, err := url.Parse(linkURL)
		if err != nil || parsed.IsAbs() {
			return match // leave absolute URLs and unparseable URLs alone
		}

		resolved := base.ResolveReference(parsed)
		return fmt.Sprintf("%s(%s%s)", linkText, resolved.String(), title)
	})

	return result, nil
}

// transformStripHTML removes HTML blocks and inline HTML from markdown.
func transformStripHTML(md string) string {
	// Remove HTML block-level tags (lines that start with <)
	htmlBlockPattern := regexp.MustCompile(`(?m)^<[^>]+>.*</[^>]+>\s*$`)
	result := htmlBlockPattern.ReplaceAllString(md, "")

	// Remove self-closing HTML tags
	selfClosingPattern := regexp.MustCompile(`(?m)^<[^>]+/>\s*$`)
	result = selfClosingPattern.ReplaceAllString(result, "")

	// Remove inline HTML tags
	inlineHTMLPattern := regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	result = inlineHTMLPattern.ReplaceAllString(result, "")

	return result
}

// transformNormalizeWhitespace normalizes excessive whitespace in markdown.
func transformNormalizeWhitespace(md string) string {
	lines := strings.Split(md, "\n")
	var result []string
	consecutiveBlank := 0

	for _, line := range lines {
		// Trim trailing whitespace from each line
		line = strings.TrimRight(line, " \t")

		if strings.TrimSpace(line) == "" {
			consecutiveBlank++
			// Allow at most one blank line
			if consecutiveBlank <= 1 {
				result = append(result, "")
			}
		} else {
			consecutiveBlank = 0
			result = append(result, line)
		}
	}

	output := strings.Join(result, "\n")

	// Ensure file ends with exactly one newline
	output = strings.TrimRight(output, "\n") + "\n"

	return output
}

// slugifyHeading converts heading text to a URL-safe ID.
func slugifyHeading(text string) string {
	var buf bytes.Buffer
	for _, r := range strings.ToLower(text) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			buf.WriteRune(r)
		} else if r == ' ' || r == '-' {
			buf.WriteRune('-')
		}
	}
	result := buf.String()
	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")
	// Collapse multiple hyphens
	multiHyphen := regexp.MustCompile(`-{2,}`)
	result = multiHyphen.ReplaceAllString(result, "-")
	return result
}
