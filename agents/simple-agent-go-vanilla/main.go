package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"simple-agent/bedrock"
	"simple-agent/memory"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcore/types"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/google/uuid"
)

// s returns a safe string for *string values.
func s(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// j prints any Go value as pretty JSON (best-effort).
func j(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v) // fallback
	}
	return string(b)
}

// convertEventsToMessages converts AgentCore events to Bedrock messages
// Takes the last maxMessages events and converts them to alternating user/assistant messages
// Ensures the conversation starts with a user message and messages properly alternate
func convertEventsToMessages(events []types.Event, maxMessages int) ([]brtypes.Message, error) {
	var messages []brtypes.Message

	// Events come in reverse chronological order (newest first), so we need to reverse them
	// to get chronological order (oldest first)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]

		// Extract conversational payload
		if len(event.Payload) == 0 {
			continue
		}

		for _, payload := range event.Payload {
			conversational, ok := payload.(*types.PayloadTypeMemberConversational)
			if !ok {
				continue
			}

			// Extract text content
			textContent, ok := conversational.Value.Content.(*types.ContentMemberText)
			if !ok {
				continue
			}

			// Convert role
			var role brtypes.ConversationRole
			switch conversational.Value.Role {
			case types.RoleUser:
				role = brtypes.ConversationRoleUser
			case types.RoleAssistant:
				role = brtypes.ConversationRoleAssistant
			default:
				continue
			}

			// Create message
			messages = append(messages, brtypes.Message{
				Role: role,
				Content: []brtypes.ContentBlock{
					&brtypes.ContentBlockMemberText{
						Value: textContent.Value,
					},
				},
			})
		}
	}

	// Take only the last maxMessages
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}

	// Ensure conversation starts with a user message
	// If it starts with assistant, remove messages until we find a user message
	for len(messages) > 0 && messages[0].Role != brtypes.ConversationRoleUser {
		messages = messages[1:]
	}

	// Ensure proper alternation: user -> assistant -> user -> assistant
	// Remove any consecutive messages with the same role
	if len(messages) > 1 {
		filtered := []brtypes.Message{messages[0]}
		for i := 1; i < len(messages); i++ {
			// Only add if role is different from previous
			if messages[i].Role != filtered[len(filtered)-1].Role {
				filtered = append(filtered, messages[i])
			}
		}
		messages = filtered
	}

	return messages, nil
}

func main() {
	ctx := context.Background()

	// Create the Bedrock client
	client, err := bedrock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Memory client
	ac, err := memory.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create memory client: %v", err)
	}

	memoryID := os.Getenv("MEMORY_ID")
	if memoryID == "" {
		log.Fatal("MEMORY_ID environment variable is required")
	}

	// Get strategy name and construct the full strategy ID
	// Memory ID format: memory_<suffix> (e.g., memory_itp2x-N5PjG7Adx9)
	// Strategy ID format: <strategy_name>_<suffix> (e.g., preference_builtin_itp2x-N5PjG7Adx9)
	memoryStrategyName := os.Getenv("MEMORY_USER_PREFERENCES_STRATEGY_NAME")
	if memoryStrategyName == "" {
		log.Fatal("MEMORY_USER_PREFERENCES_STRATEGY_NAME environment variable is required")
	}

	// Extract suffix from memory ID (everything after "memory_")
	memorySuffix := strings.TrimPrefix(memoryID, "memory_")
	memoryStrategyID := memoryStrategyName + "_" + memorySuffix

	modelID := "eu.amazon.nova-pro-v1:0"

	systemPrompt := "You are a friendly AI assistant. You are here to help users with there questions. You rely on your knowledge of the world to help them to the best of your abilities."

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	http.HandleFunc("/invocations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}
		prompt := body["prompt"]

		// Extract ActorId from the request
		actorID := body["actorId"]
		if actorID == "" {
			http.Error(w, "actorId is required", http.StatusBadRequest)
			return
		}

		// Extract SessionId from the request (nullable)
		sessionID := body["sessionId"]
		isNewSession := false
		if sessionID == "" {
			// Generate a new UUID if no session is provided
			sessionID = uuid.New().String()
			isNewSession = true
		}

		// for actor and sessionId fetch the memory records for the preference strategy
		preferenceMemory, err := ac.GetMemoryRecordsText(ctx, memoryID, actorID, memoryStrategyID)

		if err != nil {
			log.Printf("Warning: failed to load preference memory: %v", err)
			// Continue with empty history
			preferenceMemory = ""
		}

		log.Printf("Preference Memory: %s", preferenceMemory)

		// Build the system prompt with preferences
		enhancedSystemPrompt := systemPrompt
		if preferenceMemory != "" {
			enhancedSystemPrompt = systemPrompt + "\n\nUser Preferences:\n" + preferenceMemory
		}

		// Load the last 30 messages from memory
		events, err := ac.ListEvents(ctx, memoryID, actorID, sessionID, true)
		if err != nil {
			log.Printf("Warning: failed to load events from memory: %v", err)
			// Continue with empty history
			events = nil
		}

		// Check if this is a new session (no events in memory)
		if len(events) == 0 {
			isNewSession = true
		}

		// Convert events to Bedrock messages (last 30 messages)
		messages, err := convertEventsToMessages(events, 30)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert events: %v", err), http.StatusInternalServerError)
			return
		}

		// Log session info and conversation history
		log.Printf("=== Request for Actor: %s, Session: %s ===", actorID, sessionID)
		if isNewSession {
			log.Printf("🆕 NEW SESSION STARTED")
		} else {
			log.Printf("📜 Continuing existing session with %d messages in history", len(messages))
		}

		// Print conversation history in chronological order
		if len(messages) > 0 {
			log.Printf("--- Conversation History ---")
			for i, msg := range messages {
				role := "USER"
				if msg.Role == brtypes.ConversationRoleAssistant {
					role = "ASSISTANT"
				}
				// Extract text from message
				text := ""
				for _, content := range msg.Content {
					if textBlock, ok := content.(*brtypes.ContentBlockMemberText); ok {
						text = textBlock.Value
						break
					}
				}
				// Truncate long messages for readability
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				log.Printf("[%d] %s: %s", i+1, role, text)
			}
			log.Printf("----------------------------")
		}

		// Add the current user message
		log.Printf("USER: %s", prompt)

		// If the last message in history is from the user, we need to remove it
		// to avoid consecutive user messages (which Bedrock doesn't allow)
		if len(messages) > 0 && messages[len(messages)-1].Role == brtypes.ConversationRoleUser {
			messages = messages[:len(messages)-1]
		}

		messages = append(messages, brtypes.Message{
			Role: brtypes.ConversationRoleUser,
			Content: []brtypes.ContentBlock{
				&brtypes.ContentBlockMemberText{
					Value: prompt,
				},
			},
		})

		// Call Converse with message history and enhanced system prompt (including preferences)
		output, err := client.ConverseWithMessages(ctx, modelID, messages, enhancedSystemPrompt)
		if err != nil {
			// don't kill the server; return an error to the client
			http.Error(w, fmt.Sprintf("converse error: %v", err), http.StatusBadGateway)
			return
		}

		response, err := bedrock.ExtractTextResponse(output)
		if err != nil {
			http.Error(w, fmt.Sprintf("extract response error: %v", err), http.StatusBadGateway)
			return
		}

		// Log the assistant response
		responsePreview := response
		if len(responsePreview) > 100 {
			responsePreview = responsePreview[:100] + "..."
		}
		log.Printf("ASSISTANT: %s", responsePreview)

		// Save the user message to memory
		_, err = ac.CreateEvent(ctx, memoryID, actorID, sessionID, "user", prompt, nil)
		if err != nil {
			log.Printf("Warning: failed to save user message to memory: %v", err)
		}

		// Save the assistant response to memory
		_, err = ac.CreateEvent(ctx, memoryID, actorID, sessionID, "assistant", response, nil)
		if err != nil {
			log.Printf("Warning: failed to save assistant response to memory: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result":    response,
			"actorId":   actorID,
			"sessionId": sessionID,
		})
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
