// Copyright (c) 2023, Geert JM Vanderkelen

package gogql

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golistic/xt"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

//type TestUserModel struct {
//	ID    int    `json:"id"`
//	Name  string `json:"name"`
//	Email string `json:"email"`
//}
//
//type UpdateTestUserModelInput struct {
//	ID    int    `json:"id"`
//	Name  string `json:"name"`
//	Email string `json:"email"`
//}

var schema = `
schema {
    mutation: Mutation
}

type Mutation {
	updateUser(input: UpdateUserInput!): User
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
`

func testContext(t *testing.T, query string, vars Variables, ctx context.Context) context.Context {
	srcMutation := &ast.Source{
		Name:    "test",
		Input:   query,
		BuiltIn: false,
	}
	doc, err := parser.ParseQuery(srcMutation)
	xt.OK(t, err)
	_ = doc

	rqCtx := &graphql.OperationContext{
		RawQuery:             query,
		Variables:            vars,
		Doc:                  doc,
		OperationName:        doc.Operations[0].Name,
		DisableIntrospection: true,
		ResolverMiddleware: func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			return nil, nil
		},
		RootResolverMiddleware: func(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
			return nil
		},
		Operation: doc.Operations[0],
	}
	xt.OK(t, rqCtx.Validate(ctx))
	return graphql.WithOperationContext(ctx, rqCtx)
}

func TestGQLGenInputObjectFields(t *testing.T) {
	srcSchema := &ast.Source{
		Name:    "TestSchema",
		Input:   schema,
		BuiltIn: false,
	}
	schemaDoc, err := parser.ParseSchema(srcSchema)
	xt.OK(t, err)
	_ = schemaDoc

	var cases = map[string]struct {
		query     string
		exp       map[string][]string
		expErr    error
		args      []string
		variables Variables
	}{
		"with operation name": {
			query: `mutation UpdateUser { updateUser(input: {id: 123, name: "Marta"}) { id } }`,
			exp:   map[string][]string{"input": {"id", "name"}},
		},
		"multiple arguments": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}, input2: {foo: "bar"}) { id } }`,
			exp:   map[string][]string{"input": {"id", "name"}, "input2": {"foo"}},
			args:  []string{"input", "input2"},
		},
		"using variables": {
			query: `mutation ($input2: UpdateUserInput!) { updateUser(input: {id: 123, name: "Marta"}, input2: $input2) { id } }`,
			exp:   map[string][]string{"input": {"id", "name"}, "input2": {"foo"}},
			args:  []string{"input", "input2"},
			variables: Variables{
				"input2": Variables{"foo": "bar"},
			},
		},
		"input argument is object": {
			query: `mutation { updateUser(input: {id: 123, name: "Marta"}, foo: 1234) { id } }`,
			exp:   map[string][]string{"input": {"id", "name"}},
			args:  []string{"input", "foo"},
		},
		"no input arguments": {
			query: `mutation { updateUser(foo: 1234) { id } }`,
			exp:   nil,
			args:  []string{"input", "foo"},
		},
		"not a mutation": {
			query:  `query { user(id: 123) { id } }`,
			exp:    nil,
			expErr: fmt.Errorf("field context missing"),
		},
		"multiple mutations": {
			query: `
mutation ($inputA: UpdateUserInput!, $inputB: UpdateUserInput!) {
	updateUser(input: $inputA) { id }
	updateUser(input: $inputB) { id }
}
`,
			exp: map[string][]string{"inputA": {"id", "name"}, "inputB": {"givenName", "id"}},
			variables: Variables{
				"inputA": Variables{"id": "123", "name": "Marta"},
				"inputB": Variables{"id": "987", "givenName": "Tomas"},
			},
		},
		"multiple mutations but no variables": {
			query: `
mutation {
	updateUser(input: {id: "123", name: "Marta"}) { id }
	updateUser(input: {id: "987", givenName: "Tomas"}) { id }
}
`,
			exp: map[string][]string{"inputA": {"id", "name"}, "inputB": {"givenName", "id"}},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx = testContext(t, c.query, c.variables, ctx)
			have, err := GQLGenInputObjectFields(ctx, c.args...)

			if c.expErr != nil {
				xt.Eq(t, c.expErr, errors.Unwrap(err))
			} else {
				xt.OK(t, err)
				xt.Eq(t, c.exp, have)
			}
		})
	}

}
