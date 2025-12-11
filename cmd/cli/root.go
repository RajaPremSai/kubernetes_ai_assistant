package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	openaiAPIURLv1        = "https://api.openai.com/v1"
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	openAIDeploymentName  = flag.String("openai-deployment-name", env.GetOr("OPENAI_DEPLOYMENT_NAME", env.String, "gpt-3.5-turbo-0301"), "This deployment name used")
	version               = "dev"
	openAIAPIKey          = flag.String("openai-api-key", env.GetOr("OPENAI_API_KEY", env.String, ""), "This is required")
	debug                 = flag.Bool("debug", env.GetOr("DEBUG", strconv.ParseBool, false), "whether to print debug logs. Defaults to false")
)

func InitAndExecute() {
	if *openAIAPIKey == "" {
		fmt.Println("Please provide an OpenAI key.")
		os.Exit(1)
	}
	if err := RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubernetes_ai_assistant",
		Short:        "kubernetes_ai_assistant",
		Long:         "kubernetes_ai_assistant is a plugin for kubectl",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if *debug {
				log.SetLevel(log.DebugLevel)
				printDebugFlags()
			}
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("prompt must be provided")
			}

			err := run(args)
			if err != nil {
				return err
			}
			return nil
		},
	}
	kubernetesConfigFlags.AddFlags(cmd.PersitentFlags())
	return cmd
}

func printDebugFlags() {
	log.Debugf("openai-endpoint:%s", *openAIEndpoint)
	log.Debugf("openai-deployment-name:%s", *openAIDeploymentName)
	log.Debugf("azure-openai-map: %s", *azureModelMap)
	log.Debugf("temperature:%f", *temperature)
	log.Debugf("use-k8s-api:%t", *usek8sAPI)
	log.Debugf("k8s-openapi-url:%s", *k8sOpenAPIURL)
}

func run(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	oaiClients, err := newOAIClients()
	if err != nil {
		return err
	}
}
