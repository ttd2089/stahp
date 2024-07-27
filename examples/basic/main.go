// This is an example of the most basic use of the [github.com/ttd2089/stahp] package.
//
// The business logic of the application is encapsulated in an app type and the route targets are
// its methods.
//
// Request parsing and response writing is handled inline where the routes are instantiated. A more
// realistic use of the package would include DI, generic parser builders, declarative validation,
// middleware, etc., but the purpose of this example is to show the core functionality of the
// package with no distractions.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/ttd2089/stahp"
	"golang.org/x/exp/maps"
)

func main() {

	a := app{}

	http.Handle(
		"POST /users/{$}",
		stahp.Route(
			// Our target function `addUser` takes a `createUserReq` and returns a `user`...
			a.addUser,
			// ... so we need a parser to get a `createUserReq` from the request...
			func(r *http.Request) (createUserReq, error) {
				var req createUserReq
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&req); err != nil {
					return createUserReq{}, err
				}
				return req, nil
			},
			// ... and a responder to handle the `user` or any errors that occur.
			stahp.NewResponder(
				// A responder needs to handle the value returned from the target function,...
				func(w http.ResponseWriter, resp user) {
					w.WriteHeader(http.StatusCreated)
					enc := json.NewEncoder(w)
					if err := enc.Encode(resp); err != nil {
						log.Printf("[ERROR] %s\n", err)
					}
				},
				// ... any error returned from the parser,...
				func(w http.ResponseWriter, parseErr error) {
					http.Error(w, parseErr.Error(), http.StatusBadRequest)
				},
				// ... and any error returned from the target function.
				func(w http.ResponseWriter, err error) {
					if errors.Is(err, errUserAlreadyExists) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			),
		),
	)

	http.Handle(
		"GET /users/{$}",
		stahp.Route(
			a.getUsers,
			stahp.NoReqParser,
			stahp.NewResponder(
				func(w http.ResponseWriter, resp []userSummary) {
					w.WriteHeader(http.StatusOK)
					enc := json.NewEncoder(w)
					if err := enc.Encode(resp); err != nil {
						log.Printf("[ERROR] %s\n", err)
					}
				},
				func(w http.ResponseWriter, parseErr error) {
					log.Printf("[CRITICAL] unexpected parse error from NoReqTarget GET /users: %s", parseErr)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				},
				func(w http.ResponseWriter, err error) {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			),
		),
	)

	http.Handle(
		"GET /users/{id}",
		stahp.Route(
			a.getUser,
			func(r *http.Request) (int, error) {
				id, err := strconv.Atoi(r.PathValue("id"))
				if err != nil {
					return 0, errUserNotFound
				}
				return id, nil
			},
			stahp.NewResponder(
				func(w http.ResponseWriter, resp user) {
					w.WriteHeader(http.StatusOK)
					enc := json.NewEncoder(w)
					if err := enc.Encode(resp); err != nil {
						log.Printf("[ERROR] %s\n", err)
					}
				},
				func(w http.ResponseWriter, parseErr error) {
					if errors.Is(parseErr, errUserNotFound) {
						http.Error(w, parseErr.Error(), http.StatusNotFound)
						return
					}
					http.Error(w, parseErr.Error(), http.StatusInternalServerError)
				},
				func(w http.ResponseWriter, err error) {
					if errors.Is(err, errUserNotFound) {
						http.Error(w, err.Error(), http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			),
		),
	)

	http.Handle(
		"POST /users/{user_id}/posts/{$}",
		stahp.Route(
			a.addPost,
			func(r *http.Request) (createPostReq, error) {
				id, err := strconv.Atoi(r.PathValue("user_id"))
				if err != nil {
					return createPostReq{}, errUserNotFound
				}
				req := createPostReq{
					userID: id,
				}
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&req.post); err != nil {
					return createPostReq{}, err
				}
				return req, nil
			},
			stahp.NewResponder(
				func(w http.ResponseWriter, resp post) {
					w.WriteHeader(http.StatusOK)
					enc := json.NewEncoder(w)
					if err := enc.Encode(resp); err != nil {
						log.Printf("[ERROR] %s\n", err)
					}
				},
				func(w http.ResponseWriter, parseErr error) {
					if errors.Is(parseErr, errUserNotFound) {
						http.Error(w, parseErr.Error(), http.StatusNotFound)
						return
					}
					http.Error(w, parseErr.Error(), http.StatusInternalServerError)
				},
				func(w http.ResponseWriter, err error) {
					if errors.Is(err, errUserNotFound) {
						http.Error(w, err.Error(), http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			),
		),
	)

	log.Printf("[INFO] listening on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Printf("fatal: %s\n", err.Error())
	}
}

type createUserReq struct {
	Name string `json:"name"`
}

type user struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Posts []post `json:"posts"`
}

type userSummary struct {
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Posts []postSummary `json:"posts"`
}

type createPostReq struct {
	userID int
	post   struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
}

type post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type postSummary struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type userRecord struct {
	name  string
	posts map[int]post
}

var errUserAlreadyExists = errors.New("user already exists")
var errUserNotFound = errors.New("user not found")
var errUserPostAlreadyExists = errors.New("user post already exists")

type app struct {
	db        map[int]userRecord
	userIDSeq int
	postIDSeq int
	mu        sync.Mutex
}

func (a *app) addUser(_ context.Context, req createUserReq) (user, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, existing := range a.db {
		if req.Name == existing.name {
			return user{}, errUserAlreadyExists
		}
	}
	a.userIDSeq++
	if a.db == nil {
		a.db = make(map[int]userRecord)
	}
	u := user{
		ID:    a.userIDSeq,
		Name:  req.Name,
		Posts: make([]post, 0),
	}
	a.db[u.ID] = userRecord{
		name:  u.Name,
		posts: make(map[int]post),
	}
	return u, nil
}

func (a *app) getUsers(_ context.Context, _ struct{}) ([]userSummary, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	summaries := make([]userSummary, 0, len(a.db))
	for uid, u := range a.db {
		postSummaries := make([]postSummary, 0, len(u.posts))
		for pid, p := range u.posts {
			postSummaries = append(postSummaries, postSummary{
				ID:    pid,
				Title: p.Title,
			})
		}
		summaries = append(summaries, userSummary{
			ID:    uid,
			Name:  u.name,
			Posts: postSummaries,
		})
	}
	return summaries, nil
}

func (a *app) getUser(_ context.Context, id int) (user, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if u, ok := a.db[id]; ok {
		return user{
			ID:    id,
			Name:  u.name,
			Posts: maps.Values(u.posts),
		}, nil
	}
	return user{}, errUserNotFound
}

func (a *app) addPost(_ context.Context, req createPostReq) (post, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	u, ok := a.db[req.userID]
	if !ok {
		return post{}, errUserNotFound
	}
	for _, existing := range u.posts {
		if req.post.Title == existing.Title {
			return post{}, errUserPostAlreadyExists
		}
	}
	a.postIDSeq++
	p := post{
		ID:    a.postIDSeq,
		Title: req.post.Title,
		Body:  req.post.Body,
	}
	u.posts[p.ID] = p
	a.db[req.userID] = u
	return p, nil
}
