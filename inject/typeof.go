package inject

import (
	"reflect"
)

// TypeOf returns the [reflect.Type] of T.
//
// The [reflect.TypeOf] function returns the [reflect.Type] of the VALUE given rather than the type
// of the expresion. For example, if a variable of an interface type holds nil and that variable is
// passed to [reflect.TypeOf] then the returned [reflect.Type] will be nil, the type of the value,
// instead of the interface type, the type of the expression.
func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}
