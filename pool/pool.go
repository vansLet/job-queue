package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TaskAdder interface {
	AddTask(func()) error
}

type DefaultPool struct {
	handleErr   func(int32, any)
	wg          *sync.WaitGroup
	mtx         *sync.RWMutex
	workerId    atomic.Int32
	is_close    atomic.Bool
	channelTask chan func()
}

func New(num int, err func(int32, any)) *DefaultPool {
	if num < 2 {
		panic("num less than 2")
	}

	st := make(chan func(), num)
	dp := DefaultPool{
		handleErr:   err,
		mtx:         &sync.RWMutex{},
		wg:          &sync.WaitGroup{},
		workerId:    atomic.Int32{},
		is_close:    atomic.Bool{},
		channelTask: st,
	}
	for range num {
		dp.addWorker()
	}
	return &dp
}

func (dp *DefaultPool) reduceWorkerId() {
	i := dp.workerId.Add(-1)
	fmt.Println("ID: ", i, " Worker Close")
	dp.wg.Done()
}
func (dp *DefaultPool) Len() int {
	return int(dp.workerId.Load())
}

func (dp *DefaultPool) addWorker() {
	dp.wg.Add(1)
	go func() {
		id := dp.workerId.Add(1)
		fmt.Println("ID: ", id, " Worker Open")
		defer dp.reduceWorkerId()
		for task := range dp.channelTask {
			dp.execTask(id, task)
		}
	}()
}

func (dp *DefaultPool) catchPanic(id int32) {
	if err := recover(); err != nil && !dp.is_close.Load() {
		if dp.handleErr != nil {
			dp.handleErr(id, err)
		}
	}
}
func (dp *DefaultPool) execTask(id int32, task func()) {
	defer dp.catchPanic(id)
	task()
}
func (dp *DefaultPool) Close() {
	dp.is_close.Store(true)
	time.Sleep(time.Millisecond * 500)
	close(dp.channelTask)
	dp.wg.Wait()
}

func (dp *DefaultPool) AddTask(send func()) error {
	if send == nil {
		panic("param send is nil")
	}
	if dp.is_close.Load() {

	}
	dp.channelTask <- send
	return nil
}
