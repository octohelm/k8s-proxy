package secureproxy

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func ResolveKubeConfig() (*rest.Config, error) {
	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		clientConfig, err = localConfig()
		if err != nil {
			return nil, err
		}
	}
	return clientConfig, nil
}

func localConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	apiConfig, err := rules.Load()
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
}
