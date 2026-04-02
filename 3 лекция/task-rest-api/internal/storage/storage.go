package storage

import (
	"sync"

	"task-rest-api/internal/models"
)

type Storage struct {
	tasks      map[string]models.Task
	tasksMutex sync.RWMutex
}

func New() *Storage {
	return &Storage{
		tasks: make(map[string]models.Task),
	}
}

func (s *Storage) List(completedFilter string) []models.Task {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	result := make([]models.Task, 0)
	for _, task := range s.tasks {
		if completedFilter != "" {
			filter := completedFilter == "true"
			if task.Completed != filter {
				continue
			}
		}
		result = append(result, task)
	}

	return result
}

func (s *Storage) Create(task models.Task) models.Task {
	s.tasksMutex.Lock()
	s.tasks[task.ID] = task
	s.tasksMutex.Unlock()

	return task
}

func (s *Storage) Get(id string) (models.Task, bool) {
	s.tasksMutex.RLock()
	task, exists := s.tasks[id]
	s.tasksMutex.RUnlock()

	return task, exists
}

func (s *Storage) Update(id string, task models.Task) (models.Task, bool) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return models.Task{}, false
	}

	task.ID = id
	s.tasks[id] = task

	return task, true
}

func (s *Storage) Delete(id string) bool {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return false
	}

	delete(s.tasks, id)
	return true
}

func (s *Storage) DeleteAll() {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	s.tasks = make(map[string]models.Task)
}