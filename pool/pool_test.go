package pool_test

import (
	"fmt"
	"jobqueue/pool"
	"sync"
	"testing"
	"time"
)

func TestSendPool(t *testing.T) {
	pool := pool.New(5, nil)
	wg := &sync.WaitGroup{}
	t.Run("testing worker add", func(t *testing.T) {
		for i := range 100 {
			wg.Add(1)
			err := pool.AddTask(t.Context(), func() {
				defer wg.Done()
				println("Hello For Num: ", i)
				time.Sleep(time.Millisecond)
			})
			if err != nil {
				fmt.Println("data i: ", i, " cannot send")
				fmt.Println(err)
			}
		}
	})
	// pool.Close()

	t.Run("Test Subimt After Stop and Call again Stop", func(t *testing.T) {
		_ = pool.AddTask(t.Context(), func() {
			panic("")
		})
		// if err == nil {
		// 	t.Fatal("err harus ada")
		// }

		_ = pool.AddTask(t.Context(), func() {
			// panic("")
			fmt.Println("Hello World")
		})

		pool.Close()

	})
	wg.Wait()

}
