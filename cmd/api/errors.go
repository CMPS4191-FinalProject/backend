package main

import (
	"fmt"
	"net/http"
)

func (c *serverConfig) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := c.writeResponseJSON(w, status, env, nil)
	if err != nil {
		c.logger.Error(err.Error(), nil)
		w.WriteHeader(500)
	}
}

func (c *serverConfig) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := ERROR_NOTFOUND
	c.errorResponse(w, r, http.StatusNotFound, message)
}

func (c *serverConfig) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	c.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}
