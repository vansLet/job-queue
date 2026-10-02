package pool_test

import (
	"jobqueue/pool"
	"testing"
	"time"
)

func TestSendPool(t *testing.T) {
	t.Parallel()
	pool := pool.New(3, nil)
	t.Run("testing worker add", func(t *testing.T) {
		for i := range 50 {
			pool.AddTask(func() {
				println("Hello For Num: ", i)
				time.Sleep(time.Millisecond * 200)
			})
		}
	})
	t.Run("check len worker", func(t *testing.T) {
		if pool.Len() != 3 {
			t.Errorf("len: %d not equal 3 ", pool.Len())
		}
	})

	pool.Close()

}
