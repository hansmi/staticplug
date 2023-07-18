# Static plugins for Go binaries

[![Latest release](https://img.shields.io/github/v/release/hansmi/staticplug)][releases]
[![CI workflow](https://github.com/hansmi/staticplug/actions/workflows/ci.yaml/badge.svg)](https://github.com/hansmi/staticplug/actions/workflows/ci.yaml)
[![Go reference](https://pkg.go.dev/badge/github.com/hansmi/staticplug.svg)](https://pkg.go.dev/github.com/hansmi/staticplug)

The `staticplug` Go module implements compile-time plugins. Each plugin is
a type that gets compiled into the executable. Once the executable is built its
plugins and their functionality is fixed.

Modules wanting to make use of plugins need to create a registry:

```golang
var Registry = staticplug.NewRegistry()
```

Plugins register with the registry as part of the process initialization:

```golang
func init() {
  pkg.Registry.MustRegister(&myPlugin{})
}
```

The extendable module discovers plugins using implemented interfaces:

```golang
func Do() {
  plugins, err := Registry.PluginsImplementing((*Calculator)(nil))
  …
  for _, info := range plugins {
    p, err := info.New()
    …
    log.Printf("sum = %d", p.(Calculator).Add(1, 2))
  }
}
```

[releases]: https://github.com/hansmi/staticplug/releases/latest

<!-- vim: set sw=2 sts=2 et : -->
