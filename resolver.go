package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/globalsign/mgo/bson"

	"github.com/dukfaar/goUtils/eventbus"
	dukgraphql "github.com/dukfaar/goUtils/graphql"
	"github.com/dukfaar/goUtils/permission"
	"github.com/dukfaar/goUtils/relay"
	"github.com/dukfaar/recipeBackend/recipe"
	graphql "github.com/graph-gophers/graphql-go"
)

type Resolver struct {
}

func (r *Resolver) Recipes(ctx context.Context, args struct {
	First  *int32
	Last   *int32
	Before *string
	After  *string
}) (*recipe.ConnectionResolver, error) {
	recipeService := ctx.Value("recipeService").(recipe.Service)

	var totalChannel = make(chan int)
	go func() {
		var total, _ = recipeService.Count()
		totalChannel <- total
	}()

	var recipesChannel = make(chan []recipe.Model)
	go func() {
		result, _ := recipeService.List(args.First, args.Last, args.Before, args.After)
		recipesChannel <- result
	}()

	var (
		start string
		end   string
	)

	var recipes = <-recipesChannel

	if len(recipes) == 0 {
		start, end = "", ""
	} else {
		start, end = recipes[0].ID.Hex(), recipes[len(recipes)-1].ID.Hex()
	}

	hasPreviousPageChannel, hasNextPageChannel := relay.GetHasPreviousAndNextPage(len(recipes), start, end, recipeService)

	return &recipe.ConnectionResolver{
		Models: recipes,
		ConnectionResolver: relay.ConnectionResolver{
			relay.Connection{
				Total:           int32(<-totalChannel),
				From:            start,
				To:              end,
				HasNextPage:     <-hasNextPageChannel,
				HasPreviousPage: <-hasPreviousPageChannel,
			},
		},
	}, nil
}

func setDataOnModel(model *recipe.Model, input *recipe.MutationInput) {
	model.Inputs = make([]recipe.InputElement, len(*input.Inputs))
	model.Outputs = make([]recipe.OutputElement, len(*input.Outputs))

	for i := range *input.Inputs {
		model.Inputs[i] = recipe.InputElement{recipe.InOutElement{
			ItemID: bson.ObjectIdHex(string((*input.Inputs)[i].ItemID)),
			Amount: int32((*input.Inputs)[i].Amount),
		}}
	}
	for i := range *input.Outputs {
		model.Outputs[i] = recipe.OutputElement{recipe.InOutElement{
			ItemID: bson.ObjectIdHex(string((*input.Outputs)[i].ItemID)),
			Amount: int32((*input.Outputs)[i].Amount),
		}}
	}
}

func (r *Resolver) CreateRecipe(ctx context.Context, args struct {
	Input *recipe.MutationInput
}) (*recipe.Resolver, error) {
	recipeService := ctx.Value("recipeService").(recipe.Service)

	inputModel := recipe.Model{}
	setDataOnModel(&inputModel, args.Input)

	newModel, err := recipeService.Create(&inputModel)

	if err == nil {
		return &recipe.Resolver{
			Model: newModel,
		}, nil
	}

	return nil, err
}

func (r *Resolver) UpdateRecipe(ctx context.Context, args struct {
	Id    string
	Input *recipe.MutationInput
}) (*recipe.Resolver, error) {
	recipeService := ctx.Value("recipeService").(recipe.Service)

	inputModel := recipe.Model{}
	setDataOnModel(&inputModel, args.Input)

	newModel, err := recipeService.Update(args.Id, &inputModel)

	if err == nil {
		return &recipe.Resolver{
			Model: newModel,
		}, nil
	}

	return nil, err
}

func (r *Resolver) DeleteRecipe(ctx context.Context, args struct {
	Id string
}) (*graphql.ID, error) {
	recipeService := ctx.Value("recipeService").(recipe.Service)

	deletedID, err := recipeService.DeleteByID(args.Id)
	result := graphql.ID(deletedID)

	if err == nil {
		return &result, nil
	}

	return nil, err
}

func (r *Resolver) Recipe(ctx context.Context, args struct {
	Id string
}) (*recipe.Resolver, error) {
	recipeService := ctx.Value("recipeService").(recipe.Service)

	queryRecipe, err := recipeService.FindByID(args.Id)

	if err == nil {
		return &recipe.Resolver{
			Model: queryRecipe,
		}, nil
	}

	return nil, err
}

func fetchFFXIVNamespace(ctx context.Context) (string, error) {
	fetcher := ctx.Value("apigatewayfetcher").(dukgraphql.Fetcher)

	namespaceResult, err := fetcher.Fetch(dukgraphql.Request{
		Query: "query { namespaceByName(name: \"FFXIV\") { _id name } }",
	})

	if err != nil {
		fmt.Printf("Error fetching namespace: %v\n", err)
		return "", err
	}

	namespaceResponse := dukgraphql.Response{namespaceResult}

	return namespaceResponse.GetObject("namespaceByName").GetString("_id"), nil
}

func fetchFFXIVItemByName(ctx context.Context, name string, namespaceId string) (string, error) {
	fetcher := ctx.Value("apigatewayfetcher").(dukgraphql.Fetcher)

	itemResult, err := fetcher.Fetch(dukgraphql.Request{
		Query: "query { findItem(name: \"" + name + "\", namespaceId: \"" + namespaceId + "\") { _id } }",
	})

	if err != nil {
		fmt.Printf("Error fetching item: %v\n", err)
		return "", err
	}

	itemResponse := dukgraphql.Response{itemResult}

	return itemResponse.GetObject("findItem").GetString("_id"), nil
}

var rcItemMap = make(map[string]string)

func ConvertRcItemIDToItemServiceID(ctx context.Context, id string, namespaceId string) (string, error) {
	if _, ok := rcItemMap[id]; ok {
		return rcItemMap[id], nil
	}

	rcItemResponse, err := http.Get("https://rc.dukfaar.com/api/item/" + id)
	if err != nil {
		fmt.Printf("Error getting item: %v\n", err)
		return "", err
	}
	defer rcItemResponse.Body.Close()

	var itemData struct {
		Id   string `json:"_id"`
		Name string `json:"name"`
	}
	err = json.NewDecoder(rcItemResponse.Body).Decode(&itemData)

	if err != nil {
		fmt.Printf("Error parsing item data: %v\n", err)
		return "", err
	}

	newId, err := fetchFFXIVItemByName(ctx, itemData.Name, namespaceId)

	if err != nil {
		fmt.Printf("Error fetching service item: %v\n", err)
		return "", err
	}

	rcItemMap[id] = newId

	return newId, nil
}

func ConvertRecipe(ctx context.Context, recipe map[string]interface{}, namespaceId string) error {
	delete(recipe, "_id")
	recipe["namespace"] = namespaceId

	inputs := recipe["inputs"].([]interface{})
	for inputIndex := range inputs {
		input := inputs[inputIndex].(map[string]interface{})
		delete(input, "_id")
		newId, err := ConvertRcItemIDToItemServiceID(ctx, input["item"].(string), namespaceId)
		if err != nil {
			return err
		}
		input["item"] = newId
	}

	outputs := recipe["outputs"].([]interface{})
	for outputIndex := range outputs {
		output := outputs[outputIndex].(map[string]interface{})
		delete(output, "_id")
		newId, err := ConvertRcItemIDToItemServiceID(ctx, output["item"].(string), namespaceId)
		if err != nil {
			return err
		}
		output["item"] = newId
	}

	return nil
}

func (r *Resolver) RcRecipeImport(ctx context.Context) (string, error) {
	err := permission.Check(ctx, "mutation.rcRecipeImport")
	if err != nil {
		return "No Permission", err
	}

	rcRecipeResponse, err := http.Get("https://rc.dukfaar.com/api/recipe")

	if err != nil {
		fmt.Printf("Error getting recipes: %v\n", err)
		return "Error reading from RC", err
	}
	defer rcRecipeResponse.Body.Close()

	var recipeData []interface{}
	err = json.NewDecoder(rcRecipeResponse.Body).Decode(&recipeData)

	if err != nil {
		fmt.Printf("Error reading recipes: %v\n", err)
		return "Error parsing data from RC", err
	}

	eventbus := ctx.Value("eventbus").(eventbus.EventBus)
	namespaceId, err := fetchFFXIVNamespace(ctx)
	if err != nil {
		return "Error fetching namespace", err
	}

	go func() {
		for index := range recipeData {
			err := ConvertRecipe(ctx, recipeData[index].(map[string]interface{}), namespaceId)

			if err == nil {
				eventbus.Emit("import.recipe", recipeData[index])
			}
		}
	}()

	return "OK", nil
}
