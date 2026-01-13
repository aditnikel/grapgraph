# grapgraph

`grapgraph` is a Go application that demonstrates how to model and query financial data using a graph database, specifically [RedisGraph](https://redis.io/docs/data-types/graph/) (via the `redis-stack` or custom Redis image).

It simulates a financial network with users, accounts, and merchants, tracking money transfers and detecting suspicious patterns using Cypher queries.

## Features

- **Graph Data Modeling**: Nodes for `User`, `Account`, and `Merchant` with relationships like `:OWNS`, `:BELONGS_TO`, and `:TRANSFER`.
- **Data Seeding**: Automatically populates the graph with sample entities and transactions.
- **Traceability**: Demonstrates how to trace money flow through multiple hops (e.g., finding the path of funds from one account to another).
- **Go & Redis**: Uses the `go-redis` client to interact with Redis.

## Prerequisites

- **Go**: Version 1.20 or higher.
- **Docker & Docker Compose**: For running the Redis instance.

## Setup

1.  **Clone the repository:**

    ```bash
    git clone <repository-url>
    cd grapgraph
    ```

2.  **Configure Environment Variables:**
    Copy the example environment file to `.env`:

    ```bash
    cp .env.exampe .env
    ```

    Review the `.env` file to ensure the Redis settings match your local environment.

3.  **Prepare FalkorDB Module:**

    - Download the FalkorDB module (Linux assets) from [FalkorDB v4.16.0 Releases](https://github.com/FalkorDB/FalkorDB/releases/tag/v4.16.0).
    - Place the `falkordb.so` file into the `deps/` directory in the project root.

    > **Note**: If a pre-compiled `falkordb.so` is not available for your architecture in the assets, you may need to build it from source or extract it from the official Docker image.

4.  **Build and Start Redis:**
    Build the custom Docker image and start the container:
    ```bash
    docker compose build
    docker compose up -d
    ```
    This starts a Redis instance (custom image `redis-8.2.3-custom`) exposed on port `6379`.

## Usage

Run the application using Go:

```bash
go run main.go
```

### What it does

1.  **Connects**: Establishes a connection to the Redis instance defined in `.env`.
2.  **Seeds**: Runs a Cypher query to create sample nodes (Alice, Bob, Carol, Dave, Shops, Casinos) and edges (Transfers).
3.  **Traces**: Executes a graph query to trace money leaving account `a1` (Alice's Checking) and prints the resulting paths.

## Project Structure

- `main.go`: Entry point of the application. Handles initialization and high-level execution flow.
- `graph/`: Contains the core logic.
  - `client.go`: Redis client wrapper and connection logic.
  - `seed.go`: Contains the large Cypher query used to seed the initial graph data.
  - `queries.go`: Helper functions for specific graph queries (e.g., `TraceMoney`).
- `docker/`: Docker configuration files.
- `docker-compose.yaml`: Definition for the Redis service.
