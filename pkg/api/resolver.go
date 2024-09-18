package api

import (
	"errors"
	"fmt"
	"reflect"
)

// The Resolver interface and Resolve functions are adapted from github.com/ttd2089/garlic. They
// are reproduced rather than referenced to avoid the dependency since we only want to provide a
// simple interface for consuming dynamically resolved values.
//
// https://github.com/ttd2089/garlic/blob/0532da5448be7eadedbc6058082d00ae855b5b0a/pkg/di/resolve.go

// ErrNilResolver is returned when a function receives a nil [Resolver].
var ErrNilResolver = errors.New("resolver must not be nil")

// A Resolver resolves instances of a requested type.
type Resolver interface {

	// Resolve provides an instance of the requested type if one is registered. Implementations
	// MUST ensure that the values returned are assignable to the requested type.
	Resolve(reflect.Type) (any, error)

	// NewScope providers a new [Resolver]. Implementations MUST return a [Resolver] that resolves
	// request-scoped values that are distinct from those resolved by any other [Resolver] and
	// reused for every resolution of the same type by itself.
	NewScope() Resolver
}

// Resolve obtains an instance of the requested type from a [Resolver]. An [error] is returned when
// the [Resolver] returns an [error] or a value that is not assignable to T.
func Resolve[T any](resolver Resolver) (T, error) {
	if resolver == nil {
		var zero T
		return zero, ErrNilResolver
	}

	var zero T
	typ := reflect.TypeFor[T]()

	resolved, err := resolver.Resolve(typ)
	if err != nil {
		return zero, fmt.Errorf("resolve: %w", err)
	}

	typed, ok := resolved.(T)
	if !ok {
		return zero, fmt.Errorf("resolve: Resolver returned %T when %T was requested", resolved, typed)
	}

	return typed, nil
}
