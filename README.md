# Tour Management Platform

Microservices platform for creating and experiencing guided tours with real-time GPS tracking.

## Services

- **Tour Service** (Go) - Tour and key point management
- **Stakeholders Service** (Go) - Authentication
- **Blog Service** (Go) - Blogs and comments
- **Follower Service** (Go) - Social graph (Neo4j)
- **Purchase Service** (Python) - Shopping cart
- **Gateway Service** (Go) - API gateway
- **Frontend Service** (Vue.js 3) - SPA with maps

**Stack:** Go | Python | Vue.js | MongoDB | Neo4j | Docker

## Setup

1. Clone repository:
```bash
git clone https://github.com/IvanNovakovic/SOA_Proj.git
cd SOA_Proj
```

2. Add API key to `frontend-service/.env`:
```
VITE_OPENROUTE_API_KEY=your_key
```

3. Run:
```bash
docker-compose up --build
```

4. Access:
- App: http://localhost:8087
- API: http://localhost:8080

## Development

```bash
# Rebuild service
docker-compose up -d --build service-name

# View logs
docker logs service-name

# Stop
docker-compose down
```