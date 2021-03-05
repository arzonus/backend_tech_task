package application_memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PxyUp/backend_tech_task/internal/application"

	"github.com/tidwall/buntdb"
	"golang.org/x/sync/singleflight"
)

type Repository struct {
	application.Repository
	db *buntdb.DB

	// using for avoiding multiple changing in one moment
	sg singleflight.Group
}

func NewRepository(repository application.Repository) (*Repository, error) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return nil, err
	}

	if err := db.CreateIndex("status", "*", buntdb.IndexJSON("Status")); err != nil {
		return nil, err
	}
	if err := db.CreateIndex("user_id", "*", buntdb.IndexJSON("UserID")); err != nil {
		return nil, err
	}
	if err := db.CreateIndex("created_at", "*", buntdb.IndexJSON("CreatedAt")); err != nil {
		return nil, err
	}
	if err := db.CreateIndex("updated_at", "*", buntdb.IndexJSON("UpdatedAt")); err != nil {
		return nil, err
	}

	r := &Repository{
		Repository: repository,
		db:         db,
	}

	if err := r.warmCache(context.TODO()); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) warmCache(ctx context.Context) error {
	apps, err := r.Repository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("couldn't warm cache: %w", err)
	}
	return r.SetMultiple(apps...)
}

func (r *Repository) Update(ctx context.Context, params *application.UpdateParams) (*application.Application, error) {
	v, err, _ := r.sg.Do(params.ID, func() (interface{}, error) {
		app, err := r.Repository.Update(ctx, params)
		if err != nil {
			return nil, err
		}

		if err := r.Set(app); err != nil {
			return nil, err
		}
		return app, nil
	})
	if err != nil {
		return nil, err
	}

	app, ok := v.(*application.Application)
	if !ok {
		return nil, fmt.Errorf("update couldn't return app")
	}

	return app, nil
}

func (r *Repository) Create(ctx context.Context, app *application.Application) error {
	_, err, _ := r.sg.Do(app.ID, func() (interface{}, error) {
		if err := r.Repository.Create(ctx, app); err != nil {
			return nil, err
		}

		if err := r.Set(app); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (r *Repository) FindByID(ctx context.Context, id string) (*application.Application, error) {
	var m = new(ApplicationModel)

	if err := r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(id)
		if err != nil {
			return err
		}

		return json.Unmarshal([]byte(v), m)
	}); err != nil {
		if !errors.Is(err, buntdb.ErrNotFound) {
			app, err := r.Repository.FindByID(ctx, id)
			if err != nil {
				return nil, err
			}

			return app, r.Set(app)
		}
		return nil, err
	}

	return m.Parse(), nil
}

func (r *Repository) FindByFilters(
	ctx context.Context,
	filter *application.GetByFilterParams,
) ([]application.Application, error) {
	var apps []application.Application

	err := r.db.View(func(tx *buntdb.Tx) error {
		var (
			// storing results of search of whole query
			results map[string]string
			// storing results of filter search
			tmp = make(map[string]string)
		)

		merge := func() func(key, value string) bool {
			// if it's first filter search, save apps to results

			if results == nil {
				results = make(map[string]string)
				return func(key, value string) bool {
					results[key] = value
					return true
				}
			}

			// other searches will be merged with results
			// and stored result of merging to tmp
			return func(key, value string) bool {
				if _, ok := results[key]; ok {
					tmp[key] = value
				}
				return true
			}
		}

		// after filter search, need to cleanup tmp
		// and store tmp to results
		swap := func() {
			results = tmp
			tmp = make(map[string]string)
		}

		if filter.Status != nil {
			if err := tx.AscendEqual(
				"status",
				fmt.Sprintf(`{"Status": %d}`, filter.Status.Int32()),
				merge(),
			); err != nil {
				return err
			}
			swap()
		}

		if filter.UserID != nil {
			if err := tx.AscendEqual(
				"user_id",
				fmt.Sprintf(`{"UserID": "%s"}`, *filter.UserID),
				merge(),
			); err != nil {
				return err
			}
			swap()
		}

		if filter.CreatedAt != nil {
			if err := tx.AscendRange(
				"created_at",
				fmt.Sprintf(`{"CreatedAt": %d}`, filter.CreatedAt.Start.Unix()),
				fmt.Sprintf(`{"CreatedAt": %d}`, filter.CreatedAt.End.Unix()),
				merge(),
			); err != nil {
				return err
			}
		}

		if filter.UpdatedAt != nil {
			if err := tx.AscendRange(
				"updated_at",
				fmt.Sprintf(`{"UpdatedAt": %d}`, filter.UpdatedAt.Start.Unix()),
				fmt.Sprintf(`{"UpdatedAt": %d}`, filter.UpdatedAt.End.Unix()),
				merge(),
			); err != nil {
				return err
			}
			swap()
		}

		if results == nil {
			return nil
		}

		var m ApplicationModel
		for _, value := range results {
			if err := json.Unmarshal([]byte(value), &m); err != nil {
				return err
			}
			apps = append(apps, *m.Parse())
		}

		return nil
	})
	if err != nil || len(apps) == 0 {
		apps, err := r.Repository.FindByFilters(ctx, filter)
		if err != nil {
			return nil, err
		}

		return apps, r.SetMultiple(apps...)
	}

	return apps, nil
}

func (r *Repository) Set(app *application.Application) error {
	b, err := json.Marshal(NewApplicationModel(app))
	if err != nil {
		return err
	}

	return r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(app.ID, string(b), nil)
		return err
	})
}

func (r *Repository) SetMultiple(apps ...application.Application) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		for _, app := range apps {
			b, err := json.Marshal(NewApplicationModel(&app))
			if err != nil {
				return err
			}

			if _, _, err := tx.Set(app.ID, string(b), nil); err != nil {
				return err
			}
		}
		return nil
	})
}

type ApplicationModel struct {
	application.Application
	CreatedAt int64
	UpdatedAt int64
}

func (m ApplicationModel) Parse() *application.Application {
	m.Application.CreatedAt = time.Unix(m.CreatedAt, 0)
	m.Application.UpdatedAt = time.Unix(m.UpdatedAt, 0)
	return &m.Application
}

func (m ApplicationModel) Value() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NewApplicationModel(app *application.Application) *ApplicationModel {
	return &ApplicationModel{
		Application: *app,
		CreatedAt:   app.CreatedAt.Unix(),
		UpdatedAt:   app.UpdatedAt.Unix(),
	}
}
