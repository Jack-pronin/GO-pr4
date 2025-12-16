package task

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

const dataFile = "tasks.json"

type Repo struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]*Task
}

func NewRepo() *Repo {
	return &Repo{items: make(map[int64]*Task)}
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
	r.save()
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
	r.save()
	return t, nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	r.save()
	return nil
}

func (r *Repo) ListFiltered(done *bool, page, limit int) []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*Task, 0)

	for _, t := range r.items {
		if done != nil && t.Done != *done {
			continue
		}
		items = append(items, t)
	}

	if limit <= 0 || page <= 0 {
		return items
	}

	start := (page - 1) * limit
	if start >= len(items) {
		return []*Task{}
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}

func (r *Repo) Load() {
	file, err := os.Open(dataFile)
	if err != nil {
		return
	}
	defer file.Close()

	_ = json.NewDecoder(file).Decode(&r.items)
}

func (r *Repo) save() {
	file, err := os.Create(dataFile)
	if err != nil {
		return
	}
	defer file.Close()

	_ = json.NewEncoder(file).Encode(r.items)
}
