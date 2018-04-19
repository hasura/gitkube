package providers

import (
	"fmt"
)

func GetProvider(providerName string) (Provider, error) {
	switch providerName {
	case "github":
		return &GithubProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported git provider: %s", providerName)
	}
}
