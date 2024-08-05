package inject

import (
	"errors"
	"fmt"
	"reflect"
)

// A ServiceResolver resolves instances of a requested type.
type ServiceResolver interface {

	// Resolve provides an instance of the requested type if one is registered. Implementations
	// MUST ensure that the values returned are assignable to the requested type.
	Resolve(reflect.Type) (any, error)
}

// Resolve obtains an instance of the requested type from a [ServiceResolver]. An error is returned
// when the [ServiceResolver] returns an error and when the value returned by the [ServiceResolver]
// is not assignable to T.
func Resolve[T any](resolver ServiceResolver) (T, error) {
	var zero T
	if resolver == nil {
		return zero, errors.New("cannot resolve instances from nil ServiceResolver")
	}
	type_ := reflect.TypeFor[T]()
	resolved, err := resolver.Resolve(type_)
	if err != nil {
		return zero, err
	}
	typed, ok := resolved.(T)
	if !ok {
		return typed, fmt.Errorf("ServiceResolver returned %T when %T was requested", resolved, zero)
	}
	return typed, nil
}

// MustResolve obtains an intance of the requested type from a [ServiceResolver] and panics when
// the [ServiceResolver] returns an error and when the value returned by the [ServiceResolver] is
// not assignable to T.
func MustResolve[T any](resolver ServiceResolver) T {
	service, err := Resolve[T](resolver)
	if err != nil {
		panic(err)
	}
	return service
}
