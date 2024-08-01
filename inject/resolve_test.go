package inject

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {

	t.Run("requests the type of T from the ServiceResolver", func(t *testing.T) {
		var asFooer fooer
		// If we just take the TypeOf defaultT directly we'll get nil for interface types but if we
		// get the element (pointed to) type of a pointer to T we'll get T the actual T.
		expected := reflect.TypeOf(&asFooer).Elem()
		sp := mockResolver{}
		sp.returns(&assignableToFooer{}, nil)
		_, _ = Resolve[fooer](&sp)
		if sp.requestedTypes[0] != expected {
			t.Fatalf("expected %v; got %v", expected, sp.requestedTypes[0])
		}
	})

	t.Run("returns errors from the ServiceResolver", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		sp := mockResolver{}
		sp.returns(0, expectedErr)
		if _, actualErr := Resolve[int](&sp); !errors.Is(actualErr, expectedErr) {
			t.Fatalf("expected %v; got %v", expectedErr, actualErr)
		}
	})

	t.Run("returns error when the returned value is not assignable to requested to type", func(t *testing.T) {
		sp := mockResolver{}
		sp.returns(0, nil)
		if _, err := Resolve[string](&sp); err == nil {
			t.Fatal("expected error; got <nil>")
		}
	})

	t.Run("returns resolved value when assignable to requested type", func(t *testing.T) {
		expected := &assignableToFooer{}
		sp := mockResolver{}
		sp.returns(expected, nil)
		actual, err := Resolve[fooer](&sp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual != expected {
			t.Fatalf("expected %v; got %v", expected, actual)
		}
	})
}

type fooer interface {
	Foo()
}

type assignableToFooer struct{}

func (assignableToFooer) Foo() {}

type mockResolver struct {
	returnValues []struct {
		v   any
		err error
	}
	requestedTypes []reflect.Type
}

func (m *mockResolver) returns(v any, err error) {
	m.returnValues = append(m.returnValues, struct {
		v   any
		err error
	}{
		v:   v,
		err: err,
	})
}

func (m *mockResolver) Resolve(type_ reflect.Type) (any, error) {
	m.requestedTypes = append(m.requestedTypes, type_)
	if len(m.returnValues) == 0 {
		panic("no return values configured")
	}
	r := m.returnValues[0]
	m.returnValues = m.returnValues[1:]
	return r.v, r.err
}
