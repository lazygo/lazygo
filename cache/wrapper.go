package cache

import (
	"reflect"
	"time"
)

type DataResult interface {
	ToObject(val interface{}) error
}

type wrapper struct {
	Data struct {
		Meta     interface{} `json:"meta"`
		Deadline int64       `json:"deadline"`
	}
	handler func(wp *wrapper) error
}

func (wp *wrapper) Pack(value interface{}, ttl time.Duration) error {
	wp.Data.Meta = value
	wp.Data.Deadline = time.Now().Unix() + int64(ttl.Seconds())
	return nil
}

func (wp *wrapper) PackFunc(value func() (interface{}, error), ttl time.Duration) error {
	var err error
	res, err := value()
	if err != nil {
		return err
	}
	resV := reflect.ValueOf(res)
	metaV := reflect.ValueOf(wp.Data.Meta)
	metaV.Elem().Set(resV.Elem())
	wp.Data.Deadline = time.Now().Unix() + int64(ttl.Seconds())
	return nil
}

func (wp *wrapper) ToObject(val interface{}) error {
	wp.Data.Meta = val
	return wp.handler(wp)
}
