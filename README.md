# Grapgraph

`Grapgraph` is a high-performance fraud detection and financial network analysis platform built with **Go** and **FalkorDB** (Redis Graph). It allows for real-time ingestion of financial events and complex multi-hop relationship querying to detect suspicious patterns like money laundering, account takeovers, and shared entity fraud.

---

## 🚀 Key Features

- **Real-time Event Ingestion**: Aggregated upserts of user events (Logins, Payments, Withdrawals, etc.) into a graph structure.
- **Industrial-Grade API Framework**: Rebuilt with **Goa v3** for design-first API development, providing automatic validation and OpenAPI documentation.
- **Modern Spider-Web UI**: A premium D3.js powered dark-mode visualization tool with glassmorphism panels, relationship directionality (arrows), and entity icons.
- **Real-time Event Ingestion**: Aggregated upserts of user events (Logins, Payments, Withdrawals, etc.) into a graph structure.
- **Budget-Aware Subgraph Analysis**: Performant multi-hop traversal (up to 3 hops) with neighbor ranking and automatic results truncation.
- **Rich Seed Data**: Includes realistic fraud scenarios (Money mules, Bot networks, Account takeovers).

---

## 🏗️ Architecture

- **Goa v3 Layer**: Defines the API DSL in `/design`, generating type-safe controllers and transport encoders/decoders.
- **Business Services**: Core logic in `src/service/` handles graph traversals and ingestion rules.
- **FalkorDB Repository**: High-performance interaction with Redis Graph using localized Cypher blocks.
- **Web Frontend**: Modern single-page application using D3.js and responsive CSS.

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
- `design/`: Goa v3 API Design DSL.
- `gen/`: Re-generatable Goa boilerplate (HTTP, endpoints, types).
- `src/`: Core logic organized by domain.
  - `goa_services/`: Wrappers implementing Goa interfaces.
  - `service/`: Core business logic (Ingest and Graph).
  - `graph/`: Redis interaction and Cypher templates.
  - `model/`: Shared domain objects and DTOs.
  - `observability/`: Logging and monitoring.
- `web/`: Modern D3 visualization interface.

---

## 📜 License

[MIT License](LICENSE)
