package recipe

import (
	"github.com/dukfaar/goUtils/relay"
	"github.com/globalsign/mgo/bson"
	graphql "github.com/graph-gophers/graphql-go"
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
	ID                    bson.ObjectId   `json:"_id,omitempty" bson:"_id,omitempty"`
	Inputs                []InputElement  `json:"inputs,omitempty"`
	Outputs               []OutputElement `json:"outputs,omitempty"`
	NamespaceID           *bson.ObjectId  `json:"namespaceId,omitempty" bson:"namespaceId,omitempty"`
	CraftingLevel         *int32          `json:"craftingLevel,omitempty" bson:"craftingLevel,omitempty"`
	CraftingJobID         *bson.ObjectId  `json:"craftingJob,omitempty" bson:"craftingJob,omitempty"`
	Masterbook            *int32          `json:"masterbook,omitempty" bson:"masterbook,omitempty"`
	RequiredControl       *int32          `json:"requiredControl,omitempty" bson:"requiredControl,omitempty"`
	RequiredCraftsmanship *int32          `json:"requiredCraftsmanship,omitempty" bson:"requiredCraftsmanship,omitempty"`
	Stars                 *int32          `json:"stars,omitempty" bson:"stars,omitempty"`
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
	namespaceId: ID
	inputs: [RecipeInput]
	outputs: [RecipeOutput]
	craftingLevel: Int
	craftingJobId: ID
	masterbook: Int
	requiredControl: Int
	requiredCraftsmanship: Int
	stars: Int
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
