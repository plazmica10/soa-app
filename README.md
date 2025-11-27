# Tour Management Platform

A microservices-based platform for creating, managing, and experiencing guided tours with real-time location tracking and social features.

## Architecture

### Services
- **Tour Service** (Go): Tour creation, key points, GPS-based execution tracking
- **Stakeholders Service** (Go): User authentication and profiles
- **Blog Service** (Go): Blog posts, comments, likes
- **Follower Service** (Go): Social graph with Neo4j
- **Purchase Service** (Python): Shopping cart and purchases
- **Gateway Service** (Go): API gateway with JWT auth
- **Frontend Service** (Vue.js 3): SPA with interactive maps

### Infrastructure
MongoDB | Neo4j | Jaeger | Nginx

## Key Features

- Create tours with multiple key points and GPS coordinates
- Real-time position tracking with 30m proximity detection
- Interactive maps with street-based routing (OpenRouteService)
- Social features: follow users, blogs, comments, likes
- E-commerce: shopping cart and token-based purchases
- Tour execution with automatic progress tracking

## Technology Stack

**Backend:** Go, Python/FastAPI, gRPC, MongoDB, Neo4j  
**Frontend:** Vue.js 3, Leaflet.js, OpenRouteService  
**DevOps:** Docker, Docker Compose, Jaeger

## Quick Start

1. Clone repository:
```bash
git clone https://github.com/IvanNovakovic/SOA_Proj.git
cd SOA_Proj
```

2. Add OpenRouteService API key to `frontend-service/.env`:
```
VITE_OPENROUTE_API_KEY=your_api_key_here
```

3. Start services:
```bash
docker-compose up --build
```

4. Access:
- Frontend: http://localhost:8087
- API Gateway: http://localhost:8080
- Jaeger UI: http://localhost:16686

## Project Structure

```
SOA_Proj/
├── tour-service/         # Tour management
├── stakeholders-service/ # Authentication
├── blog-service/        # Blogs and comments
├── follower-service/    # Social graph
├── purchase-service/    # E-commerce
├── gateway-service/     # API gateway
├── frontend-service/    # Vue.js SPA
├── protos/             # gRPC definitions
└── docker-compose.yml  # Orchestration
```

## API Examples

**Authentication**
```bash
POST /api/register
POST /api/login
```

**Tours**
```bash
GET  /api/tours
POST /api/tours
GET  /api/tours/:id/keypoints
POST /api/executions
```

**Social**
```bash
POST /api/follow/:userId
GET  /api/followers/:userId
POST /api/blogs/:id/comments
```

**Shopping**
```bash
POST /api/cart
POST /api/checkout
```

## Development

```bash
# Rebuild specific service
docker-compose up -d --build tour-service

# View logs
docker logs tour-service

# Stop all
docker-compose down
```

## License

University project for Service-Oriented Architecture course.