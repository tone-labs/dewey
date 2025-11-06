# Generic Filter Builders

This package provides generic, reusable filter builders that eliminate the need for repetitive per-field filter implementations.

## Problem

Previously, implementing filters for a model with N fields and K filter operations required N × K method implementations. For example, a User model with 7 fields required 91 nearly-identical method implementations (~560 lines of code).

## Solution

Generic filter builders parameterized by database column type. Each column type (string, bool, time, etc.) has standard filter behavior implemented once and reused across all fields of that type.

## Complexity Reduction

- **Before**: N models × M fields × K operations = O(N × M × K)
- **After**: N models × M fields = O(N × M)

For the User model: **563 lines → 116 lines** (79% reduction)

## Usage

### 1. Import the Package

```go
import "github.com/tone-labs/dewey/filter"
```

### 2. Define Filter Builders for Your Model

Use the declarative API to create a filter builder map. You need to provide the Combinators which contain the Or/And functions used for creating accurate impossible predicates:

```go
// Define the Combinators
combinators := filter.Combinators[predicate.User]{
    Or:  user.Or,
    And: user.And,
}

var userFilterBuilders = filter.BuildFilterMap(
    combinators, // Pass combinators as first argument

    // Non-nullable string field
    filter.StringField("email", filter.StringPredicates[predicate.User]{
        Eq:         user.EmailEQ,
        Ne:         user.EmailNEQ,
        Gt:         user.EmailGT,
        Gte:        user.EmailGTE,
        Lt:         user.EmailLT,
        Lte:        user.EmailLTE,
        In:         user.EmailIn,
        Nin:        user.EmailNotIn,
        Contains:   user.EmailContainsFold,
        StartsWith: user.EmailHasPrefix,
        EndsWith:   user.EmailHasSuffix,
    }),

    // Nullable string field
    filter.NullableStringField("first_name", filter.StringPredicates[predicate.User]{
        Eq:         user.FirstNameEQ,
        Ne:         user.FirstNameNEQ,
        Gt:         user.FirstNameGT,
        Gte:        user.FirstNameGTE,
        Lt:         user.FirstNameLT,
        Lte:        user.FirstNameLTE,
        In:         user.FirstNameIn,
        Nin:        user.FirstNameNotIn,
        Contains:   user.FirstNameContainsFold,
        StartsWith: user.FirstNameHasPrefix,
        EndsWith:   user.FirstNameHasSuffix,
        IsNil:      user.FirstNameIsNil,
        IsNotNil:   user.FirstNameNotNil,
    }),

    // Boolean field
    filter.BoolField("is_active", filter.BoolPredicates[predicate.User]{
        Eq: user.IsActiveEQ,
        Ne: user.IsActiveNEQ,
    }),

    // Time/timestamp field
    filter.TimeField("created_at", filter.TimePredicates[predicate.User]{
        Eq:  user.CreatedAtEQ,
        Ne:  user.CreatedAtNEQ,
        Gt:  user.CreatedAtGT,
        Gte: user.CreatedAtGTE,
        Lt:  user.CreatedAtLT,
        Lte: user.CreatedAtLTE,
        In:  user.CreatedAtIn,
        Nin: user.CreatedAtNotIn,
    }),
)
```

### 3. Use with Dewey's Filter System

The filter builders integrate seamlessly with Dewey's structured filter API:

```go
query = filter.ApplyStructuredFilters(
    query,
    cfg,
    filterGroup,
    userFilterBuilders,  // Your declaratively-defined builder map
    predicates,
)
```

## Available Field Types

### String Fields

- **`StringField(name, predicates)`** - Non-nullable string
- **`NullableStringField(name, predicates)`** - Nullable string

Supports: Eq, Ne, Gt, Gte, Lt, Lte, In, Nin, Contains, StartsWith, EndsWith, IsNull, IsNotNull

For non-nullable fields, `IsNull` returns a mathematically impossible predicate (`And(Eq(value), Ne(value))` - always false), and `IsNotNull` returns a tautology (`Or(Eq(value), Ne(value))` - always true). This provides accurate semantics regardless of field content.

### Boolean Fields

- **`BoolField(name, predicates)`**

Supports: Eq, Ne, In, Nin (comparison and string operators are no-ops)

The library automatically derives "always true" and "always false" predicates using `Or(Eq(true), Eq(false))` (tautology - matches any boolean value) and `And(Eq(true), Eq(false))` (impossible condition - can't be both).

### Time/Timestamp Fields

- **`TimeField(name, predicates)`** - Non-nullable time
- **`NullableTimeField(name, predicates)`** - Nullable time

Supports: Eq, Ne, Gt, Gte, Lt, Lte, In, Nin, IsNull, IsNotNull

For non-nullable fields, `IsNull` returns a mathematically impossible predicate (`And(Eq(zeroTime), Ne(zeroTime))` - always false), and `IsNotNull` returns a tautology (`Or(Eq(zeroTime), Ne(zeroTime))` - always true).

Handles automatic parsing of ISO dates (YYYY-MM-DD) and RFC3339 timestamps.

## Benefits

1. **Massive code reduction** - ~80% less boilerplate per model
2. **Type safety** - Compile-time guarantees via generics
3. **Consistency** - All fields of the same type behave identically
4. **Maintainability** - Bug fixes in one place benefit all fields
5. **Extensibility** - New field types = one generic builder implementation
6. **ORM agnostic** - Works with any ORM (Ent, GORM, sqlc, etc.)

## Architecture

The library uses Go generics to provide type-safe, reusable filter builders:

```
FieldFilterBuilder[P]  (interface)
    ↑
    ├── StringFilterBuilder[P]    (handles all string fields)
    ├── BoolFilterBuilder[P]      (handles all boolean fields)
    └── TimeFilterBuilder[P]      (handles all time fields)
```

Each generic builder:
1. Takes predicate functions from your ORM
2. Handles type conversions (e.g., `any` → `string`, `time.Time`)
3. Implements all 13 filter operations
4. Returns ORM-specific predicates

## Future Extensions

Potential additional field types:

- `IntField` / `FloatField` - Numeric fields
- `UUIDField` - UUID fields
- `EnumField` - Enumerated types
- `JSONField` - JSON/JSONB fields

Each new type would be implemented once and reused across all models.
