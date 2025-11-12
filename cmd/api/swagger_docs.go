package main

// This file contains all Swagger/OpenAPI annotations for the Soil Quality Monitor API

// Health Check Endpoints

// HealthCheckHandler godoc
// @Summary     Health check endpoint
// @Description Get the health status of the API
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200 {object} HealthCheck
// @Router      /healthcheck [get]

// Authentication Endpoints

// RegisterHandler godoc
// @Summary     Register a new user
// @Description Register a new user account
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       user body types.UserCreateRequest true "User registration data"
// @Success     201 {object} AuthResponse
// @Failure     400 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /auth/join [post]

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

// User Management Endpoints

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

// UpdateUserHandler godoc
// @Summary     Update user
// @Description Update user information
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

// DeleteUserHandler godoc
// @Summary     Delete user
// @Description Delete a user account
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "User ID"
// @Success     204 "User deleted successfully"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [delete]

// Node Management Endpoints

// CreateNodeHandler godoc
// @Summary     Create a new node
// @Description Create a new IoT monitoring node
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       node body types.Node true "Node data"
// @Success     201 {object} NodeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes [post]

// GetNodesHandler godoc
// @Summary     Get all nodes
// @Description Get list of all IoT monitoring nodes
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       limit query int false "Number of items per page"
// @Param       offset query int false "Number of items to skip"
// @Param       sort_by query string false "Field to sort by" Enums(device_id, status)
// @Param       sort_order query string false "Sort order" Enums(asc, desc)
// @Success     200 {array} NodeResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes [get]

// GetNodeHandler godoc
// @Summary     Get node by ID
// @Description Get a specific node by device ID
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Device ID"
// @Success     200 {object} NodeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /nodes/{id} [get]

// UpdateNodeHandler godoc
// @Summary     Update node
// @Description Update node status (ONLINE/OFFLINE/ERROR)
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

// DeleteNodeHandler godoc
// @Summary     Delete node
// @Description Delete an IoT monitoring node
// @Tags        nodes
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Device ID"
// @Success     204 "Node deleted successfully"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodes/{id} [delete]

// Node Data Endpoints

// CreateNodeDataHandler godoc
// @Summary     Record sensor data
// @Description Create a new sensor reading from IoT node
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       data body types.NodeDataCreateRequest true "Sensor data"
// @Success     201 {object} NodeDataItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata [post]

// GetNodeDataHandler godoc
// @Summary     Get all sensor data
// @Description Get list of all sensor readings with pagination
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       limit query int false "Number of items per page"
// @Param       offset query int false "Number of items to skip"
// @Param       sort_by query string false "Field to sort by" Enums(id, user_id, device_id, moisture_content, timestamp)
// @Param       sort_order query string false "Sort order" Enums(asc, desc)
// @Success     200 {array} NodeDataItem
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata [get]

// GetNodeDataByIDHandler godoc
// @Summary     Get sensor data by ID
// @Description Get a specific sensor reading by ID
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Data ID"
// @Success     200 {object} NodeDataItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /nodedata/{id} [get]

// GetNodeDataByDeviceIDHandler godoc
// @Summary     Get sensor data by device
// @Description Get all sensor readings for a specific device
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       deviceId path int true "Device ID"
// @Success     200 {array} NodeDataItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/device/{deviceId} [get]

// GetNodeDataByUserIDHandler godoc
// @Summary     Get sensor data by user
// @Description Get all sensor readings for a specific user
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       userId path int true "User ID"
// @Success     200 {array} NodeDataItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/user/{userId} [get]

// DeleteNodeDataHandler godoc
// @Summary     Delete sensor data
// @Description Delete a sensor reading
// @Tags        node-data
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       id path int true "Data ID"
// @Success     204 "Sensor data deleted successfully"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /nodedata/{id} [delete]

// Node Favorites Endpoints

// CreateNodeFavoriteHandler godoc
// @Summary     Add node to favorites
// @Description Add an IoT node to user's favorites list
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       favorite body types.NodeFavoriteCreateRequest true "Favorite data"
// @Success     201 {object} NodeFavoriteItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse "Node already in favorites"
// @Failure     500 {object} ErrorResponse
// @Router      /favorites [post]

// GetNodeFavoritesHandler godoc
// @Summary     Get all favorites
// @Description Get list of all favorite nodes
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     200 {array} NodeFavoriteItem
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /favorites [get]

// GetNodeFavoritesByUserIDHandler godoc
// @Summary     Get user's favorites
// @Description Get all favorite nodes for a specific user
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       userId path int true "User ID"
// @Success     200 {array} NodeFavoriteItem
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /favorites/user/{userId} [get]

// DeleteNodeFavoriteHandler godoc
// @Summary     Remove node from favorites
// @Description Remove an IoT node from user's favorites list
// @Tags        favorites
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Param       userId path int true "User ID"
// @Param       deviceId path int true "Device ID"
// @Success     204 "Favorite removed successfully"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse "Favorite not found"
// @Failure     500 {object} ErrorResponse
// @Router      /favorites/user/{userId}/device/{deviceId} [delete]
