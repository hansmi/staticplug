package staticplug

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var errPluginFailure = errors.New("plugin failed")

type fakeCalculator interface {
	add(int, int) int
}

type fakePlugin struct {
	name string
	prio int
	inst Plugin
}

var _ Plugin = (*fakePlugin)(nil)

func (p *fakePlugin) PluginInfo() PluginInfo {
	return PluginInfo{
		Name:     p.name,
		Priority: p.prio,
		New: func() (Plugin, error) {
			if p.inst == nil {
				return &fakePlugin{}, nil
			}

			return p.inst, nil
		},
	}
}

type calcPlugin struct{}

var _ Plugin = (*calcPlugin)(nil)
var _ fakeCalculator = (*calcPlugin)(nil)

func (p *calcPlugin) PluginInfo() PluginInfo {
	return PluginInfo{
		Name: "calc",
		New: func() (Plugin, error) {
			return &calcPlugin{}, nil
		},
	}
}

func (p *calcPlugin) add(a, b int) int {
	return a + b
}

type failingPlugin struct{}

var _ Plugin = (*failingPlugin)(nil)

func (p *failingPlugin) PluginInfo() PluginInfo {
	return PluginInfo{
		Name: "failing",
		New: func() (Plugin, error) {
			return nil, errPluginFailure
		},
	}
}

func TestRegistryRegister(t *testing.T) {
	for _, tc := range []struct {
		name      string
		reg       *Registry
		p         Plugin
		wantErr   error
		wantNames []string
	}{
		{
			name:      "simple",
			p:         &fakePlugin{name: "fake"},
			wantNames: []string{"fake"},
		},
		{
			name:    "missing name",
			p:       &fakePlugin{},
			wantErr: os.ErrInvalid,
		},
		{
			name: "wrong type",
			p: &fakePlugin{
				name: "bad type",
				inst: &calcPlugin{},
			},
			wantErr: os.ErrInvalid,
		},
		{
			name:    "instantiation fails",
			p:       &failingPlugin{},
			wantErr: errPluginFailure,
		},
		{
			name: "already registered",
			reg: func() *Registry {
				r := NewRegistry()
				if err := r.Register(&fakePlugin{name: "test"}); err != nil {
					t.Fatal(err)
				}
				return r
			}(),
			p: &fakePlugin{
				name: "test",
			},
			wantErr:   os.ErrInvalid,
			wantNames: []string{"test"},
		},
		{
			name: "multiple",
			reg: func() *Registry {
				r := NewRegistry()
				for _, name := range []string{"first", "second"} {
					if err := r.Register(&fakePlugin{name: name}); err != nil {
						t.Fatal(err)
					}
				}
				return r
			}(),
			p: &fakePlugin{
				name: "another",
			},
			wantNames: []string{"another", "first", "second"},
		},
		{
			name: "priority",
			reg: func() *Registry {
				r := NewRegistry()
				r.MustRegister(&fakePlugin{name: "hundred", prio: 100})
				r.MustRegister(&fakePlugin{name: "twenty", prio: 20})
				r.MustRegister(&fakePlugin{name: "minus", prio: -1})
				return r
			}(),
			p: &fakePlugin{
				name: "another",
			},
			wantNames: []string{"minus", "another", "twenty", "hundred"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.reg

			if r == nil {
				r = NewRegistry()
			}

			err := r.Register(tc.p)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if got, want := len(r.Plugins()), len(tc.wantNames); got != want {
				t.Errorf("Got %d plugins, want %d", got, want)
			}

			if diff := cmp.Diff(tc.wantNames, r.PluginNames(), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Plugin name diff (-want +got):\n%s", diff)
			}

			for _, name := range tc.wantNames {
				if got, ok := r.PluginByName(name); !ok {
					t.Errorf("Plugin %q not found", name)
				} else if diff := cmp.Diff(name, got.Name); diff != "" {
					t.Errorf("Plugin name diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestRegistryMustRegister(t *testing.T) {
	defer func() {
		got := recover()

		if got == nil || !errors.Is(got.(error), errPluginFailure) {
			t.Errorf("MustRegister() didn't fail as expected, got %v", got)
		}
	}()

	r := NewRegistry()
	r.MustRegister(&failingPlugin{})
}

func TestRegistryPluginsImplementing(t *testing.T) {
	reg := NewRegistry()

	for _, p := range []Plugin{
		&fakePlugin{name: "first"},
		&fakePlugin{name: "second"},
		&calcPlugin{},
	} {
		if err := reg.Register(p); err != nil {
			t.Fatalf("register(%+v) failed: %v", p, err)
		}
	}

	for _, tc := range []struct {
		name      string
		iface     any
		wantErr   error
		wantNames []string
	}{
		{
			name:    "nil",
			wantErr: os.ErrInvalid,
		},
		{
			name:    "bad type, TypeOf",
			iface:   reflect.TypeOf(""),
			wantErr: os.ErrInvalid,
		},
		{
			name:    "bad type, string",
			iface:   "",
			wantErr: os.ErrInvalid,
		},
		{
			name:  "no match",
			iface: reflect.TypeOf((*interface{ notImplemented() })(nil)).Elem(),
		},
		{
			name:      "plugin",
			iface:     reflect.TypeOf((*Plugin)(nil)).Elem(),
			wantNames: []string{"calc", "first", "second"},
		},
		{
			name:      "one match",
			iface:     reflect.TypeOf((*fakeCalculator)(nil)).Elem(),
			wantNames: []string{"calc"},
		},
		{
			name:      "one match",
			iface:     (*fakeCalculator)(nil),
			wantNames: []string{"calc"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := reg.PluginsImplementing(tc.iface)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			var names []string

			for _, p := range got {
				names = append(names, p.Name)
			}

			sort.Strings(names)

			if diff := cmp.Diff(tc.wantNames, names, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Plugin name diff (-want +got):\n%s", diff)
			}
		})
	}
}
