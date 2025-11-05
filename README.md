# 📚 Dewey - Your REST API Librarian

**Database-agnostic query utilities for Go REST APIs**

Dewey is a lightweight, zero-dependency library that helps you organize, filter, sort, and paginate your REST API queries. Just like a librarian helps you find the right books, Dewey helps you find the right data.

## Why Dewey?

Named after the Dewey Decimal System, this library brings the same organizational principles to your REST APIs:
- **Categorize** - Filter data with multi-token search
- **Sort** - Order results by any field
- **Paginate** - Browse through results systematically
- **No lock-in** - Works with any ORM or query builder

## Installation

```bash
go get github.com/tone-labs/dewey
```

## Philosophy

Most Go REST APIs reinvent pagination, filtering, and sorting for every project. Dewey extracts these patterns into reusable, type-safe utilities that work with *any* database layer.

**Key principles:**
- **Zero external dependencies** - Only Go stdlib
- **ORM-agnostic** - Works with Ent, GORM, sqlc, sqlx, or raw SQL
- **No reflection** - Uses explicit adapter functions for type safety
- **No magic** - Clear, simple code you can understand and debug

## Quick Start

### With Ent

```go
import (
    "github.com/tone-labs/dewey/pagination"
    "github.com/tone-labs/dewey/sort"
    "github.com/tone-labs/dewey/filter"
)

func ListUsers(ctx context.Context, client *ent.Client, limit, offset int) ([]*ent.User, error) {
    // 1. Create configs (one-time setup, can be package-level variables)
    paginationCfg := pagination.Config[*ent.UserQuery]{
        Limit: func(q *ent.UserQuery, n int) *ent.UserQuery {
            return q.Limit(n)
        },
        Offset: func(q *ent.UserQuery, n int) *ent.UserQuery {
            return q.Offset(n)
        },
    }

    // 2. Build and apply
    query := client.User.Query()
    query = pagination.Apply(query, paginationCfg, limit, offset)

    return query.All(ctx)
}
```

## Packages

### 📖 `pagination` - Browse your data catalog

Dewey helps you navigate through large result sets with offset/limit pagination.

```go
cfg := pagination.Config[*ent.UserQuery]{
    Limit: func(q *ent.UserQuery, n int) *ent.UserQuery {
        return q.Limit(n)
    },
    Offset: func(q *ent.UserQuery, n int) *ent.UserQuery {
        return q.Offset(n)
    },
}

query = pagination.Apply(query, cfg, 25, 0) // First page, 25 items
```

**Features:**
- Simple offset/limit pagination
- `Page[T]` helper type with metadata (total, hasNext, hasPrev, pageNumber)
- Zero or negative values are ignored (allows optional pagination)

### 🗂️ `sort` - Organize by any criteria

Like the Dewey Decimal System organizes books, sort organizes your query results.

```go
// Define which fields can be sorted (JSON name -> DB column)
fields := sort.Fields{
    "email":      "users.email",
    "created_at": "users.created_at",
}

// Create order builder for your ORM
type EntOrderBuilder struct{}

func (EntOrderBuilder) Asc(field string) any {
    return func(s *sql.Selector) {
        s.OrderBy(sql.Asc(field))
    }
}

func (EntOrderBuilder) Desc(field string) any {
    return func(s *sql.Selector) {
        s.OrderBy(sql.Desc(field))
    }
}

// Apply sorting
cfg := sort.Config[*ent.UserQuery]{
    Order: func(q *ent.UserQuery, opts ...any) *ent.UserQuery {
        entOpts := make([]func(*sql.Selector), len(opts))
        for i, opt := range opts {
            entOpts[i] = opt.(func(*sql.Selector))
        }
        return q.Order(entOpts...)
    },
}

query = sort.Apply(query, cfg, fields, EntOrderBuilder{}, "email", "desc")
```

**Features:**
- Map clean JSON field names to DB columns
- Type-safe order builders
- Multi-field sorting with `ApplyMultiple`
- Unknown fields are safely ignored

### 🔍 `filter` - Find exactly what you need

Dewey's card catalog helps you search through your data efficiently. The filter package provides three approaches:

#### **1. Structured Filters** (NEW) - Django/React-Admin style filtering

Perfect for building UI-driven filters with specific operators.

```go
// Define field filter builders for each filterable field
type UserEmailFilterBuilder struct{}

func (b UserEmailFilterBuilder) Eq(value any) predicate.User {
    return user.EmailEQ(value.(string))
}

func (b UserEmailFilterBuilder) Contains(value string) predicate.User {
    return user.EmailContainsFold(value)
}

// Implement other operators as needed...

// Apply structured filters
group := filter.FilterGroup{
    Filters: []filter.Filter{
        {Field: "email", Operator: filter.OpContains, Value: "john"},
        {Field: "age", Operator: filter.OpGte, Value: 18},
        {Field: "is_active", Operator: filter.OpEq, Value: true},
    },
    Logic: "and", // or "or"
}

fieldBuilders := map[string]filter.FieldFilterBuilder[predicate.User]{
    "email":     UserEmailFilterBuilder{},
    "age":       UserAgeFilterBuilder{},
    "is_active": UserActiveFilterBuilder{},
}

query = filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, predicates)
```

**Supported operators:**
- `eq`, `ne` - Equality/inequality
- `gt`, `gte`, `lt`, `lte` - Comparisons
- `in`, `nin` - Array membership
- `contains`, `startswith`, `endswith` - String matching
- `null`, `nnull` - Null checks

**Logic modes:**
- `"and"` - All filters must match (default)
- `"or"` - Any filter can match

#### **2. Full-Text Search** - Quick keyword search

Great for global search boxes that search across multiple fields.

```go
// Define searchable fields
searchFields := filter.SearchFields[predicate.User]{
    "email":      user.EmailContainsFold,
    "first_name": user.FirstNameContainsFold,
    "last_name":  user.LastNameContainsFold,
}

// Apply search
query = filter.ApplySearch(query, cfg, searchFields, predicates, "john doe")
```

**Search behavior:**
- Tokenizes input on whitespace: `"john doe"` → `["john", "doe"]`
- Each token matches ANY field (OR logic): `email ILIKE '%john%' OR first_name ILIKE '%john%'`
- All tokens must match (AND logic): `(... token1 ...) AND (... token2 ...)`
- Case-insensitive by default

#### **3. Utility Functions**

```go
// Filter by multiple IDs (for Refine.js getMany)
query = filter.ApplyIDs(query, cfg, builder, ids)

// Apply any custom predicate
query = filter.ApplyWhere(query, cfg, customPredicate)
```

**Combining approaches:**
You can use structured filters AND search together:

```go
// Apply structured filters for specific fields
query = filter.ApplyStructuredFilters(query, cfg, group, fieldBuilders, predicates)

// Then apply search for keyword matching
query = filter.ApplySearch(query, cfg, searchFields, predicates, searchTerm)

// Both work together with AND logic
```

## Complete Example

Here's a real-world handler using Dewey's full toolkit:

```go
package handlers

import (
    "context"
    "github.com/tone-labs/dewey/pagination"
    "github.com/tone-labs/dewey/sort"
    "github.com/tone-labs/dewey/filter"
)

// Package-level configs (create once, reuse everywhere)
var (
    userPaginationCfg = pagination.Config[*ent.UserQuery]{
        Limit:  func(q *ent.UserQuery, n int) *ent.UserQuery { return q.Limit(n) },
        Offset: func(q *ent.UserQuery, n int) *ent.UserQuery { return q.Offset(n) },
    }

    userSortCfg = sort.Config[*ent.UserQuery]{
        Order: func(q *ent.UserQuery, opts ...any) *ent.UserQuery {
            entOpts := make([]func(*sql.Selector), len(opts))
            for i, opt := range opts {
                entOpts[i] = opt.(func(*sql.Selector))
            }
            return q.Order(entOpts...)
        },
    }

    userFilterCfg = filter.Config[*ent.UserQuery, predicate.User]{
        Where: func(q *ent.UserQuery, p predicate.User) *ent.UserQuery {
            return q.Where(p)
        },
    }

    userSortFields = sort.Fields{
        "email":      user.FieldEmail,
        "created_at": user.FieldCreatedAt,
    }

    userSearchFields = filter.SearchFields[predicate.User]{
        user.FieldEmail:     user.EmailContainsFold,
        user.FieldFirstName: user.FirstNameContainsFold,
    }

    userPredicates = filter.PredicateBuilder[predicate.User]{
        IDIn: func(ids ...any) predicate.User {
            uuids := make([]uuid.UUID, len(ids))
            for i, id := range ids {
                uuids[i] = id.(uuid.UUID)
            }
            return user.IDIn(uuids...)
        },
        Or:  user.Or,
        And: user.And,
    }
)

type ListUsersInput struct {
    Search  string   `query:"search"`
    SortBy  string   `query:"sort_by"`
    SortDir string   `query:"sort_dir"`
    Limit   int      `query:"limit" default:"25"`
    Offset  int      `query:"offset" default:"0"`
    IDs     []string `query:"id"` // For getMany support
}

func ListUsers(ctx context.Context, client *ent.Client, input *ListUsersInput) ([]User, int, error) {
    query := client.User.Query()

    // Apply ID filter (if present)
    if len(input.IDs) > 0 {
        ids := make([]any, len(input.IDs))
        for i, idStr := range input.IDs {
            id, _ := uuid.Parse(idStr)
            ids[i] = id
        }
        query = filter.ApplyIDs(query, userFilterCfg, userPredicates, ids)
    }

    // Apply search
    query = filter.ApplySearch(query, userFilterCfg, userSearchFields, userPredicates, input.Search)

    // Apply sorting
    query = sort.Apply(query, userSortCfg, userSortFields, EntOrderBuilder{}, input.SortBy, input.SortDir)

    // Get total (before pagination)
    total, err := query.Count(ctx)
    if err != nil {
        return nil, 0, err
    }

    // Apply pagination
    query = pagination.Apply(query, userPaginationCfg, input.Limit, input.Offset)

    // Execute
    users, err := query.All(ctx)
    if err != nil {
        return nil, 0, err
    }

    return users, total, nil
}
```

## Adapters for Other ORMs

### GORM

```go
// Pagination
cfg := pagination.Config[*gorm.DB]{
    Limit: func(db *gorm.DB, n int) *gorm.DB {
        return db.Limit(n)
    },
    Offset: func(db *gorm.DB, n int) *gorm.DB {
        return db.Offset(n)
    },
}

// Sorting
type GormOrderBuilder struct{}

func (GormOrderBuilder) Asc(field string) any {
    return field + " ASC"
}

func (GormOrderBuilder) Desc(field string) any {
    return field + " DESC"
}

sortCfg := sort.Config[*gorm.DB]{
    Order: func(db *gorm.DB, opts ...any) *gorm.DB {
        for _, opt := range opts {
            db = db.Order(opt)
        }
        return db
    },
}
```

### sqlc / sqlx / Raw SQL

For raw SQL, you'd typically build queries manually. The utilities are less applicable, but you can still use the `Page[T]` helper for response formatting.

## Design Decisions

### Why "Dewey"?

The Dewey Decimal Classification system revolutionized how libraries organize information. Similarly, Dewey the library helps you organize and access your API data efficiently. Plus, it's short, memorable, and fun! 📚

### Why no reflection?

Reflection makes code hard to debug and understand. Explicit adapter functions are:
- Easier to debug (just set a breakpoint)
- Easier to understand (see exactly what's happening)
- Type-safe (compiler catches errors)
- More idiomatic Go

### Why `any` for order options?

Different ORMs use different types for ordering:
- Ent: `func(*sql.Selector)`
- GORM: `string`
- sqlc: Manual SQL construction

Using `any` allows the OrderBuilder to return whatever type your ORM expects, and your `Order` function casts it appropriately.

### Why separate configs for each utility?

**Flexibility**. Some queries only need pagination, others need all three utilities. Separating them allows you to:
- Use only what you need
- Test utilities independently
- Compose them in any order

## Contributing

Dewey is intentionally minimal. Before adding features, ask:
1. Does this work with *all* ORMs, or just one?
2. Can users implement this themselves in 5 lines?
3. Does this violate the "zero dependencies" principle?

If yes to any, it probably doesn't belong in Dewey's catalog.

## License

MIT

## Acknowledgments

Built for the [Strawberry](https://github.com/tone-labs/strawberry) project - a modern Go + React admin framework. Dewey handles the data organization while Strawberry handles the UI.

---

**"The librarian is the heart of the school."** - And Dewey is the heart of your REST API. 📚✨
