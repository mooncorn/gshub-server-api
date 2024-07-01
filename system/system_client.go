package system

type SystemController interface {
	GetInstanceID() (string, error)
}

type SystemClient struct {
	controller SystemController
}

func NewSystemClient() *SystemClient {
	// fetch instance id

	return &SystemClient{}
}
