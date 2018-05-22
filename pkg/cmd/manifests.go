package cmd

import (
	api "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var crd = apiextensionsv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "remotes.gitkube.sh",
	},
	Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
		Group:   api.SchemeGroupVersion.Group,
		Version: api.SchemeGroupVersion.Version,
		Scope:   apiextensionsv1beta1.NamespaceScoped,
		Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
			Plural:     "remotes",
			Singular:   "remote",
			Kind:       "Remote",
			ShortNames: []string{"rem"},
		},
	},
}
