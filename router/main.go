package router

import "net/http"

type App struct{}

type ServiceLifetime int

const (
	Singleton ServiceLifetime = iota
	Scoped
	Transient
)

// RegisterType registers the type of the given T as the concrete type to satisfy the type
// parameter T when instances are resolved by the target [App]. After the instance is resolved,
// every exported field will be initialized by the [App]. Note that the given instance of T is not
// used even for types registered with Singleton lifetime.
func RegisterType[T any](*App, ServiceLifetime, T) {}

// RegisterFunc registers the given function as the factory to satisfy the type parameter T when
// instances are resolved by the target [App]. Instances resolved by functions will not have their
// exported fields populated.
func RegisterFunc[T any](*App, ServiceLifetime, func(*App) T) {}

// Resolve gets an instance of T from the given [App].
func Resolve[T any](*App) T {
	return *new(T)
}

// A Middlware is a stage in a pipeline that handles an [http.Request]. Each Middleware is wrapped
// around the next [http.Handler] in the pipeline to produce a new [http.Handler] which has the
// option to perform work before and after invoking the next [http.Handler], or skpping the
// remainder of the pipeline in cases such as an authorization middleware the determines a request
// is not allowed.
type Middleware interface {
	Wrap(next http.Handler) http.Handler
}

// A MiddwareFunc is a function that directly implements the [Middleware] interface.
type MiddlewareFunc func(next http.Handler) http.Handler

func (fn MiddlewareFunc) Wrap(next http.Handler) http.Handler {
	return fn(next)
}

// A MiddlewareBuilder builds instances of Middleware.
type MiddlewareBuilder interface {

	// Build builds a Middleware.
	Build() Middleware
}

// Add registers a MiddlewareBuilder to build a Middleware for the request pipeline of the [App].
func (a *App) Add(ServiceLifetime, MiddlewareBuilder) *App {
	return a
}

func Example() {

	a := &App{}

	RegisterType[Fooer](a, Singleton, &Oof{})

	RegisterType(a, Scoped, &Barer{})

	RegisterFunc(a, Transient, func(a *App) Bazer {
		return &Zab{
			Fooer: Resolve[Fooer](a),
		}
	})

	RegisterType[Raygo](a, Singleton, &Ogyar{})
	RegisterType(a, Scoped, &AuthzContextSetter{})
	RegisterFunc(a, Scoped, func(a *App) AuthzContxtProvider {
		return Resolve[*AuthzContextSetter](a)
	})

	a.Add(Singleton, &MetricsMiddleware{}).
		Add(Scoped, &AuthzMiddleware{}).
		Add(Singleton, &TracingMiddleware{})
}

type MetricsMiddleware struct{}

func (a *MetricsMiddleware) Build() Middleware {
	return nil
}

type TracingMiddleware struct{}

func (a *TracingMiddleware) Build() Middleware {
	return nil
}

type AuthzMiddleware struct {
	Raygo              Raygo
	AuthzContextSetter AuthzContextSetter
}

func (a *AuthzMiddleware) Build() Middleware {
	return nil
}

type Raygo interface {
	Allowed(*http.Request) (bool, error)
}

type Ogyar struct{}

func (r *Ogyar) Allowed(*http.Request) (bool, error) {
	return true, nil
}

type AuthzContext struct {
	UserID string
	Roles  []string
}

type AuthzContxtProvider interface {
	Get() AuthzContext
}

type AuthzContextSetter struct {
	authzContext AuthzContext
}

func (a *AuthzContextSetter) Set(authzContext AuthzContext) {
	a.authzContext = authzContext
}

func (a *AuthzContextSetter) Get() AuthzContext {
	return a.authzContext
}

type Fooer interface {
	Foo()
}

type Oof struct{}

func (*Oof) Foo() {}

type Barer struct{}

type Bazer interface {
	Baz()
}

type Zab struct {
	Fooer Fooer
}

func (*Zab) Baz() {}
