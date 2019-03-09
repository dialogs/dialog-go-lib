package rand

import (
	"math/rand"
	"sync"
	"time"
)

var global *Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	global = NewRand(source)
}

// Int63 returns a random value with using a mutex (for exclude a repetitions)
// from global randomizer
func Int63() int64 {
	return global.Int63()
}

// RawInt63 returns a random value (will be a repetitions and errors)
// Error example: panic: runtime error: index out of range
// from global randomizer
func RawInt63() int64 {
	return global.RawInt63()
}

// A Rand is wrapper of native randomizer
type Rand struct {
	rand *rand.Rand
	mu   sync.RWMutex
}

// NewRand create new randomizer
func NewRand(source rand.Source) *Rand {
	return &Rand{
		rand: rand.New(source),
	}
}

// Int63 returns a random value with using a mutex (for exclude a repetitions)
func (r *Rand) Int63() (val int64) {
	r.mu.Lock()
	val = r.rand.Int63()
	r.mu.Unlock()
	return
}

// RawInt63 returns a random value (will be a repetitions and errors)
// Error example: panic: runtime error: index out of range
func (r *Rand) RawInt63() (val int64) {
	return r.rand.Int63()
}
