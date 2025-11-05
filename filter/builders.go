package filter

import (
	"fmt"
	"time"
)

// StringPredicates contains all the predicate functions needed for string field filtering.
// These are typically provided by your ORM (e.g., Ent's user.EmailEQ, user.EmailContainsFold, etc.)
type StringPredicates[P any] struct {
	// Comparison operators
	Eq  func(string) P
	Ne  func(string) P
	Gt  func(string) P
	Gte func(string) P
	Lt  func(string) P
	Lte func(string) P

	// Array operators
	In  func(...string) P
	Nin func(...string) P

	// String-specific operators
	Contains   func(string) P
	StartsWith func(string) P
	EndsWith   func(string) P

	// Null operators (optional - only needed for nullable fields)
	IsNil    func() P
	IsNotNil func() P
}

// StringFilterBuilder implements FieldFilterBuilder for string fields.
// It handles type conversions and delegates to ORM-specific predicate functions.
type StringFilterBuilder[P any] struct {
	predicates StringPredicates[P]
	nullable   bool
}

// NewStringFilter creates a filter builder for a non-nullable string field.
// The alwaysTrue and alwaysFalse predicates are derived automatically:
//   - alwaysTrue: Ne("") - a non-empty required field is never equal to empty string
//   - alwaysFalse: Eq("") - a non-empty required field is always not equal to empty string
//
// Example for Ent:
//
//	builder := NewStringFilter(
//	    StringPredicates[predicate.User]{
//	        Eq:         user.EmailEQ,
//	        Ne:         user.EmailNEQ,
//	        Gt:         user.EmailGT,
//	        Gte:        user.EmailGTE,
//	        Lt:         user.EmailLT,
//	        Lte:        user.EmailLTE,
//	        In:         user.EmailIn,
//	        Nin:        user.EmailNotIn,
//	        Contains:   user.EmailContainsFold,
//	        StartsWith: user.EmailHasPrefix,
//	        EndsWith:   user.EmailHasSuffix,
//	    },
//	)
func NewStringFilter[P any](predicates StringPredicates[P]) FieldFilterBuilder[P] {
	return &StringFilterBuilder[P]{
		predicates: predicates,
		nullable:   false,
	}
}

// NewNullableStringFilter creates a filter builder for a nullable string field.
//
// Example for Ent:
//
//	builder := NewNullableStringFilter(
//	    StringPredicates[predicate.User]{
//	        Eq:         user.FirstNameEQ,
//	        Ne:         user.FirstNameNEQ,
//	        // ... other predicates
//	        IsNil:      user.FirstNameIsNil,
//	        IsNotNil:   user.FirstNameNotNil,
//	    },
//	)
func NewNullableStringFilter[P any](predicates StringPredicates[P]) FieldFilterBuilder[P] {
	return &StringFilterBuilder[P]{
		predicates: predicates,
		nullable:   true,
	}
}

func (b *StringFilterBuilder[P]) Eq(value any) P {
	return b.predicates.Eq(value.(string))
}

func (b *StringFilterBuilder[P]) Ne(value any) P {
	return b.predicates.Ne(value.(string))
}

func (b *StringFilterBuilder[P]) Gt(value any) P {
	return b.predicates.Gt(value.(string))
}

func (b *StringFilterBuilder[P]) Gte(value any) P {
	return b.predicates.Gte(value.(string))
}

func (b *StringFilterBuilder[P]) Lt(value any) P {
	return b.predicates.Lt(value.(string))
}

func (b *StringFilterBuilder[P]) Lte(value any) P {
	return b.predicates.Lte(value.(string))
}

func (b *StringFilterBuilder[P]) In(values []any) P {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = v.(string)
	}
	return b.predicates.In(strs...)
}

func (b *StringFilterBuilder[P]) Nin(values []any) P {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = v.(string)
	}
	return b.predicates.Nin(strs...)
}

func (b *StringFilterBuilder[P]) Contains(value string) P {
	return b.predicates.Contains(value)
}

func (b *StringFilterBuilder[P]) StartsWith(value string) P {
	return b.predicates.StartsWith(value)
}

func (b *StringFilterBuilder[P]) EndsWith(value string) P {
	return b.predicates.EndsWith(value)
}

func (b *StringFilterBuilder[P]) IsNull() P {
	if b.nullable {
		return b.predicates.IsNil()
	}
	// Non-nullable field - return predicate that never matches
	// For strings: Eq("") means "field equals empty string", which is always false for required fields
	return b.predicates.Eq("")
}

func (b *StringFilterBuilder[P]) IsNotNull() P {
	if b.nullable {
		return b.predicates.IsNotNil()
	}
	// Non-nullable field - return predicate that always matches
	// For strings: Ne("") means "field not equal empty string", which is always true for required fields
	return b.predicates.Ne("")
}

// BoolPredicates contains all the predicate functions needed for boolean field filtering.
type BoolPredicates[P any] struct {
	Eq  func(bool) P
	Ne  func(bool) P
	Or  func(...P) P // Used for In/Nin operations
	And func(...P) P // Used for Nin operations
}

// BoolFilterBuilder implements FieldFilterBuilder for boolean fields.
// Most comparison and string operations are not meaningful for booleans,
// so they default to simple equality checks or no-ops.
type BoolFilterBuilder[P any] struct {
	predicates BoolPredicates[P]
}

// NewBoolFilter creates a filter builder for a boolean field.
// Boolean fields in databases are typically non-nullable with defaults.
// The alwaysTrue and alwaysFalse predicates are derived automatically:
//   - alwaysTrue: Or(Eq(true), Eq(false)) - matches both possible values
//   - alwaysFalse: And(Eq(true), Eq(false)) - impossible condition
//
// Example for Ent:
//
//	builder := NewBoolFilter(
//	    BoolPredicates[predicate.User]{
//	        Eq:  user.IsActiveEQ,
//	        Ne:  user.IsActiveNEQ,
//	        Or:  user.Or,
//	        And: user.And,
//	    },
//	)
func NewBoolFilter[P any](predicates BoolPredicates[P]) FieldFilterBuilder[P] {
	return &BoolFilterBuilder[P]{
		predicates: predicates,
	}
}

func (b *BoolFilterBuilder[P]) Eq(value any) P {
	return b.predicates.Eq(value.(bool))
}

func (b *BoolFilterBuilder[P]) Ne(value any) P {
	return b.predicates.Ne(value.(bool))
}

// Comparison operators aren't meaningful for booleans, so we treat them as equality
func (b *BoolFilterBuilder[P]) Gt(value any) P {
	return b.predicates.Eq(value.(bool))
}

func (b *BoolFilterBuilder[P]) Gte(value any) P {
	return b.predicates.Eq(value.(bool))
}

func (b *BoolFilterBuilder[P]) Lt(value any) P {
	return b.predicates.Eq(value.(bool))
}

func (b *BoolFilterBuilder[P]) Lte(value any) P {
	return b.predicates.Eq(value.(bool))
}

func (b *BoolFilterBuilder[P]) In(values []any) P {
	if len(values) == 0 {
		// No values to match - always false
		return b.predicates.And(b.predicates.Eq(true), b.predicates.Eq(false))
	}

	hasTrue, hasFalse := false, false
	for _, v := range values {
		if v.(bool) {
			hasTrue = true
		} else {
			hasFalse = true
		}
	}

	if hasTrue && hasFalse {
		// Both values included - always matches
		return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
	} else if hasTrue {
		return b.predicates.Eq(true)
	} else {
		return b.predicates.Eq(false)
	}
}

func (b *BoolFilterBuilder[P]) Nin(values []any) P {
	if len(values) == 0 {
		// No values to exclude - always true
		return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
	}

	hasTrue, hasFalse := false, false
	for _, v := range values {
		if v.(bool) {
			hasTrue = true
		} else {
			hasFalse = true
		}
	}

	if hasTrue && hasFalse {
		// Both values excluded - never matches
		return b.predicates.And(b.predicates.Eq(true), b.predicates.Eq(false))
	} else if hasTrue {
		return b.predicates.Eq(false)
	} else {
		return b.predicates.Eq(true)
	}
}

// String operations aren't applicable to booleans - return always-true predicate
func (b *BoolFilterBuilder[P]) Contains(value string) P {
	return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
}

func (b *BoolFilterBuilder[P]) StartsWith(value string) P {
	return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
}

func (b *BoolFilterBuilder[P]) EndsWith(value string) P {
	return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
}

func (b *BoolFilterBuilder[P]) IsNull() P {
	// Boolean fields are typically non-nullable - return always-false
	return b.predicates.And(b.predicates.Eq(true), b.predicates.Eq(false))
}

func (b *BoolFilterBuilder[P]) IsNotNull() P {
	// Boolean fields are typically non-nullable - return always-true
	return b.predicates.Or(b.predicates.Eq(true), b.predicates.Eq(false))
}

// TimePredicates contains all the predicate functions needed for time/timestamp field filtering.
type TimePredicates[P any] struct {
	Eq  func(time.Time) P
	Ne  func(time.Time) P
	Gt  func(time.Time) P
	Gte func(time.Time) P
	Lt  func(time.Time) P
	Lte func(time.Time) P
	In  func(...time.Time) P
	Nin func(...time.Time) P

	// For nullable time fields
	IsNil    func() P
	IsNotNil func() P
}

// TimeFilterBuilder implements FieldFilterBuilder for time/timestamp fields.
// It handles parsing time values from strings and time.Time objects.
type TimeFilterBuilder[P any] struct {
	predicates TimePredicates[P]
	nullable   bool
}

// NewTimeFilter creates a filter builder for a non-nullable time/timestamp field.
// The alwaysTrue and alwaysFalse predicates are derived automatically:
//   - alwaysTrue: Gte(zeroTime) - all times are >= the zero time
//   - alwaysFalse: Lt(zeroTime) - no times are < the zero time
//
// Example for Ent:
//
//	builder := NewTimeFilter(
//	    TimePredicates[predicate.User]{
//	        Eq:  user.CreatedAtEQ,
//	        Ne:  user.CreatedAtNEQ,
//	        Gt:  user.CreatedAtGT,
//	        Gte: user.CreatedAtGTE,
//	        Lt:  user.CreatedAtLT,
//	        Lte: user.CreatedAtLTE,
//	        In:  user.CreatedAtIn,
//	        Nin: user.CreatedAtNotIn,
//	    },
//	)
func NewTimeFilter[P any](predicates TimePredicates[P]) FieldFilterBuilder[P] {
	return &TimeFilterBuilder[P]{
		predicates: predicates,
		nullable:   false,
	}
}

// NewNullableTimeFilter creates a filter builder for a nullable time field.
func NewNullableTimeFilter[P any](predicates TimePredicates[P]) FieldFilterBuilder[P] {
	return &TimeFilterBuilder[P]{
		predicates: predicates,
		nullable:   true,
	}
}

// parseTimeValue parses a time value from various formats (string, time.Time).
// Supports both ISO date strings (YYYY-MM-DD) and RFC3339 timestamps.
func parseTimeValue(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try RFC3339 first (full timestamp)
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		// Try ISO date format (YYYY-MM-DD)
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("unable to parse time value: %s", v)
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type: %T", value)
	}
}

func (b *TimeFilterBuilder[P]) Eq(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Eq(t)
}

func (b *TimeFilterBuilder[P]) Ne(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Ne(t)
}

func (b *TimeFilterBuilder[P]) Gt(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Gt(t)
}

func (b *TimeFilterBuilder[P]) Gte(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Gte(t)
}

func (b *TimeFilterBuilder[P]) Lt(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Lt(t)
}

func (b *TimeFilterBuilder[P]) Lte(value any) P {
	t, _ := parseTimeValue(value)
	return b.predicates.Lte(t)
}

func (b *TimeFilterBuilder[P]) In(values []any) P {
	times := make([]time.Time, len(values))
	for i, v := range values {
		t, _ := parseTimeValue(v)
		times[i] = t
	}
	return b.predicates.In(times...)
}

func (b *TimeFilterBuilder[P]) Nin(values []any) P {
	times := make([]time.Time, len(values))
	for i, v := range values {
		t, _ := parseTimeValue(v)
		times[i] = t
	}
	return b.predicates.Nin(times...)
}

// String operations aren't applicable to time fields - return always-true predicate
func (b *TimeFilterBuilder[P]) Contains(value string) P {
	return b.predicates.Gte(time.Time{})
}

func (b *TimeFilterBuilder[P]) StartsWith(value string) P {
	return b.predicates.Gte(time.Time{})
}

func (b *TimeFilterBuilder[P]) EndsWith(value string) P {
	return b.predicates.Gte(time.Time{})
}

func (b *TimeFilterBuilder[P]) IsNull() P {
	if b.nullable {
		return b.predicates.IsNil()
	}
	// Non-nullable field - return always-false
	// All times are >= zero time, so no times are < zero time
	return b.predicates.Lt(time.Time{})
}

func (b *TimeFilterBuilder[P]) IsNotNull() P {
	if b.nullable {
		return b.predicates.IsNotNil()
	}
	// Non-nullable field - return always-true
	// All times are >= zero time
	return b.predicates.Gte(time.Time{})
}
