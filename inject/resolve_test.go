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
		resolver := mockResolver{}
		resolver.returns(&assignableToFooer{}, nil)
		_, _ = Resolve[fooer](&resolver)
		if resolver.requestedTypes[0] != expected {
			t.Fatalf("expected %v; got %v", expected, resolver.requestedTypes[0])
		}
	})

	t.Run("returns errors from the ServiceResolver", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		resolver := mockResolver{}
		resolver.returns(0, expectedErr)
		if _, actualErr := Resolve[int](&resolver); !errors.Is(actualErr, expectedErr) {
			t.Fatalf("expected %v; got %v", expectedErr, actualErr)
		}
	})

	t.Run("returns error when the returned value is not assignable to requested to type", func(t *testing.T) {
		resolver := mockResolver{}
		resolver.returns(0, nil)
		if _, err := Resolve[string](&resolver); err == nil {
			t.Fatal("expected error; got <nil>")
		}
	})

	t.Run("returns resolved value when assignable to requested type", func(t *testing.T) {
		expected := &assignableToFooer{}
		resolver := mockResolver{}
		resolver.returns(expected, nil)
		actual, err := Resolve[fooer](&resolver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual != expected {
			t.Fatalf("expected %v; got %v", expected, actual)
		}
	})
}

type mockResolver struct {
	returnValues []struct {
		v   any
		err error
	}
	requestedTypes []reflect.Type
}

func (mock *mockResolver) returns(v any, err error) {
	mock.returnValues = append(mock.returnValues, struct {
		v   any
		err error
	}{
		v:   v,
		err: err,
	})
}

func (mock *mockResolver) Resolve(type_ reflect.Type) (any, error) {
	mock.requestedTypes = append(mock.requestedTypes, type_)
	if len(mock.returnValues) == 0 {
		panic("no return values configured")
	}
	r := mock.returnValues[0]
	mock.returnValues = mock.returnValues[1:]
	return r.v, r.err
}
