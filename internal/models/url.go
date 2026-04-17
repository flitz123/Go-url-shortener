package models

type URL struct {
	Code     string `json:"code"`
	Original string `json:"original"`
	Clicks   int    `json:"clicks"`
}
