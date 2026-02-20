package command

import (
	"github.com/arfrfrr/arngit/internal/core"
	"github.com/arfrfrr/arngit/internal/ui"
)

// Context holds the execution context for a command.
type Context struct {
	Engine *core.Engine
	UI     *ui.Renderer
	Args   []string
	Flags  map[string]interface{}
}

// GetFlag returns a flag value.
func (c *Context) GetFlag(name string) interface{} {
	if c.Flags == nil {
		return nil
	}
	return c.Flags[name]
}

// GetStringFlag returns a string flag value.
func (c *Context) GetStringFlag(name string, defaultVal string) string {
	if v := c.GetFlag(name); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// GetBoolFlag returns a bool flag value.
func (c *Context) GetBoolFlag(name string) bool {
	if v := c.GetFlag(name); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetIntFlag returns an int flag value.
func (c *Context) GetIntFlag(name string, defaultVal int) int {
	if v := c.GetFlag(name); v != nil {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return defaultVal
}
