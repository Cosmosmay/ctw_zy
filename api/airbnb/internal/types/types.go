// Code generated by goctl. DO NOT EDIT.
package types

type Request struct {
	Name string `path:"name,options=you|me"`
}

type Response struct {
	Message string `json:"message"`
}

type Task struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}