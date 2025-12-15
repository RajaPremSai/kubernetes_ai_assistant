package cli

import (
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

const defaultNamespace = "default"

func applyManifest(completion string) error {
	
}

func getKubeConfig() string {
	var kubeConfig string
	if *kubernetesConfigFlags.KubeConfig == "" {
		kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	} else {
		kubeConfig = *kubernetesConfigFlags.KubeConfig
	}
	return kubeConfig

}

func getConfig(kubeConfig string) (api.Config, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return api.Config{}, err
	}
	return config, nil
}

func getCurrentContextName() (string, error) {
	kubeConfig := getKubeConfig()
	config, err := getConfig(kubeConfig)
	if err != nil {
		return "", err
	}
	currentContext := config.CurrentContext
	return currentContext, nil
}
