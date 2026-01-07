package task

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type ErrorMessage struct {
	Msg string `json:"msg"`
}

var allTasks []Task

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(allTasks)
}

func PostTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorMessage{Msg: "Failed to read request body"})
		return
	}

	var task Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorMessage{Msg: "Failed to parse request body"})
		return
	}

	if task.Desc == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorMessage{Msg: "desc should not be empty"})
		return
	}

	if len(allTasks) == 0 {
		task.ID = 1
	} else {
		longitud := len(allTasks)
		task.ID = allTasks[longitud-1].ID + 1
	}

	allTasks = append(allTasks, task)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nil)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")

	taskID, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorMessage{Msg: "Invalid id"})
		return
	}

	for i, t := range allTasks {
		if t.ID == taskID {
			allTasks = append(allTasks[:i], allTasks[i+1:]...)
		}
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(nil)
}
