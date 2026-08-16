# Relay campaign events service

## Project layout

```
.
├── cmd/server/main.go        entrypoint — wires and starts the HTTP server
├── internal/events/
│   ├── model.go               Event, Stats, Outcome types + validation
│   ├── store.go                concurrency-safe in-memory store + stats
│   ├── handlers.go            POST /events, GET /campaigns/{id}/stats
│   └── store_test.go          tests for dedup, conflict, concurrency
├── debugging/                 Part 3 — see debugging/README.md (not attempted, see NOTES.md)
├── NOTES.md                    Parts 1, 4, 5 + completion summary
├── BUGS.md                    Part 3 findings (not completed — time box, see NOTES.md)
└── AI_USAGE.md                 AI usage disclosure
```

## Run

```
go run ./cmd/server
```

Listens on `:8080`.

## Test

```
go test ./...
go test -race ./...
```

## Endpoints

### POST /events

Accepts a JSON array of events. Returns 200 with a per-event outcome even if some
events in the batch are invalid or duplicates — the batch itself is never rejected
wholesale.

```
curl -X POST localhost:8080/events \
  -H 'Content-Type: application/json' \
  -d '[
    {"event_id":"evt_1","campaign_id":"cmp_a","contact_id":"ct_1","type":"sent","timestamp":"2026-08-10T09:00:00Z"},
    {"event_id":"evt_2","campaign_id":"cmp_a","contact_id":"ct_1","type":"delivered","timestamp":"2026-08-10T09:00:05Z"}
  ]'
```

Response:
```json
{
  "accepted": 2,
  "rejected": 0,
  "results": [
    {"event_id": "evt_1", "outcome": "stored"},
    {"event_id": "evt_2", "outcome": "stored"}
  ]
}
```

Posting the exact same array again returns `"outcome": "duplicate"` for both, and
stats are unchanged.

### GET /campaigns/{campaign_id}/stats

```
curl localhost:8080/campaigns/cmp_a/stats
```

```json
{
  "campaign_id": "cmp_a",
  "sent": 1,
  "delivered": 1,
  "opened": 0,
  "clicked": 0,
  "unique_contacts": 1,
  "last_event_at": "2026-08-10T09:00:05Z"
}
```

## Not implemented (see NOTES.md)

- `GET /campaigns/{id}/events` (optional in the brief, skipped)
- Persistent storage — in-memory only, resets on restart (deliberate for stated
  scale; see the Part 4 scale memo in NOTES.md)
- Part 3 debugging exercise — skipped under the time box; see NOTES.md for why and
  what I'd do first given more time
