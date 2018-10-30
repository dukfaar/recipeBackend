package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/dukfaar/goUtils/env"
	"github.com/dukfaar/goUtils/eventbus"
	dukGraphql "github.com/dukfaar/goUtils/graphql"
	dukHttp "github.com/dukfaar/goUtils/http"
	"github.com/dukfaar/goUtils/permission"
	"github.com/dukfaar/recipeBackend/recipe"

	"github.com/globalsign/mgo"

	"github.com/gorilla/websocket"

	graphql "github.com/graph-gophers/graphql-go"
	graphqlRelay "github.com/graph-gophers/graphql-go/relay"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func createApiGatewayFetcher() dukGraphql.Fetcher {
	url := env.GetDefaultEnvVar("API_GATEWAY_HOST", "localhost") + ":" + env.GetDefaultEnvVar("API_GATEWAY_PORT", "8090")
	path := env.GetDefaultEnvVar("API_GATEWAY_PATH", "/graphql")

	apiGatewayFetcher, err := dukGraphql.NewHttpFetcher(url, path)

	if err != nil {
		panic(err)
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	loginApiGatewayFetcher := dukGraphql.NewClientLoginHttpFetcher(apiGatewayFetcher, clientID, clientSecret)

	return loginApiGatewayFetcher
}

func main() {
	dbSession, err := mgo.Dial(env.GetDefaultEnvVar("DB_HOST", "localhost"))
	if err != nil {
		panic(err)
	}
	defer dbSession.Close()

	log.Println("Connected to database")

	db := dbSession.DB("recipe")

	nsqEventbus := eventbus.NewNsqEventBus(env.GetDefaultEnvVar("NSQD_TCP_URL", "localhost:4150"), env.GetDefaultEnvVar("NSQLOOKUP_HTTP_URL", "localhost:4161"))

	permissionService := permission.NewService()
	loginApiGatewayFetcher := createApiGatewayFetcher()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "db", db)
	ctx = context.WithValue(ctx, "recipeService", recipe.NewMgoService(db, nsqEventbus))
	ctx = context.WithValue(ctx, "permissionService", permissionService)
	ctx = context.WithValue(ctx, "eventbus", nsqEventbus)
	ctx = context.WithValue(ctx, "apigatewayfetcher", loginApiGatewayFetcher)

	schema := graphql.MustParseSchema(Schema, &Resolver{})

	http.Handle("/graphql", dukHttp.AddContext(ctx, dukHttp.Authenticate(&graphqlRelay.Handler{
		Schema: schema,
	})))

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

	serviceInfo := eventbus.ServiceInfo{
		Name:                  "recipe",
		Hostname:              env.GetDefaultEnvVar("PUBLISHED_HOSTNAME", "servicebackend"),
		Port:                  env.GetDefaultEnvVar("PUBLISHED_PORT", "8080"),
		GraphQLHttpEndpoint:   "/graphql",
		GraphQLSocketEndpoint: "/socket",
		SchemaExtensions: []eventbus.SchemaExtension{{
			Type: "RecipeInput",
			Fields: []eventbus.FieldType{
				{
					Name: "item",
					Type: "Item",
					Resolve: eventbus.ResolveType{
						By: "item",
						FieldArguments: map[string]string{
							"id": "itemId",
						},
					},
				},
			}}, {
			Type: "RecipeOutput",
			Fields: []eventbus.FieldType{
				{
					Name: "item",
					Type: "Item",
					Resolve: eventbus.ResolveType{
						By: "item",
						FieldArguments: map[string]string{
							"id": "itemId",
						},
					},
				},
			},
		}},
	}

	result, err := loginApiGatewayFetcher.Fetch(dukGraphql.Request{
		Query: permission.Query,
	})
	if err != nil {
		panic(err)
	}
	queryResult := dukGraphql.Response{result}

	permission.ParseQueryResponse(queryResult, permissionService)
	permissionService.BuildAllUserPermissionData()

	permission.AddAuthEventsHandlers(nsqEventbus, permissionService)

	nsqEventbus.Emit("service.up", serviceInfo)

	nsqEventbus.On("service.up", "recipe", func(msg []byte) error {
		newService := eventbus.ServiceInfo{}
		json.Unmarshal(msg, &newService)

		if newService.Name == "apigateway" {
			nsqEventbus.Emit("service.up", serviceInfo)
		}

		return nil
	})

	eventDBSession := dbSession.Clone()
	eventDB := eventDBSession.DB("recipe")
	defer eventDBSession.Close()
	eventRecipeService := recipe.NewMgoService(eventDB, nsqEventbus)

	nsqEventbus.On("import.recipe", "recipe", CreateRCEventImporter(eventRecipeService, loginApiGatewayFetcher))

	http.Handle("/metrics", promhttp.Handler())

	dukGraphql.EmitRegisterEvents("registerQuery", schema.Inspect().QueryType(), nsqEventbus)
	dukGraphql.EmitRegisterEvents("registerMutation", schema.Inspect().MutationType(), nsqEventbus)
	dukGraphql.EmitRegisterEvents("registerSubscription", schema.Inspect().SubscriptionType(), nsqEventbus)
	dukGraphql.EmitRegisterTypeEvents("registerType", schema.Inspect().Types(), nsqEventbus)

	log.Fatal(http.ListenAndServe(":"+env.GetDefaultEnvVar("PORT", "8080"), nil))
}
