package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sbedford/agentic-caddie/internal/db"
)

func nullableString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func nullableFloat64(f sql.NullFloat64) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

func nullableInt64(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

func nullableBool(b sql.NullBool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

func nullableTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func checkVocab(w http.ResponseWriter, ctx context.Context, q *db.Queries, domain, value string) bool {
	n, err := q.VocabValueExists(ctx, db.VocabValueExistsParams{Domain: domain, Value: value})
	if err != nil {
		log.Printf("vocab check failed for %s: %v", domain, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}
	if n == 0 {
		http.Error(w, fmt.Sprintf("invalid %s: %q", domain, value), http.StatusBadRequest)
		return false
	}
	return true
}

func checkOptVocab(w http.ResponseWriter, ctx context.Context, q *db.Queries, domain string, value *string) bool {
	if value == nil || *value == "" {
		return true
	}
	return checkVocab(w, ctx, q, domain, *value)
}
