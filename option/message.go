package option

import "github.com/sagernet/sing/common/json"

type TypedMessage[T any] struct {
	Value T
}

func (t TypedMessage[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Value)
}

func (t TypedMessage[T]) UnmarshalJSON(b []byte) error {
	return json.UnmarshalDisallowUnknownFields(b, &t.Value)
}
