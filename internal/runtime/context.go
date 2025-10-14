package runtime

// contextKey is the type used for context keys in the wonda runtime.
// Using a custom type prevents collisions with context keys from other packages.
type contextKey string

// Context keys used throughout the simulation runtime.
const (
	// AgentNameKey is the context key for storing the current agent's name.
	AgentNameKey contextKey = "agent_name"
)
