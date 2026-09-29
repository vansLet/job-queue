package worker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrSizeThread error = errors.New("minimum size thread 2 and max 10")
var ErrChannelClose error = errors.New("channel send job close")

type workerErr struct {
	id   int
	data any
}

type job func()

// struktur data Pool yang hanya bertugas mengontorl workerPool, jika langsung menggukannya tidak akan bisa mengontrol workerPool.
// Untuk Menggunakannya pakai function New().
type Pool struct {
	Count   int
	W_panic bool
	stop    atomic.Bool
	send    chan job
	wg      *sync.WaitGroup
	Err     chan workerErr
}

// menggembalikan Pool dan membuat workerPool yang berjumlah num.
// jika with_panic true wajib memanggil Error() biar tidak terjadi deadlock
// dan num akan ditambah 1 secara otomatis
func New(num int, withPanic bool) *Pool {
	if num < 2 || num > 20 {
		panic(ErrSizeThread)
	}
	if withPanic {
		num += 1 // extra go buat handle error
	}
	pool := &Pool{
		Count:   num,
		W_panic: withPanic,
		wg:      &sync.WaitGroup{},
		stop:    atomic.Bool{},
		send:    make(chan job, 10),
		Err:     make(chan workerErr, 5),
	}
	for i := range pool.Count {
		createWorkerJob(i+1, pool)
	}
	return pool
}

// Kirim f ke workerPool.
//
// Channel ditutup tidak akan bisa menerima job dan mengembalikan error.
//
// Jika f == nil panic!
func (p *Pool) SpawnJob(f func()) error {
	if f == nil {
		panic("f not implement")
	}
	if p.stop.Load() {
		return ErrChannelClose
	}
	p.send <- f
	return nil
}

// f berisi worker id dan error panic dari workerPool,
//
// jika with_panic true: Error() harus dipakai jika tidak akan menyebabkan deadlock, jika false jangan dipakai.
func (p *Pool) Error(f func(fromId int, err any)) {
	if !p.W_panic {
		return
	}
	_ = p.SpawnJob(func() {
		for err_job := range p.Err {
			if err_job.data != nil {
				f(err_job.id, err_job)
			}
		}
	})
}

// close channel & stop menerima job
func (p *Pool) Close() {
	if !p.stop.Load() {
		p.stop.Store(true)
		time.Sleep(time.Millisecond * 500)
		close(p.send)
		close(p.Err)
	}
	p.Wait()
}

// menuggu semua proses didalam workerPool.
//
//	warning: Sebelum digunakan close semua workerPool menggukan Close(), jika tidak akan menyebabkan deadlock
func (p *Pool) Wait() {
	p.wg.Wait()
}

func createWorkerJob(id int, pool *Pool) {
	pool.wg.Add(1)
	println("Success Create worker Id: ", id)
	go func(pool *Pool) {
		defer pool.wg.Done()
		for job := range pool.send {
			if r := runJob(job); r != nil {
				if !pool.stop.Load() && pool.W_panic {
					pool.Err <- workerErr{id, r}
				}
			}
			time.Sleep(time.Millisecond * 500)
		}
	}(pool)
}
func runJob(j job) (err any) {
	defer func() {
		if r := recover(); r != nil {
			err = r
			return
		}
	}()
	if j != nil {
		j()
	}
	return nil
}
