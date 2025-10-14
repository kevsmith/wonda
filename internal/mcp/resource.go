package mcp

// Resource represents stateful data that can be accessed by tools.
// In MCP terminology, resources are data sources that tools can read from.
type Resource struct {
	// URI is a unique identifier for this resource (e.g., "world://state")
	URI string

	// Name is a human-readable name for this resource
	Name string

	// Description explains what data this resource provides
	Description string

	// MimeType indicates the format of the resource data
	MimeType string

	// Read retrieves the current state of this resource
	Read func() (interface{}, error)
}
