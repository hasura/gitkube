/*
Copyright 2018 Hasura.io

*/

package fake

import (
	v1alpha1 "github.com/hasura/gitkube/pkg/client/clientset/versioned/typed/gitkube/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeGitkubeV1alpha1 struct {
	*testing.Fake
}

func (c *FakeGitkubeV1alpha1) Remotes(namespace string) v1alpha1.RemoteInterface {
	return &FakeRemotes{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeGitkubeV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
