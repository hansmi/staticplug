package staticplug

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type fakeInterface interface {
	MyFunc() string
}

func TestTypeOfInterface(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   any
		want    reflect.Type
		wantErr error
	}{
		{
			name:    "nil",
			wantErr: os.ErrInvalid,
		},
		{
			name:    "string",
			value:   "value",
			wantErr: os.ErrInvalid,
		},
		{
			name:  "error type",
			value: (*error)(nil),
			want:  reflect.TypeOf((*error)(nil)).Elem(),
		},
		{
			name:  "custom",
			value: (*fakeInterface)(nil),
			want:  reflect.TypeOf((*fakeInterface)(nil)).Elem(),
		},
		{
			name:    "custom non-ptr",
			value:   (fakeInterface)(nil),
			wantErr: os.ErrInvalid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TypeOfInterface(tc.value)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got, cmp.Comparer(func(a, b reflect.Type) bool {
				return a == b
			})); diff != "" {
				t.Errorf("TypeOfInterface() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMustTypeOfInterface(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   any
		wantErr error
	}{
		{
			name:    "nil",
			wantErr: os.ErrInvalid,
		},
		{
			name:  "success",
			value: (*fakeInterface)(nil),
		},
		{
			name:    "non-ptr",
			value:   (fakeInterface)(nil),
			wantErr: os.ErrInvalid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				got := recover()

				if diff := cmp.Diff(tc.wantErr, got, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("Panic diff (-want +got):\n%s", diff)
				}
			}()

			MustTypeOfInterface(tc.value)
		})
	}
}
