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
	c.router.POST(v("/auth/join"), c.RegisterHandler)                // User registration
	c.router.POST(v("/auth/login"), c.LoginHandler)                  // User login
	c.router.POST(v("/auth/logout"), c.LogoutHandler)                // User logout
	c.router.POST(v("/auth/verify"), c.requireAuth(c.VerifyHandler)) // Verify account

	// Users (protected routes)
	c.router.GET(v("/auth/me"), c.requireAuth(c.GetCurrentUserHandler))              // Get current user
	c.router.POST(v("/users"), c.requireAdmin(c.CreateUserHandler))                  // Create user (admin only)
	c.router.GET(v("/users"), c.requireAdmin(c.GetUsersHandler))                     // List users (admin only)
	c.router.GET(v("/users/:id"), c.requireOwnerOrAdmin(c.GetUserHandler))           // Get user by ID (self or admin)
	c.router.PUT(v("/users/:id"), c.requireOwnerOrAdmin(c.UpdateUserHandler))        // Update user (self or admin)
	c.router.DELETE(v("/users/:id"), c.requireAdmin(c.DeleteUserHandler))            // Delete user (admin only)

	// Nodes (IoT devices) - protected routes
	c.router.POST(v("/nodes"), c.requireAuth(c.CreateNodeHandler))       // Create node
	c.router.GET(v("/nodes"), c.requireAuth(c.GetNodesHandler))          // List nodes
	c.router.GET(v("/nodes/:id"), c.requireAuth(c.GetNodeHandler))       // Get node by ID
	c.router.PUT(v("/nodes/:id"), c.requireAuth(c.UpdateNodeHandler))    // Update node status
	c.router.DELETE(v("/nodes/:id"), c.requireAuth(c.DeleteNodeHandler)) // Delete node

	// Node Data (sensor readings) - protected routes
	c.router.POST(v("/nodedata"), c.requireAuth(c.CreateNodeDataHandler))                                      // Create sensor reading
	c.router.GET(v("/nodedata"), c.requireAuth(c.GetNodeDataHandler))                                          // List all sensor data
	c.router.GET(v("/nodedata/device/:deviceId"), c.requireAuth(c.GetNodeDataByDeviceIDHandler))               // Get data by device
	c.router.GET(v("/nodedata/user/:userId"), c.requireOwnerOrAdmin(c.GetNodeDataByUserIDHandler))             // Get data by user (owner or admin)
	c.router.DELETE(v("/nodedata/device/:deviceId"), c.requireAuth(c.DeleteNodeDataHandler))                   // Delete sensor data

	// Node Favorites - protected routes
	c.router.POST(v("/favorites"), c.requireAuth(c.CreateNodeFavoriteHandler))                                 // Add favorite
	c.router.GET(v("/favorites"), c.requireAuth(c.GetNodeFavoritesHandler))                                    // List all favorites
	c.router.GET(v("/favorites/user/:userId"), c.requireOwnerOrAdmin(c.GetNodeFavoritesByUserIDHandler))       // Get user's favorites (owner or admin)
	c.router.DELETE(v("/favorites/:deviceId"), c.requireAuth(c.DeleteNodeFavoriteHandler))                     // Remove favorite

	// WebSocket faucet endpoint (protected)
	c.router.GET(v("/faucet"), c.requireAuth(c.FaucetHandler))

	// Metrics endpoints (metrics available to all authenticated users, messages admin only)
	c.router.GET(v("/metrics"), c.requireAuth(c.GetMetricsHandler))
	c.router.GET(v("/metrics/messages"), c.requireAdmin(c.GetMessageLogHandler))

	// Open-Meteo API integration (free weather and soil data, no API key required)
	c.router.GET(v("/forecast/soil"), c.requireAuth(c.GetSoilMoistureForecastHandler))     // Get soil moisture forecast
	c.router.GET(v("/forecast/weather"), c.requireAuth(c.GetWeatherForecastHandler))       // Get weather forecast
	c.router.GET(v("/forecast/compare"), c.requireAuth(c.CompareSensorWithForecastHandler)) // Compare sensor with forecast

	// Dashboard routes (static HTML served with authentication check)
	c.router.GET("/dashboard", c.DashboardHandler)
	c.router.GET("/dashboard/login", c.DashboardLoginHandler)

	return c.middleware(c.router)
}
