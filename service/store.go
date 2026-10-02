package service

import (
	"errors"
	"jobqueue/model"
	"jobqueue/pool"
	"sync"
	"sync/atomic"
	"time"
)

type DataTask struct {
	id   atomic.Int32
	data []model.Task
	pool pool.TaskAdder
	mx   *sync.RWMutex
}

var ErrNotFound error = errors.New("task not found")

func NewData(worker pool.TaskAdder) *DataTask {
	return &DataTask{id: atomic.Int32{}, data: make([]model.Task, 0, 10), pool: worker, mx: &sync.RWMutex{}}
}

func (d *DataTask) CreateTask(job model.Task) (model.Task, error) {
	j := job
	j.ID = int(d.id.Add(1))
	j.Status = model.Pending
	j.CreateAt = time.Now()

	d.pool.AddTask(func() {
		d.mx.Lock()
		defer d.mx.Unlock()
		c := j
		c.Status = model.Complete
		d.data = append(d.data, c)
	})
	return j, nil
}

func (d *DataTask) GetAll() []model.Task {
	d.mx.Lock()
	defer d.mx.Unlock()
	cdata := make([]model.Task, len(d.data))
	copy(cdata, d.data)
	return cdata
}

func (d *DataTask) GetById(id int) (model.Task, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	for _, t := range d.data {
		if t.ID == id {
			return t, nil
		}
	}
	return model.Task{}, ErrNotFound
}

func (d *DataTask) Delete(id int) error {
	for i, t := range d.data {
		if t.ID == id {
			d.mx.Lock()
			defer d.mx.Unlock()
			d.data = append(d.data[:i], d.data[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
