package main

import (
	"github.com/dukfaar/goUtils/relay"
	"github.com/dukfaar/recipeBackend/recipe"
)

var Schema string = `
		schema {
			query: Query
			mutation: Mutation
		}

		type Query {
			recipes(first: Int, last: Int, before: String, after: String): RecipeConnection!
			recipe(id: ID!): Recipe!
		}

		input RecipeMutationInOutInput {
			itemId: ID!
			amount: Int!
		}

		input RecipeMutationInput {
			inputs: [RecipeMutationInOutInput] 
			outputs: [RecipeMutationInOutInput]
		}

		type Mutation {
			createRecipe(input: RecipeMutationInput): Recipe!
			updateRecipe(id: ID!, input: RecipeMutationInput): Recipe!
			deleteRecipe(id: ID!): ID
		}` +
	relay.PageInfoGraphQLString +
	recipe.GraphQLType
