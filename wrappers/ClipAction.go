package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// wrappers.clip_action Python module
var clipActionModule *python.PyObject

// init performs setup before running
func init() {
	// Create the gym.wrappers.clip_action Python module
	wrappersModule := python.PyImport_ImportModule("gym.wrappers.clip_action")
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.clip_action")
	}
	defer wrappersModule.DecRef()
	clipActionModule = python.PyImport_AddModule("gym.wrappers.clip_action")
	wrappersModule.IncRef()
}

// ClipAction wraps a gogym.Environment and clips the continuous action
// within the valid bounds.
//
// https://github.com/openai/gym/blob/master/gym/wrappers/clip_action.py
type ClipAction struct {
	gogym.Environment
}

// NewClipAction returns a new gogym.Environment that clips the actions
// taken in env.
func NewClipAction(env gogym.Environment) (gogym.Environment, error) {
	// Call the ClipAction constructor with the argument environment
	newEnv := clipActionModule.CallMethodArgs("ClipAction", env.Env())
	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("clipAction: could not wrap environment")
	}

	env.Env().DecRef()

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("ClipAction(%v)", env.Name()),
		env.ContinuousAction(),
		env.ActionSpace(),
		env.ObservationSpace(),
	)

	return &ClipAction{
		Environment: newGymEnv,
	}, nil
}
