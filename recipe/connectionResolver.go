package recipe

import (
	"github.com/dukfaar/goUtils/relay"
)

type ConnectionResolver struct {
	Models []Model
	relay.ConnectionResolver
}

func (r *ConnectionResolver) Edges() *[]*EdgeResolver {
	l := make([]*EdgeResolver, len(r.Models))
	for i := range r.Models {
		l[i] = &EdgeResolver{
			Model: &r.Models[i],
		}
	}
	return &l
}
