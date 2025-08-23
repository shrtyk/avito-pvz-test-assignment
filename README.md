# Avito PVZ Test Assignment

[![CI](https://github.com/shrtyk/avito-pvz-test-assignment/actions/workflows/ci.yml/badge.svg)](https://github.com/shrtyk/avito-pvz-test-assignment/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/shrtyk/avito-pvz-test-assignment/graph/badge.svg?token=8HU8XM22KZ)](https://codecov.io/gh/shrtyk/avito-pvz-test-assignment)
[![Go Report Card](https://goreportcard.com/badge/github.com/shrtyk/avito-pvz-test-assignment)](https://goreportcard.com/report/github.com/shrtyk/avito-pvz-test-assignment)

A backend service for managing package pickup points (PVZ) and goods reception, built with Go. It features a RESTful HTTP API (with OpenAPI docs) and a gRPC API.

## Features

- **User Management**: JWT + Refresh Token authentication.
- **RBAC**: `moderator` and `employee` roles.
- **PVZ & Reception Workflow**: Create/manage PVZs, open/close receptions, add/delete products (LIFO).
- **API**: REST and gRPC endpoints.
- **Monitoring**: Prometheus metrics.
- **Testing**: Unit, integration, and k6 load tests.

## Tech Stack

- **Go**
- **PostgreSQL**
- **Docker**
- **REST (go-chi)**
- **gRPC**
- **Prometheus**
- **k6**

## Getting Started

1.  **Prerequisites:** Go 1.24+, Docker, Make.

2.  **Clone the repository:**

    ```sh
    git clone https://github.com/shrtyk/avito-pvz-test-assignment.git && cd avito-pvz-test-assignment
    ```

3.  **Initial Setup:**

    ```sh
    # Creates .env from .env_example and generates RSA keys
    make setup
    ```

4.  **Run the application:**

    ```sh
    # Starts all services (app, db, prometheus) and applies migrations
    make docker/up
    ```

    The services will be available at:

    - HTTP API: `http://localhost:8080`
    - gRPC API: `localhost:3000`
    - Prometheus: `http://localhost:9000`

5.  **Stop the application:**
    ```sh
    # Stops and removes all running containers and volumes
    make docker/down
    ```

## Development Commands

The `Makefile` provides several helpers for development. For a full list of commands, please see the `Makefile`.

```sh
# Run all unit and integration tests
make test
```

```sh
# Run the linter
make linter/run
```

```sh
# Generate mocks, DTOs, and RSA keys
make generate
```

```sh
# Run k6 load test
make load-test/run
```

## Performance

Load tests were executed using k6. The test scenario simulates a ramp-up from 50 to 1000 virtual users over 30 seconds, sustaining the load for 1 minute.

The following results were achieved on an **Intel Core i9-9900KF** CPU:

| Metric                  | Result       |
| ----------------------- | ------------ |
| **Requests per Second** | ~5,040 req/s |
| **p(95) Latency**       | 40.05ms      |
| **Success Rate**        | 100%         |
| **Total Requests**      | ~510,000     |

All performance thresholds passed:

- `p(95) < 100ms`: **✓ Passed** (40.05ms)
- `checks > 99.99%`: **✓ Passed** (100%)
