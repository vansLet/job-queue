package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jobqueue/model"
	"jobqueue/pool"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type DatabaseTasks interface {
	Add(ctx context.Context, new model.Task) (model.Task, error)
	Gets(ctx context.Context) []model.Task
	GetById(ctx context.Context, id int) (model.Task, error)
	Remove(ctx context.Context, id int) error
}

type changeStatus struct {
	id     int
	status model.Status
}

type DefaultDBTask struct {
	db    *sql.DB
	pool  pool.TaskAdder
	close bool
	notif chan changeStatus
	wg    *sync.WaitGroup
	mtx   *sync.RWMutex
}

func (d *DefaultDBTask) IsClose() bool {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.close
}

func (d *DefaultDBTask) Close() {
	if d.IsClose() {
		return
	}
	d.mtx.Lock()
	d.close = true
	d.mtx.Unlock()

	close(d.notif)
	d.wg.Wait()
	d.db.Close()
}

func (d *DefaultDBTask) setStatus(ctx context.Context, id int, status model.Status) error {
	query := `
		UPDATE tasks 
			SET status = ? 
			WHERE tasks.id = ?;
	`
	maxRetry := 5
	for i := range maxRetry {
		r, err := d.db.ExecContext(ctx, query, status.String(), id)
		if err == nil {
			rid, err := r.RowsAffected()
			if err == nil {
				if rid >= 1 {
					return nil
				}
			}
		}
		delay, err := time.ParseDuration(fmt.Sprintf("%d000ms", i+1))
		if err != nil {
			panic(err)
		}
		<-time.After(delay)
	}
	return fmt.Errorf("update task id:%d failed ", id)
}

func (d *DefaultDBTask) asyncUpdateStatus() error {
	ctx := context.Background()
	d.wg.Add(1)
	err := d.pool.AddTask(ctx, func() {
		defer func() {
			fmt.Println("Worker DB Closing")
			d.wg.Done()
		}()
		fmt.Println("Worker DB Is Running")
		for change := range d.notif {
			err := d.setStatus(ctx, change.id, change.status)
			if err != nil {
				_ = d.setStatus(ctx, change.id, model.Failed)
			}
		}
	})
	return err
}

func (d *DefaultDBTask) runUpdateStatusBackground() {
	go func() {
		for !d.IsClose() {
			d.wg.Wait()
			if d.IsClose() {
				break
			}
			_ = d.asyncUpdateStatus()
		}
	}()
}

func (d *DefaultDBTask) sendToUpdateStatus(ctx context.Context, status changeStatus) error {
	if d.IsClose() {
		return pool.ErrChannelClose
	}
	select {
	case d.notif <- status:
		return nil
	case <-ctx.Done():
		e := ctx.Err()
		if errors.Is(e, context.DeadlineExceeded) {
			return ErrDBTimeOut
		} else if errors.Is(e, context.Canceled) {
			return ErrCanceldDB
		}
		return e
	}

}

func (d *DefaultDBTask) Add(ctx context.Context, new model.Task) (model.Task, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return new, err
	}
	defer tx.Rollback()
	new.CreateAt = time.Now()

	query := `
		INSERT INTO tasks (type, payload, status, create_at) 
		VALUES( ?, ?, ?, ?)
	`

	r, err := tx.ExecContext(
		ctx,
		query,
		new.Type, new.Payload, model.Pending.String(), new.CreateAt,
	)
	if err != nil {
		return model.Task{}, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return model.Task{}, err
	}
	new.ID = int(id)
	if err = d.sendToUpdateStatus(ctx,
		changeStatus{new.ID,
			model.Complete,
		}); err != nil {
		return model.Task{}, err
	}
	err = tx.Commit()
	if err != nil {
		return model.Task{}, err
	}
	return new, nil
}

func (d *DefaultDBTask) Gets(ctx context.Context) []model.Task {
	query := `
		SELECT id, type, payload, status, create_at FROM tasks;
	`
	tasks := make([]model.Task, 0, 10)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return []model.Task{}
	}
	defer rows.Close()
	var status string
	var t model.Task
	for rows.Next() {
		if rows.Err() != nil {
			return tasks
		}
		err := rows.Scan(&t.ID, &t.Type, &t.Payload, &status, &t.CreateAt)
		if err != nil {
			return tasks
		}
		s, err := model.ToStatus(status)
		if err != nil {
			t.Status = model.Pending
		} else {
			t.Status = s
		}
		tasks = append(tasks, t)

	}

	return tasks
}
func (d *DefaultDBTask) GetById(ctx context.Context, id int) (model.Task, error) {
	query := `
		SELECT id, type, payload, status, create_at FROM tasks Where id = ?;
	`
	row := d.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		return model.Task{}, ErrTaskNotFound
	}
	var status string
	var t model.Task
	err := row.Scan(&t.ID, &t.Type, &t.Payload, &status, &t.CreateAt)
	if err != nil {
		return model.Task{}, ErrTaskNotFound
	}
	s, err := model.ToStatus(status)
	if err != nil {
		t.Status = model.Pending
	} else {
		t.Status = s
	}
	return t, nil
}

func (d *DefaultDBTask) Remove(ctx context.Context, id int) error {
	query := `
		DELETE FROM tasks WHERE id = ? ;
	`
	r, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	i, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if i >= 1 {
		return nil
	}
	return ErrTaskNotFound
}

func Init(path string, p pool.TaskAdder) (*DefaultDBTask, error) {
	if p == nil {
		panic("pool cannot nil")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(time.Second * 30)
	db.SetConnMaxLifetime(time.Minute * 30)

	qeury := `
		CREATE TABLE IF NOT EXISTS tasks(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			payload TEXT NOT NULL,
			status VARCHAR(20) NOT NULL,
			create_at DATETIME NOT NULL
		);
	`

	if _, err = db.Exec(qeury); err != nil {
		db.Close()
		return nil, err
	}
	d := &DefaultDBTask{
		db:    db,
		pool:  p,
		notif: make(chan changeStatus, 5),
		close: false,
		wg:    &sync.WaitGroup{},
		mtx:   &sync.RWMutex{},
	}
	d.runUpdateStatusBackground()

	return d, nil
}
