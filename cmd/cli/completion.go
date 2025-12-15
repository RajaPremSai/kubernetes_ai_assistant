package cli

import (
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type oaiClients struct {
	openAIClient openai.Client
}

func newOAIClients() (oaiClients, error) {
	var config openai.ClientConfig
	config = openai.DefaultConfig(*openAIAPIKey)
	if openAIEndpoint != &openaiAPIURLv1 {
		if strings.Contains(*openAIEndpoint, "openai.azure.com") {
			config = openai.DefaultAzureConfig(*openAIAPIKey, *openAIEndpoint)
			if len(*azureModelMap) != 0 {
				config.AzureModelMapperFunc = func(model string) string {
					return (*azureModelMap)[model]
				}
			}
		}
	} else {
		config.BaseURL = *openAIEndpoint
	}
	config.APIVersion = "2023-07-01-preview"
	clients := oaiClients{
		openAIClient: *openai.NewClientWithConfig(config),
	}
	return clients, nil
}
