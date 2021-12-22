package codes

// Success200 OK
// swagger:model
type Success200 struct {
	// Code text
	Status string `json:"Status" example:"OK"`
}

// Success201 Created
// swagger:model
type Success201 struct {
	// Code text
	Status string `json:"Status" example:"Created"`
}

// Error500 Internal Server Error
// swagger:model
type Error500 struct {
	// Error text
	Error string `json:"Error" example:"Internal Server Error"`
}

// Error502 Bad Gateway
// swagger:model
type Error502 struct {
	// Error text
	Error string `json:"Error" example:"Bad Gateway"`
}

// Error503 Service Unavailable
// swagger:model
type Error503 struct {
	// Error text
	Error string `json:"Error" example:"Service Unavailable"`
}

// Error400 Internal Server Error
// swagger:model
type Error400 struct {
	// Error text
	Error string `json:"Error" example:"Internal Server Error"`
}

// Error401 Unauthorized
// swagger:model
type Error401 struct {
	// Error text
	Error string `json:"Error" example:"Unauthorized"`
}

// Error403 Forbidden
// swagger:model
type Error403 struct {
	// Error text
	Error string `json:"Error" example:"Forbidden"`
}

// Error409 Conflict
// swagger:model
type Error409 struct {
	// Error text
	Error string `json:"Error" example:"Conflict"`
}

// Error424 Failed Dependency
// swagger:model
type Error424 struct {
	// Error text
	Error string `json:"Error" example:"Failed Dependency"`
}

// Error422 Unprocessable Entity
// swagger:model
type Error422 struct {
	// Error text
	Error string `json:"Error" example:"Unprocessable Entity"`
}
