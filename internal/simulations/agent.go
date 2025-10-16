package simulations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/prompts"
	"github.com/poiesic/wonda/internal/scenarios"
)

// ToolExecutor interface for executing tool calls during agent reasoning.
type ToolExecutor interface {
	ExecuteTool(ctx context.Context, toolCall *mcp.ToolCall) *mcp.ToolResult
}

// AgentState represents the runtime state of an agent during simulation.
type AgentState struct {
	// Physical state
	Position  string
	Condition int // 0-100

	// Emotional state
	Emotion          string
	EmotionIntensity int // 0-10
}

// Agent represents an active participant in a simulation.
// It binds a character definition with an LLM client and maintains runtime state.
type Agent struct {
	// Identity
	Name      string
	Character *scenarios.Character

	// LLM Interface
	Client Client

	// Runtime State
	State AgentState

	// Configuration
	Model    string
	Provider string
}

// NewAgent creates a new agent from a character definition and LLM client.
func NewAgent(name string, character *scenarios.Character, client Client, provider, model string) *Agent {
	return &Agent{
		Name:      name,
		Character: character,
		Client:    client,
		Provider:  provider,
		Model:     model,
		State: AgentState{
			Position:         "unknown",
			Condition:        100,
			Emotion:          "neutral",
			EmotionIntensity: 5,
		},
	}
}

// ApplyInitialState updates the agent's state from scenario initial state overrides.
func (a *Agent) ApplyInitialState(initial *scenarios.InitialState) {
	if initial == nil {
		return
	}

	if initial.Position != "" {
		a.State.Position = initial.Position
	}
	if initial.Condition > 0 {
		a.State.Condition = initial.Condition
	}
	if initial.Emotion != "" {
		a.State.Emotion = initial.Emotion
	}
	if initial.EmotionIntensity > 0 {
		a.State.EmotionIntensity = initial.EmotionIntensity
	}
}

// SceneContext contains scene information to be included in prompts.
type SceneContext struct {
	Location   string
	Time       string
	Atmosphere string
	Backstory  string
}

// Think sends a prompt to the agent's LLM and returns the response.
// It includes the character personality, current state, and available tools.
// The agent discovers goals and world state through MCP tools.
// This method handles the tool execution loop internally - if the LLM requests tool calls,
// they are executed and the results are sent back to the LLM until a final response is obtained.
func (a *Agent) Think(ctx context.Context, situation string, sceneCtx *SceneContext, tools []map[string]interface{}, executor ToolExecutor) (ChatResponse, error) {
	if a.Client == nil {
		return ChatResponse{}, fmt.Errorf("agent %s has no LLM client", a.Name)
	}

	// Build the initial prompt using template
	systemPrompt, err := a.buildPrompt(situation, sceneCtx)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Start with initial message
	messages := []Message{
		{Role: "user", Content: systemPrompt},
	}

	// Tool execution loop - max 50 iterations to allow for complex workflows like voting
	maxIterations := 50
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Call LLM
		req := ChatRequest{
			Messages: messages,
			Model:    a.Model,
			Tools:    tools,
		}

		response, err := a.Client.Chat(ctx, req)
		if err != nil {
			return ChatResponse{}, fmt.Errorf("LLM call failed: %w", err)
		}

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			return response, nil
		}

		// Add assistant's response (with tool calls) to messages
		// For OpenAI format, we need to preserve tool call information
		messages = append(messages, Message{
			Role:    "assistant",
			Content: response.Message,
			// TODO: May need to add ToolCalls field to Message struct
		})

		// Execute tools and collect results
		turnEnded := false
		for _, toolCall := range response.ToolCalls {
			// Execute the tool
			mcpToolCall := &mcp.ToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Name,
				Arguments: toolCall.Arguments,
			}
			result := executor.ExecuteTool(ctx, mcpToolCall)

			// Check if this tool ends the turn
			if result.EndsTurn {
				turnEnded = true
			}

			// Add tool result to messages
			// Format the result as JSON for better LLM parsing
			var resultContent string
			if result.IsError {
				resultContent = fmt.Sprintf("Tool '%s' error: %v", toolCall.Name, result.Content)
			} else {
				// Marshal result to JSON
				resultJSON, err := json.MarshalIndent(result.Content, "", "  ")
				if err != nil {
					// Fallback to string representation
					resultContent = fmt.Sprintf("Tool '%s' returned: %v", toolCall.Name, result.Content)
				} else {
					resultContent = fmt.Sprintf("Tool '%s' returned:\n%s", toolCall.Name, string(resultJSON))
				}
			}

			messages = append(messages, Message{
				Role:    "tool",
				Content: resultContent,
			})
		}

		// If a turn-ending tool was called, stop the loop
		if turnEnded {
			return response, nil
		}
	}

	// If we hit max iterations, return what we have
	return ChatResponse{
		Message: "Error: Maximum tool execution iterations reached",
	}, fmt.Errorf("maximum tool execution iterations (%d) reached", maxIterations)
}

// buildPrompt creates the full prompt using the template system.
// The prompt template is loaded from the prompts package.
// If sceneCtx is provided (typically on turn 1), it includes scene information.
func (a *Agent) buildPrompt(situation string, sceneCtx *SceneContext) (string, error) {
	// Get prompt template
	promptTemplate, err := prompts.GetPrompt("agent_turn")
	if err != nil {
		return "", fmt.Errorf("failed to load agent turn prompt: %w", err)
	}

	tmpl, err := template.New("agent_turn").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Name         string
		Character    *scenarios.Character
		State        AgentState
		Situation    string
		SceneContext *SceneContext
	}{
		Name:         a.Name,
		Character:    a.Character,
		State:        a.State,
		Situation:    situation,
		SceneContext: sceneCtx,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// String returns a string representation of the agent.
func (a *Agent) String() string {
	archetype := "unknown"
	if a.Character != nil && a.Character.External != nil {
		archetype = a.Character.External.Archetype
	}
	return fmt.Sprintf("Agent{Name: %s, Character: %s, Provider: %s, Model: %s}",
		a.Name, archetype, a.Provider, a.Model)
}
