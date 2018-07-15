package recipe

import (
	"github.com/dukfaar/goUtils/relay"
	graphql "github.com/graph-gophers/graphql-go"
	"gopkg.in/mgo.v2/bson"
)

type InOutElement struct {
	ItemID bson.ObjectId `json:"itemID,omitempty" bson:"_id,omitempty"`
	Amount int32         `json:"amount,omitempty"`
}

type InputElement struct {
	InOutElement
}

type OutputElement struct {
	InOutElement
}

type Model struct {
	ID      bson.ObjectId   `json:"_id,omitempty" bson:"_id,omitempty"`
	Inputs  []InputElement  `json:"inputs,omitempty"`
	Outputs []OutputElement `json:"outputs,omitempty"`
}

type MutationInOutElement struct {
	ItemID graphql.ID
	Amount int32
}

type MutationInput struct {
	Inputs  *[]*MutationInOutElement
	Outputs *[]*MutationInOutElement
}

var GraphQLType = `
type Recipe {
	_id: ID
	inputs: [RecipeInput]
	outputs: [RecipeOutput]
}

type RecipeInput {
	itemId: ID
	amount: Int
}

type RecipeOutput {
	itemId: ID
	amount: Int
}
` +
	relay.GenerateConnectionTypes("Recipe")
