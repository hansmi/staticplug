package staticplug

// Plugin is the interface shared by all plugin implementations. Most plugins
// will implement additional interfaces for specific functionality.
type Plugin interface {
	PluginInfo() PluginInfo
}

// PluginInfo represents a registered plugin.
type PluginInfo struct {
	Name string

	// New returns a new instance of the plugin.
	New func() (Plugin, error)
}
