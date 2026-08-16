// Package events implements storage and stats for Relay campaign events.
//
// Storage: in-memory, guarded by a single sync.RWMutex. This is a
// deliberate choice for the stated scale (a few thousand events/day across
// ~50 campaigns) — see NOTES.md Part 4 for what changes at 100M events/day.
//
// Dedup key: event_id alone, per the provider contract ("event_id ...
// identifies one event"). A duplicate event_id with an identical payload is
// treated as a harmless retry. A duplicate event_id with a *different*
// payload is flagged as a conflict rather than silently overwritten or
// dropped, since that pattern suggests an upstream bug worth surfacing.
package events

import (
	"sync"
	"time"
)

// Store is a concurrency-safe in-memory event store.
type Store struct {
	mu     sync.RWMutex
	events map[string]Event // event_id -> event
}

// NewStore returns an empty Store ready for use.
func NewStore() *Store {
	return &Store{events: make(map[string]Event)}
}

// Ingest validates and stores a single event, returning what happened.
// Safe for concurrent use.
func (s *Store) Ingest(e Event) EventResult {
	if reason, ok := Validate(e); !ok {
		return EventResult{EventID: e.EventID, Outcome: OutcomeInvalid, Reason: reason}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.events[e.EventID]
	if !exists {
		s.events[e.EventID] = e
		return EventResult{EventID: e.EventID, Outcome: OutcomeStored}
	}
	if SameEvent(existing, e) {
		return EventResult{EventID: e.EventID, Outcome: OutcomeDuplicate}
	}
	return EventResult{
		EventID: e.EventID,
		Outcome: OutcomeConflict,
		Reason:  "event_id already exists with different field values",
	}
}

// StatsFor computes stats for a single campaign by scanning stored events.
// At the stated scale this is fine; see NOTES.md Part 4 for why this would
// need to become a pre-aggregated counter at higher volume.
func (s *Store) StatsFor(campaignID string) Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := Stats{CampaignID: campaignID}
	contacts := make(map[string]bool)
	var latest time.Time

	for _, e := range s.events {
		if e.CampaignID != campaignID {
			continue
		}
		switch e.Type {
		case "sent":
			st.Sent++
		case "delivered":
			st.Delivered++
		case "opened":
			st.Opened++
		case "clicked":
			st.Clicked++
		}
		contacts[e.ContactID] = true
		if ts, err := time.Parse(time.RFC3339, e.Timestamp); err == nil && ts.After(latest) {
			latest = ts
		}
	}
	st.UniqueContacts = len(contacts)
	if !latest.IsZero() {
		st.LastEventAt = latest.Format(time.RFC3339)
	}
	return st
}

// Len returns the total number of distinct stored events. Used by tests.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}
