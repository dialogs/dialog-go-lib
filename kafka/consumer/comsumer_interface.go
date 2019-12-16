package consumer

import (
	"time"
)

type ConsumerI interface {
	Delay(time.Duration) error
}
