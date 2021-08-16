package gogym

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

// Space describes a space of actions, observations, etc. It is the Go
// equivalent of a description of the gym.spaces package. Each space
// must be constructed from its Python equivalent. That is, the
// constructor for each space takes in the Python version of the space
// and converts it to a Go version.
type Space interface {
	// Sample takes a sample from within the spaces bounds
	Sample() []*mat.VecDense

	// Contains returns whether x is in the space
	Contains(x interface{}) bool

	// Seed seeds the sampler for the space
	Seed(uint64)

	// Low returns the lower bounds of the space
	Low() []*mat.VecDense

	// High returns the upper bounds of the space
	High() []*mat.VecDense
}

// FromPythonSpace converts a Python Open AI Gym space to a Go
// equivalent
func FromPythonSpace(space *python.PyObject) (Space, error) {
	var value Space
	var err error
	switch space.Type() {
	case boxSpace:
		value, err = NewBoxSpace(space)

	case discreteSpace:
		value, err = NewDiscreteSpace(space)

	case dictSpace:
		value, err = NewDictSpace(space)

	// case tupleSpace:
	// 	value, err = NewTupleSpace(space)

	default:
		return nil, fmt.Errorf("fromPythonSpace: space %v not yet "+
			"implemented", space.Type())
	}
	if err != nil {
		return nil, fmt.Errorf("fromPythonSpace: could not convert space: %v",
			err)
	}
	return value, nil
}
