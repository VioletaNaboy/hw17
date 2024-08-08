// main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func main() {
	mux := http.NewServeMux()

	cache := NewCache()

	tasks := TaskResource{
		cache: cache,
	}

	mux.HandleFunc("GET /tasks", tasks.GetAll)
	mux.HandleFunc("POST /tasks/add", tasks.CreateOne)
	mux.HandleFunc("PUT /tasks/update", tasks.UpdateOne)
	mux.HandleFunc("DELETE /tasks/delete", tasks.DeleteOne)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Failed to listen and serve: %v\n", err)
	}
}

type TaskResource struct {
	cache *Cache
}

func (t *TaskResource) GetAll(w http.ResponseWriter, r *http.Request) {
	cacheKey := "tasks_all"
	var tasks []Task

	err := t.cache.Get(cacheKey, &tasks)
	if err == nil {
		json.NewEncoder(w).Encode(tasks)
		return
	}

	// Redis cache miss; return empty list
	tasks = []Task{}

	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
		return
	}

	t.cache.Set(cacheKey, tasks, 10*time.Minute)
}

func (t *TaskResource) CreateOne(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Request body is missing", http.StatusBadRequest)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Failed to decode task", http.StatusBadRequest)
		return
	}

	tasksKey := "tasks_all"
	var tasks []Task
	err = t.cache.Get(tasksKey, &tasks)
	if err == nil {
		task.ID = len(tasks) + 1
		tasks = append(tasks, task)
		t.cache.Set(tasksKey, tasks, 10*time.Minute)
	} else {
		task.ID = 1
		tasks = []Task{task}
		t.cache.Set(tasksKey, tasks, 10*time.Minute)
	}

	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}

func (t *TaskResource) UpdateOne(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Request body is missing", http.StatusBadRequest)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Failed to decode task", http.StatusBadRequest)
		return
	}

	tasksKey := "tasks_all"
	var tasks []Task
	err = t.cache.Get(tasksKey, &tasks)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	updated := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			updated = true
			break
		}
	}

	if !updated {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	t.cache.Set(tasksKey, tasks, 10*time.Minute)

	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}

func (t *TaskResource) DeleteOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tasksKey := "tasks_all"
	var tasks []Task
	err = t.cache.Get(tasksKey, &tasks)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	updated := false
	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			updated = true
			break
		}
	}

	if !updated {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	t.cache.Set(tasksKey, tasks, 10*time.Minute)

	w.WriteHeader(http.StatusNoContent)
}
