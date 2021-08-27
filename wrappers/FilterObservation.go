package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// wrappers.filter_observation Python module
var filterObservationModule *python.PyObject

// init performs setup before running
func init() {
	// Create the gym.wrappers.filter_observation Python module
	wrappersModule := python.PyImport_ImportModule("gym.wrappers." +
		"filter_observation")
	defer wrappersModule.DecRef()
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.filter_observation")
	}
	filterObservationModule = python.PyImport_AddModule("gym.wrappers." +
		"filter_observation")
	wrappersModule.IncRef()
}

// FilterObservation filters DictSpace environment observations
// by their keys.
//
// https://github.com/openai/gym/blob/master/gym/wrappers/filter_
// observation.py
type FilterObservation struct {
	gogym.Environment
	wrapped gogym.Environment

	keys []string
}

// NewFilterObservation returns a new gogym.Environment that filters
// DictSpace state observations by the specified keys.
//
// This function wil panic if the observation space of env is not a
// DictSpace.
func NewFilterObservation(env gogym.Environment,
	keys ...string) (gogym.Environment, error) {
	// Ensure observation space is a DictSpace
	_, isDictSpace := env.ObservationSpace().(*gogym.DictSpace)
	if !isDictSpace {
		return nil, fmt.Errorf("newFilterObservation: could not wrap " +
			"environment with non-DictSpace observation space")
	}

	// Create the arguments for the filter keys
	pythonArgs := make([]*python.PyObject, len(keys)+1)
	pythonArgs[0] = env.Env()
	for i, key := range keys {
		if i == 0 {
			continue
		}
		pythonArgs[i] = python.PyUnicode_FromString(key)
		defer pythonArgs[i].DecRef()
	}

	// Call the FilterObservation constructor with the argument environment
	newEnv := filterObservationModule.CallMethodArgs("FilterObservation",
		pythonArgs...)
	defer newEnv.DecRef()
	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}

		return nil, fmt.Errorf("newFilterObservation: could not wrap " +
			"environment")
	}

	// Create the new observation space
	pyObservationSpace := newEnv.GetAttrString("observation_space")
	defer pyObservationSpace.DecRef()
	obsSpace, err := gogym.SpaceFromPyObject(pyObservationSpace)
	if err != nil {
		return nil, fmt.Errorf("newFilterObservation: could not get Python "+
			"observation space: %v", err)
	}

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("FilterObservation(%v)", env.Name()),
		env.ContinuousAction(),
		env.ActionSpace(),
		obsSpace.(*gogym.BoxSpace),
	)

	return &FilterObservation{
		Environment: newGymEnv,
		wrapped:     env,
		keys:        keys,
	}, nil
}

// Observation returns the filtered observation of x
func (f *FilterObservation) Observation(
	x map[string]interface{}) map[string]interface{} {

	return f.filterObservation(x)
}

// filterObservation filters obs by the filter keys of f
func (f *FilterObservation) filterObservation(
	obs map[string]interface{}) map[string]interface{} {

	newObs := make(map[string]interface{})
	for _, key := range f.keys {
		newObs[key] = obs[key]
	}
	return newObs
}
