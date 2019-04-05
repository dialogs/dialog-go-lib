package enum

import "fmt"

// Enum is object with relationships of enumerates and strings values
type Enum struct {
	name           string
	mapIndexString map[interface{}]string
	mapStringIndex map[string]interface{}
}

// New creates enumeration
func New(name string) *Enum {
	return &Enum{
		name:           name,
		mapIndexString: make(map[interface{}]string),
		mapStringIndex: make(map[string]interface{}),
	}
}

// Name returns enumeration name
func (e *Enum) Name() string {
	return e.name
}

// Add new relationship of an enumeration and a string value
// If an index already exists, Handle panics.
func (e *Enum) Add(index interface{}, str string) *Enum {
	if _, ok := e.GetByIndex(index); ok {
		panic(fmt.Sprintf("%v: index already exists: '%v'", e.Name(), index))
	}

	e.mapIndexString[index] = str
	e.mapStringIndex[str] = index
	return e
}

// GetByString returns a enumeration value by string key
func (e *Enum) GetByString(val string) (interface{}, bool) {
	index, ok := e.mapStringIndex[val]
	return index, ok
}

// GetByIndex returns a string value by a enumeration value
func (e *Enum) GetByIndex(val interface{}) (string, bool) {
	str, ok := e.mapIndexString[val]
	return str, ok
}

// StringKeys returns all strings keys of enumeration
func (e *Enum) StringKeys() []string {

	list := make([]string, 0, len(e.mapStringIndex))
	for k := range e.mapStringIndex {
		list = append(list, k)
	}
	return list
}
