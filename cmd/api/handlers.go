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

func (c *serverConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var user types.User
	if err := c.readRequestJSON(w, r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (c *serverConfig) GetUsersHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	users, err := c.db.GetUsersWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		users = []types.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (c *serverConfig) GetUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	user, err := c.db.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (c *serverConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	var updatedUser types.User
	if err := c.readRequestJSON(w, r, &updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.ValidateUser(updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedUser.UserID = id
	if err := c.db.UpdateUser(id, updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (c *serverConfig) DeleteUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	if err := c.db.DeleteUser(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *serverConfig) CreateNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var node types.Node
	if err := c.readRequestJSON(w, r, &node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateNode(node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.CreateNode(node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

func (c *serverConfig) GetNodesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	nodes, err := c.db.GetNodesWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(nodes) == 0 {
		nodes = []types.Node{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func (c *serverConfig) GetNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	node, err := c.db.GetNodeByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

func (c *serverConfig) UpdateNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	var updatedNode types.Node
	if err := c.readRequestJSON(w, r, &updatedNode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.ValidateNode(updatedNode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedNode.DeviceID = id
	if err := c.db.UpdateNode(id, updatedNode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedNode)
}

func (c *serverConfig) DeleteNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	if err := c.db.DeleteNode(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *serverConfig) CreateNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var nodeData types.NodeData
	if err := c.readRequestJSON(w, r, &nodeData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateNodeData(nodeData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.CreateNodeData(nodeData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nodeData)
}

func (c *serverConfig) GetNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	nodeData, err := c.db.GetNodeDataWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

func (c *serverConfig) GetNodeDataByIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node data ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	nodeData, err := c.db.GetNodeDataByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

func (c *serverConfig) GetNodeDataByDeviceIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the device ID from the URL parameters
	idStr := ps.ByName("deviceId")
	deviceID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID format", http.StatusBadRequest)
		return
	}

	nodeData, err := c.db.GetNodeDataByDeviceID(deviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

func (c *serverConfig) GetNodeDataByUserIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("userId")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	nodeData, err := c.db.GetNodeDataByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

func (c *serverConfig) DeleteNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node data ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	if err := c.db.DeleteNodeData(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Node Favorites Handlers
func (c *serverConfig) CreateNodeFavoriteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var favorite types.NodeFavorite
	if err := c.readRequestJSON(w, r, &favorite); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.ValidateNodeFavorite(favorite); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.CreateNodeFavorite(favorite); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(favorite)
}

func (c *serverConfig) GetNodeFavoritesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	favorites, err := c.db.GetNodeFavorites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(favorites) == 0 {
		favorites = []types.NodeFavorite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (c *serverConfig) GetNodeFavoritesByUserIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("userId")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	favorites, err := c.db.GetNodeFavoritesByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(favorites) == 0 {
		favorites = []types.NodeFavorite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (c *serverConfig) DeleteNodeFavoriteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID and device ID from the URL parameters
	userIDStr := ps.ByName("userId")
	deviceIDStr := ps.ByName("deviceId")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		http.Error(w, "Invalid device ID format", http.StatusBadRequest)
		return
	}

	if err := c.db.DeleteNodeFavorite(userID, deviceID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Authentication Handlers
func (c *serverConfig) RegisterHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req types.UserCreateRequest
	if err := c.readRequestJSON(w, r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, _ := c.db.GetUserByUsername(req.Username)
	if existingUser != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := types.User{
		Username: req.Username,
		Password: EncodePasswordHash(passwordHash),
	}

	if err := database.ValidateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.db.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(&user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return user info and token (excluding password)
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id":  user.UserID,
			"username": user.Username,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (c *serverConfig) LoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.readRequestJSON(w, r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get user by username
	user, err := c.db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Decode stored password hash
	storedHash, err := DecodePasswordHash(user.Password)
	if err != nil {
		http.Error(w, "Invalid stored password format", http.StatusInternalServerError)
		return
	}

	// Verify password
	isValid, err := VerifyPassword(req.Password, storedHash)
	if err != nil {
		http.Error(w, "Failed to verify password", http.StatusInternalServerError)
		return
	}

	if !isValid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return user info and token (excluding password)
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id":  user.UserID,
			"username": user.Username,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
