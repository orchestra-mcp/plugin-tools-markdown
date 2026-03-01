package internal

import (
	"github.com/orchestra-mcp/plugin-tools-markdown/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ToolsPlugin registers all markdown parsing tools.
type ToolsPlugin struct{}

// RegisterTools registers all 8 markdown tools with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	builder.RegisterTool("md_parse",
		"Parse markdown string into typed JSON AST",
		tools.MdParseSchema(), tools.MdParse())

	builder.RegisterTool("md_parse_file",
		"Parse markdown file into AST (reads from path)",
		tools.MdParseFileSchema(), tools.MdParseFile())

	builder.RegisterTool("md_parse_frontmatter",
		"Extract YAML frontmatter from markdown",
		tools.MdParseFrontmatterSchema(), tools.MdParseFrontmatter())

	builder.RegisterTool("md_render_html",
		"Render markdown to HTML",
		tools.MdRenderHtmlSchema(), tools.MdRenderHtml())

	builder.RegisterTool("md_render_plaintext",
		"Render markdown to plain text (strip formatting)",
		tools.MdRenderPlaintextSchema(), tools.MdRenderPlaintext())

	builder.RegisterTool("md_toc",
		"Extract table of contents from markdown",
		tools.MdTocSchema(), tools.MdToc())

	builder.RegisterTool("md_lint",
		"Validate markdown against style rules",
		tools.MdLintSchema(), tools.MdLint())

	builder.RegisterTool("md_transform",
		"Apply transforms to markdown (resolve links, inject anchors)",
		tools.MdTransformSchema(), tools.MdTransform())
}
