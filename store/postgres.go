package store

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"time"
)

type postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)

	return &postgres{
		db: db,
	}, nil
}

func (s *postgres) Put(req Request) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	err = s.put(tx, req)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *postgres) Get(req Request) error {
	return s.get(s.db, req)
}

func (s *postgres) put(x sqlx.Ext, req Request) error {
	if err := req.validate(); err != nil {
		return err
	}

	resource, err := req.resource()
	if err != nil {
		return err
	}

	_, err = sqlx.NamedExec(
		x,
		`insert into resources (id, customer_id, request_type, request_data, response_type, response_data)
		 values (:id, :customer_id, :request_type, :request_data, :response_type, :response_data)
	         on conflict on constraint resources_pkey do update set (id, customer_id, request_type, request_data, response_type, response_data) =
		 (:id, :customer_id, :request_type, :request_data, :response_type, :response_data)`,
		resource,
	)

	return err
}

func (s *postgres) get(x sqlx.Ext, req Request) error {
	if err := req.validate(); err != nil {
		return err
	}

	resource, err := req.resource()
	if err != nil {
		return err
	}

	err = sqlx.Get(
		x,
		resource,
		`select * from resources where id = $1 and customer_id = $2`,
		resource.Id,
		resource.CustomerId,
	)
	if err != nil {
		return err
	}

	// this shouldn't happen haha
	if resource.UpdatedAt == nil {
		return errMissingUpdated
	}

	// if we don't have a max age set, give it a default
	if req.MaxAge == nil {
		req.MaxAge = &opsee_types.Timestamp{}
		err = req.MaxAge.Scan(time.Now().UTC().Add(-1 * DefaultTTL))
		if err != nil {
			return err
		}
	}

	// stuff in the db is expired, just ignore it
	if resource.UpdatedAt.Millis() > req.MaxAge.Millis() {
		return errResourceExpired
	}

	// ok we good, try 2 re-hydrate
	err = json.Unmarshal(resource.ResponseData, req.Output)
	if err != nil {
		return err
	}

	return nil
}
