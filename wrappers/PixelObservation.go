package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// OpenAI Gym wrappers module
var pixelModule *python.PyObject

// init performs setup before running
func init() {
	wrappersModule := python.PyImport_ImportModule("gym.wrappers.pixel_observation")
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.pixel_observation")
	}
	defer wrappersModule.DecRef()
	pixelModule = python.PyImport_AddModule("gym.wrappers.pixel_observation")
	wrappersModule.IncRef()
}

// PixelObservation wraps a gogym.Environment to provide pixel
// observations.
//
// Warning: this implementation panics for the same reason the
// gogym.*GymEnv.Render() method panics.
type PixelObservation struct {
	gogym.Environment
}

// NewPixelObservation returns a new gogym.Environment with pixel
// observations.
func NewPixelObservation(env *gogym.GymEnv, pixelsOnly bool,
	pixelKeys string) (gogym.Environment, error) {
	// Construct the arguments to the PixelObservation constructor
	var pythonPixelsOnly *python.PyObject
	if pixelsOnly {
		pythonPixelsOnly = python.PyBool_FromLong(1)
	} else {
		pythonPixelsOnly = python.PyBool_FromLong(0)
	}
	defer pythonPixelsOnly.DecRef()

	// Create the new Python Gym Environment
	newEnv := pixelModule.CallMethodArgs(
		"PixelObservationWrapper",
		env.Env(),
		pythonPixelsOnly,
		nil,
		python.PyUnicode_FromString(pixelKeys),
	)
	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("pixelObservations: could not wrap environment")
	}

	env.Env().DecRef()

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("Pixel(%v)", env.Name()),
		env.ContinuousAction(),
	)

	return &PixelObservation{
		Environment: newGymEnv,
	}, nil
}
