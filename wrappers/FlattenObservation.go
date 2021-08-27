package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// wrappers.flatten_observation Python module
var flattenObservationModule *python.PyObject

// init performs setup before running
func init() {
	// Create the gym.wrappers.flatten_observation Python module
	wrappersModule := python.PyImport_ImportModule("gym.wrappers." +
		"flatten_observation")
	defer wrappersModule.DecRef()
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.flatten_observation")
	}
	flattenObservationModule = python.PyImport_AddModule("gym.wrappers." +
		"flatten_observation")
	wrappersModule.IncRef()
}

// FlattenObservation wraps a gogym.Environment and flattens the
// observations.
//
// The observation space of a FlattenObservation wrapper should always
// be a BoxSpace.
//
// https://github.com/openai/gym/blob/master/gym/wrappers/flatten_
// observation.py
type FlattenObservation struct {
	gogym.Environment
	wrapped gogym.Environment
}

// NewFlattenObservation returns a new gogym.Environment that flattens
// state observations
func NewFlattenObservation(env gogym.Environment) (gogym.Environment, error) {
	// Call the FlattenObservation constructor with the argument environment
	newEnv := flattenObservationModule.CallMethodArgs("FlattenObservation",
		env.Env())
	defer newEnv.DecRef()
	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		return nil, fmt.Errorf("newFlattenObservation: could not wrap " +
			"environment")
	}

	// Create the new observation space
	pyObservationSpace := newEnv.GetAttrString("observation_space")
	defer pyObservationSpace.DecRef()
	obsSpace, err := gogym.SpaceFromPyObject(pyObservationSpace)
	if err != nil {
		return nil, fmt.Errorf("newFlattenObservation: could not get Python "+
			"observation space: %v",
			err)
	}

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("FlattenObservation(%v)", env.Name()),
		env.ContinuousAction(),
		env.ActionSpace(),
		obsSpace.(*gogym.BoxSpace),
	)

	return &FlattenObservation{
		Environment: newGymEnv,
		wrapped:     env,
	}, nil
}

// Observation returns a flattened version of some observation x.
// The argument x must be a valid argument to gogym.Flatten.
func (f *FlattenObservation) Observation(x interface{}) ([]float64, error) {
	return gogym.Flatten(f.ObservationSpace(), x)
}

// Close performs cleanup of environment resources
func (f *FlattenObservation) Close() {
	// Close the wrapped environment
	f.wrapped.Close()

	// Close this environment
	f.Environment.Close()
}
