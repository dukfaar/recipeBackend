package recipe

import (
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/dukfaar/goUtils/eventbus"
	"github.com/dukfaar/goUtils/service"
)

type Service interface {
	Create(*Model) (*Model, error)
	DeleteByID(id string) (string, error)
	FindByID(id string) (*Model, error)
	Update(string, interface{}) (*Model, error)

	HasElementBeforeID(id string) (bool, error)
	HasElementAfterID(id string) (bool, error)

	Count() (int, error)

	List(first *int32, last *int32, before *string, after *string) ([]Model, error)

	HasElementBeforeIDWithQuery(bson.M, string) (bool, error)
	HasElementAfterIDWithQuery(bson.M, string) (bool, error)
	CountWithQuery(bson.M) (int, error)

	MakeBaseQuery() bson.M
	MakeListQuery(query bson.M, before *string, after *string)

	PerformQuery(query bson.M) *Model
	PerformListQuery(query bson.M, first *int32, last *int32, before *string, after *string) ([]Model, error)
}

type MgoService struct {
	service.BaseMgoServiceWithQuery
	db       *mgo.Database
	eventbus eventbus.EventBus
}

func NewMgoService(db *mgo.Database, eventbus eventbus.EventBus) *MgoService {
	return &MgoService{
		BaseMgoServiceWithQuery: service.BaseMgoServiceWithQuery{
			Collection: db.C("recipes"),
		},
		db:       db,
		eventbus: eventbus,
	}
}

func (s *MgoService) Create(model *Model) (*Model, error) {
	model.ID = bson.NewObjectId()

	err := s.Collection.Insert(model)

	if err == nil {
		s.eventbus.Emit("recipe.created", model)
	}

	return model, err
}

func (s *MgoService) PerformQuery(query bson.M) *Model {
	var result Model
	s.Collection.Find(query).One(&result)
	return &result
}

func (s *MgoService) PerformListQuery(query bson.M, first *int32, last *int32, before *string, after *string) ([]Model, error) {
	var (
		skip  int
		limit int
	)

	if first != nil {
		limit = int(*first)
	}

	if last != nil {
		count, _ := s.Collection.Find(query).Count()

		limit = int(*last)
		skip = count - limit
	}

	var result []Model
	err := s.Collection.Find(query).Skip(skip).Limit(limit).All(&result)
	return result, err
}

func (s *MgoService) Update(id string, input interface{}) (*Model, error) {
	err := s.Collection.UpdateId(bson.ObjectIdHex(id), input)

	if err != nil {
		return nil, err
	}

	result, err := s.FindByID(id)

	if err != nil {
		return nil, err
	}

	s.eventbus.Emit("recipe.updated", result)

	return result, err
}

func (s *MgoService) DeleteByID(id string) (string, error) {
	err := s.Collection.RemoveId(bson.ObjectIdHex(id))

	if err == nil {
		s.eventbus.Emit("recipe.deleted", id)
	}

	return id, err
}

func (s *MgoService) FindByID(id string) (*Model, error) {
	var result Model

	err := s.Collection.FindId(bson.ObjectIdHex(id)).One(&result)

	return &result, err
}

func (s *MgoService) List(first *int32, last *int32, before *string, after *string) ([]Model, error) {
	query := bson.M{}

	if after != nil {
		query["_id"] = bson.M{
			"$gt": bson.ObjectIdHex(*after),
		}
	}

	if before != nil {
		query["_id"] = bson.M{
			"$lt": bson.ObjectIdHex(*before),
		}
	}

	var (
		skip  int
		limit int
	)

	if first != nil {
		limit = int(*first)
	}

	if last != nil {
		count, _ := s.Collection.Find(query).Count()

		limit = int(*last)
		skip = count - limit
	}

	var result []Model
	err := s.Collection.Find(query).Skip(skip).Limit(limit).All(&result)
	return result, err
}
