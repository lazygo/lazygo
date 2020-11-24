package cache

import "errors"

const (
	TypeString = iota + 1
	TypeByteArray
	TypeMap
	TypeMapSlice
)

type Map map[string]interface{}
type MapSlice []map[string]interface{}
type StringCall func() string
type ByteArrayCall func() []byte
type MapCall func() Map
type MapSliceCall func() MapSlice

type wrapper struct {
	DataType int32 `json:"type"`
	Metadata interface{} `json:"metadata"`
	Deadline int64 `json:"deadline"`
}

func (wp *wrapper) Pack(value interface{}) error {
	switch value.(type) {
	case string:
		wp.Metadata = value.(string)
		wp.DataType = TypeString
	case []byte:
		wp.Metadata = value.([]byte)
		wp.DataType = TypeByteArray
	case Map:
		wp.Metadata = value.(Map)
		wp.DataType = TypeMap
	case MapSlice:
		wp.Metadata = value.(MapSlice)
		wp.DataType = TypeMapSlice
	case StringCall:
		callback := value.(StringCall)
		wp.Metadata = callback()
		wp.DataType = TypeString
	case ByteArrayCall:
		callback := value.(ByteArrayCall)
		wp.Metadata = callback()
		wp.DataType = TypeByteArray
	case MapCall:
		callback := value.(MapCall)
		wp.Metadata = callback()
		wp.DataType = TypeMap
	case MapSliceCall:
		callback := value.(MapSliceCall)
		wp.Metadata = callback()
		wp.DataType = TypeMapSlice
	default:
		return errors.New("data type error")
	}
	return nil
}

func (wp *wrapper) GetType() int32 {
	return wp.DataType
}

func (wp *wrapper) ToString() (string, error) {
	if wp.DataType == TypeString {
		return wp.Metadata.(string), nil
	}
	if wp.DataType == TypeByteArray {
		return string(wp.Metadata.([]byte)), nil
	}
	return "", errors.New("data type error")
}

func (wp *wrapper) ToByteArray() ([]byte, error) {
	if wp.DataType == TypeString {
		return []byte(wp.Metadata.(string)), nil
	}
	if wp.DataType == TypeByteArray {
		return wp.Metadata.([]byte), nil
	}
	return nil, errors.New("data type error")
}

func (wp *wrapper) ToMap() (Map, error) {
	if wp.DataType == TypeMap {
		return wp.Metadata.(Map), nil
	}
	return nil, errors.New("data type error")
}

func (wp *wrapper) ToMapSlice() (MapSlice, error) {
	if wp.DataType == TypeMapSlice {
		return wp.Metadata.(MapSlice), nil
	}
	return nil, errors.New("data type error")
}