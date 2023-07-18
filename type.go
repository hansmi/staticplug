package staticplug

import (
	"fmt"
	"os"
	"reflect"
)

// TypeOfInterface returns the reflection type representing the dynamic type of
// iface. Construct a nil value via (*Iface)(nil).
//
// Values of [reflect.Type] are used directly.
//
// Logically this function is equivalent to the following:
//
//	reflect.TypeOf((*Iface)(nil)).Elem()
func TypeOfInterface(iface any) (reflect.Type, error) {
	if t, ok := iface.(reflect.Type); ok {
		if t != nil && t.Kind() == reflect.Interface {
			return t, nil
		}

		return nil, fmt.Errorf("%w: type must be an interface, got %+v", os.ErrInvalid, t)
	}

	var msgKind string

	t := reflect.TypeOf(iface)

	if t != nil {
		if t.Kind() == reflect.Ptr {
			if t := t.Elem(); t.Kind() == reflect.Interface {
				return t, nil
			}
		}

		msgKind = fmt.Sprintf(" (kind %v)", t.Kind())
	}

	return nil, fmt.Errorf("%w: pointer to interface type is required, got %+v%s", os.ErrInvalid, t, msgKind)
}

// MustTypeOfInterface wraps [TypeOfInterface], converting all errors to
// panics.
func MustTypeOfInterface(iface any) reflect.Type {
	return must1(TypeOfInterface(iface))
}
