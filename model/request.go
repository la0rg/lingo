package model

type APKGRequest struct {
	Words []Word `json:"words"`
}

type APKGResponse struct {
	URL string `json:"url"`
}
