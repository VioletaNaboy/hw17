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

	tasksMemoryStore := make([]Task, 0)

	tasks := TaskResource{
		cache:       cache,
		memoryStore: &tasksMemoryStore,
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
	cache       *Cache
	memoryStore *[]Task
}

func (t *TaskResource) GetAll(w http.ResponseWriter, r *http.Request) {
	cacheKey := "tasks_all"
	var tasks []Task

	err := t.cache.Get(cacheKey, &tasks)
	if err == nil {
		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
		}
		fmt.Println("Retrieved tasks from cache.")
		return
	} else {
		fmt.Printf("Cache miss or error: %v\n", err)
	}

	tasks = *t.memoryStore

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
		return
	}

	if err := t.cache.Set(cacheKey, tasks, 10*time.Minute); err != nil {
		fmt.Printf("Failed to set cache: %v\n", err)
	}
	fmt.Println("Retrieved tasks from memory store and updated cache.")
}

func (t *TaskResource) CreateOne(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Request body is missing", http.StatusBadRequest)
		return
	}

	var task Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, "Failed to decode task", http.StatusBadRequest)
		return
	}

	tasks := *t.memoryStore
	task.ID = len(tasks) + 1
	tasks = append(tasks, task)
	*t.memoryStore = tasks

	fmt.Printf("Task created: %+v\n", task)

	if err := t.cache.Set("tasks_all", tasks, 10*time.Minute); err != nil {
		fmt.Printf("Failed to set cache: %v\n", err)
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
	fmt.Println("Task added to memory store and cache updated.")
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

	tasks := *t.memoryStore
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

	*t.memoryStore = tasks

	fmt.Printf("Task updated: %+v\n", task)

	if err := t.cache.Set("tasks_all", tasks, 10*time.Minute); err != nil {
		fmt.Printf("Failed to set cache: %v\n", err)
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
	fmt.Println("Task updated in memory store and cache updated.")
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

	tasks := *t.memoryStore
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

	*t.memoryStore = tasks

	fmt.Printf("Task deleted with ID: %d\n", id)

	if err := t.cache.Set("tasks_all", tasks, 10*time.Minute); err != nil {
		fmt.Printf("Failed to set cache: %v\n", err)
	}

	w.WriteHeader(http.StatusNoContent)
	fmt.Println("Task deleted from memory store and cache updated.")
}
