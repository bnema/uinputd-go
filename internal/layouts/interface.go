package layouts

// RegistryInterface defines the interface for the layout registry.
// This interface allows for mocking in tests while Layout interface
// is already defined in layout.go
type RegistryInterface interface {
	// Get retrieves a layout by name
	Get(name string) (Layout, error)

	// Register adds a layout to the registry
	Register(layout Layout)

	// Available returns a list of all registered layout names
	Available() []string

	// Default returns the default layout
	Default() Layout
}

// Compile-time check to ensure Registry implements RegistryInterface
var _ RegistryInterface = (*Registry)(nil)
