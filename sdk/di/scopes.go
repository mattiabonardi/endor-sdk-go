package di

// Scope defines the lifecycle management strategy for dependencies
type Scope int

const (
	// Singleton creates a single shared instance across all resolutions (default)
	Singleton Scope = iota
	// Transient creates a new instance for each resolution
	Transient
	// Scoped creates instance bound to request/context scope (future: Epic 3)
	Scoped
)

// String returns string representation of the scope
func (s Scope) String() string {
	switch s {
	case Singleton:
		return "Singleton"
	case Transient:
		return "Transient"
	case Scoped:
		return "Scoped"
	default:
		return "Unknown"
	}
}
