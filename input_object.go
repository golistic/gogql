// Copyright (c) 2023, Geert JM Vanderkelen

package gogql

import (
	"context"
	"strings"

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
func GQLGenInputObjectFields(ctx context.Context, arguments ...string) map[string][]string {
	if graphql.GetOperationContext(ctx) == nil ||
		graphql.GetOperationContext(ctx).Operation == nil {
		return nil
	}

	if graphql.GetOperationContext(ctx).Operation.Operation != ast.Mutation {
		return nil
	}

	if len(arguments) == 0 {
		arguments = []string{"input"}
	}

	if len(graphql.GetOperationContext(ctx).Operation.SelectionSet) == 0 {
		return nil
	}

	fields := map[string][]string{}

	sel := graphql.GetOperationContext(ctx).Operation.SelectionSet[0]
	for _, argName := range arguments {
		arg := sel.(*ast.Field).Arguments.ForName(argName)
		if arg == nil {
			continue
		}

		if strings.HasPrefix(arg.Value.String(), "$") {
			for k := range graphql.GetOperationContext(ctx).Variables[arg.Value.Raw].(Variables) {
				fields[argName] = append(fields[argName], k)
			}
		} else {
			for _, c := range sel.(*ast.Field).Arguments.ForName(argName).Value.Children {
				fields[argName] = append(fields[argName], c.Name)
			}
		}
	}

	if len(fields) == 0 {
		return nil
	}

	return fields
}
