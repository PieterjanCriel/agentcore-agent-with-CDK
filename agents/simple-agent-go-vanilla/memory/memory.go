package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcore"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcore/types"
)

// Client wraps the Bedrock AgentCore client
type Client struct {
	client *bedrockagentcore.Client
}

// NewClient creates a new AgentCore client for eu-central-1
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	c := bedrockagentcore.NewFromConfig(cfg)
	return &Client{client: c}, nil
}

// NewClientWithRegion lets you use a custom region
func NewClientWithRegion(ctx context.Context, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	c := bedrockagentcore.NewFromConfig(cfg)
	return &Client{client: c}, nil
}

// ──────────────────────────────────────────────────────────────
// High-level helpers
// ──────────────────────────────────────────────────────────────

// CreateEvent pushes a conversational message into an AgentCore memory
func (c *Client) CreateEvent(
	ctx context.Context,
	memoryID string,
	actorID string,
	sessionID string,
	role string,
	text string,
	metadata map[string]string,
) (*types.Event, error) {

	var sdkRole types.Role
	switch role {
	case "user":
		sdkRole = types.RoleUser
	case "assistant":
		sdkRole = types.RoleAssistant
	default:
		return nil, fmt.Errorf("unsupported role: %s", role)
	}

	// metadata → AWS format
	var md map[string]types.MetadataValue
	if len(metadata) > 0 {
		md = make(map[string]types.MetadataValue, len(metadata))
		for k, v := range metadata {
			md[k] = &types.MetadataValueMemberStringValue{Value: v}
		}
	}

	input := &bedrockagentcore.CreateEventInput{
		MemoryId:       aws.String(memoryID),
		ActorId:        aws.String(actorID),
		SessionId:      aws.String(sessionID),
		EventTimestamp: aws.Time(time.Now().UTC()),
		Metadata:       md,
		Payload: []types.PayloadType{
			&types.PayloadTypeMemberConversational{
				Value: types.Conversational{
					Role:    sdkRole,
					Content: &types.ContentMemberText{Value: text},
				},
			},
		},
	}

	out, err := c.client.CreateEvent(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("CreateEvent failed: %w", err)
	}

	return out.Event, nil
}

// ListEvents returns all events for a given memory/actor/session
func (c *Client) ListEvents(
	ctx context.Context,
	memoryID string,
	actorID string,
	sessionID string,
	includePayloads bool,
) ([]types.Event, error) {

	input := &bedrockagentcore.ListEventsInput{
		MemoryId:        aws.String(memoryID),
		ActorId:         aws.String(actorID),
		SessionId:       aws.String(sessionID),
		IncludePayloads: aws.Bool(includePayloads),
	}

	p := bedrockagentcore.NewListEventsPaginator(c.client, input)

	var events []types.Event
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListEvents failed: %w", err)
		}
		events = append(events, page.Events...)
	}

	return events, nil
}

// GetEvent retrieves a single event
func (c *Client) GetEvent(ctx context.Context, memoryID, eventID string) (*types.Event, error) {
	input := &bedrockagentcore.GetEventInput{
		EventId:  aws.String(eventID),
		MemoryId: aws.String(memoryID),
	}

	out, err := c.client.GetEvent(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("GetEvent failed: %w", err)
	}

	return out.Event, nil
}

// GetMemoryRecordsText fetches up to 5 memory records for a given
// memory strategy, memory ID, and actor ID, and returns the textual
// content joined into a single string.
// It mirrors the Python helper:
//
//	path = f"/strategies/{memory_strategy_id}/actors/{actor_id}"
//	list_memory_records(..., namespace=path, max_results=5)
//
// Only text memory content is included in the returned string.
func (c *Client) GetMemoryRecordsText(
	ctx context.Context,
	memoryID string,
	actorID string,
	memoryStrategyID string,
) (string, error) {

	// Same namespace convention as your Python version
	namespace := fmt.Sprintf("/strategies/%s/actors/%s", memoryStrategyID, actorID)

	input := &bedrockagentcore.ListMemoryRecordsInput{
		MemoryId:         aws.String(memoryID),
		Namespace:        aws.String(namespace),
		MemoryStrategyId: aws.String(memoryStrategyID),
		MaxResults:       aws.Int32(5), // default to 5
	}

	out, err := c.client.ListMemoryRecords(ctx, input)
	if err != nil {
		return "", fmt.Errorf("ListMemoryRecords failed: %w", err)
	}

	var b strings.Builder

	for _, rec := range out.MemoryRecordSummaries {
		switch v := rec.Content.(type) {
		case *types.MemoryContentMemberText:
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(v.Value)
		default:
			// Ignore non-text content types for now
		}
	}

	return b.String(), nil
}
