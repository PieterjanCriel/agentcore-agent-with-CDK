package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Client wraps the Bedrock Runtime client
type Client struct {
	client *bedrockruntime.Client
}

// NewClient creates a new Bedrock client for eu-central-1 region
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return &Client{client: client}, nil
}

// NewClientWithRegion creates a new Bedrock client for a custom region
func NewClientWithRegion(ctx context.Context, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return &Client{client: client}, nil
}

// Converse makes a Converse API call to AWS Bedrock
func (c *Client) Converse(ctx context.Context, modelID string, userMessage string) (*bedrockruntime.ConverseOutput, error) {
	messages := []types.Message{
		{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: userMessage,
				},
			},
		},
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  &modelID,
		Messages: messages,
	}

	output, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("converse API call failed: %w", err)
	}

	return output, nil
}

// ConverseWithSystem makes a Converse API call with a system prompt
func (c *Client) ConverseWithSystem(ctx context.Context, modelID string, systemPrompt string, userMessage string) (*bedrockruntime.ConverseOutput, error) {
	messages := []types.Message{
		{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: userMessage,
				},
			},
		},
	}

	systemMessages := []types.SystemContentBlock{
		&types.SystemContentBlockMemberText{
			Value: systemPrompt,
		},
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  &modelID,
		Messages: messages,
		System:   systemMessages,
	}

	output, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("converse API call failed: %w", err)
	}

	return output, nil
}

// ConverseWithMessages makes a Converse API call with multiple messages (for multi-turn conversations)
// systemPrompt is optional - pass empty string to omit
func (c *Client) ConverseWithMessages(ctx context.Context, modelID string, messages []types.Message, systemPrompt string) (*bedrockruntime.ConverseOutput, error) {
	input := &bedrockruntime.ConverseInput{
		ModelId:  &modelID,
		Messages: messages,
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		input.System = []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{
				Value: systemPrompt,
			},
		}
	}

	output, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("converse API call failed: %w", err)
	}

	return output, nil
}

// ConverseWithConfig makes a Converse API call with full configuration options
func (c *Client) ConverseWithConfig(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	output, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("converse API call failed: %w", err)
	}

	return output, nil
}

// ExtractTextResponse is a helper function to extract text from a Converse response
func ExtractTextResponse(output *bedrockruntime.ConverseOutput) (string, error) {
	if output == nil || output.Output == nil {
		return "", fmt.Errorf("output is nil")
	}

	message, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected output type")
	}

	if len(message.Value.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	textBlock, ok := message.Value.Content[0].(*types.ContentBlockMemberText)
	if !ok {
		return "", fmt.Errorf("first content block is not text")
	}

	return textBlock.Value, nil
}
