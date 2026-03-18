# Go-Starter – Product Catalog API

## What it does

- **Search** – `GET /search/:query` – Proxies to Unbxd search API.
- **Ingest** – `POST /ingest` – Upsert products (JSON single or array) into PostgreSQL.
- **Browse** – `GET /browse/:id` – One product by ID. `GET /browse/category/:catlevel1Name/:catlevel2Name/:catlevel3Name/:categoryType` – List by category (`page`, `sort`, `minPrice`, `maxPrice`).
- **Delete** – `DELETE /delete_prod/:id` or `DELETE /delete_prod` with body `{"ids":["id1","id2"]}`.

Uses **PostgreSQL** (required) and **Redis** (optional cache). Health: `/pong`, `/monitor`.

## Quick start

1. Start dependencies:
   ```bash
   docker compose up -d
   ```
2. Create the table:
   ```bash
   psql -h localhost -U netcoreunbxd -d unbxd_internship -f migrations/001_init_catalog.sql
   ```
3. Run:
   ```bash
   cp .env.example .env   # edit if needed
   make gobuildmac       # or make gobuild for Linux
   ./bin/gostarter.bin start
   ```
4. Try APIs (e.g. ingest from `catalog.json`, then browse/search).
