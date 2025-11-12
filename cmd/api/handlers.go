package main

import (
	"encoding/json"
	"net/http"
	"qotd/cmd/api/database"
	"qotd/cmd/api/types"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type HealthCheck struct {
	Status      string `json:"status"`
	Environment string `json:"environment,omitempty"`
}

func (c *serverConfig) HealthCheckHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := envelope{
		"status": "alive",
		"system_info": map[string]string{
			"environment": c.env,
			"version":     c.version,
		},
	}
	err := c.writeResponseJSON(w, http.StatusOK, data, nil)
	if err != nil {
		c.logger.Error(err.Error())
		http.Error(w, ERROR_INTERNAL, http.StatusInternalServerError)
	}
}

func (c *serverConfig) CreateQuoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var quote types.Quote
	if err := c.readRequestJSON(w, r, &quote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateQuote(quote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.WriteQuote(quote); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quote)
}

func (c *serverConfig) GetQuotesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	quotes, err := c.db.GetQuotesWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(quotes) == 0 {
		quotes = []types.Quote{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func (c *serverConfig) UpdateQuoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the quote ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	var updatedQuote types.Quote
	if err := c.readRequestJSON(w, r, &updatedQuote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.ValidateQuote(updatedQuote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedQuote.ID = id
	if err := c.db.ModifyQuote(id, updatedQuote); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedQuote)
}

func (c *serverConfig) DeleteQuoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the quote ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	if err := c.db.DeleteQuote(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *serverConfig) CreateCommentHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var comment types.Comment
	if err := c.readRequestJSON(w, r, &comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateComment(comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.WriteComment(comment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (c *serverConfig) GetCommentsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	comments, err := c.db.GetCommentsWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(comments) == 0 {
		comments = []types.Comment{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (c *serverConfig) UpdateCommentHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the comment ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	var updatedComment types.Comment
	if err := c.readRequestJSON(w, r, &updatedComment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.ValidateComment(updatedComment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedComment.ID = id
	if err := c.db.ModifyComment(id, updatedComment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedComment)
}

func (c *serverConfig) DeleteCommentHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the comment ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	if err := c.db.DeleteComment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
