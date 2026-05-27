# User Service

Manages public user profiles and master profiles. Does NOT handle authentication.

## Responsibilities

- CRUD operations on user profiles
- Master profile management (description, experience, rating)
- Role switching (user <-> master)
- Listens for `user.created` and `review.created` Kafka events

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /internal/users/{id} | Get user profile |
| PATCH | /internal/users/{id} | Update own profile |
| GET | /internal/masters/{id} | Get master profile |
| PATCH | /internal/masters/{id} | Update master profile |
| POST | /internal/users/{id}/roles/master | Enable master role |
| DELETE | /internal/users/{id}/roles/master | Disable master role |
| GET | /internal/health | Health check |

## Configuration

See `.env.example` for all environment variables.

## Running

```bash
go run ./cmd/api
```
