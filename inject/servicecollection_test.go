package inject

import (
	"reflect"
	"testing"
)

func TestRegisterType(t *testing.T) {
	// The effect of registering a type is observed through its resolution so the tests here will
	// tend to involve setting up one or more registrations, building a service provider, then
	// resolving the target service to ensure correctness.

	t.Run("register transient struct", func(t *testing.T) {
		sc := ServiceCollection{}
		RegisterType(&sc, Transient, structWithNoExportedFields{})
		sp, err := sc.Build()
		if err != nil {
			t.Fatalf("unexpected error building ServiceProvider: %v", err)
		}
		resolved, err := sp.Resolve(TypeOf[structWithNoExportedFields]())
		if err != nil {
			t.Fatalf("unexpected error resolving service: %v", err)
		}
		if _, ok := resolved.(structWithNoExportedFields); !ok {
			t.Fatalf(
				"resolved service has unexpected type: want %v; got %v",
				TypeOf[structWithNoExportedFields](),
				reflect.TypeOf(resolved))
		}
	})
}

type structWithNoExportedFields struct{}
