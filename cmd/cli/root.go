package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/janeczku/go-spinner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/walles/env"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	apply     = "Apply"
	dontApply = "Don't Apply"
	reprompt  = "Reprompt"
)

var (
	openaiAPIURLv1        = "https://api.openai.com/v1"
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	openAIDeploymentName  = flag.String("openai-deployment-name", env.GetOr("OPENAI_DEPLOYMENT_NAME", env.String, "gpt-3.5-turbo-0301"), "This deployment name used")
	openAIEndpoint        = flag.String("openai-endpoint", env.GetOr("OPENAI_ENDPOINT", env.String, openaiAPIURLv1), "The endpoint for OpenAI service. Defaults to"+openaiAPIURLv1+". Set this to your Local AI endpoint or Azure OpenAI Service, if needed.")
	version               = "dev"
	azureModelMap         = flag.StringTotString("azure-openai-map", env.GetOr("AZURE_OPENAI_MAP", env.Map(env.String, "=", env.String, ""), map[string]string{}), "This is azure ")
	openAIAPIKey          = flag.String("openai-api-key", env.GetOr("OPENAI_API_KEY", env.String, ""), "This is required")
	debug                 = flag.Bool("debug", env.GetOr("DEBUG", strconv.ParseBool, false), "whether to print debug logs. Defaults to false")
	raw                  = flag.Bool("raw", false, "Prints the raw YAML output immediately. Defaults to false.")
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
	kubernetesConfigFlags.AddFlags(cmd.PersistentFlags())
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
	var action, completion string

	for action != apply {
		args = append(args, action)
		s := spinner.NewSpinner("Processing...")
		if !*debug && !*raw {
			s.SetCharset([]string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"})
			s.Start()
		}
		completion, err = gptCompletion(ctx, oaiClients, args, *openAIDeploymentName)
		if err != nil {
			return err
		}
		s.Stop()
		if *raw {
			fmt.Println(completion)
			return nil
		}
		text := fmt.Sprintln("Attempting to apply the following manifest:\n%s", completion)
		fmt.Println(text)
		action, err := userActionPrompt()
		if err != nil {
			return err
		}
		if action == dontApply {
			return nil
		}
	}
	return applyManifest(completion)
}

func userActionPrompt()(string,error)