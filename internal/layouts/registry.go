package layouts

import (
	"fmt"
	"sync"
)

// Layout name constants
const (
	NameUS = "us" // US QWERTY
	NameUK = "uk" // UK QWERTY
	NameFR = "fr" // French AZERTY
	NameDE = "de" // German QWERTZ
	NameES = "es" // Spanish QWERTY
	NameIT = "it" // Italian QWERTY
)

// Registry manages available keyboard layouts.
type Registry struct {
	mu      sync.RWMutex
	layouts map[string]Layout
}

// NewRegistry creates a new layout registry with default layouts.
func NewRegistry() *Registry {
	r := &Registry{
		layouts: make(map[string]Layout),
	}

	// Register default layouts
	r.Register(NewUS())
	r.Register(NewFR())
	r.Register(NewDE())
	r.Register(NewES())
	r.Register(NewUK())
	r.Register(NewIT())

	return r
}

// Register adds a layout to the registry.
func (r *Registry) Register(layout Layout) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.layouts[layout.Name()] = layout
}

// Get retrieves a layout by name.
func (r *Registry) Get(name string) (Layout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	layout, ok := r.layouts[name]
	if !ok {
		return nil, fmt.Errorf("layout %q not found (available: %v)", name, r.Available())
	}

	return layout, nil
}

// Available returns a list of available layout names.
func (r *Registry) Available() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.layouts))
	for name := range r.layouts {
		names = append(names, name)
	}
	return names
}

// Default returns the default US layout.
func (r *Registry) Default() Layout {
	return NewUS()
}
