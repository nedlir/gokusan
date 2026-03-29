# Gokusan

A secure, event-driven file sharing platform built with Go microservices, React, and Kafka.

![architecture](readme/architecture.png)

## Getting Started

```bash
docker compose up --build
```

| Service  | URL                   |
|----------|-----------------------|
| Client   | http://localhost:5173 |
| Gateway  | http://localhost:8000 |
| Kong Admin | http://localhost:8001 |

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full system design, service breakdown, and request flow diagrams.
