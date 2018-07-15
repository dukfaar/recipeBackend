package main

import (
	"context"

	"gopkg.in/mgo.v2/bson"

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
