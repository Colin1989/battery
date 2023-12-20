package facade

// Marshaler represents a marshal interface
type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
}

// Unmarshaler represents a Unmarshal interface
type Unmarshaler interface {
	Unmarshal([]byte, interface{}) error
}

// Serializer is the interface that groups the basic Marshal and Unmarshal methods.
type Serializer interface {
	Marshaler
	Unmarshaler
	GetName() string
}
