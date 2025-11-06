package filter

// FieldBuilder represents a configured filter builder for a single field.
// It contains a factory function that constructs the actual builder with combinators injected.
type FieldBuilder[P any] struct {
	Name   string
	Create func(combinators Combinators[P]) FieldFilterBuilder[P]
}

// BuildFilterMap constructs a map of field names to filter builders from a list of field configurations.
// This provides a declarative way to register all filters for a model.
//
// The Combinators provide the Or/And functions needed to construct mathematically accurate
// "always true" and "always false" predicates for IsNull/IsNotNull on non-nullable fields.
//
// Example usage:
//
//	combinators := Combinators[predicate.User]{
//	    Or:  user.Or,
//	    And: user.And,
//	}
//
//	filterBuilders := BuildFilterMap(
//	    combinators,
//	    StringField("email", StringPredicates{...}),
//	    NullableStringField("first_name", StringPredicates{...}),
//	    BoolField("is_active", BoolPredicates{...}),
//	    TimeField("created_at", TimePredicates{...}),
//	)
func BuildFilterMap[P any](
	combinators Combinators[P],
	fields ...FieldBuilder[P],
) map[string]FieldFilterBuilder[P] {
	result := make(map[string]FieldFilterBuilder[P], len(fields))
	for _, field := range fields {
		result[field.Name] = field.Create(combinators)
	}
	return result
}

// StringField creates a FieldBuilder for a non-nullable string field.
// The builder will be constructed with combinators injected by BuildFilterMap.
//
// Example:
//
//	StringField("email", StringPredicates[predicate.User]{
//	    Eq:         user.EmailEQ,
//	    Ne:         user.EmailNEQ,
//	    Gt:         user.EmailGT,
//	    Gte:        user.EmailGTE,
//	    Lt:         user.EmailLT,
//	    Lte:        user.EmailLTE,
//	    In:         user.EmailIn,
//	    Nin:        user.EmailNotIn,
//	    Contains:   user.EmailContainsFold,
//	    StartsWith: user.EmailHasPrefix,
//	    EndsWith:   user.EmailHasSuffix,
//	})
func StringField[P any](
	name string,
	predicates StringPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name: name,
		Create: func(combinators Combinators[P]) FieldFilterBuilder[P] {
			return newStringFilterWithCombinators(predicates, combinators, false)
		},
	}
}

// NullableStringField creates a FieldBuilder for a nullable string field.
//
// Example:
//
//	NullableStringField("first_name", StringPredicates[predicate.User]{
//	    Eq:         user.FirstNameEQ,
//	    Ne:         user.FirstNameNEQ,
//	    // ...
//	    IsNil:      user.FirstNameIsNil,
//	    IsNotNil:   user.FirstNameNotNil,
//	})
func NullableStringField[P any](
	name string,
	predicates StringPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name: name,
		Create: func(combinators Combinators[P]) FieldFilterBuilder[P] {
			return newStringFilterWithCombinators(predicates, combinators, true)
		},
	}
}

// BoolField creates a FieldBuilder for a boolean field.
// The builder will be constructed with combinators injected by BuildFilterMap.
//
// Example:
//
//	BoolField("is_active", BoolPredicates[predicate.User]{
//	    Eq:  user.IsActiveEQ,
//	    Ne:  user.IsActiveNEQ,
//	})
func BoolField[P any](
	name string,
	predicates BoolPredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name: name,
		Create: func(combinators Combinators[P]) FieldFilterBuilder[P] {
			return newBoolFilterWithCombinators(predicates, combinators)
		},
	}
}

// TimeField creates a FieldBuilder for a non-nullable time/timestamp field.
//
// Example:
//
//	TimeField("created_at", TimePredicates[predicate.User]{
//	    Eq:  user.CreatedAtEQ,
//	    Ne:  user.CreatedAtNEQ,
//	    Gt:  user.CreatedAtGT,
//	    Gte: user.CreatedAtGTE,
//	    Lt:  user.CreatedAtLT,
//	    Lte: user.CreatedAtLTE,
//	    In:  user.CreatedAtIn,
//	    Nin: user.CreatedAtNotIn,
//	})
func TimeField[P any](
	name string,
	predicates TimePredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name: name,
		Create: func(combinators Combinators[P]) FieldFilterBuilder[P] {
			return newTimeFilterWithCombinators(predicates, combinators, false)
		},
	}
}

// NullableTimeField creates a FieldBuilder for a nullable time field.
func NullableTimeField[P any](
	name string,
	predicates TimePredicates[P],
) FieldBuilder[P] {
	return FieldBuilder[P]{
		Name: name,
		Create: func(combinators Combinators[P]) FieldFilterBuilder[P] {
			return newTimeFilterWithCombinators(predicates, combinators, true)
		},
	}
}
