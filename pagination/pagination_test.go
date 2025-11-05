package pagination_test

import (
	"testing"

	"github.com/tone-labs/dewey/pagination"
)

// MockQuery simulates a query builder for testing
type MockQuery struct {
	limit  int
	offset int
}

func TestApply(t *testing.T) {
	cfg := pagination.Config[*MockQuery]{
		Limit: func(q *MockQuery, n int) *MockQuery {
			q.limit = n
			return q
		},
		Offset: func(q *MockQuery, n int) *MockQuery {
			q.offset = n
			return q
		},
	}

	tests := []struct {
		name           string
		limit          int
		offset         int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "basic pagination",
			limit:          25,
			offset:         0,
			expectedLimit:  25,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			limit:          25,
			offset:         25,
			expectedLimit:  25,
			expectedOffset: 25,
		},
		{
			name:           "zero limit ignored",
			limit:          0,
			offset:         10,
			expectedLimit:  0,
			expectedOffset: 10,
		},
		{
			name:           "negative limit ignored",
			limit:          -1,
			offset:         10,
			expectedLimit:  0,
			expectedOffset: 10,
		},
		{
			name:           "zero offset ignored",
			limit:          25,
			offset:         0,
			expectedLimit:  25,
			expectedOffset: 0,
		},
		{
			name:           "negative offset ignored",
			limit:          25,
			offset:         -1,
			expectedLimit:  25,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &MockQuery{}
			result := pagination.Apply(query, cfg, tt.limit, tt.offset)

			if result.limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, result.limit)
			}
			if result.offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, result.offset)
			}
		})
	}
}

func TestPage(t *testing.T) {
	t.Run("HasNextPage", func(t *testing.T) {
		tests := []struct {
			name     string
			page     pagination.Page[string]
			expected bool
		}{
			{
				name:     "has next page",
				page:     pagination.NewPage([]string{"a", "b"}, 100, 25, 0),
				expected: true,
			},
			{
				name:     "no next page",
				page:     pagination.NewPage([]string{"a", "b"}, 100, 25, 75),
				expected: false,
			},
			{
				name:     "last page partial",
				page:     pagination.NewPage([]string{"a"}, 76, 25, 75),
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.page.HasNextPage()
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("HasPrevPage", func(t *testing.T) {
		tests := []struct {
			name     string
			page     pagination.Page[string]
			expected bool
		}{
			{
				name:     "first page",
				page:     pagination.NewPage([]string{"a", "b"}, 100, 25, 0),
				expected: false,
			},
			{
				name:     "second page",
				page:     pagination.NewPage([]string{"a", "b"}, 100, 25, 25),
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.page.HasPrevPage()
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("PageNumber", func(t *testing.T) {
		tests := []struct {
			name     string
			page     pagination.Page[string]
			expected int
		}{
			{
				name:     "first page",
				page:     pagination.NewPage([]string{"a"}, 100, 25, 0),
				expected: 1,
			},
			{
				name:     "second page",
				page:     pagination.NewPage([]string{"a"}, 100, 25, 25),
				expected: 2,
			},
			{
				name:     "third page",
				page:     pagination.NewPage([]string{"a"}, 100, 25, 50),
				expected: 3,
			},
			{
				name:     "zero limit defaults to page 1",
				page:     pagination.NewPage([]string{"a"}, 100, 0, 50),
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.page.PageNumber()
				if result != tt.expected {
					t.Errorf("expected %d, got %d", tt.expected, result)
				}
			})
		}
	})

	t.Run("TotalPages", func(t *testing.T) {
		tests := []struct {
			name     string
			page     pagination.Page[string]
			expected int
		}{
			{
				name:     "exact pages",
				page:     pagination.NewPage([]string{"a"}, 100, 25, 0),
				expected: 4,
			},
			{
				name:     "partial last page",
				page:     pagination.NewPage([]string{"a"}, 101, 25, 0),
				expected: 5,
			},
			{
				name:     "single page",
				page:     pagination.NewPage([]string{"a"}, 10, 25, 0),
				expected: 1,
			},
			{
				name:     "zero limit defaults to 1 page",
				page:     pagination.NewPage([]string{"a"}, 100, 0, 0),
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.page.TotalPages()
				if result != tt.expected {
					t.Errorf("expected %d, got %d", tt.expected, result)
				}
			})
		}
	})
}
