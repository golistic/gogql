// Copyright (c) 2023, Geert JM Vanderkelen

package gogql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golistic/xgo/xt"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/golistic/gogql"
)

// type TestUserModel struct {
//	ID    int    `json:"id"`
//	Name  string `json:"name"`
//	Email string `json:"email"`
// }
//
// type UpdateTestUserModelInput struct {
//	ID    int    `json:"id"`
//	Name  string `json:"name"`
//	Email string `json:"email"`
// }

var schema = `
schema {
    query: Query
    mutation: Mutation
}

type Query {
	user(id: ID!): User
}

type Mutation {
	updateUser(input: UpdateUserInput!, settings: SettingsInput): User
}

type User {
	id: ID!
	name: String!
	email: String!
}

input UpdateUserInput {
	id: ID!
	name: String
	givenName: String
	email: String
}

input SettingsInput {
	notify: Boolean
	newsletter: Boolean
}
`

func TestGQLGenInputObjectFields(t *testing.T) {
	srcSchema := &ast.Source{
		Name:    "TestSchema",
		Input:   schema,
		BuiltIn: true,
	}

	schemaAst, err := gqlparser.LoadSchema(srcSchema)
	xt.OK(t, err)

	var cases = map[string]struct {
		query     string
		want      map[string][]string
		wantErr   error
		args      []string
		variables map[string]any
	}{
		"with operation name": {
			query: `mutation UpdateUser { updateUser(input: {id: 123, name: "Marta"}) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}},
			args:  []string{"input"},
		},
		"multiple arguments": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}, settings: {notify: true}) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}, "settings": {"notify"}},
			args:  []string{"input", "settings"},
		},
		"multiple arguments (get all)": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}, settings: {notify: true}) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}, "settings": {"notify"}},
		},
		"argument missing from field arguments": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}},
			args:  []string{"input", "settings"},
		},
		"using variables for one argument": {
			query: `mutation ($settings: SettingsInput!) { updateUser(input: {id: 123, name: "Marta"}, settings: $settings) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}, "settings": {"newsletter"}},
			args:  []string{"input", "settings"},
			variables: map[string]any{
				"settings": map[string]any{"newsletter": false},
			},
		},
		"input argument is object": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}) { id } }`,
			want:  map[string][]string{"input": {"id", "name"}},
			args:  []string{"input", "foo"},
		},
		"not a mutation": {
			query:   `query { user(id: 123) { id } }`,
			want:    nil,
			wantErr: fmt.Errorf("not a mutation"),
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			doc, gqlErrors := gqlparser.LoadQuery(schemaAst, c.query)
			if gqlErrors != nil {
				t.Fatal("cannot load schema", gqlErrors.Error())
			}

			// recreate the context like used by gqlgen
			operation := doc.Operations[0]
			field := operation.SelectionSet[0].(*ast.Field)

			opCtx := &graphql.OperationContext{
				Variables: c.variables,
				Operation: operation,
			}

			fieldCtx := &graphql.FieldContext{
				Field: graphql.CollectedField{
					Field: field,
				},
			}

			ctx := context.Background()
			ctx = graphql.WithOperationContext(ctx, opCtx)
			ctx = graphql.WithFieldContext(ctx, fieldCtx)

			// test using arguments
			have, err := gogql.ProvidedFieldsForCurrentMutation(ctx, c.args...)

			if c.wantErr != nil {
				xt.ErrorIs(t, c.wantErr, err)
			} else {
				xt.OK(t, err)
				xt.Eq(t, c.want, have)
			}
		})
	}

}
