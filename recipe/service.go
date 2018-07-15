package recipe

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/dukfaar/goUtils/eventbus"
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
}

type MgoService struct {
	db         *mgo.Database
	collection *mgo.Collection
	eventbus   eventbus.EventBus
}

func NewMgoService(db *mgo.Database, eventbus eventbus.EventBus) *MgoService {
	return &MgoService{
		db:         db,
		collection: db.C("recipes"),
		eventbus:   eventbus,
	}
}

func (s *MgoService) Create(model *Model) (*Model, error) {
	model.ID = bson.NewObjectId()

	err := s.collection.Insert(model)

	if err == nil {
		s.eventbus.Emit("recipe.created", model)
	}

	return model, err
}

func (s *MgoService) Update(id string, input interface{}) (*Model, error) {
	err := s.collection.UpdateId(bson.ObjectIdHex(id), input)

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
	err := s.collection.RemoveId(bson.ObjectIdHex(id))

	if err == nil {
		s.eventbus.Emit("recipe.deleted", id)
	}

	return id, err
}

func (s *MgoService) FindByID(id string) (*Model, error) {
	var result Model

	err := s.collection.FindId(bson.ObjectIdHex(id)).One(&result)

	return &result, err
}

func (s *MgoService) HasElementBeforeID(id string) (bool, error) {
	query := bson.M{}

	query["_id"] = bson.M{
		"$lt": bson.ObjectIdHex(id),
	}

	count, err := s.collection.Find(query).Count()
	return count > 0, err
}

func (s *MgoService) HasElementAfterID(id string) (bool, error) {
	query := bson.M{}

	query["_id"] = bson.M{
		"$gt": bson.ObjectIdHex(id),
	}

	count, err := s.collection.Find(query).Count()
	return count > 0, err
}

func (s *MgoService) Count() (int, error) {
	count, err := s.collection.Find(bson.M{}).Count()
	return count, err
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
		count, _ := s.collection.Find(query).Count()

		limit = int(*last)
		skip = count - limit
	}

	var result []Model
	err := s.collection.Find(query).Skip(skip).Limit(limit).All(&result)
	return result, err
}
