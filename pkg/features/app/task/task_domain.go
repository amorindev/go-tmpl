package task

type Task struct {
    ID int `json:"id"`
    Desc string `json:"desc"`
    IsCompleted bool `json:"is_completed"`
}