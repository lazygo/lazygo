package request

type ProfileRequest struct {
	UID uint64 `json:"uid" bind:"ctx"`
}

type ProfileResponse struct {
}

func (r *ProfileRequest) Verify() error {
	return nil
}

func (r *ProfileRequest) Clear() {
}
