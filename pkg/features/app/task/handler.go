package task

import (
	"net/http"

	"github.com/amorindev/go-tmpl/pkg/shared/api/middlewares"
)

func NewTaskHandler(server *http.ServeMux, authMdw *middlewares.AuthMiddleware) {
	server.HandleFunc("GET /tasks", authMdw.AccessTokenMdw(GetTasks))
	server.HandleFunc("POST /tasks", authMdw.AccessTokenMdw(PostTask))
	server.HandleFunc("DELETE /tasks/{id}", authMdw.AccessTokenMdw(DeleteTask))
}
