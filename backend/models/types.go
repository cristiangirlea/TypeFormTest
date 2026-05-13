package models

import "time"

type QuestionType string

const (
	QuestionTypeText QuestionType = "text"
)

type Question struct {
	ID   string       `json:"id"`
	Text string       `json:"text"`
	Type QuestionType `json:"type"`
}

type Form struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions,omitzero"`
	ShareSlug string     `json:"shareSlug,omitzero"`
	CreatedAt time.Time  `json:"createdAt,omitzero"`
}

type Answer struct {
	QuestionID string `json:"questionID"`
	Value      string `json:"value"`
}

type Response struct {
	ID        string    `json:"id"`
	FormID    string    `json:"formID"`
	Answers   []Answer  `json:"answers,omitzero"`
	CreatedAt time.Time `json:"createdAt,omitzero"`
}
