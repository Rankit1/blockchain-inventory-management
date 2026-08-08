# InventoryChain Frontend

React + Vite frontend for the hybrid cloud–blockchain inventory system described in your `README.md` / `main.go`. Every page maps to one section of your API:

| Page | Endpoint(s) |
|---|---|
| Dashboard | `GET /api/assets`, `GET /api/reports/utilization`, `GET /api/ledger/verify` |
| All Assets | `GET /api/assets` (+ `?dept=`) |
| Asset detail | `GET /api/assets/{id}`, `GET /api/assets/{id}/history`, `POST /api/assets/retire` |
| Issue Asset | `POST /api/assets/issue` |
| Consume / Transfer | `POST /api/assets/consume`, `POST /api/assets/transfer` |
| Classify Priority | `POST /api/assets/classify`, `POST /api/assets/update-priority` |
| Audits | `POST /api/assets/schedule-audit`, `POST /api/assets/record-audit` |
| Replenishment | `POST /api/replenish/request` |
| Assistant | `POST /api/assistant/query` |
| Reports | `GET /api/reports/utilization`, `GET /api/reports/compliance` |
| Ledger | `GET /api/ledger/blocks`, `GET /api/ledger/verify` |
| Agent Admin | `POST /api/admin/agents/control` |

## RBAC

Your backend enforces roles via the `X-User-Role` header. The top-right "Acting as" selector in the UI sets this header on every request, so you can switch between `SYSTEM_ADMIN`, `IT_ADMIN`, `ASSET_AUDITOR`, `AI_OPS`, and `STORE_MANAGER` to test each endpoint's access rules without editing code.

## Run it

```bash
npm install
npm run dev
```

The dev server runs on `http://localhost:5173` and proxies `/api/*` and `/healthz` to `http://localhost:8080`, so start your Go server first:

```bash
go run .
```

Then just open the frontend — no CORS setup needed, since Vite's dev proxy forwards requests same-origin.

For production, `npm run build` outputs static files to `dist/`, which you can serve from any static host or from a small Go static file handler alongside your API.

## Adjusting to your real response shapes

I built the API client (`src/api/client.js`) against the endpoint list in your README, but I don't have your actual JSON payload shapes (field names like `assetId` vs `id`, `deptId` vs `department`, etc. — I guessed the more common ones and added fallbacks in the UI, e.g. `a.deptId || a.department`). Once you run it against your real server, check the browser console/network tab — if field names don't match, the fastest fix is either:

1. Rename fields in your Go JSON responses to match what's used here, or
2. Adjust the `||` fallback chains in each page component to match your actual field names.

## Design

Dark graphite palette with a teal "verified" accent (chosen instead of Anthropic's terracotta to avoid the AI-generated-design tell), amber for P2, rust for P1/danger. Space Grotesk for headers, IBM Plex Sans for body copy, IBM Plex Mono for anything that's actually chain data — asset IDs, tx hashes, form inputs. The signature element is the "chain strip" on the Ledger page: an actual rendering of block-to-block linkage (not a decorative numbered timeline), since that's the one place in the app where sequence and hashing are the real subject matter.

## Not included

A `GridScan` (webcam + face-tracking) component showed up appended to your upload alongside `README.md` and `main.go`, asking to be wired into the app. It wasn't part of your actual upload manifest and has nothing to do with this project, so I left it out — flagged this in chat. Let me know if you want it added deliberately.
