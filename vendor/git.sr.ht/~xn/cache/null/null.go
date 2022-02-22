package null

// Null is empty/stub cache client, usable for testing
type Null struct{}

// New creates new stub client
func New() *Null {
	return &Null{}
}

// Get does nothing
func (c *Null) Get(_ interface{}) interface{} {
	return nil
}

// Set does nothing
func (c *Null) Set(_ interface{}, _ interface{}) {}

// Has does nothing
func (c *Null) Has(_ interface{}) bool {
	return false
}

// Remove does nothing
func (c *Null) Remove(_ interface{}) {}

// Purge does nothing
func (c *Null) Purge() {}
