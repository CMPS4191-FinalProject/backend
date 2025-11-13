package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (c *serverConfig) routes() http.Handler {
	c.router.NotFound = http.HandlerFunc(c.notFoundResponse)
	c.router.MethodNotAllowed = http.HandlerFunc(c.methodNotAllowedResponse)

	c.router.GET(v("/healthcheck"), func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c.HealthCheckHandler(w, r, ps)
	})

	// Swagger UI route (public)
	c.router.GET("/swagger/*any", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		httpSwagger.WrapHandler(w, r)
	})

	// Authentication routes (public)
	c.router.POST(v("/auth/join"), c.RegisterHandler) // User registration
	c.router.POST(v("/auth/login"), c.LoginHandler)   // User login

	// Users (protected routes)
	c.router.POST(v("/users"), c.requireAuth(c.CreateUserHandler))       // Create user
	c.router.GET(v("/users"), c.requireAuth(c.GetUsersHandler))          // List users
	c.router.GET(v("/users/:id"), c.requireAuth(c.GetUserHandler))       // Get user by ID
	c.router.PUT(v("/users/:id"), c.requireAuth(c.UpdateUserHandler))    // Update user
	c.router.DELETE(v("/users/:id"), c.requireAuth(c.DeleteUserHandler)) // Delete user

	// Nodes (IoT devices) - protected routes
	c.router.POST(v("/nodes"), c.requireAuth(c.CreateNodeHandler))       // Create node
	c.router.GET(v("/nodes"), c.requireAuth(c.GetNodesHandler))          // List nodes
	c.router.GET(v("/nodes/:id"), c.requireAuth(c.GetNodeHandler))       // Get node by ID
	c.router.PUT(v("/nodes/:id"), c.requireAuth(c.UpdateNodeHandler))    // Update node status
	c.router.DELETE(v("/nodes/:id"), c.requireAuth(c.DeleteNodeHandler)) // Delete node

	// Node Data (sensor readings) - protected routes
	c.router.POST(v("/nodedata"), c.requireAuth(c.CreateNodeDataHandler))                        // Create sensor reading
	c.router.GET(v("/nodedata"), c.requireAuth(c.GetNodeDataHandler))                            // List all sensor data
	c.router.GET(v("/nodedata/device/:deviceId"), c.requireAuth(c.GetNodeDataByDeviceIDHandler)) // Get data by device
	c.router.GET(v("/nodedata/user/:userId"), c.requireAuth(c.GetNodeDataByUserIDHandler))       // Get data by user
	c.router.DELETE(v("/nodedata/device/:deviceId"), c.requireAuth(c.DeleteNodeDataHandler))     // Delete sensor data

	// Node Favorites - protected routes
	c.router.POST(v("/favorites"), c.requireAuth(c.CreateNodeFavoriteHandler))                   // Add favorite
	c.router.GET(v("/favorites"), c.requireAuth(c.GetNodeFavoritesHandler))                      // List all favorites
	c.router.GET(v("/favorites/user/:userId"), c.requireAuth(c.GetNodeFavoritesByUserIDHandler)) // Get user's favorites
	c.router.DELETE(v("/favorites/:deviceId"), c.requireAuth(c.DeleteNodeFavoriteHandler))       // Remove favorite

	// WebSocket faucet endpoint (protected)
	c.router.GET(v("/faucet"), c.requireAuth(c.FaucetHandler))

	return c.middleware(c.router)
}
