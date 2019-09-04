// Serde is a package for serializing and deserializing data
package serde

type Serializer interface {
	Encode(val interface{}) ([]byte, error)
}

type Deserializer interface {
	Decode(data []byte) (interface{}, error)
}

type Serde interface {
	Serializer
	Deserializer
}
