package request

type LoginRequest struct {
	Username string `json:"username" bind:"form"`
	Password string `json:"password" bind:"form"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (r *LoginRequest) Verify() error {
	return nil
}

func (r *LoginRequest) Clear() {
}
