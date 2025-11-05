package filter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tone-labs/dewey/filter"
)

// MockQuery simulates a query builder for testing
type MockQuery struct {
	predicates []string
}

// MockPredicate represents a WHERE condition
type MockPredicate string

// Mock predicate builder functions
func mockIDIn(ids ...any) MockPredicate {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("%v", id)
	}
	return MockPredicate(fmt.Sprintf("ID IN (%s)", strings.Join(idStrs, ",")))
}

func mockOr(predicates ...MockPredicate) MockPredicate {
	strs := make([]string, len(predicates))
	for i, p := range predicates {
		strs[i] = string(p)
	}
	return MockPredicate(fmt.Sprintf("(%s)", strings.Join(strs, " OR ")))
}

func mockAnd(predicates ...MockPredicate) MockPredicate {
	strs := make([]string, len(predicates))
	for i, p := range predicates {
		strs[i] = string(p)
	}
	return MockPredicate(fmt.Sprintf("(%s)", strings.Join(strs, " AND ")))
}

func mockContainsFold(field string) func(string) MockPredicate {
	return func(value string) MockPredicate {
		return MockPredicate(fmt.Sprintf("%s ILIKE '%%%s%%'", field, value))
	}
}

func TestApplyIDs(t *testing.T) {
	cfg := filter.Config[*MockQuery, MockPredicate]{
		Where: func(q *MockQuery, p MockPredicate) *MockQuery {
			q.predicates = append(q.predicates, string(p))
			return q
		},
	}

	builder := filter.PredicateBuilder[MockPredicate]{
		IDIn: mockIDIn,
		Or:   mockOr,
		And:  mockAnd,
	}

	tests := []struct {
		name     string
		ids      []any
		expected string
	}{
		{
			name:     "single ID",
			ids:      []any{"123"},
			expected: "ID IN (123)",
		},
		{
			name:     "multiple IDs",
			ids:      []any{"123", "456", "789"},
			expected: "ID IN (123,456,789)",
		},
		{
			name:     "empty IDs ignored",
			ids:      []any{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := filter.ApplyIDs(query, cfg, builder, tt.ids)

			if tt.expected == "" {
				if len(result.predicates) != 0 {
					t.Errorf("expected no predicates, got %v", result.predicates)
				}
				return
			}

			if len(result.predicates) != 1 {
				t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
			}

			if result.predicates[0] != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.predicates[0])
			}
		})
	}
}

func TestApplySearch(t *testing.T) {
	cfg := filter.Config[*MockQuery, MockPredicate]{
		Where: func(q *MockQuery, p MockPredicate) *MockQuery {
			q.predicates = append(q.predicates, string(p))
			return q
		},
	}

	builder := filter.PredicateBuilder[MockPredicate]{
		IDIn: mockIDIn,
		Or:   mockOr,
		And:  mockAnd,
	}

	fields := filter.SearchFields[MockPredicate]{
		"email":      mockContainsFold("email"),
		"first_name": mockContainsFold("first_name"),
		"last_name":  mockContainsFold("last_name"),
	}

	tests := []struct {
		name     string
		search   string
		expected string
	}{
		{
			name:     "single token searches all fields with OR",
			search:   "john",
			expected: "(email ILIKE '%john%' OR first_name ILIKE '%john%' OR last_name ILIKE '%john%')",
		},
		{
			name:   "multiple tokens use AND between tokens, OR within fields",
			search: "john doe",
			expected: "((email ILIKE '%john%' OR first_name ILIKE '%john%' OR last_name ILIKE '%john%') AND " +
				"(email ILIKE '%doe%' OR first_name ILIKE '%doe%' OR last_name ILIKE '%doe%'))",
		},
		{
			name:   "three tokens",
			search: "john doe smith",
			expected: "((email ILIKE '%john%' OR first_name ILIKE '%john%' OR last_name ILIKE '%john%') AND " +
				"(email ILIKE '%doe%' OR first_name ILIKE '%doe%' OR last_name ILIKE '%doe%') AND " +
				"(email ILIKE '%smith%' OR first_name ILIKE '%smith%' OR last_name ILIKE '%smith%'))",
		},
		{
			name:     "empty search ignored",
			search:   "",
			expected: "",
		},
		{
			name:     "whitespace-only search ignored",
			search:   "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := filter.ApplySearch(query, cfg, fields, builder, tt.search)

			if tt.expected == "" {
				if len(result.predicates) != 0 {
					t.Errorf("expected no predicates, got %v", result.predicates)
				}
				return
			}

			if len(result.predicates) != 1 {
				t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
			}

			if result.predicates[0] != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, result.predicates[0])
			}
		})
	}
}

func TestApplyWhere(t *testing.T) {
	cfg := filter.Config[*MockQuery, MockPredicate]{
		Where: func(q *MockQuery, p MockPredicate) *MockQuery {
			q.predicates = append(q.predicates, string(p))
			return q
		},
	}

	query := &MockQuery{}
	predicate := MockPredicate("is_active = true")

	result := filter.ApplyWhere(query, cfg, predicate)

	if len(result.predicates) != 1 {
		t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
	}

	if result.predicates[0] != "is_active = true" {
		t.Errorf("expected 'is_active = true', got %s", result.predicates[0])
	}
}
