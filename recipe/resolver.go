package recipe

import graphql "github.com/graph-gophers/graphql-go"

type InputElementResolver struct {
	InputElement *InputElement
}

type OutputElementResolver struct {
	OutputElement *OutputElement
}

type Resolver struct {
	Model *Model
}

func (r *Resolver) ID() *graphql.ID {
	id := graphql.ID(r.Model.ID.Hex())
	return &id
}

func (r *Resolver) Inputs() *[]*InputElementResolver {
	l := make([]*InputElementResolver, len(r.Model.Inputs))
	for i, input := range r.Model.Inputs {
		l[i] = &InputElementResolver{InputElement: &input}
	}
	return &l
}

func (r *Resolver) Outputs() *[]*OutputElementResolver {
	l := make([]*OutputElementResolver, len(r.Model.Outputs))
	for i, output := range r.Model.Outputs {
		l[i] = &OutputElementResolver{OutputElement: &output}
	}
	return &l
}

func (r *InputElementResolver) ItemID() *graphql.ID {
	id := graphql.ID(r.InputElement.ItemID.Hex())
	return &id
}

func (r *InputElementResolver) Amount() *int32 {
	result := r.InputElement.Amount
	return &result
}

func (r *OutputElementResolver) ItemID() *graphql.ID {
	id := graphql.ID(r.OutputElement.ItemID.Hex())
	return &id
}

func (r *OutputElementResolver) Amount() *int32 {
	result := r.OutputElement.Amount
	return &result
}
