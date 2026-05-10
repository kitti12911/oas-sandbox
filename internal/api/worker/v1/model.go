package workerv1

type SubmitJobInput struct {
	Body SubmitJobRequest
}

type SubmitJobRequest struct {
	ID      string         `json:"id"                example:"job-1"       doc:"Job ID"`
	Type    string         `json:"type"              example:"debug.print" doc:"Job type"`
	Payload map[string]any `json:"payload,omitempty"                      doc:"Job payload"`
}

type SubmitJobOutput struct {
	Body SubmitJobResult
}

type SubmitJobResult struct {
	ID string `json:"id" example:"job-1" doc:"Submitted job ID"`
}
