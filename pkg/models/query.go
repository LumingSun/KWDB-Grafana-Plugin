package models

// QueryModel is the backend representation of the frontend query JSON.
type QueryModel struct {
	Mode       string   `json:"mode"`
	Format     string   `json:"format"`
	RawSql     string   `json:"rawSql"`
	TimeColumn string   `json:"timeColumn"`
	Tags       []string `json:"tags"`
}
