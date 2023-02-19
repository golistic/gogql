// Copyright (c) 2023, Geert JM Vanderkelen

package gogql

import (
	"context"
	"fmt"
	"sort"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// GQLGenInputObjectFields retrieves the fields of the Input Object types used as arguments
// when executing mutations.
// For example, `mutation { updateUser(input: {id: 123, name: "Marta"}) { id }` would return
// the fields "id" and "name" returning a map with key 'input', and slice of strings.
// If no arguments are specified, 'input' will be returned (if available).
// This also works when variables are used.
// Nil is returned when no operation is in context, or when the operation is not
// a mutation, or no arguments were available.
func GQLGenInputObjectFields(ctx context.Context, arguments ...string) (map[string][]string, error) {
	baseErr := "getting input type fields (%w)"
	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil || fieldCtx.Field.Arguments == nil {
		return nil, fmt.Errorf(baseErr, fmt.Errorf("field context missing"))
	}

	res := map[string][]string{}
	fieldArgs := graphql.GetFieldContext(ctx).Field.Arguments

	for _, argument := range arguments {
		inputArgValue := fieldArgs.ForName(argument).Value
		var fields []string
		switch inputArgValue.Kind {
		case ast.Variable:
			operationCtx := graphql.GetOperationContext(ctx)
			if operationCtx == nil {
				return nil, fmt.Errorf(baseErr, fmt.Errorf("operation context missing"))
			}
			varName := inputArgValue.String()[1:] // always prefixed with $
			for k := range graphql.GetOperationContext(ctx).Variables[varName].(map[string]any) {
				fields = append(fields, k)
			}
		case ast.ObjectValue:
			for _, c := range inputArgValue.Children {
				fields = append(fields, c.Name)
			}
		}

		sort.Strings(fields)
		res[argument] = fields
	}

	if len(res) == 0 {
		return nil, nil
	}

	return res, nil
}
