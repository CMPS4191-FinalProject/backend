# Really basic soil moisture monitor API

A simple Go API that pulls data from an IoT device and returns the soil moisture content of them

## Features

- **Real-time Soil Monitoring**: Monitor soil moisture levels from IoT devices
- **User Authentication**: Secure JWT-based authentication system
- **WebSocket Support**: Real-time data updates
- **Weather Integration**: Free weather and soil forecasts via Open-Meteo API
- **Sensor Validation**: Compare your sensor readings with professional forecasts
- **Comprehensive API**: Full REST API with Swagger documentation

## How to Run

1. Make sure you have Go installed (https://golang.org/dl/)
2. Open a terminal in this project directory.
3. Run:

   go run main.go

4. Visit http://localhost:4000/swagger/index.html in your browser to interact with the API

## Testing

The project includes comprehensive unit and integration tests.

### Run all tests:
```bash
go test ./cmd/api -v
```

### Run specific test:
```bash
go test ./cmd/api -v -run TestHealthCheckHandler
```

## API Documentation

### Swagger/OpenAPI
Interactive API documentation is available at http://localhost:4000/swagger/index.html when the server is running.

### Postman Collection
Import the `postman_collection.json` file into Postman to quickly test all API endpoints.

## cURL Examples

### Health Check
```bash
curl -X GET http://localhost:4000/v1/healthcheck
```

### Authentication

#### Register a new user
```bash
curl -X POST http://localhost:4000/v1/auth/join \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "SecurePassword123!",
    "email": "user@example.com"
  }'
```

#### Login
```bash
curl -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "user_id": 1,
    "username": "admin"
  }
}
```

Save the token for authenticated requests.

#### Logout
```bash
curl -X POST http://localhost:4000/v1/auth/logout
```

### Users (requires authentication)

#### Get current user
```bash
curl -X GET http://localhost:4000/v1/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get all users (admin only)
```bash
curl -X GET "http://localhost:4000/v1/users?limit=10&offset=0&sort_by=user_id&sort_order=asc" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get user by ID
```bash
curl -X GET http://localhost:4000/v1/users/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Update user
```bash
curl -X PUT http://localhost:4000/v1/users/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "updateduser",
    "password": "NewPassword123!"
  }'
```

#### Delete user (admin only)
```bash
curl -X DELETE http://localhost:4000/v1/users/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### IoT Nodes (Monitoring Devices)

#### Create a new node
```bash
curl -X POST http://localhost:4000/v1/nodes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "ONLINE"
  }'
```

#### Get all nodes
```bash
curl -X GET "http://localhost:4000/v1/nodes?limit=10&offset=0&sortBy=device_id&sortOrder=asc" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get node by ID
```bash
curl -X GET http://localhost:4000/v1/nodes/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Update node status
```bash
curl -X PUT http://localhost:4000/v1/nodes/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "OFFLINE"
  }'
```

#### Delete node
```bash
curl -X DELETE http://localhost:4000/v1/nodes/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Sensor Data (Node Data)

#### Create sensor reading
```bash
curl -X POST http://localhost:4000/v1/nodedata \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": 1,
    "moisture_content": 45.5
  }'
```

#### Get all sensor data
```bash
curl -X GET "http://localhost:4000/v1/nodedata?limit=10&offset=0&sortBy=timestamp&sortOrder=desc" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get sensor data by device ID
```bash
curl -X GET http://localhost:4000/v1/nodedata/device/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get sensor data by user ID
```bash
curl -X GET http://localhost:4000/v1/nodedata/user/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Delete sensor data
```bash
curl -X DELETE http://localhost:4000/v1/nodedata/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Favorites

#### Add node to favorites
```bash
curl -X POST http://localhost:4000/v1/favorites \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": 1
  }'
```

#### Get all favorites
```bash
curl -X GET http://localhost:4000/v1/favorites \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get user's favorites
```bash
curl -X GET http://localhost:4000/v1/favorites/user/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Remove from favorites
```bash
curl -X DELETE http://localhost:4000/v1/favorites/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Metrics (Admin only)

#### Get WebSocket metrics
```bash
curl -X GET http://localhost:4000/v1/metrics \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get message log
```bash
curl -X GET http://localhost:4000/v1/metrics/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Weather and Soil Forecasts (Open-Meteo Integration)

The API integrates with Open-Meteo, a free weather and soil data API that requires no API key. This allows you to:
- Get soil moisture forecasts for any location
- Compare sensor readings with forecast data
- Get weather forecasts including soil temperature

#### Get soil moisture forecast
```bash
curl -X GET "http://localhost:4000/v1/forecast/soil?lat=40.7128&lon=-74.0060&days=7" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Parameters:**
- `lat` - Latitude (required)
- `lon` - Longitude (required)
- `days` - Forecast days, 1-16 (optional, default: 7)

**Response:** Hourly soil moisture forecast at different depths (0-1cm, 1-3cm, 3-9cm, 9-27cm, 27-81cm)

#### Get weather forecast
```bash
curl -X GET "http://localhost:4000/v1/forecast/weather?lat=40.7128&lon=-74.0060&days=3" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Parameters:**
- `lat` - Latitude (required)
- `lon` - Longitude (required)
- `days` - Forecast days, 1-16 (optional, default: 7)

**Response:** Hourly weather forecast including temperature, humidity, precipitation, and soil temperature

#### Compare sensor data with forecast
```bash
curl -X GET "http://localhost:4000/v1/forecast/compare?device_id=1&lat=40.7128&lon=-74.0060" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Parameters:**
- `device_id` - Your device ID (required)
- `lat` - Latitude of the device location (required)
- `lon` - Longitude of the device location (required)

**Response:** Comparison of your sensor readings with forecast data, including insights and recommendations

## Environment Variables

Configure the application using environment variables (see `.envrc.example`):
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (development/production)
- `API_VERSION` - API version prefix (default: v1)
- `DB_TYPE` - Database type (IN_MEMORY or POSTGRES)
- `DB_DSN` - PostgreSQL connection string (required if DB_TYPE=POSTGRES)

## Third-Party Integrations

### Open-Meteo API (Free Weather and Soil Data)

The application integrates with [Open-Meteo](https://open-meteo.com/), a free and open-source weather API that provides:
- **No API key required** - Completely free to use
- Soil moisture forecasts at multiple depths
- Weather forecasts (temperature, humidity, precipitation)
- Soil temperature data
- Hourly forecasts up to 16 days

This integration allows users to:
1. Compare their IoT sensor readings with professional weather forecasts
2. Get insights on irrigation timing based on upcoming weather
3. Validate sensor accuracy against forecast data
4. Plan watering schedules based on precipitation forecasts

#### Example Use Case

Monitor your garden in New York City and compare with forecast data:

```bash
# 1. Login and get your token
TOKEN=$(curl -s -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# 2. Create a monitoring node
curl -X POST http://localhost:4000/v1/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"ONLINE"}'

# 3. Add sensor data (45.5% moisture)
curl -X POST http://localhost:4000/v1/nodedata \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_id":1,"moisture_content":45.5}'

# 4. Get weather forecast for your location (NYC)
curl -X GET "http://localhost:4000/v1/forecast/weather?lat=40.7128&lon=-74.0060&days=3" \
  -H "Authorization: Bearer $TOKEN"

# 5. Compare your sensor with professional forecast
curl -X GET "http://localhost:4000/v1/forecast/compare?device_id=1&lat=40.7128&lon=-74.0060" \
  -H "Authorization: Bearer $TOKEN"
```

The comparison will provide insights like:
- Whether your sensor reads higher/lower than forecast
- Upcoming moisture trends (increasing/decreasing)
- Irrigation recommendations based on forecast data
3. Validate sensor accuracy against forecast data
4. Plan watering schedules based on precipitation forecasts
