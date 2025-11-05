package filter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tone-labs/dewey/filter"
)

// MockFieldFilterBuilder implements FieldFilterBuilder for testing
type MockFieldFilterBuilder struct {
	fieldName string
}

func (b MockFieldFilterBuilder) Eq(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s = %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) Ne(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s != %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) Gt(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s > %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) Gte(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s >= %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) Lt(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s < %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) Lte(value any) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s <= %v", b.fieldName, value))
}

func (b MockFieldFilterBuilder) In(values []any) MockPredicate {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return MockPredicate(fmt.Sprintf("%s IN (%s)", b.fieldName, strings.Join(strs, ",")))
}

func (b MockFieldFilterBuilder) Nin(values []any) MockPredicate {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return MockPredicate(fmt.Sprintf("%s NOT IN (%s)", b.fieldName, strings.Join(strs, ",")))
}

func (b MockFieldFilterBuilder) Contains(value string) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s ILIKE '%%%s%%'", b.fieldName, value))
}

func (b MockFieldFilterBuilder) StartsWith(value string) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s ILIKE '%s%%'", b.fieldName, value))
}

func (b MockFieldFilterBuilder) EndsWith(value string) MockPredicate {
	return MockPredicate(fmt.Sprintf("%s ILIKE '%%%s'", b.fieldName, value))
}

func (b MockFieldFilterBuilder) IsNull() MockPredicate {
	return MockPredicate(fmt.Sprintf("%s IS NULL", b.fieldName))
}

func (b MockFieldFilterBuilder) IsNotNull() MockPredicate {
	return MockPredicate(fmt.Sprintf("%s IS NOT NULL", b.fieldName))
}

func TestApplyStructuredFilters_SingleFilter(t *testing.T) {
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

	fieldBuilders := map[string]filter.FieldFilterBuilder[MockPredicate]{
		"email": MockFieldFilterBuilder{fieldName: "email"},
		"age":   MockFieldFilterBuilder{fieldName: "age"},
		"name":  MockFieldFilterBuilder{fieldName: "name"},
	}

	tests := []struct {
		name     string
		group    filter.FilterGroup
		expected string
	}{
		{
			name: "eq operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpEq, Value: "john@example.com"},
				},
				Logic: "and",
			},
			expected: "email = john@example.com",
		},
		{
			name: "ne operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpNe, Value: "spam@example.com"},
				},
				Logic: "and",
			},
			expected: "email != spam@example.com",
		},
		{
			name: "gt operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpGt, Value: 18},
				},
				Logic: "and",
			},
			expected: "age > 18",
		},
		{
			name: "gte operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpGte, Value: 18},
				},
				Logic: "and",
			},
			expected: "age >= 18",
		},
		{
			name: "lt operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpLt, Value: 65},
				},
				Logic: "and",
			},
			expected: "age < 65",
		},
		{
			name: "lte operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpLte, Value: 65},
				},
				Logic: "and",
			},
			expected: "age <= 65",
		},
		{
			name: "in operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "name", Operator: filter.OpIn, Value: []any{"John", "Jane", "Bob"}},
				},
				Logic: "and",
			},
			expected: "name IN (John,Jane,Bob)",
		},
		{
			name: "nin operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "name", Operator: filter.OpNin, Value: []any{"Spam", "Bot"}},
				},
				Logic: "and",
			},
			expected: "name NOT IN (Spam,Bot)",
		},
		{
			name: "contains operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpContains, Value: "john"},
				},
				Logic: "and",
			},
			expected: "email ILIKE '%john%'",
		},
		{
			name: "startswith operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpStartsWith, Value: "john"},
				},
				Logic: "and",
			},
			expected: "email ILIKE 'john%'",
		},
		{
			name: "endswith operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpEndsWith, Value: "@example.com"},
				},
				Logic: "and",
			},
			expected: "email ILIKE '%@example.com'",
		},
		{
			name: "null operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpNull, Value: nil},
				},
				Logic: "and",
			},
			expected: "email IS NULL",
		},
		{
			name: "nnull operator",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpNnull, Value: nil},
				},
				Logic: "and",
			},
			expected: "email IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := filter.ApplyStructuredFilters(query, cfg, tt.group, fieldBuilders, builder)

			if len(result.predicates) != 1 {
				t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
			}

			if result.predicates[0] != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, result.predicates[0])
			}
		})
	}
}

func TestApplyStructuredFilters_MultipleFilters_AND(t *testing.T) {
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

	fieldBuilders := map[string]filter.FieldFilterBuilder[MockPredicate]{
		"email": MockFieldFilterBuilder{fieldName: "email"},
		"age":   MockFieldFilterBuilder{fieldName: "age"},
		"name":  MockFieldFilterBuilder{fieldName: "name"},
	}

	tests := []struct {
		name     string
		group    filter.FilterGroup
		expected string
	}{
		{
			name: "two filters with AND",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpGte, Value: 18},
					{Field: "age", Operator: filter.OpLte, Value: 65},
				},
				Logic: "and",
			},
			expected: "(age >= 18 AND age <= 65)",
		},
		{
			name: "three filters with AND",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "age", Operator: filter.OpGte, Value: 18},
					{Field: "email", Operator: filter.OpContains, Value: "john"},
					{Field: "name", Operator: filter.OpNe, Value: "Admin"},
				},
				Logic: "and",
			},
			expected: "(age >= 18 AND email ILIKE '%john%' AND name != Admin)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := filter.ApplyStructuredFilters(query, cfg, tt.group, fieldBuilders, builder)

			if len(result.predicates) != 1 {
				t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
			}

			if result.predicates[0] != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, result.predicates[0])
			}
		})
	}
}

func TestApplyStructuredFilters_MultipleFilters_OR(t *testing.T) {
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

	fieldBuilders := map[string]filter.FieldFilterBuilder[MockPredicate]{
		"email": MockFieldFilterBuilder{fieldName: "email"},
		"name":  MockFieldFilterBuilder{fieldName: "name"},
	}

	tests := []struct {
		name     string
		group    filter.FilterGroup
		expected string
	}{
		{
			name: "two filters with OR",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpContains, Value: "john"},
					{Field: "name", Operator: filter.OpContains, Value: "john"},
				},
				Logic: "or",
			},
			expected: "(email ILIKE '%john%' OR name ILIKE '%john%')",
		},
		{
			name: "three filters with OR",
			group: filter.FilterGroup{
				Filters: []filter.Filter{
					{Field: "email", Operator: filter.OpEq, Value: "admin@example.com"},
					{Field: "email", Operator: filter.OpEq, Value: "root@example.com"},
					{Field: "name", Operator: filter.OpEq, Value: "SuperAdmin"},
				},
				Logic: "or",
			},
			expected: "(email = admin@example.com OR email = root@example.com OR name = SuperAdmin)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := filter.ApplyStructuredFilters(query, cfg, tt.group, fieldBuilders, builder)

			if len(result.predicates) != 1 {
				t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
			}

			if result.predicates[0] != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, result.predicates[0])
			}
		})
	}
}

func TestApplyStructuredFilters_EdgeCases(t *testing.T) {
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

	fieldBuilders := map[string]filter.FieldFilterBuilder[MockPredicate]{
		"email": MockFieldFilterBuilder{fieldName: "email"},
		"age":   MockFieldFilterBuilder{fieldName: "age"},
	}

	t.Run("empty filter group", func(t *testing.T) {
		query := &MockQuery{}
		group := filter.FilterGroup{
			Filters: []filter.Filter{},
			Logic:   "and",
		}

		result := filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, builder)

		if len(result.predicates) != 0 {
			t.Errorf("expected no predicates for empty group, got %d", len(result.predicates))
		}
	})

	t.Run("unknown field ignored", func(t *testing.T) {
		query := &MockQuery{}
		group := filter.FilterGroup{
			Filters: []filter.Filter{
				{Field: "unknown_field", Operator: filter.OpEq, Value: "test"},
			},
			Logic: "and",
		}

		result := filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, builder)

		if len(result.predicates) != 0 {
			t.Errorf("expected no predicates for unknown field, got %d", len(result.predicates))
		}
	})

	t.Run("mix of known and unknown fields", func(t *testing.T) {
		query := &MockQuery{}
		group := filter.FilterGroup{
			Filters: []filter.Filter{
				{Field: "email", Operator: filter.OpEq, Value: "john@example.com"},
				{Field: "unknown", Operator: filter.OpEq, Value: "test"},
				{Field: "age", Operator: filter.OpGte, Value: 18},
			},
			Logic: "and",
		}

		result := filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, builder)

		if len(result.predicates) != 1 {
			t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
		}

		expected := "(email = john@example.com AND age >= 18)"
		if result.predicates[0] != expected {
			t.Errorf("\nexpected: %s\ngot:      %s", expected, result.predicates[0])
		}
	})

	t.Run("default to AND if logic not specified", func(t *testing.T) {
		query := &MockQuery{}
		group := filter.FilterGroup{
			Filters: []filter.Filter{
				{Field: "age", Operator: filter.OpGte, Value: 18},
				{Field: "age", Operator: filter.OpLte, Value: 65},
			},
			Logic: "", // Empty, should default to AND
		}

		result := filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, builder)

		if len(result.predicates) != 1 {
			t.Fatalf("expected 1 predicate, got %d", len(result.predicates))
		}

		expected := "(age >= 18 AND age <= 65)"
		if result.predicates[0] != expected {
			t.Errorf("\nexpected: %s\ngot:      %s", expected, result.predicates[0])
		}
	})
}
