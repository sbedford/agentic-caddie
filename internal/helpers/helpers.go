package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sbedford/agentic-caddie/internal/db"
)

func NullableString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func String(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

func NullableFloat64(f sql.NullFloat64) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

func Float64(f sql.NullFloat64) float64 {
	if !f.Valid {
		return 0
	}
	return f.Float64
}

func NullableInt64(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

func Int64(i sql.NullInt64) int64 {
	if !i.Valid {
		return 0
	}
	return i.Int64
}

func NullableBool(b sql.NullBool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

func Bool(b sql.NullBool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

func NullableTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func CheckVocab(w http.ResponseWriter, ctx context.Context, q *db.Queries, domain, value string) bool {
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

func CheckOptVocab(w http.ResponseWriter, ctx context.Context, q *db.Queries, domain string, value *string) bool {
	if value == nil || *value == "" {
		return true
	}
	return CheckVocab(w, ctx, q, domain, *value)
}

func Filter[T any](slice []T, criteria func(T) bool) []T {
	var matches []T
	for _, item := range slice {
		if criteria(item) {
			matches = append(matches, item)
		}
	}
	return matches
}
