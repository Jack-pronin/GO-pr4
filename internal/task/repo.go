package task

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Repo struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]*Task
}

func NewRepo() *Repo {
	r := &Repo{items: make(map[int64]*Task)}
	r.loadFromFile()
	return r
}

func (r *Repo) loadFromFile() {
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		return
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return
	}

	for _, task := range tasks {
		r.items[task.ID] = task
		if task.ID > r.seq {
			r.seq = task.ID
		}
	}
}

func (r *Repo) saveToFile() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		tasks = append(tasks, t)
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		return
	}

	os.WriteFile("tasks.json", data, 0644)
}

func (r *Repo) List() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, t)
	}
	return out
}

func (r *Repo) GetWithPagination(page, limit int, done *bool) ([]*Task, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allTasks := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		if done == nil || t.Done == *done {
			allTasks = append(allTasks, t)
		}
	}

	total := len(allTasks)

	start := (page - 1) * limit
	if start >= total {
		return []*Task{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return allTasks[start:end], total
}

func (r *Repo) Get(id int64) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *Repo) Create(title string) *Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	now := time.Now()
	t := &Task{ID: r.seq, Title: title, CreatedAt: now, UpdatedAt: now, Done: false}
	r.items[t.ID] = t

	go r.saveToFile()

	return t
}

func (r *Repo) Update(id int64, title string, done bool) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	t.Title = title
	t.Done = done
	t.UpdatedAt = time.Now()

	go r.saveToFile()

	return t, nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)

	go r.saveToFile()

	return nil
}
