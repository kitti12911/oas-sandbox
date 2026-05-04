package system

type HealthOutput struct {
	Body struct {
		Status  string `json:"status"  example:"ok"          doc:"Health status"`
		Service string `json:"service" example:"oas-sandbox" doc:"Service name"`
	}
}
