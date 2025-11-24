package main

import (
	"fmt"
	"net/http"
)

func (c *serverConfig) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := c.writeResponseJSON(w, status, env, nil)
	if err != nil {
		c.logger.Error(err.Error())
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

func (c *serverConfig) badRequestResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	c.errorResponse(w, r, http.StatusBadRequest, message)
}

func (c *serverConfig) unauthorizedResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	c.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (c *serverConfig) forbiddenResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	c.errorResponse(w, r, http.StatusForbidden, message)
}

func (c *serverConfig) internalServerErrorResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	c.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (c *serverConfig) tooManyRequestsResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	c.errorResponse(w, r, http.StatusTooManyRequests, message)
}
