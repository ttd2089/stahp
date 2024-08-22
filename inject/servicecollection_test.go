package inject

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestServiceCollection(t *testing.T) {

	t.Run("RegisterType", func(t *testing.T) {

		t.Run("scoped struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType(&services, Scoped, struct{}{})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("singleton struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType(&services, Singleton, struct{}{})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("transient struct does not return error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterType(&services, Transient, struct{}{})
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}
		})

		for _, lifetime := range []ServiceLifetime{Transient, Scoped, Singleton} {
			t.Run(fmt.Sprintf("%s pointer to struct does not return error", lifetime), func(t *testing.T) {
				services := ServiceCollection{}
				err := RegisterType(&services, lifetime, &struct{}{})
				if err != nil {
					t.Fatalf("unexpected error %q", err)
				}
			})
		}
	})

	t.Run("RegisterFunc", func(t *testing.T) {

		t.Run("scoped struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[interface{}](&services, Scoped, func(ServiceResolver) (struct{}, error) {
				return struct{}{}, nil
			})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("singleton struct returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[interface{}](&services, Singleton, func(ServiceResolver) (struct{}, error) {
				return struct{}{}, nil
			})
			if !errors.Is(err, ErrNonTransientStruct) {
				t.Fatalf("expected %q; got %q", ErrNonTransientStruct, err)
			}
		})

		t.Run("transient struct does not return error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[interface{}](&services, Transient, func(ServiceResolver) (struct{}, error) {
				return struct{}{}, nil
			})
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}
		})

		for _, lifetime := range []ServiceLifetime{Transient, Scoped, Singleton} {
			t.Run(fmt.Sprintf("%s pointer to struct does not return error", lifetime), func(t *testing.T) {
				services := ServiceCollection{}
				err := RegisterFunc[interface{}](&services, lifetime, func(ServiceResolver) (*struct{}, error) {
					return &struct{}{}, nil
				})
				if err != nil {
					t.Fatalf("unexpected error %q", err)
				}
			})
		}

		t.Run("unassignable impl returns error", func(t *testing.T) {
			services := ServiceCollection{}
			err := RegisterFunc[string](&services, Scoped, func(ServiceResolver) (*struct{}, error) {
				return &struct{}{}, nil
			})
			if !errors.Is(err, ErrInvalidImplementation) {
				t.Fatalf("expected %q; got %q", ErrInvalidImplementation, err)
			}
		})
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
						RegisterType(&services, Transient, struct{}{})
						return services
					}(),
				},
				{
					name: "from func",
					services: func() ServiceCollection {
						services := ServiceCollection{}
						RegisterFunc[struct{}](&services, Transient, func(ServiceResolver) (struct{}, error) {
							return struct{}{}, nil
						})
						return services
					}(),
				},
			}

			for _, tt := range testCases {
				t.Run(tt.name, func(t *testing.T) {
					provider, _ := tt.services.Build()
					resolved, err := provider.Resolve(reflect.TypeFor[struct{}]())
					if err != nil {
						t.Fatalf("unexpected error from ServiceProvider.Resolve: %q", err)
					}
					if _, ok := resolved.(struct{}); !ok {
						t.Fatalf("expected %q; got %q", reflect.TypeFor[struct{}](), reflect.TypeOf(resolved))
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
							RegisterType(&services, Transient, &struct{}{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*struct{}](&services, Transient, func(ServiceResolver) (*struct{}, error) {
								return &struct{}{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolved, err := provider.Resolve(reflect.TypeFor[*struct{}]())
						if err != nil {
							t.Fatalf("unexpected error from ServiceProvider.Resolve: %q", err)
						}
						instance, ok := resolved.(*struct{})
						if !ok {
							t.Fatalf("expected %q; got %q", reflect.TypeFor[*struct{}](), reflect.TypeOf(resolved))
						}
						if instance == nil {
							t.Fatalf("expected non-nil pointer; got nil")
						}
					})
				}
			})
		}

		t.Run("instance management", func(t *testing.T) {

			// distinctCapableStruct is required to observe whether pointers point to the same
			// instance or not because pointers to zero-length structs can be equal even when the
			// pointed values are distinct.
			//
			// From the Golang spec:
			// Pointer types are comparable. Two pointer values are equal if they point to the same
			// variable or if both have value nil. Pointers to distinct zero-size variables may or
			// may not be equal.
			type distinctCapableStruct bytes.Reader

			t.Run("transient instances from the same provider are distinct", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Transient, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Transient, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func() *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve()
						b := resolve()
						if a == b {
							t.Fatalf("instances are the same: %p %p", a, b)
						}
					})
				}
			})

			t.Run("scoped instances from the same provider are the same", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Scoped, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Scoped, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func() *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve()
						b := resolve()
						if a != b {
							t.Fatalf("instances are distinct: %p %p", a, b)
						}
					})
				}
			})

			t.Run("singleton instances from the same provider are the same", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Singleton, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Singleton, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func() *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve()
						b := resolve()
						if a != b {
							t.Fatalf("instances are distinct: %p %p", a, b)
						}
					})
				}
			})

			t.Run("scoped instances from child scope provider are distinct", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Scoped, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Scoped, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func(provider *ServiceProvider) *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve(&provider)
						childScope := provider.NewScope()
						b := resolve(&childScope)
						if a == b {
							t.Fatalf("instances are the same: %p %p", a, b)
						}
					})
				}
			})

			t.Run("singleton instances from child scope provider are the same", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Singleton, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Singleton, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func(provider *ServiceProvider) *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve(&provider)
						childScope := provider.NewScope()
						b := resolve(&childScope)
						if a != b {
							t.Fatalf("instances are distinct: %p %p", a, b)
						}
					})
				}
			})

			t.Run("singleton instances from grandchild scope provider are the same", func(t *testing.T) {

				testCases := []struct {
					name     string
					services ServiceCollection
				}{
					{
						name: "from type",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterType(&services, Singleton, &distinctCapableStruct{})
							return services
						}(),
					},
					{
						name: "from func",
						services: func() ServiceCollection {
							services := ServiceCollection{}
							RegisterFunc[*distinctCapableStruct](&services, Singleton, func(ServiceResolver) (*distinctCapableStruct, error) {
								return &distinctCapableStruct{}, nil
							})
							return services
						}(),
					},
				}

				for _, tt := range testCases {
					t.Run(tt.name, func(t *testing.T) {
						provider, _ := tt.services.Build()
						resolve := func(provider *ServiceProvider) *distinctCapableStruct {
							resolved, _ := provider.Resolve(reflect.TypeFor[*distinctCapableStruct]())
							return resolved.(*distinctCapableStruct)
						}
						a := resolve(&provider)
						childScope := provider.NewScope()
						grandchildScope := childScope.NewScope()
						b := resolve(&grandchildScope)
						if a != b {
							t.Fatalf("instances are distinct: %p %p", a, b)
						}
					})
				}
			})
		})
	})
}
