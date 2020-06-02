package utils

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitTimeout(t *testing.T) {
	wg := &sync.WaitGroup{}

	f := func(wg *sync.WaitGroup, ts time.Duration) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(ts)
		}()
	}

	f(wg, time.Second*2)
	succ := WaitTimeout(wg, time.Second)
	assert.Equal(t, succ, false)

	f(wg, time.Millisecond*500)
	succ = WaitTimeout(wg, time.Second)
	assert.Equal(t, succ, true)
}
