package inject

import (
	"reflect"
	"testing"
)

func TestTypeOf(t *testing.T) {

	testCases := []struct {
		testName    string
		expected    reflect.Type
		notExpected []reflect.Type
		actual      reflect.Type
	}{
		{
			testName: "int",
			expected: reflect.TypeOf(int(42)),
			actual:   TypeOf[int](),
		},
		{
			testName: "int alias",
			expected: reflect.TypeOf(intAlias(int(42))),
			notExpected: []reflect.Type{
				reflect.TypeOf(int(42)),
			},
			actual: TypeOf[intAlias](),
		},
		{
			testName: "pointer to int",
			expected: reflect.TypeOf(new(int)),
			actual:   TypeOf[*int](),
		},
		{
			testName: "pointer to in aliast",
			expected: reflect.TypeOf(new(intAlias)),
			notExpected: []reflect.Type{
				reflect.TypeOf(new(int)),
			},
			actual: TypeOf[*intAlias](),
		},
		{
			testName: "empty struct",
			expected: reflect.TypeOf(struct{}{}),
			actual:   TypeOf[struct{}](),
		},
		{
			testName: "struct alias",
			expected: reflect.TypeOf(emptyStructAlias{}),
			notExpected: []reflect.Type{
				reflect.TypeOf(struct{}{}),
			},
			actual: TypeOf[emptyStructAlias](),
		},
		{
			testName: "pointer to struct",
			expected: reflect.TypeOf(new(struct{})),
			actual:   TypeOf[*struct{}](),
		},
		{
			testName: "pointer to struct alias",
			expected: reflect.TypeOf(new(emptyStructAlias)),
			notExpected: []reflect.Type{
				reflect.TypeOf(new(struct{})),
			},
			actual: TypeOf[*emptyStructAlias](),
		},
		{
			testName: "interface",
			expected: reflect.TypeOf(new(interface{})).Elem(),
			notExpected: []reflect.Type{
				func() reflect.Type {
					var x interface{}
					return reflect.TypeOf(x)
				}(),
			},
			actual: TypeOf[interface{}](),
		},
		{
			testName: "interface alias",
			expected: reflect.TypeOf(new(emptyInterfaceAlias)).Elem(),
			notExpected: []reflect.Type{
				func() reflect.Type {
					var x emptyInterfaceAlias
					return reflect.TypeOf(x)
				}(),
			},
			actual: TypeOf[emptyInterfaceAlias](),
		},
		{
			testName: "pointer to interface",
			expected: reflect.TypeOf(new(interface{})),
			actual:   TypeOf[*interface{}](),
		},
		{
			testName: "pointer to interface alias",
			expected: reflect.TypeOf(new(emptyInterfaceAlias)),
			notExpected: []reflect.Type{
				reflect.TypeOf(new(interface{})),
			},
			actual: TypeOf[*emptyInterfaceAlias](),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("expected %v; got %v", tt.expected, tt.actual)
			}
			for _, notExpected := range tt.notExpected {
				if tt.actual == notExpected {
					t.Fatalf("got unexpected value %v", notExpected)
				}
			}
		})
	}
}

type intAlias int

type emptyStructAlias struct{}

type emptyInterfaceAlias interface{}
