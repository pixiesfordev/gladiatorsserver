package utility

import (
	"math"
	"sync"
)

type accumulator struct {
	value int64
	mutex sync.Mutex
}

func NewAccumulator() *accumulator {
	return &accumulator{
		value: 0,
	}
}

func (a *accumulator) GetNextIdx() int64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if (a.value + 1) <= math.MaxInt64 {
		a.value += 1
	} else {
		a.value = 0
	}

	return a.value
}
