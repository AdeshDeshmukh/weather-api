# Weather API (Go + Redis + Visual Crossing)

A learning-focused project that demonstrates:
- Calling a 3rd-party API (Visual Crossing)
- Caching responses in Redis with TTL expiration
- Managing configuration via environment variables
- Applying IP-based rate limiting
- Returning meaningful API errors

## Architecture

Client -> Go API -> Redis cache -> Visual Crossing API

Flow:
1. Client calls `GET /weather?city=...`
2. API checks Redis key `weather:<city>`
3. If cache hit: return cached payload
4. If cache miss: call Visual Crossing, store in Redis with TTL, return data

## Prerequisites

- Go 1.22+
- Docker (optional, for local Redis)

## Setup

1. Start Redis:

```bash
docker compose up -d
```

2. Prepare environment variables:

```bash
cp .env.example .env
```

Then edit `.env` and set `VISUAL_CROSSING_API_KEY`.

3. Install dependencies:

```bash
go mod tidy
```

4. Run the API:

```bash
go run .
```

Server starts on `http://localhost:8080` by default.

## Endpoints

- `GET /health`
- `GET /weather/hardcoded`
- `GET /weather?city=london`

### Example requests

```bash
curl "http://localhost:8080/health"
curl "http://localhost:8080/weather/hardcoded"
curl "http://localhost:8080/weather?city=london"
```

## Environment Variables

- `PORT`: server port (default `8080`)
- `VISUAL_CROSSING_API_KEY`: required API key
- `VISUAL_CROSSING_BASE_URL`: provider base URL
- `REDIS_ADDR`: Redis host:port
- `REDIS_PASSWORD`: Redis password (if any)
- `REDIS_DB`: Redis DB index (default `0`)
- `CACHE_TTL_HOURS`: cache expiration in hours (default `12`)
- `RATE_LIMIT_RPS`: requests per second per IP (default `5`)
- `RATE_LIMIT_BURST`: burst per IP (default `10`)

## What You Learn

- How to separate concerns into config, service, and transport layers
- How to protect your API key with env vars
- How cache-aside strategy reduces 3rd-party API calls
- How TTL-based cache invalidation keeps data fresh automatically
- How to implement basic abuse prevention with rate limiting
