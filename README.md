# Grapgraph

`Grapgraph` is a high-performance fraud detection and financial network analysis platform built with **Go** and **FalkorDB** (Redis Graph). It allows for real-time ingestion of financial events and complex multi-hop relationship querying to detect suspicious patterns like money laundering, account takeovers, and shared entity fraud.

---

## 🚀 Key Features

- **Real-time Event Ingestion**: Aggregated upserts of user events (Logins, Payments, Withdrawals, etc.) into a graph structure.
- **Multi-Hop Subgraph Analysis**: Budget-aware graph traversal (up to 3 hops) to detect relationships between users and shared entities (Devices, Wallets, Banks).
- **Advanced Cypher Templates**: Utilization of localized Cypher blocks for conditional aggregations and time-windowed metrics.
- **Budget-Aware Traversal**: Automatic truncation and neighbor ranking to ensure high performance even on massive graphs.
- **Modern Tech Stack**:
  - **Go 1.25**: Leveraging the latest Go performance and features.
  - **Rueidis**: High-performance Redis client with support for FalkorDB commands.
  - **Chi Router**: Clean and lightweight HTTP routing.
  - **Structured Logging**: JSON-based logging with configurable levels and context.
  - **Panic Recovery**: Robust middleware to catch and log panics with full stack traces.

---

## 🏗️ Architecture

- **Ingest Service**: Processes customer events and maintains an aggregated graph of relationships between users and entities.
- **Graph Service**: Handles complex traversal requests, applying time windows, min-event counts, and ranking metrics to extract relevant subgraphs.
- **FalkorDB**: High-speed graph database built on Redis, enabling Cypher query execution at memory speeds.

---

## 🛠️ Prerequisites

- **Go**: Version 1.25 or higher.
- **Docker & Docker Compose**: For running the FalkorDB cluster.
- **redis-cli**: Optional, for manual graph inspection.

---

## ⚙️ Setup

1. **Clone the repository**:

   ```bash
   git clone <repository-url>
   cd grapgraph
   ```

2. **Configure Environment Variables**:

   ```bash
   cp .env.exampe .env
   ```

   _Note: Default settings are optimized for the provided Docker Compose setup._

3. **Start FalkorDB Cluster**:

   ```bash
   docker compose up -d
   ```

   This starts a 3-node FalkorDB cluster exposed on ports `7001-7003`.

4. **Run the Application**:

   ```bash
   go run cmd/api/main.go
   ```

5. **Seed Demo Data (Optional)**:
   ```bash
   go run cmd/seed/main.go --reset
   ```

---

## 📡 API Documentation

### 🏁 Health Check

`GET /healthz`

- Returns `200 OK` if the database connection is healthy.

### 📥 Ingest Event

`POST /v1/ingest/event`

```json
{
  "user_id": "u_123",
  "event_type": "PAYMENT",
  "event_timestamp": "2024-03-20T10:00:00Z",
  "merchant_id_mpan": "m_777",
  "total_transaction_amount": 150.5
}
```

### 🔍 Query Subgraph

`POST /v1/graph/subgraph`

```json
{
  "root": { "type": "USER", "key": "u_123" },
  "hops": 2,
  "time_window": {
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-12-31T23:59:59Z"
  },
  "min_event_count": 2,
  "rank_neighbors_by": "event_count_30d"
}
```

### 📋 Metadata

`GET /v1/graph/metadata`

- Returns available node types, edge types, and ranking metrics.

---

## 📂 Project Structure

- `cmd/`: Application entry points (`api`, `seed`).
- `src/`: Core logic organized by domain.
  - `graph/`: Redis interaction and Cypher query templates.
  - `service/`: Business logic for Ingest and Subgraph analysis.
  - `httpapi/`: HTTP handlers, router, and middleware.
  - `model/`: DTOs and shared enums.
  - `observability/`: Logging and monitoring utilities.
- `test/`: Integration tests.

---

## 📜 License

[MIT License](LICENSE)
