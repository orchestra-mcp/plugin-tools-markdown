package toolsmarkdown

import (
	"github.com/orchestra-mcp/plugin-tools-markdown/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all markdown tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)
}
