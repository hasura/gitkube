package providers

type Provider interface {
	BuildSpecFromPayload(payload interface{}) (PushSpec, error)
}

type PushSpec struct {
	Ref    string
	GitUrl string
}
