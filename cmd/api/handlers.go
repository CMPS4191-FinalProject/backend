package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"qotd/cmd/api/database"
	"qotd/cmd/api/types"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

type HealthCheck struct {
	Status      string `json:"status"`
	Environment string `json:"environment,omitempty"`
}

// HealthCheckHandler godoc
// @Summary     Health check endpoint
// @Description Get the health status of the API
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200 {object} HealthCheck
// @Router      /healthcheck [get]
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
		c.internalServerErrorResponse(w, r, ERROR_INTERNAL)
	}
}

// CreateUserHandler godoc
// @Summary     Create a new user
// @Description Create a new user (admin only)
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       user body types.User true "User data"
// @Success     201 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users [post]
func (c *serverConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var user types.User
	if err := c.readRequestJSON(w, r, &user); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := database.ValidateUser(user); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := c.db.CreateUser(user); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUsersHandler godoc
// @Summary     Get all users
// @Description Get list of users with pagination and sorting
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       limit query int false "Number of items per page"
// @Param       offset query int false "Number of items to skip"
// @Param       sort_by query string false "Field to sort by" Enums(user_id, username)
// @Param       sort_order query string false "Sort order" Enums(asc, desc)
// @Success     200 {array} UserResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users [get]
func (c *serverConfig) GetUsersHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	users, err := c.db.GetUsersWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(users) == 0 {
		users = []types.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUserHandler godoc
// @Summary     Get user by ID
// @Description Get a specific user by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "User ID"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /users/{id} [get]
func (c *serverConfig) GetUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}

	// Authorization is handled by middleware (requireOwnerOrAdmin)
	user, err := c.db.GetUserByID(id)
	if err != nil {
		c.notFoundResponse(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserHandler godoc
// @Summary     Update a user
// @Description Update an existing user's information
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "User ID"
// @Param       user body types.UserUpdateRequest true "Updated user data"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [put]
func (c *serverConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}

	// Authorization is handled by middleware (requireOwnerOrAdmin)
	currentUser, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}

	var updatedUser types.User
	if err := c.readRequestJSON(w, r, &updatedUser); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	// Prevent non-admin users from changing their role
	existingUser, err := c.db.GetUserByID(id)
	if err != nil {
		c.notFoundResponse(w, r)
		return
	}

	if currentUser.Role != types.RoleAdmin {
		updatedUser.Role = existingUser.Role // Keep existing role
	}

	if err := database.ValidateUser(updatedUser); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}
	updatedUser.UserID = id
	if err := c.db.UpdateUser(id, updatedUser); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

// GetCurrentUserHandler godoc
// @Summary     Get current user
// @Description Get the currently authenticated user's information
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     200 {object} UserResponse
// @Failure     401 {object} ErrorResponse
// @Router      /users/me [get]
func (c *serverConfig) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}

	// Get full user data from database
	fullUser, err := c.db.GetUserByID(user.UserID)
	if err != nil {
		c.notFoundResponse(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullUser)
}

// DeleteUserHandler godoc
// @Summary     Delete a user
// @Description Delete a user by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "User ID"
// @Success     204 "No Content"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [delete]
func (c *serverConfig) DeleteUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}
	if err := c.db.DeleteUser(id); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateNodeHandler godoc
// @Summary     Create a new monitoring node
// @Description Create a new IoT monitoring device/node
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       node body types.NodeCreateRequest true "Node data"
// @Success     201 {object} NodeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes [post]
func (c *serverConfig) CreateNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var node types.Node
	if err := c.readRequestJSON(w, r, &node); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := database.ValidateNode(node); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := c.db.CreateNode(node); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

// GetNodesHandler godoc
// @Summary     Get all monitoring nodes
// @Description Get a list of all IoT monitoring devices/nodes with pagination
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       limit query int false "Limit number of results"
// @Param       offset query int false "Offset for pagination"
// @Param       sortBy query string false "Sort by field (device_id, status)"
// @Param       sortOrder query string false "Sort order (asc, desc)"
// @Success     200 {array} NodeResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes [get]
func (c *serverConfig) GetNodesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	nodes, err := c.db.GetNodesWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(nodes) == 0 {
		nodes = []types.Node{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// GetNodeHandler godoc
// @Summary     Get a monitoring node by ID
// @Description Get a specific IoT monitoring device/node by its device ID
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Device ID"
// @Success     200 {object} NodeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes/{id} [get]
func (c *serverConfig) GetNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}

	node, err := c.db.GetNodeByID(id)
	if err != nil {
		c.notFoundResponse(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// UpdateNodeHandler godoc
// @Summary     Update a monitoring node
// @Description Update an existing IoT monitoring device/node's status
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Device ID"
// @Param       node body types.NodeUpdateRequest true "Updated node data"
// @Success     200 {object} NodeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes/{id} [put]
func (c *serverConfig) UpdateNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}
	var updatedNode types.Node
	if err := c.readRequestJSON(w, r, &updatedNode); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}
	if err := database.ValidateNode(updatedNode); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}
	updatedNode.DeviceID = id
	if err := c.db.UpdateNode(id, updatedNode); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedNode)
}

// DeleteNodeHandler godoc
// @Summary     Delete a monitoring node
// @Description Delete an IoT monitoring device/node by its device ID
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Device ID"
// @Success     204 "No Content"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes/{id} [delete]
func (c *serverConfig) DeleteNodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}
	if err := c.db.DeleteNode(id); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateNodeDataHandler godoc
// @Summary     Create sensor data reading
// @Description Create a new sensor data reading from an IoT monitoring device
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       nodedata body types.NodeDataCreateRequest true "Sensor data"
// @Success     201 {object} NodeDataResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata [post]
func (c *serverConfig) CreateNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var nodeDataCreateRequest types.NodeDataCreateRequest
	if err := c.readRequestJSON(w, r, &nodeDataCreateRequest); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	user, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}

	nodeData := types.NodeData{
		UserID:          user.UserID,
		DeviceID:        nodeDataCreateRequest.DeviceID,
		MoistureContent: nodeDataCreateRequest.MoistureContent,
		Timestamp:       time.Now(),
	}

	if err := database.ValidateNodeData(nodeData); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := c.db.CreateNodeData(nodeData); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	// Broadcast new sensor data to WebSocket clients
	BroadcastSensorData(nodeData.UserID, nodeData.DeviceID, nodeData.MoistureContent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nodeData)
}

// GetNodeDataHandler godoc
// @Summary     Get all sensor data
// @Description Get a list of all sensor readings from IoT monitoring devices with pagination
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       limit query int false "Limit number of results"
// @Param       offset query int false "Offset for pagination"
// @Param       sortBy query string false "Sort by field (id, user_id, device_id, moisture_content, timestamp)"
// @Param       sortOrder query string false "Sort order (asc, desc)"
// @Success     200 {array} NodeDataResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata [get]
func (c *serverConfig) GetNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse pagination and sorting parameters
	limit, offset := parsePaginationParams(r)
	sortBy, sortOrder := parseSortParams(r)

	nodeData, err := c.db.GetNodeDataWithPagination(limit, offset, sortBy, sortOrder)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

// GetNodeDataByIDHandler godoc
// @Summary     Get sensor data by ID
// @Description Get a specific sensor reading by its ID
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Node Data ID"
// @Success     200 {object} NodeDataResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/{id} [get]
// func (c *serverConfig) GetNodeDataByIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	// Extract the node data ID from the URL parameters
// 	idStr := ps.ByName("id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid ID format", http.StatusBadRequest)
// 		return
// 	}

// 	nodeData, err := c.db.GetNodeDataByID(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(nodeData)
// }

// GetNodeDataByDeviceIDHandler godoc
// @Summary     Get sensor data by device ID
// @Description Get all sensor readings from a specific IoT monitoring device
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       deviceId path int true "Device ID"
// @Success     200 {array} NodeDataResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/device/{deviceId} [get]
func (c *serverConfig) GetNodeDataByDeviceIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the device ID from the URL parameters
	idStr := ps.ByName("deviceId")
	deviceID, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid device ID format")
		return
	}

	nodeData, err := c.db.GetNodeDataByDeviceID(deviceID)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

// GetNodeDataByUserIDHandler godoc
// @Summary     Get sensor data by user ID
// @Description Get all sensor readings associated with a specific user
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       userId path int true "User ID"
// @Success     200 {array} NodeDataResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/user/{userId} [get]
func (c *serverConfig) GetNodeDataByUserIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("userId")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid user ID format")
		return
	}

	// Authorization is handled by middleware (requireOwnerOrAdmin)
	nodeData, err := c.db.GetNodeDataByUserID(userID)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(nodeData) == 0 {
		nodeData = []types.NodeData{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeData)
}

// DeleteNodeDataHandler godoc
// @Summary     Delete sensor data
// @Description Delete a specific sensor reading by its ID
// @Tags        nodedata
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Node Data ID"
// @Success     204 "No Content"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/{id} [delete]
func (c *serverConfig) DeleteNodeDataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the node data ID from the URL parameters
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid ID format")
		return
	}
	if err := c.db.DeleteNodeData(id); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateNodeFavoriteHandler godoc
// @Summary     Add node to favorites
// @Description Add an IoT monitoring device to a user's favorites
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       favorite body types.NodeFavoriteCreateRequest true "Favorite data"
// @Success     201 {object} NodeFavoritesResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse "Node already in favorites"
// @Failure     500 {object} ErrorResponse
// @Router      /favorites [post]
// Node Favorites Handlers
func (c *serverConfig) CreateNodeFavoriteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var favorite_request types.NodeFavoriteCreateRequest
	if err := c.readRequestJSON(w, r, &favorite_request); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	user, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}
	// Cast this into a types.NodeFavorite type
	favorite := types.NodeFavorite{
		DeviceID: favorite_request.DeviceID,
		UserID:   user.UserID,
	}

	if err := database.ValidateNodeFavorite(favorite); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := c.db.CreateNodeFavorite(favorite); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(favorite)
}

// GetNodeFavoritesHandler godoc
// @Summary     Get all node favorites
// @Description Get a list of all user's favorite IoT monitoring devices
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     200 {array} NodeFavoritesResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /favorites [get]
func (c *serverConfig) GetNodeFavoritesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	favorites, err := c.db.GetNodeFavorites()
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(favorites) == 0 {
		favorites = []types.NodeFavorite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

// GetNodeFavoritesByUserIDHandler godoc
// @Summary     Get user's node favorites
// @Description Get all favorite IoT monitoring devices for a specific user by user ID
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       user_id path int true "User ID"
// @Success     200 {array} NodeFavoritesResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /favorites/user/{user_id} [get]
func (c *serverConfig) GetNodeFavoritesByUserIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID from the URL parameters
	idStr := ps.ByName("userId")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid user ID format")
		return
	}

	// Authorization is handled by middleware (requireOwnerOrAdmin)
	favorites, err := c.db.GetNodeFavoritesByUserID(userID)
	if err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}

	if len(favorites) == 0 {
		favorites = []types.NodeFavorite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

// DeleteNodeFavoriteHandler godoc
// @Summary     Delete a node favorite
// @Description Remove an IoT monitoring device from user's favorites list
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       deviceId path int true "Device ID"
// @Success     204 "No Content"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /favorites/{deviceId} [delete]
func (c *serverConfig) DeleteNodeFavoriteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract the user ID and device ID from the URL parameters
	user, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}
	userID := user.UserID
	deviceIDStr := ps.ByName("deviceId")

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.badRequestResponse(w, r, "Invalid device ID format")
		return
	}

	if err := c.db.DeleteNodeFavorite(userID, deviceID); err != nil {
		c.internalServerErrorResponse(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Authentication Handlers
// RegisterHandler godoc
// @Summary     Register a new user
// @Description Register a new user account
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       user body types.UserCreateRequest true "User registration data"
// @Success     201 {object} map[string]string "Check your email"
// @Failure     400 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /auth/join [post]
func (c *serverConfig) RegisterHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req types.UserCreateRequest
	if err := c.readRequestJSON(w, r, &req); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		c.badRequestResponse(w, r, "Username, password, and email are required")
		return
	}

	// Check if user already exists
	existingUser, _ := c.db.GetUserByUsername(req.Username)
	if existingUser != nil {
		c.conflictResponse(w, r, "Username already exists")
		return
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Failed to hash password")
		return
	}

	// Create user
	user := types.User{
		Username:   req.Username,
		Password:   EncodePasswordHash(passwordHash),
		IsVerified: false, // User starts as unverified
		Role:       types.RoleUser,
	}

	if err := database.ValidateUser(user); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	if err := c.db.CreateUser(user); err != nil {
		c.internalServerErrorResponse(w, r, "Failed to create user")
		return
	}

	// Fetch the created user to get the correct user_id
	createdUser, err := c.db.GetUserByUsername(req.Username)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Failed to retrieve created user")
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(createdUser)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Failed to generate token")
		return
	}

	// Return user info and token (excluding password)
	response := map[string]interface{}{
		"message": "Check your email to verify your account.",
	}

	from := strings.TrimSpace(getEnvAsString("SMTP_SENDER_EMAIL", ""))
	sender_username := strings.TrimSpace(getEnvAsString("SMTP_SENDER_USERNAME", ""))
	password := getEnvAsString("SMTP_SENDER_PASSWORD", "")
	to := strings.TrimSpace(req.Email)
	smtpHost := getEnvAsString("SMTP_HOST", "smtp.gmail.com")
	smtpPort := getEnvAsString("SMTP_PORT", "587")

	if from == "" || password == "" {
		c.logger.Error("SMTP credentials missing: check SMTP_SENDER_EMAIL and SMTP_SENDER_PASSWORD env vars")
		c.internalServerErrorResponse(w, r, "Email service not configured")
		return
	}

	if _, err := mail.ParseAddress(from); err != nil {
		c.logger.Error("invalid SMTP sender address", "address", from, "error", err)
		c.internalServerErrorResponse(w, r, "Email service misconfigured")
		return
	}

	if _, err := mail.ParseAddress(to); err != nil {
		c.logger.Error("invalid recipient email address", "address", to, "error", err)
		c.badRequestResponse(w, r, "Invalid recipient email address")
		return
	}

	// Message content
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Verification Token\r\n" +
		"\r\n" +
		token + "\r\n")

	// Authentication
	auth := smtp.PlainAuth("", sender_username, password, smtpHost)

	// Send the email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LoginHandler godoc
// @Summary     User login
// @Description Authenticate user and get JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       credentials body LoginRequest true "User login credentials"
// @Success     200 {object} AuthResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /auth/login [post]
func (c *serverConfig) LoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.readRequestJSON(w, r, &req); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		c.badRequestResponse(w, r, "Username and password are required")
		return
	}

	// Get user by username
	user, err := c.db.GetUserByUsername(req.Username)
	if err != nil {
		c.unauthorizedResponse(w, r, "Invalid credentials")
		return
	}

	// Decode stored password hash
	storedHash, err := DecodePasswordHash(user.Password)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Invalid stored password format")
		return
	}

	// Verify password
	isValid, err := VerifyPassword(req.Password, storedHash)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Failed to verify password")
		return
	}

	if !isValid {
		c.unauthorizedResponse(w, r, "Invalid credentials")
		return
	}

	// Check if user is verified
	if !user.IsVerified {
		c.forbiddenResponse(w, r, "Account not verified. Please verify your account first.")
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user)
	if err != nil {
		c.internalServerErrorResponse(w, r, "Failed to generate token")
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

	// Set JWT token in secure cookie
	cookie := &http.Cookie{
		Name:     "authorization",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   getEnvAsBool("HTTPS_ENABLED", false), // Use HTTPS in production
		SameSite: http.SameSiteNoneMode,                // Allow cross-origin requests
		MaxAge:   int((24 * time.Hour).Seconds()),      // 24 hours in seconds
	}

	// In development, allow less restrictive SameSite
	if getEnvAsString("ENVIRONMENT", "development") == "development" {
		cookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler godoc
// @Summary     User logout
// @Description Logout user and clear authentication cookie
// @Tags        auth
// @Success     200 {object} map[string]string
// @Failure     500 {object} ErrorResponse
// @Router      /auth/logout [post]
func (c *serverConfig) LogoutHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Clear the authorization cookie
	cookie := &http.Cookie{
		Name:     "authorization",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   getEnvAsBool("HTTPS_ENABLED", false),
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1, // Immediately expire the cookie
	}

	// In development, use same settings as login
	if getEnvAsString("ENVIRONMENT", "development") == "development" {
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = false
	}

	http.SetCookie(w, cookie)

	response := map[string]string{
		"message": "Successfully logged out",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyHandler godoc
// @Summary     Verify user account
// @Description Verify a user account using the JWT token received during registration
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       verification body types.UserVerifyRequest true "Verification data"
// @Success     200 {object} OkResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /auth/verify [post]
func (c *serverConfig) VerifyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get user from context (requireAuth middleware has already validated the token)
	var verificationRequest types.UserVerifyRequest
	if err := c.readRequestJSON(w, r, &verificationRequest); err != nil {
		c.badRequestResponse(w, r, err.Error())
		return
	}

	// Extract the token from the request
	tokenString := verificationRequest.VerificationCode
	// Parse and validate the token
	user, err := GetUserIDFromJWT(tokenString)
	if err != nil {
		c.unauthorizedResponse(w, r, "Invalid or expired token: "+err.Error())
		return
	}

	// Verify the user in the database
	if err := c.db.VerifyUser(user); err != nil {
		c.internalServerErrorResponse(w, r, "Failed to verify user: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "Account successfully verified",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetricsHandler godoc
// @Summary     Get WebSocket metrics
// @Description Get real-time WebSocket metrics (admin only)
// @Tags        metrics
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /metrics [get]
func (c *serverConfig) GetMetricsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	metrics := GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetMessageLogHandler godoc
// @Summary     Get WebSocket message log
// @Description Get recent WebSocket messages (admin only, hawk-eye view)
// @Tags        metrics
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     200 {array} MessageLogEntry
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /metrics/messages [get]
func (c *serverConfig) GetMessageLogHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageLog := GetMessageLog()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messageLog)
}

// DashboardHandler serves the admin dashboard
func (c *serverConfig) DashboardHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.ServeFile(w, r, "cmd/api/templates/dashboard.html")
}

// DashboardLoginHandler serves the login page (redirect to main dashboard which handles login)
func (c *serverConfig) DashboardLoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
