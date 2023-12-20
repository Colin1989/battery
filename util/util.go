package util

import "github.com/colin1989/battery/facade"

// SerializeOrRaw serializes the interface if its not an array of bytes already
func SerializeOrRaw(serializer facade.Serializer, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	data, err := serializer.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}
