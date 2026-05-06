
# Goshort

## Description

Trivial URL shortening on REST API service.
It allows you to shorten long URLs and redirect users to the original link.
  
## System requirements

You need to have [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed in oder to build and run the project. No additional tools required.

## How to run with Docker

Define environment variables. You can copy environment from [example](https://github.com/n0f4ph4mst3r/goshort/blob/master/.env.sample)

    cp .env.sample .env

Perform

	sudo docker-compose up

Access the application via http://localhost:8080.

## How to run manually

### Tools

To develop the app manually, you need the following tools installed:

- [Go](https://go.dev/) (version 1.25.1 or newer)

- [Postgres](https://www.postgresql.org/) and [Redis](https://redis.io/) databases (you can run them via Docker Compose)

### Start the dev DB

If you don’t have Postgres or Redis installed locally, you can start them with Docker Compose:

    sudo docker-compose -f docker-compose.yml up postgres redis

This will start local Postgres and Redis instances.

### Start the server

Run following command:

    go run ./cmd/main.go

This will compile and start the backend server. After that, the REST API service will be available with your [config](https://github.com/n0f4ph4mst3r/goshort/blob/master/config/sample.yaml) and [env](https://github.com/n0f4ph4mst3r/goshort/blob/master/.env.sample). Check

### Usage

This service supports two authentication modes:
1. **Basic Auth**: Default mode (Username: `myuser`, Password: `qwerty`).
2. **JWT (SSO)**: Uses RS256 algorithm. For JWT authorization, you must provide a public key from the [TokenCraft SSO service](https://github.com/n0f4ph4mst3r/TokenCraft).

Use these endpoints for manage your urls

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| **POST** | `/api/url` | Basic/Bearer | Saves a long URL with an alias |
| **GET** | `/api/url/{alias}` | None | Redirects to the original URL |
| **DELETE** | `/api/url/{alias}` | Basic/Bearer | Deletes record from DB and cache|

### Examples

**POST Request Body:**

```json
{
    "url": "[https://duckduckgo.com](https://duckduckgo.com)",
    "alias": "goDuck"
}
```

**CURL Create (Basic Auth):**

```bash
	curl -X POST {{HOST_URL}}/api/url -u myuser:qwerty -d '{"url":"[https://example.com](https://example.com)","alias":"ex"}'
```

**CURL Delete (JWT):**

```bash
	curl -X DELETE {{HOST_URL}}/api/url/goDuck -H "Authorization: Bearer <your_jwt_token>"
```