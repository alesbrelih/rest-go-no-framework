package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

var pathRegex *regexp.Regexp

type Todo struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type todoHandlers struct {
	mu    sync.Mutex // good practice to keep mutex near the data its trying to protect
	store map[int64]Todo
}

func (h *todoHandlers) todos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		break
	case "POST":
		h.post(w, r)
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *todoHandlers) todo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getOne(w, r)
		break
	case "DELETE":
		h.delete(w, r)
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *todoHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("application/json required"))
		return
	}

	var todo Todo
	err = json.Unmarshal(bodyBytes, &todo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock() // i like this

	h.store[todo.Id] = todo
}

func (h *todoHandlers) get(w http.ResponseWriter, r *http.Request) {
	todos := make([]Todo, len(h.store))

	h.mu.Lock()
	i := 0
	for _, todo := range h.store {
		todos[i] = todo
		i++
	}
	h.mu.Unlock()

	jsonBytes, err := json.Marshal(todos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jsonBytes)
}

func (h *todoHandlers) getOne(w http.ResponseWriter, r *http.Request) {

	if !pathRegex.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid path, should be /todo/{id}"))
		return
	}

	idParam := pathRegex.FindStringSubmatch(r.URL.Path)[1]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Bad stuff happened"))
		return
	}

	h.mu.Lock()
	todo, found := h.store[id]
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Todo with %v does not exist", id)))
		return
	}
	h.mu.Unlock()

	jsonBytes, err := json.Marshal(todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Bad stuff happened"))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jsonBytes)
}

func (h *todoHandlers) delete(w http.ResponseWriter, r *http.Request) {

	if !pathRegex.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid path, should be /todo/{id}"))
		return
	}

	idParam := pathRegex.FindStringSubmatch(r.URL.Path)[1]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Bad stuff happened"))
		return
	}

	h.mu.Lock()
	_, found := h.store[id]
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Todo with %v does not exist", id)))
		return
	}
	h.mu.Unlock()

	delete(h.store, id)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func newTodoHandlers() *todoHandlers {
	// using pointer since it will be data storage
	return &todoHandlers{
		store: map[int64]Todo{
			1: {
				Id:    1,
				Title: "Do dishes",
				Done:  false,
			},
			2: {
				Id:    2,
				Title: "Sweep",
				Done:  true,
			},
		},
	}
}

func init() {
	var err error
	pathRegex, err = regexp.Compile("^/?todo/([\\d]+)/?$")
	if err != nil {
		panic(err.Error())
	}
}

// What does this means, anyway?

// In simple terms, value receiver makes a copy of the type and pass it to the function. The function stack now holds an equal object but at a different location on memory.

// Pointer receiver passes the address of a type to the function. The function stack has a reference to the original object.
func main() {
	todoHandlers := newTodoHandlers()
	http.HandleFunc("/todo", todoHandlers.todos)
	http.HandleFunc("/todo/", todoHandlers.todo)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
