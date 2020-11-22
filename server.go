package main

import (
	"net/http"
	"regexp"

	"github.com/alesbrelih/simple-go-rest/internals/todo"
)

var pathRegex *regexp.Regexp

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
	todoHandlers := todo.NewTodoHandlers(pathRegex)
	http.HandleFunc("/todo", todoHandlers.Todos)
	http.HandleFunc("/todo/", todoHandlers.Todo)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
