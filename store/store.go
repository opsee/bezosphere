package store

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"hash/crc64"
	"reflect"
	"time"
)

var (
	DefaultTTL = 2 * time.Minute
	crcTable   = crc64.MakeTable(crc64.ISO)
)

type Store interface {
	Get(Request) error
	Put(Request) error
}

type resource struct {
	Id           string
	CustomerId   string                 `db:"customer_id"`
	RequestType  string                 `db:"request_type"`
	RequestData  []byte                 `db:"request_data"`
	ResponseType string                 `db:"response_type"`
	ResponseData []byte                 `db:"response_data"`
	CreatedAt    *opsee_types.Timestamp `db:"created_at"`
	UpdatedAt    *opsee_types.Timestamp `db:"updated_at"`
}

type Request struct {
	CustomerId string
	Input      interface{}
	Output     interface{}
	MaxAge     *opsee_types.Timestamp
}

func (req Request) validate() error {
	if req.CustomerId == "" {
		return errMissingCustomerId
	}

	if req.Input == nil {
		return errMissingAWSRequestInput
	}

	if req.Output == nil {
		return errMissingAWSOutput
	}

	return nil
}

func (req Request) resource() (*resource, error) {
	id, err := checksum(req.Input)
	if err != nil {
		return nil, err
	}

	rd, err := json.Marshal(req.Input)
	if err != nil {
		return nil, err
	}

	resd, err := json.Marshal(req.Output)
	if err != nil {
		return nil, err
	}

	return &resource{
		Id:           fmt.Sprintf("%d", id),
		CustomerId:   req.CustomerId,
		RequestType:  reflect.TypeOf(req.Input).String(),
		RequestData:  rd,
		ResponseType: reflect.TypeOf(req.Output).String(),
		ResponseData: resd,
	}, nil
}

func checksum(v interface{}) (uint64, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	if err != nil {
		return 0, err
	}

	return crc64.Checksum(buf.Bytes(), crcTable), nil
}
