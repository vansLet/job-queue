package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrChannelClose = errors.New("channel close")
)

var defaultHandlingPanic = func(i int32, err any) {
	fmt.Printf("Catch Panic From WorkerID: %d, Err: %v\n\n", i, err)
}

type TaskAdder interface {
	AddTask(context.Context, func()) error
}

type DefaultPool struct {
	handleErr   func(int32, any)
	wg          *sync.WaitGroup
	mtx         *sync.RWMutex
	addTaskWg   *sync.WaitGroup
	workerId    atomic.Int32
	is_close    bool
	channelTask chan func()
}

func New(num int, err func(int32, any)) *DefaultPool {
	if num < 2 {
		panic("num less than 2")
	}

	st := make(chan func(), (num * 2))
	dp := DefaultPool{
		handleErr:   err,
		addTaskWg:   &sync.WaitGroup{},
		mtx:         &sync.RWMutex{},
		wg:          &sync.WaitGroup{},
		workerId:    atomic.Int32{},
		is_close:    false,
		channelTask: st,
	}
	if err == nil {
		dp.handleErr = defaultHandlingPanic
	}
	for range num {
		dp.addWorker()
	}
	time.Sleep(time.Millisecond)
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
	dp.mtx.RLock()
	defer dp.mtx.RUnlock()
	if err := recover(); err != nil && !dp.is_close {
		if dp.handleErr != nil {
			dp.handleErr(id, err)
		}
		_ = err
		return
	}
}
func (dp *DefaultPool) execTask(id int32, task func()) {
	defer dp.catchPanic(id)
	task()
}
func (dp *DefaultPool) Close() {
	if dp.IsClose() {
		return
	}
	dp.mtx.Lock()
	dp.is_close = true
	dp.mtx.Unlock()

	dp.addTaskWg.Wait()
	close(dp.channelTask)

	dp.wg.Wait()
}
func (dp *DefaultPool) IsClose() bool {
	dp.mtx.RLock()
	defer dp.mtx.RUnlock()
	return dp.is_close
}
func (dp *DefaultPool) AddTask(ctx context.Context, send func()) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	if send == nil {
		panic("param send is nil")
	}
	if dp.IsClose() {
		return ErrChannelClose
	}
	dp.addTaskWg.Add(1)
	defer dp.addTaskWg.Done()
	select {
	case dp.channelTask <- send:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
