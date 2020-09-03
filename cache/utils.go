package cache

import (
	"errors"
)

const (
	TypeString = iota + 1
	TypeByteArray
	TypeMapIf
	TypeMapIfArray
)

type MapIf map[string]interface{}
type MapIfArray []map[string]interface{}
type StringCall func() string
type ByteArrayCall func() []byte
type MapIfCall func() MapIf
type MapIfArrayCall func() MapIfArray

type Wrapper struct {
	DataType int32 `json:"type"`
	Metadata interface{} `json:"metadata"`
	Deadline int64 `json:"deadline"`
}

func NewWrapper() *Wrapper {
	return &Wrapper{}
}

func (wp *Wrapper) Pack(value interface{}) error {
	switch value.(type) {
	case string:
		wp.Metadata = value.(string)
		wp.DataType = TypeString
	case []byte:
		wp.Metadata = value.([]byte)
		wp.DataType = TypeByteArray
	case MapIf:
		wp.Metadata = value.(MapIf)
		wp.DataType = TypeMapIf
	case MapIfArray:
		wp.Metadata = value.(MapIfArray)
		wp.DataType = TypeMapIfArray
	case StringCall:
		callback := value.(StringCall)
		wp.Metadata = callback()
		wp.DataType = TypeString
	case ByteArrayCall:
		callback := value.(ByteArrayCall)
		wp.Metadata = callback()
		wp.DataType = TypeByteArray
	case MapIfCall:
		callback := value.(MapIfCall)
		wp.Metadata = callback()
		wp.DataType = TypeMapIf
	case MapIfArrayCall:
		callback := value.(MapIfArrayCall)
		wp.Metadata = callback()
		wp.DataType = TypeMapIfArray
	default:
		return errors.New("data type error")
	}
	return nil
}

func (wp *Wrapper) GetType() int32 {
	return wp.DataType
}

func (wp *Wrapper) ToString() (string, error) {
	if wp.DataType == TypeString {
		return wp.Metadata.(string), nil
	}
	if wp.DataType == TypeByteArray {
		return string(wp.Metadata.([]byte)), nil
	}
	return "", errors.New("data type error")
}

func (wp *Wrapper) ToByteArray() ([]byte, error) {
	if wp.DataType == TypeString {
		return []byte(wp.Metadata.(string)), nil
	}
	if wp.DataType == TypeByteArray {
		return wp.Metadata.([]byte), nil
	}
	return nil, errors.New("data type error")
}

func (wp *Wrapper) ToMapIf() (MapIf, error) {
	if wp.DataType == TypeMapIf {
		return wp.Metadata.(MapIf), nil
	}
	return nil, errors.New("data type error")
}

func (wp *Wrapper) ToMapIfArray() (MapIfArray, error) {
	if wp.DataType == TypeMapIfArray {
		return wp.Metadata.(MapIfArray), nil
	}
	return nil, errors.New("data type error")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}