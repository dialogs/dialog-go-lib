package avro

import (
	"fmt"

	"github.com/dialogs/dialog-go-lib/enum"
)

type Kind int

const KindUnknown Kind = 0

var _KindEnum = enum.New("deserializer kind").
	Add(KindUnknown, "unknown")

func AddKind(index Kind, str string) *enum.Enum {
	return _KindEnum.Add(index, str)
}

func (k Kind) String() string {
	val, ok := _KindEnum.GetByIndex(k)
	if !ok {
		return fmt.Sprintf("invalid deserializer kind: %d", k)
	}

	return val
}
