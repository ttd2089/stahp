package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/ttd2089/garlic/pkg/di"
	"github.com/ttd2089/stahp/pkg/api"
)

type garlicRootProvider struct {
	provider di.RootProvider
}

func (p garlicRootProvider) Resolve(typ reflect.Type) (any, error) {
	return p.provider.Resolve(typ)
}

func (p garlicRootProvider) NewScope() api.Resolver {
	return garlicScope{
		scope: p.provider.NewScope(),
	}
}

type garlicScope struct {
	scope di.Scope
}

func (p garlicScope) Resolve(typ reflect.Type) (any, error) {
	return p.scope.Resolve(typ)
}

func (p garlicScope) NewScope() api.Resolver {
	return garlicScope{
		scope: p.scope.NewScope(),
	}
}

type Greeter interface {
	Greet(name string) string
}

type greeter struct {
	greeting string
}

func (g greeter) Greet(name string) string {
	greeting := g.greeting
	if greeting == "" {
		greeting = "Hello"
	}
	return fmt.Sprintf("%s, %s!", greeting, name)
}

type GreetingController struct {
	Greeter Greeter
}

type GreetingRequest struct {
	Name string `json:"name"`
}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func (g *GreetingController) HandleGreeting(ctx context.Context, req GreetingRequest) (GreetingResponse, error) {
	return GreetingResponse{
		Greeting: g.Greeter.Greet(req.Name),
	}, nil
}

type TimeRequest struct {
	Please bool `json:"please"`
}

type TimeResponse struct {
	Time time.Time `json:"time"`
}

type JSONHandler[Req any, Resp any] func(context.Context, Req) (Resp, error)

func (h JSONHandler[Req, Resp]) Endpoint(api.Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req
		dec := json.NewDecoder(r.Body)
		dec.Decode(&req)
		resp, err := h(r.Context(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "handler returned error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to serialize response JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	})
}

type JSONControllerAction[C any, Req any, Resp any] func(controller C) JSONHandler[Req, Resp]

func (action JSONControllerAction[C, Req, Resp]) Endpoint(resolver api.Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllerType := reflect.TypeFor[C]()
		resolved, err := resolver.Resolve(controllerType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve controller type %t: %v", controllerType, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		controller, ok := resolved.(C)
		if !ok {
			fmt.Fprintf(os.Stderr, "failed to resolve controller: resolver returned %T when %T was requested", resolved, controller)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handler := action(controller)
		endpoint := handler.Endpoint(resolver)
		endpoint.ServeHTTP(w, r)
	})
}

func main() {

	provider, err := buildProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build provider: %v", err)
		os.Exit(1)
		return
	}

	apiBuilder := new(api.Builder)

	apiBuilder.AddEndpoint("/time", JSONHandler[TimeRequest, TimeResponse](getTime))

	apiBuilder.AddEndpoint(
		"/greetings",
		JSONControllerAction[*GreetingController, GreetingRequest, GreetingResponse](
			func(controller *GreetingController) JSONHandler[GreetingRequest, GreetingResponse] {
				return controller.HandleGreeting
			},
		),
	)

	host, err := apiBuilder.Build(provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building application host: %v", err)
		os.Exit(1)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		host.Stop(ctx)
	}()

	if err := host.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "application host exited: %v", err)
		os.Exit(1)
	}
}

func getTime(_ context.Context, req TimeRequest) (TimeResponse, error) {
	now := time.Now()
	if !req.Please {
		skewForBeingRude := (rand.Int() % 15) - 7
		now = now.Add(time.Duration(skewForBeingRude) * time.Second)
	}
	return TimeResponse{
		Time: now,
	}, nil
}

func buildProvider() (garlicRootProvider, error) {

	registry := di.Registry{}

	registry, err := di.RegisterFactory[Greeter](registry, di.Transient, func(di.Resolver) (greeter, error) {
		return greeter{
			greeting: "Hello",
		}, nil
	})
	if err != nil {
		return garlicRootProvider{}, fmt.Errorf("register Greeter: %w", err)
	}

	registry, err = di.RegisterType[*GreetingController, *GreetingController](registry, di.Transient)
	if err != nil {
		return garlicRootProvider{}, fmt.Errorf("register GreetingController: %w", err)
	}

	rootProvider, err := registry.BuildRootProvider()
	if err != nil {
		return garlicRootProvider{}, fmt.Errorf("build root provider: %w", err)
	}

	return garlicRootProvider{
		provider: rootProvider,
	}, nil
}
