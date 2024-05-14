package model

type APKGRequest struct {
	Words []string `json:"words"`
}

type APKGResponse struct {
	URL string `json:"url"`
}
