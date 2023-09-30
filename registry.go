package staticplug

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/exp/slices"
)

type registeredPlugin struct {
	info  PluginInfo
	ptype reflect.Type
}

type Registry struct {
	mu      sync.RWMutex
	plugins []*registeredPlugin
	byName  map[string]*registeredPlugin
}

// NewRegistry instantiates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		byName: map[string]*registeredPlugin{},
	}
}

// Register adds a plugin to the registry. A new plugin instance is created in
// the course of the registration to validate the produced type.
func (r *Registry) Register(p Plugin) error {
	info := p.PluginInfo()

	if info.Name == "" {
		return fmt.Errorf("%w: plugin name missing", os.ErrInvalid)
	}

	pluginType := reflect.TypeOf(p)

	if inst, err := info.New(); err != nil {
		return fmt.Errorf("instantiating plugin %q: %w", info.Name, err)
	} else if instType := reflect.TypeOf(inst); pluginType != instType {
		return fmt.Errorf("%w: instantiating plugin %q returned type %v, want %v", os.ErrInvalid, info.Name, instType, pluginType)
	}

	rp := &registeredPlugin{
		info:  info,
		ptype: pluginType,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byName[info.Name]; ok {
		return fmt.Errorf("%w: plugin %q already registered", os.ErrInvalid, info.Name)
	}

	r.plugins = append(r.plugins, rp)
	r.byName[info.Name] = rp

	slices.SortStableFunc(r.plugins, func(a, b *registeredPlugin) int {
		// TODO: Switch to cmp.Compare.
		if a.info.Priority < b.info.Priority {
			return -1
		} else if a.info.Priority > b.info.Priority {
			return +1
		}

		return strings.Compare(a.info.Name, b.info.Name)
	})

	return nil
}

// MustRegister registers a plugin by receiving a plain and empty value of the
// plugin, i.e. without full initialization. In most cases the plugin package
// should invoke this function in its "init" function.
//
//	func init() {
//		registry.MustRegister(&myPlugin{})
//	}
func (r *Registry) MustRegister(p Plugin) {
	must0(r.Register(p))
}

// PluginNames returns the names of all registered plugins.
func (r *Registry) PluginNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))

	for _, rp := range r.plugins {
		names = append(names, rp.info.Name)
	}

	return names
}

// Plugins returns all registered plugins.
func (r *Registry) Plugins() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]PluginInfo, 0, len(r.plugins))

	for _, rp := range r.plugins {
		result = append(result, rp.info)
	}

	return result
}

// PluginByName returns a plugin by its name.
func (r *Registry) PluginByName(name string) (PluginInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rp, ok := r.byName[name]

	return rp.info, ok
}

// PluginsImplementing returns all plugins implementing a particular interface
// type. [TypeOfInterface] is used to determine the interface reflection type
// of the interface type.
func (r *Registry) PluginsImplementing(iface any) ([]PluginInfo, error) {
	ifaceType, err := TypeOfInterface(iface)
	if err != nil {
		return nil, err
	}

	var result []PluginInfo

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rp := range r.plugins {
		if rp.ptype.Implements(ifaceType) {
			result = append(result, rp.info)
		}
	}

	return result, nil
}
