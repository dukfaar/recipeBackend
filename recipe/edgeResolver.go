package recipe

import graphql "github.com/graph-gophers/graphql-go"

type EdgeResolver struct {
	Model *Model
}

func (r *EdgeResolver) Node() *Resolver {
	return &Resolver{Model: r.Model}
}

func (r *EdgeResolver) Cursor() graphql.ID {
	return graphql.ID(r.Model.ID.Hex())
}
