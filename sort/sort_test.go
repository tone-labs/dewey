package sort_test

import (
	"testing"

	"github.com/tone-labs/dewey/sort"
)

// MockQuery simulates a query builder for testing
type MockQuery struct {
	orderOpts []string
}

// MockOrderBuilder creates mock order options for testing
type MockOrderBuilder struct{}

func (MockOrderBuilder) Asc(field string) any {
	return "ASC:" + field
}

func (MockOrderBuilder) Desc(field string) any {
	return "DESC:" + field
}

func TestApply(t *testing.T) {
	cfg := sort.Config[*MockQuery]{
		Order: func(q *MockQuery, opts ...any) *MockQuery {
			for _, opt := range opts {
				q.orderOpts = append(q.orderOpts, opt.(string))
			}
			return q
		},
	}

	fields := sort.Fields{
		"email":      "users.email",
		"created_at": "users.created_at",
		"name":       "users.name",
	}

	builder := MockOrderBuilder{}

	tests := []struct {
		name     string
		sortBy   string
		sortDir  string
		expected []string
	}{
		{
			name:     "ascending sort",
			sortBy:   "email",
			sortDir:  "asc",
			expected: []string{"ASC:users.email"},
		},
		{
			name:     "descending sort",
			sortBy:   "created_at",
			sortDir:  "desc",
			expected: []string{"DESC:users.created_at"},
		},
		{
			name:     "empty sortBy ignored",
			sortBy:   "",
			sortDir:  "asc",
			expected: nil,
		},
		{
			name:     "unknown field ignored",
			sortBy:   "unknown_field",
			sortDir:  "asc",
			expected: nil,
		},
		{
			name:     "invalid direction defaults to asc",
			sortBy:   "name",
			sortDir:  "invalid",
			expected: []string{"ASC:users.name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := sort.Apply(query, cfg, fields, builder, tt.sortBy, tt.sortDir)

			if len(result.orderOpts) != len(tt.expected) {
				t.Fatalf("expected %d order opts, got %d", len(tt.expected), len(result.orderOpts))
			}

			for i, expected := range tt.expected {
				if result.orderOpts[i] != expected {
					t.Errorf("order opt %d: expected %s, got %s", i, expected, result.orderOpts[i])
				}
			}
		})
	}
}

func TestApplyMultiple(t *testing.T) {
	cfg := sort.Config[*MockQuery]{
		Order: func(q *MockQuery, opts ...any) *MockQuery {
			for _, opt := range opts {
				q.orderOpts = append(q.orderOpts, opt.(string))
			}
			return q
		},
	}

	fields := sort.Fields{
		"email":      "users.email",
		"created_at": "users.created_at",
		"name":       "users.name",
	}

	builder := MockOrderBuilder{}

	tests := []struct {
		name     string
		sorts    []sort.Criteria
		expected []string
	}{
		{
			name: "multiple sorts",
			sorts: []sort.Criteria{
				{Field: "name", Order: sort.Asc},
				{Field: "email", Order: sort.Desc},
			},
			expected: []string{"ASC:users.name", "DESC:users.email"},
		},
		{
			name: "single sort",
			sorts: []sort.Criteria{
				{Field: "created_at", Order: sort.Desc},
			},
			expected: []string{"DESC:users.created_at"},
		},
		{
			name:     "empty sorts",
			sorts:    []sort.Criteria{},
			expected: nil,
		},
		{
			name: "skip unknown fields",
			sorts: []sort.Criteria{
				{Field: "name", Order: sort.Asc},
				{Field: "unknown", Order: sort.Asc},
				{Field: "email", Order: sort.Desc},
			},
			expected: []string{"ASC:users.name", "DESC:users.email"},
		},
		{
			name: "all unknown fields",
			sorts: []sort.Criteria{
				{Field: "unknown1", Order: sort.Asc},
				{Field: "unknown2", Order: sort.Asc},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := sort.ApplyMultiple(query, cfg, fields, builder, tt.sorts)

			if len(result.orderOpts) != len(tt.expected) {
				t.Fatalf("expected %d order opts, got %d", len(tt.expected), len(result.orderOpts))
			}

			for i, expected := range tt.expected {
				if result.orderOpts[i] != expected {
					t.Errorf("order opt %d: expected %s, got %s", i, expected, result.orderOpts[i])
				}
			}
		})
	}
}
