package main

import (
	"net/http"
)

func (c *serverConfig) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer will be called when the stack unwinds
		defer func() {
			// recover() checks for panics
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				w.Header().Set("Content-Type", "application/json")
				data := envelope{"error": "Internal Server Error"}
				c.writeResponseJSON(w, http.StatusInternalServerError, data, w.Header())
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
