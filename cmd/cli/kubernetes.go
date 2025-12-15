package cli

import "k8s.io/client-go/tools/clientcmd/api"

func applyManifest(completion string) error {

}

func getKubeConfig() string {

}

func getConfig(kubeConfig string) (api.Config, error) {

}

func getCurrentContextName() (string, error) {
	kubeConfig := getKubeConfig()
	config, err := getConfig(kubeConfig)
	if err != nil {
		return "", err
	}
	currentContext := config.currentContext
	return currentContext, nil
}
