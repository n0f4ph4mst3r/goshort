
# Goshort

## Description

Trivial URL shortening on REST API service.
It allows you to shorten long URLs and redirect users to the original link.
  
## System requirements

You need to have [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed in oder to build and run the project. No additional tools required.

## Hot to run with Docker

Define environment variables. You can copy environment from [example](https://github.com/n0f4ph4mst3r/goshort/blob/master/.env.sample)

    cp .env.sample .env

Perform

	sudo docker-compose up

Access the application via http://localhost:8080.

## Hot to run manually

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

This will compile and start the backend server. After that, the REST API service will be available with your [config](https://github.com/n0f4ph4mst3r/goshort/blob/master/config/sample.yaml) and [env](https://github.com/n0f4ph4mst3r/goshort/blob/master/.env.sample).




