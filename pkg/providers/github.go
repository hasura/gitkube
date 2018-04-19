package providers

import (
	"fmt"
)

type GithubProvider struct {
}

type GithubRepository struct {
	GitUrl string `json:"git_url",omitempty`
}

type GithubPushEvent struct {
	Ref        string           `json:"ref",omitempty`
	Repository GithubRepository `json:"repository",omitempty`
}

func (gh *GithubProvider) BuildSpecFromPayload(payload interface{}) (PushSpec, error) {

	gpe, ok := payload.(GithubPushEvent)
	if !ok {
		return PushSpec{}, fmt.Errorf("unable to parse github event")
	}

	pushSpec := PushSpec{
		Ref:    gpe.Ref,
		GitUrl: gpe.Repository.GitUrl,
	}

	return pushSpec, nil
}
