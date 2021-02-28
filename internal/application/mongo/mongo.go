package application_mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PxyUp/backend_tech_task/internal/application"
	"github.com/PxyUp/backend_tech_task/internal/external"
)

type Repository struct {
	coll *mongo.Collection
}

const collectionName = "applications"

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection(collectionName)}
}

func parseInsertedID(insertedID interface{}) (primitive.ObjectID, error) {
	if id, ok := insertedID.(primitive.ObjectID); ok {
		return id, nil
	}
	return [12]byte{}, fmt.Errorf("couldn't parse inserted id")
}

func (r Repository) Create(ctx context.Context, app *application.Application) error {
	m, err := NewApplicationModel(app)
	if err != nil {
		return err
	}

	res, err := r.coll.InsertOne(ctx, m)
	if err != nil {
		return err
	}

	id, err := parseInsertedID(res.InsertedID)
	if err != nil {
		return err
	}

	app.ID = id.Hex()
	return nil
}

func (r Repository) FindByID(ctx context.Context, id string) (*application.Application, error) {
	mID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var m = new(ApplicationModel)
	if err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: mID}}).Decode(m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, application.ErrApplicationNotFound
		}
		return nil, err
	}

	return ParseApplicationModel(m)
}

func (r Repository) FindByFilters(ctx context.Context, params *application.GetByFilterParams) ([]application.Application, error) {
	var filter bson.D
	if params.Status != nil {
		filter = append(filter, bson.E{Key: "status", Value: params.Status.Int32()})
	}
	if params.UserID != nil {
		mUserID, err := primitive.ObjectIDFromHex(*params.UserID)
		if err != nil {
			return nil, err
		}

		filter = append(filter, bson.E{Key: "user_id", Value: mUserID})
	}
	if params.UpdatedAt != nil {
		filter = append(filter, bson.E{
			Key: "updated_at",
			Value: bson.M{
				"$gte": params.UpdatedAt.Start,
				"$lt":  params.UpdatedAt.End,
			},
		})
	}
	if params.CreatedAt != nil {
		filter = append(filter, bson.E{
			Key: "created_at",
			Value: bson.M{
				"$gte": params.CreatedAt.Start,
				"$lt":  params.CreatedAt.End,
			},
		})
	}

	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cur.Close(ctx); err != nil {
			log.Err(err).Msg("couldn't close cursor after find applications by filter")
		}
	}()

	var apps []application.Application
	for cur.Next(ctx) {
		var m ApplicationModel
		err := cur.Decode(&m)
		if err != nil {
			return nil, err
		}

		app, err := ParseApplicationModel(&m)
		if err != nil {
			return nil, err
		}

		apps = append(apps, *app)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (r Repository) Update(ctx context.Context, params *application.UpdateParams) (*application.Application, error) {

	mID, err := primitive.ObjectIDFromHex(params.ID)
	if err != nil {
		return nil, err
	}

	_, err = r.coll.UpdateOne(
		ctx,
		bson.D{
			bson.E{Key: "_id", Value: mID}},
		bson.D{
			{"$set", bson.D{
				bson.E{Key: "status", Value: params.Status},
				bson.E{Key: "updated_at", Value: NewDateTime(time.Now().UTC())},
			}},
		})
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, params.ID)
}

type ApplicationModel struct {
	ID             primitive.ObjectID `bson:"_id"`
	Status         int32              `bson:"status"`
	UserID         primitive.ObjectID `bson:"user_id"`
	CreatedAt      primitive.DateTime `bson:"created_at"`
	UpdatedAt      primitive.DateTime `bson:"updated_at"`
	ExternalStatus int32              `bson:"external_status"`
}

func NewDateTime(t time.Time) primitive.DateTime {
	return primitive.DateTime(t.Unix())
}

func ParseDateTime(t primitive.DateTime) time.Time {
	return time.Unix(int64(t), 0).UTC()
}

func ParseApplicationStatus(v int32) (application.Status, error) {
	s := application.NewStatus(v)
	return s, s.Validate()
}

func ParseApplicationExternalStatus(v int32) (external.Status, error) {
	s := external.NewStatus(v)
	return s, s.Validate()
}

func NewApplicationModel(app *application.Application) (*ApplicationModel, error) {
	var (
		m = &ApplicationModel{
			Status:         app.Status.Int32(),
			CreatedAt:      NewDateTime(app.CreatedAt),
			UpdatedAt:      NewDateTime(app.UpdatedAt),
			ExternalStatus: app.ExternalStatus.Int32(),
		}
		err error
	)
	m.ID, err = primitive.ObjectIDFromHex(app.ID)
	if err != nil {
		return nil, err
	}

	m.UserID, err = primitive.ObjectIDFromHex(app.UserID)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func ParseApplicationModel(m *ApplicationModel) (*application.Application, error) {
	var (
		a = &application.Application{
			ID:        m.ID.String(),
			UserID:    m.UserID.String(),
			CreatedAt: ParseDateTime(m.CreatedAt),
			UpdatedAt: ParseDateTime(m.UpdatedAt),
		}
		err error
	)
	a.Status, err = ParseApplicationStatus(m.Status)
	if err != nil {
		return nil, err
	}

	a.ExternalStatus, err = ParseApplicationExternalStatus(m.ExternalStatus)
	if err != nil {
		return nil, err
	}
	return a, nil
}
