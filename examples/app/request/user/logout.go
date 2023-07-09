package request

type LogoutRequest struct {
	Authorization string `json:"authorization" bind:"header" process:"trim,cut(32)"`
}

type LogoutResponse struct {
}

func (r *LogoutRequest) Verify() error {
	return nil
}

func (r *LogoutRequest) Clear() {
}
