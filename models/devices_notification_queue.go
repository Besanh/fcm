package models

type NotificationTask struct {
	Id      string         `json:"id"`
	UserId  string         `json:"user_id"`
	Title   string         `json:"title"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}
