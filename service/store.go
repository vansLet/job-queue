package service

import (
	"errors"
	"jobqueue/model"
	"jobqueue/worker"
	"sync"
	"time"
)

type DataTask struct {
	id   int
	data []model.Task
	pool *worker.Pool
	mx   *sync.RWMutex
}

var ErrNotFound error = errors.New("task not found")

func NewData(worker *worker.Pool) *DataTask {
	return &DataTask{id: 1, data: make([]model.Task, 0, 10), pool: worker, mx: &sync.RWMutex{}}
}

func (d *DataTask) CreateTask(job model.Task) (model.Task, error) {
	j := job
	d.mx.RLock()
	j.ID = d.id
	d.mx.RUnlock()
	//
	j.Status = model.Pending
	j.CreateAt = time.Now()
	d.pool.SpawnJob(func() {
		d.mx.Lock()
		defer d.mx.Unlock()
		c := j
		c.Status = model.Complete
		d.data = append(d.data, c)
		d.id++
	})
	return j, nil
}

func (d *DataTask) GetAll() []model.Task {
	d.mx.RLock()
	defer d.mx.RUnlock()
	return d.data
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
