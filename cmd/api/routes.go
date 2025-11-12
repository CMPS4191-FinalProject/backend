package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (c *serverConfig) routes() http.Handler {
	c.router.NotFound = http.HandlerFunc(c.notFoundResponse)
	c.router.MethodNotAllowed = http.HandlerFunc(c.methodNotAllowedResponse)

	c.router.GET(v("/healthcheck"), func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c.HealthCheckHandler(w, r, ps)
	})

	// Quotes
	c.router.POST(v("/quotes"), c.CreateQuoteHandler)       // C
	c.router.GET(v("/quotes"), c.GetQuotesHandler)          // R
	c.router.PUT(v("/quotes/:id"), c.UpdateQuoteHandler)    // U
	c.router.DELETE(v("/quotes/:id"), c.DeleteQuoteHandler) // D

	// Comments
	c.router.POST(v("/comments"), c.CreateCommentHandler)       // C
	c.router.GET(v("/comments"), c.GetCommentsHandler)          // R
	c.router.PUT(v("/comments/:id"), c.UpdateCommentHandler)    // U
	c.router.DELETE(v("/comments/:id"), c.DeleteCommentHandler) // D
	return c.middleware(c.router)
}
