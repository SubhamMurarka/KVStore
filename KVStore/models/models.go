package models

type Request struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int64  `json:"ttl"`
}

type UpdateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
