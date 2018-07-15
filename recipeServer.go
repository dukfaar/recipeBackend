package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dukfaar/goUtils/eventbus"
	dukGraphql "github.com/dukfaar/goUtils/graphql"
	dukHttp "github.com/dukfaar/goUtils/http"
	"github.com/dukfaar/recipeBackend/recipe"

	"gopkg.in/mgo.v2"

	"github.com/gorilla/websocket"

	graphql "github.com/graph-gophers/graphql-go"
	graphqlRelay "github.com/graph-gophers/graphql-go/relay"
)

func main() {
	dbSession, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer dbSession.Close()

	log.Println("Connected to database")

	db := dbSession.DB("recipe")

	eventbus := eventbus.NewNsqEventBus("localhost:4150", "localhost:4161")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "db", db)
	ctx = context.WithValue(ctx, "recipeService", recipe.NewMgoService(db, eventbus))

	schema := graphql.MustParseSchema(Schema, &Resolver{})

	http.Handle("/graphql", dukHttp.AddContext(ctx, &graphqlRelay.Handler{
		Schema: schema,
	}))

	http.Handle("/socket", dukHttp.AddContext(ctx, &dukGraphql.SocketHandler{
		Schema: schema,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
