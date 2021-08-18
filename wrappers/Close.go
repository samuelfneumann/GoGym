// Package wrappers implements Go bindings for the environment wrappers
// in OpenAI's Gym.
package wrappers

import "github.com/samuelfneumann/gogym"

// Closed indicates whether the package has been closed or not
var Closed bool = false

// Close performs cleanup of package-level resources for the wrappers
// and gogym packages.
func Close() {
	if !Closed {
		pixelModule.DecRef()
		clipActionModule.DecRef()
		flattenObservationModule.DecRef()
		rescaleActionModule.DecRef()
		filterObservationModule.DecRef()
	}
	Closed = true

	gogym.Close()
}
