# Eugene

An Eneba-style marketplace for digital game keys. Go API in `server/`, Next.js frontend in `client/`, MongoDB Atlas.

Not affiliated with Eneba. Built as a homework project.

## Running locally

### API

Needs Go 1.25+. Copy `server/.env.example` to `server/.env` and fill in the values.

Live reload with [air](https://github.com/air-verse/air), which rebuilds and restarts the API when a `.go` file changes:

```
go install github.com/air-verse/air@latest   # once
cd server
air
```

Without air:

```
cd server
go run ./cmd/api
```

Check it is up:

```
curl localhost:8080/healthz
# {"db":"ok","status":"ok"}
```
