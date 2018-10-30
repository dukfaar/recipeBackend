package main

import (
	"encoding/json"
	"fmt"

	dukgraphql "github.com/dukfaar/goUtils/graphql"
	"github.com/dukfaar/recipeBackend/recipe"
)

func CreateRCEventImporter(recipeService recipe.Service, fetcher dukgraphql.Fetcher) func(msg []byte) error {
	return func(msg []byte) error {
		var recipeData map[string]interface{}
		err := json.Unmarshal(msg, &recipeData)

		if err != nil {
			fmt.Printf("Error(%v) unmarshaling event data: %v\n", err, string(msg))
			return err
		}

		fmt.Printf("%+v\n", recipeData)

		//check if recipe exists or not
		//if it does: update
		//else create

		return nil
	}
}
