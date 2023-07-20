package staticplug

// Plugin is the interface shared by all plugin implementations. Most plugins
// will implement additional interfaces for specific functionality.
type Plugin interface {
	PluginInfo() PluginInfo
}

// PluginInfo represents a registered plugin.
type PluginInfo struct {
	// Plugin name. Must not be empty.
	Name string

	// Sorting prioritiy. Within the same priority plugins are sorted by name.
	Priority int

	// New returns a new instance of the plugin.
	New func() (Plugin, error)
}
