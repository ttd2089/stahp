package inject

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestServiceCollection(t *testing.T) {

	t.Run("RegisterType", func(t *testing.T) {

		t.Run("scoped struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType[interface{}](&services, Scoped, structWithUnexportedFields{})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("singleton struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType[interface{}](&services, Singleton, structWithUnexportedFields{})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("transient struct does not return error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType[interface{}](&services, Transient, structWithUnexportedFields{})
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}
		})

		for _, lifetime := range []ServiceLifetime{Transient, Scoped, Singleton} {
			t.Run(fmt.Sprintf("%s pointer to struct does not return error", lifetime), func(t *testing.T) {
				services := ServiceCollection{}
				err := RegisterType[interface{}](&services, lifetime, &structWithUnexportedFields{})
				if err != nil {
					t.Fatalf("unexpected error %q", err)
				}
			})
		}
	})

	t.Run("RegisterFunc", func(t *testing.T) {

		t.Run("unassignable impl returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[fooer](&services, Scoped, func(ServiceResolver) (*structWithUnexportedFields, error) {
				return &structWithUnexportedFields{}, nil
			})
			if !errors.Is(err, ErrInvalidImplementation) {
				t.Fatalf("expected %q; got %q", ErrInvalidImplementation, err)
			}
		})

		t.Run("scoped struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[fooer](&services, Scoped, func(ServiceResolver) (assignableToFooer, error) {
				return assignableToFooer{}, nil
			})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("singleton struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[fooer](&services, Singleton, func(ServiceResolver) (assignableToFooer, error) {
				return assignableToFooer{}, nil
			})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("transient struct does not return error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[fooer](&services, Transient, func(ServiceResolver) (assignableToFooer, error) {
				return assignableToFooer{}, nil
			})
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}
		})

		for _, lifetime := range []ServiceLifetime{Transient, Scoped, Singleton} {
			t.Run(fmt.Sprintf("%s pointer to struct does not return error", lifetime), func(t *testing.T) {
				services := ServiceCollection{}
				err := RegisterFunc[fooer](&services, lifetime, func(ServiceResolver) (*assignableToFooer, error) {
					return &assignableToFooer{}, nil
				})
				if err != nil {
					t.Fatalf("unexpected error %q", err)
				}
			})
		}
	})

	t.Run("Resolve", func(t *testing.T) {

		t.Run("transient struct is resolved", func(t *testing.T) {

			testCases := []struct {
				name     string
				services ServiceCollection
			}{
				{
					name: "from type",
					services: func() ServiceCollection {
						services := ServiceCollection{}
						RegisterType(&services, Transient, structWithUnexportedFields{})
						return services
					}(),
				},
				{
					name: "from func",
					services: func() ServiceCollection {
						services := ServiceCollection{}
						RegisterFunc[structWithUnexportedFields](&services, Transient, func(ServiceResolver) (structWithUnexportedFields, error) {
							return structWithUnexportedFields{}, nil
						})
						return services
					}(),
				},
			}

			for _, tt := range testCases {
				t.Run(tt.name, func(t *testing.T) {
					provider, _ := tt.services.Build()
					resolved, err := provider.Resolve(reflect.TypeFor[structWithUnexportedFields]())
					if err != nil {
						t.Fatalf("unexpected error from ServiceProvider.Resolve: %q", err)
					}
					if _, ok := resolved.(structWithUnexportedFields); !ok {
						t.Fatalf("expected %q; got %q", reflect.TypeFor[structWithUnexportedFields](), reflect.TypeOf(resolved))
					}
				})
			}
		})

		for _, lifetime := range []ServiceLifetime{Transient, Scoped, Singleton} {
			t.Run(fmt.Sprintf("%s pointer to struct is resolved with value", lifetime), func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Transient, &structWithUnexportedFields{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*structWithUnexportedFields](&services, Transient, func(ServiceResolver) (*structWithUnexportedFields, error) {
								return &structWithUnexportedFields{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolved, err := provider.Resolve(reflect.TypeFor[*structWithUnexportedFields]())
						if err != nil {
							t.Fatalf("unexpected error from ServiceProvider.Resolve: %q", err)
						}
						instance, ok := resolved.(*structWithUnexportedFields)
						if !ok {
							t.Fatalf("expected %q; got %q", reflect.TypeFor[*structWithUnexportedFields](), reflect.TypeOf(resolved))
						}
						if instance == nil {
							t.Fatalf("expected non-nil pointer; got nil")
						}
					})
				}
			})
		}

		t.Run("transient instances from the same provider are distinct", func(t *testing.T) {
			services := ServiceCollection{}
			RegisterType(&services, Transient, &structWithUnexportedFields{})
			provider, _ := services.Build()
			resolve := func() *structWithUnexportedFields {
				resolved, _ := provider.Resolve(reflect.TypeFor[*structWithUnexportedFields]())
				return resolved.(*structWithUnexportedFields)
			}
			a := resolve()
			b := resolve()
			if a == b {
				t.Fatalf("transient instances are the same: %p %p", a, b)
			}
		})

		t.Run("scoped instances from the same provider are the same", func(t *testing.T) {
			services := ServiceCollection{}
			RegisterType(&services, Scoped, &structWithUnexportedFields{})
			provider, _ := services.Build()
			resolve := func() *structWithUnexportedFields {
				resolved, _ := provider.Resolve(reflect.TypeFor[*structWithUnexportedFields]())
				return resolved.(*structWithUnexportedFields)
			}
			a := resolve()
			b := resolve()
			if a != b {
				t.Fatalf("transient instances are the distinct: %p %p", a, b)
			}
		})
	})
}
