package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// wrappers.rescale_action Python module
var rescaleActionModule *python.PyObject

// init performs setup before running
func init() {
	// Create the gym.wrappers.rescale_action Python module
	wrappersModule := python.PyImport_ImportModule("gym.wrappers.rescale_action")
	defer wrappersModule.DecRef()
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.rescale_action")
	}
	rescaleActionModule = python.PyImport_AddModule("gym.wrappers.rescale_action")
	wrappersModule.IncRef()
}

// RescaleAction wraps a gogym.Environment and rescales the continuous
// action space of the environment to a range [a, b]. The action
// space should be a BoxSpace.
//
// https://github.com/openai/gym/blob/master/gym/wrappers/rescale_
// action.py
type RescaleAction struct {
	gogym.Environment
	wrapped     gogym.Environment
	a, b        float64
	actionSpace gogym.Space
}

// NewRescaleAction returns a new gogym.Environment that rescales the
// actions taken in env.
func NewRescaleAction(env gogym.Environment, a, b float64) (gogym.Environment,
	error) {
	// Call the RescaleAction constructor with the argument environment
	low := python.PyFloat_FromDouble(a)
	high := python.PyFloat_FromDouble(b)
	newEnv := rescaleActionModule.CallMethodArgs("RescaleAction", env.Env(),
		low, high)

	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		return nil, fmt.Errorf("rescaleAction: could not wrap environment")
	}

	// Create the new action space
	pyActionSpace := newEnv.GetAttrString("action_space")
	defer pyActionSpace.DecRef()
	actionSpace, err := gogym.NewBoxSpace(pyActionSpace)
	if err != nil {
		return nil, fmt.Errorf("could not get Python action space: %v", err)
	} else if actionSpace == nil {
		return nil, fmt.Errorf("could not get Python action space")
	}

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("ClipAction(%v)", env.Name()),
		env.ContinuousAction(),
		env.ActionSpace(),
		env.ObservationSpace(),
	)

	return &RescaleAction{
		Environment: newGymEnv,
		wrapped:     env,
		a:           a,
		b:           b,
		actionSpace: actionSpace,
	}, nil
}

// ActionSpace returns the action space as a Go data structure
func (r *RescaleAction) ActionSpace() gogym.Space {
	return r.actionSpace
}

// Action rescales the argument action to the legal bounds in the
// RescaleAction environment
func (r *RescaleAction) Action(action []float64) ([]float64, error) {
	pyAction, err := gogym.F64ToList(action)
	if err != nil {
		return nil, fmt.Errorf("action: could not convert to Python List: %v",
			err)
	}

	scaledAction := r.Environment.Env().CallMethodArgs("action", pyAction)
	if scaledAction == nil && action != nil {
		return nil, fmt.Errorf("action: could not get Python action")
	}

	goAction, err := gogym.F64SliceFromIter(scaledAction)
	if err != nil {
		return nil, fmt.Errorf("action: could not convert Python List to "+
			"[]float64: %v", err)
	}

	return goAction, nil
}

// Close performs cleanup of environment resources
func (r *RescaleAction) Close() {
	// Close the wrapped environment
	r.wrapped.Close()

	// Close this environment
	r.Environment.Close()
}
