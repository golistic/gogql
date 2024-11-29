// Copyright (c) 2023, Geert JM Vanderkelen

package gogql

import (
	"context"
	"fmt"
	"slices"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// GQLGenInputObjectFields retrieves the fields of the Input Object types used as arguments
// when executing mutations.
//
// Deprecated: use ProvidedFieldsForCurrentMutation instead.
func GQLGenInputObjectFields(ctx context.Context, arguments ...string) (map[string][]string, error) {
	return ProvidedFieldsForCurrentMutation(ctx, arguments...)
}

// ProvidedFieldsForCurrentMutation retrieves the fields of the Input Object types used as arguments
// within a gqlgen generated resolver.
//
// For example, `mutation { updateUser(input: {id: 123, name: "Marta"}) { id }` would return
// the fields "id" and "name" returning a map with key 'input', and slice of strings.
// If no arguments are specified, 'input' will be returned (if available).
// This also works when variables are used.
//
// This function is intended to be used in gqlgen generated resolvers handling mutations.
func ProvidedFieldsForCurrentMutation(ctx context.Context, filterArguments ...string) (map[string][]string, error) {

	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return nil, fmt.Errorf("no operation in context")
	}

	if opCtx.Operation.Operation != ast.Mutation {
		return nil, fmt.Errorf("not a mutation")
	}

	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil {
		return nil, fmt.Errorf("no field in context")
	}

	providedFields := make(map[string][]string)
	currentFieldName := fieldCtx.Field.Name

	for _, selection := range opCtx.Operation.SelectionSet {

		field, ok := selection.(*ast.Field)
		if !ok || field.Name != currentFieldName {
			continue
		}

		for _, arg := range field.Arguments {
			if len(filterArguments) > 0 && !slices.Contains(filterArguments, arg.Name) {
				continue
			}

			var fieldNames []string

			switch arg.Value.Kind {
			case ast.Variable:
				variableName := arg.Value.Raw
				if variable, ok := opCtx.Variables[variableName]; ok {
					fieldNames = extractFieldNamesFromVariable(variable)
				} else {
					return nil, fmt.Errorf("variable %s not found in operation variables", variableName)
				}
			case ast.ObjectValue:
				for _, child := range arg.Value.Children {
					fieldNames = append(fieldNames, child.Name)
				}
			default:
				return nil, fmt.Errorf("unsupported argument value kind: %s", arg.Value.String())
			}

			providedFields[arg.Name] = fieldNames
		}
	}

	return providedFields, nil
}

// extractFieldNamesFromVariable extracts the field names from a variable object.
func extractFieldNamesFromVariable(variable any) []string {

	var fieldNames []string

	if obj, ok := variable.(map[string]any); ok {
		for key := range obj {
			fieldNames = append(fieldNames, key)
		}
	}

	return fieldNames
}
