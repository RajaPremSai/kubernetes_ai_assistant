package cli

import (
	"context"
	"strings"
)

func (c *oaiClients) openaiGptCompletion(ctx context.Context, prompt *strings.Builder, temp float32) (string, error)

func (c *oaiClients) openaiGptChatCompletion(ctx context.Context, prompt *strings.Builder, temp float32) (string, error)
