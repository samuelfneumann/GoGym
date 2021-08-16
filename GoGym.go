// Package gogym provides Go bindings for OpenAI's Python package Gym.
//
// Before running, ensure python-3.7.pc is in a directory pointed to
// by PKG_CONFIG_PATH. On Ubuntu:
// export PKG_CONFIG_PATH="$PKG_CONFIG_PATH":/usr/local/lib/pkgconfig
package gogym

// #cgo pkg-config: python-3.7
// #include <Python.h>
import "C"
import (
	"fmt"
	"os"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

// Set of open environments
var openEnvironments map[Environment]struct{} = make(map[Environment]struct{})

// Python modules
var gym *python.PyObject
var dict *python.PyObject

// Space types
var spaces *python.PyObject
var boxSpace *python.PyObject
var discreteSpace *python.PyObject
var dictSpace *python.PyObject
var tupleSpace *python.PyObject

// Closed indicates whether the package has been closed or not
var Closed bool = false

// init performs setup for the package before running
func init() {
	// Initialize the Python interpreter
	python.Py_Initialize()

	// Import gym
	gymModule := python.PyImport_ImportModule("gym")
	if gymModule == nil {
		if python.PyErr_Occurred() != nil {
			python.PyErr_Print()
		}
		panic("init: could not import gym")
	}
	defer gymModule.DecRef()
	gym = python.PyImport_AddModule("gym")
	gymModule.IncRef()

	// ! These needs to be closed after
	dict = python.PyModule_GetDict(gym)
	dict.IncRef()

	spaces = python.PyDict_GetItemString(dict, "spaces")
	spaces.IncRef()

	boxSpace = spaces.GetAttrString("Box")
	if boxSpace == nil {
		panic("init: could not get Python BoxSpace space type")
	}

	discreteSpace = spaces.GetAttrString("Discrete")
	if discreteSpace == nil {
		panic("init: could not get Python DiscreteSpace space type")
	}

	dictSpace = spaces.GetAttrString("Dict")
	if dictSpace == nil {
		panic("init: could not get Python Dict space type")
	}

	tupleSpace = spaces.GetAttrString("Tuple")
	if dictSpace == nil {
		panic("init: could not get Python Tuple space type")
	}
}

// Environment describes an OpenAI Gym environment
type Environment interface {
	// Env gets the Python OpenAI Gym environment from the Go
	// Environment
	Env() *python.PyObject

	// Name gets the name of the environment
	Name() string

	// ContinuousAction returns whether or not the environment has
	// continuous actions
	ContinuousAction() bool

	// Seed seeds the Environment and returns the seed. It is equivalent
	// to calling env.seed(seed) in Python's OpenAI Gym.
	Seed(seed int) ([]int, error)

	// ActionSpace returns the action space as a Go data structure
	ActionSpace() Space

	// ObservationSpace returns the observation space as a Go data structure
	ObservationSpace() Space

	// Step takes one environmental step given some action a and returns
	// the next observation, reward, and a flag indicating if the
	// episode has completed. It is equivalent to calling env.step(a) in
	// Python's OpenAI Gym.
	Step(a *mat.VecDense) (*mat.VecDense, float64, bool,
		error)

	// Reset resets the Environment and returns the starting state. It
	// is equivalent to calling env.reset() in Python's OpenAI Gym.
	Reset() (*mat.VecDense, error)

	// Close performs cleanup of environment resources. It should be
	// called once the environment is no longer needed.
	Close()
}

// GymEnv wraps a Python gym environment and provides Go bindings for
// interacting with that environment
type GymEnv struct {
	env              *python.PyObject
	envName          string
	continuousAction bool

	actionSpace      Space
	observationSpace Space
}

// New creates and returns a new *GymEnv
func New(env *python.PyObject, envName string, continuousAction bool,
	actionSpace, observationSpace Space) Environment {
	if Closed {
		panic("new: cannot create environment when package closed")
	}
	gymEnv := &GymEnv{
		env:              env,
		envName:          envName,
		continuousAction: continuousAction,
		actionSpace:      actionSpace,
		observationSpace: observationSpace,
	}

	openEnvironments[gymEnv] = struct{}{}
	return gymEnv
}

// Make returns a new environment with the given name. It is equivalent
// to gym.make(envName) in Python's OpenAI Gym.
func Make(envName string) (Environment, error) {
	if Closed {
		panic("make: cannot create environment when package closed")
	}
	// Get the gym.make function
	makeEnv := python.PyDict_GetItemString(dict, "make")
	makeEnv.IncRef()
	defer makeEnv.DecRef()
	if !(makeEnv != nil && python.PyCallable_Check(makeEnv)) {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		return nil, fmt.Errorf("make: error creating env %v", envName)
	}

	// Construct the arguments to the gym.make function
	args := python.PyTuple_New(1)
	defer args.DecRef()
	python.PyTuple_SetItem(args, 0, python.PyUnicode_FromString(envName))

	// Create the gym environment
	gymEnv := makeEnv.CallObject(args)
	if gymEnv == nil {
		if python.PyErr_Occurred() != nil {
			python.PyErr_Print()
		}
		return nil, fmt.Errorf("make: could not call gym.make")
	}

	// Figure out if the environment has continuous actions or not
	actionSpace := gymEnv.GetAttrString("action_space")
	defer actionSpace.DecRef()
	continuousAction := actionSpace.Type() == boxSpace

	// Construct the action space
	var goActionSpace Space
	var err error
	if continuousAction {
		goActionSpace, err = NewBoxSpace(actionSpace)
		if err != nil {
			return nil, fmt.Errorf("make: could not create BoxSpace action space "+
				"from type %v", actionSpace.Type())
		}

	} else if actionSpace.Type() == discreteSpace {
		goActionSpace, err = NewDiscreteSpace(actionSpace)
		if err != nil {
			return nil, fmt.Errorf("make: could not create DiscreteSpace action "+
				"space from type %v", actionSpace.Type())
		}

	} else {
		goActionSpace = nil
		fmt.Fprintf(os.Stderr, "make: action space %T not yet implemented",
			actionSpace.Type())
	}

	// Construct the observation space
	observationSpace := gymEnv.GetAttrString("observation_space")
	if observationSpace == nil {
		fmt.Println("\n\n Obs nil")
	}
	defer observationSpace.DecRef()
	var goObservationSpace Space
	if observationSpace.Type() == boxSpace {
		goObservationSpace, err = NewBoxSpace(observationSpace)
		if err != nil {
			return nil, fmt.Errorf("make: could not create BoxSpace observation "+
				"space from type %v", observationSpace.Type())
		}

	} else if observationSpace.Type() == discreteSpace {
		goObservationSpace, err = NewDiscreteSpace(observationSpace)
		if err != nil {
			return nil, fmt.Errorf("make: could not create DiscreteSpace "+
				"observation space from type %v", observationSpace.Type())
		}

	} else {
		goObservationSpace = nil
		fmt.Fprintf(os.Stderr, "make: observation space %T not yet "+
			"implemented", observationSpace.Type())
	}

	env := New(gymEnv, envName, continuousAction, goActionSpace,
		goObservationSpace)

	// Register the environment with the list of all environments
	openEnvironments[env] = struct{}{}

	return env, nil
}

// ActionSpace returns the action space as a Go data structure
func (g *GymEnv) ActionSpace() Space {
	return g.actionSpace
}

// ObservationSpace returns the observation space as a Go data structure
func (g *GymEnv) ObservationSpace() Space {
	return g.observationSpace
}

// Env gets the GymEnv's Python gym environment
func (g *GymEnv) Env() *python.PyObject {
	return g.env
}

// Name gets the name of the environment
func (g *GymEnv) Name() string {
	return g.envName
}

// ContinuousAction returns whether the environment uses continuous
// actions or not
func (g *GymEnv) ContinuousAction() bool {
	return g.continuousAction
}

// Seed seeds the GymEnv and returns the seed. It is equivalent to
// calling env.seed(seed) in Python's OpenAI Gym.
func (g *GymEnv) Seed(seed int) ([]int, error) {
	// Get the seed function
	seedFunc := g.env.GetAttrString("seed")
	defer seedFunc.DecRef()

	// Create the Python arguments
	args := python.PyTuple_New(1)
	defer args.DecRef()
	python.PyTuple_SetItem(args, 0, python.PyLong_FromGoInt(seed))

	// Seed the environment
	retVal := seedFunc.CallObject(args)
	defer retVal.DecRef()
	if retVal == nil {
		return nil, fmt.Errorf("seed: no seed returned from gym")
	}

	// Return the seed
	s, err := IntSliceFromIter(retVal)
	if err != nil {
		return nil, fmt.Errorf("seed: could not convert seed to Go: %v", err)
	}
	return s, nil
}

// Step takes one environmental step given some action a and returns
// the next observation, reward, and a flag indicating if the
// episode has completed. It is equivalent to calling env.step(a) in
// Python's OpenAI Gym.
func (g *GymEnv) Step(a *mat.VecDense) (*mat.VecDense, float64, bool,
	error) {
	// Get the step function
	stepFunc := g.env.GetAttrString("step")
	defer stepFunc.DecRef()

	// Create the Python arguments
	args := python.PyTuple_New(1)
	defer args.DecRef()
	if g.continuousAction {
		arr, err := F64ToList(a.RawVector().Data)
		if err != nil {
			return nil, 0, false, fmt.Errorf("step: could not convert " +
				"[]float64 to Python List")
		}
		python.PyTuple_SetItem(args, 0, arr)
	} else {
		python.PyTuple_SetItem(args, 0, python.PyLong_FromDouble(a.AtVec(0)))
	}

	// Call step in Python gym
	retVal := stepFunc.CallObject(args)
	defer retVal.DecRef()
	if retVal == nil {
		return nil, 0, false, fmt.Errorf("step: could not step in " +
			"gym environment")
	}

	// Get the observation vector
	obs := python.PyTuple_GetItem(retVal, 0)
	goObsSlice, err := F64SliceFromIter(obs)
	if err != nil {
		return nil, 0, false, fmt.Errorf("step: could not decode observation")
	}
	goObs := mat.NewVecDense(len(goObsSlice), goObsSlice)

	// Get the reward
	reward := python.PyTuple_GetItem(retVal, 1)
	goReward := python.PyFloat_AsDouble(reward)

	// Figure out if the episode is done
	done := python.PyTuple_GetItem(retVal, 2)
	goDone := python.Py_True == done

	return goObs, goReward, goDone, nil
}

// Reset resets the GymEnv and returns the starting state. It is
// equivalent to calling env.reset() in Python's OpenAI Gym.
func (g *GymEnv) Reset() (*mat.VecDense, error) {
	resetFunc := g.env.GetAttrString("reset")
	defer resetFunc.DecRef()

	state := resetFunc.CallObject(nil)
	defer state.DecRef()

	data, err := F64SliceFromIter(state)
	if err != nil {
		return nil, fmt.Errorf("reset: could not decode Python iterable: %v",
			err)
	}
	return mat.NewVecDense(len(data), data), nil
}

// Close performs cleanup of environment resources. It should be
// called once the environment is no longer needed.
func (g *GymEnv) Close() {
	// Remove g from the list of all open environments
	delete(openEnvironments, g)

	// Decrement the gym environment counter
	g.env.DecRef()
}

// Render renders the environment. It is equivalent to env.render()
// in Python's OpenAI Gym.
//
// Currently, a bug in Python causes this function to panic.
func (g *GymEnv) Render() {
	// This function does not work. The issue is trying to import
	// gym.error in the classic_control/rendering file. This should
	// be a fairly easy fix.
	renderFunc := g.env.GetAttrString("render")
	defer renderFunc.DecRef()

	render := renderFunc.CallObject(nil)
	if render == nil {
		if python.PyErr_Occurred() != nil {
			python.PyErr_Print()
		}
		panic("render: could not render")
	}
}

// Close performs cleanup of package resources. Any environments that
// have not been closed will be closed. This should be called after
// the package is no longer needed or at the end of main.
func Close() {
	if !Closed {
		// Close all open environments
		for env := range openEnvironments {
			env.Close()
		}

		// Decrement the reference count for the gym module
		gym.DecRef()

		// Decrement spaces counters
		spaces.DecRef()
		boxSpace.DecRef()
		discreteSpace.DecRef()
		dictSpace.DecRef()

		// Close Python interpreter
		python.Py_Finalize()

	}
	Closed = true
}

// An example of how to use the package
func Example() {
	// Create the environment
	env, err := Make("MountainCar-v0")
	if err != nil {
		panic(err)
	}

	// Print the environment
	fmt.Println("GymEnv:", env)
	fmt.Printf("Python gym env: ")
	print(env.Env())
	fmt.Println()

	// Reset the environment
	state, err := env.Reset()
	if err != nil {
		panic(err)
	}
	fmt.Println("Reset:", state)
	fmt.Println()

	// Seed the environment
	seed, err := env.Seed(10)
	if err != nil {
		panic(err)
	}
	fmt.Println("Seed:", seed)
	fmt.Println()

	// Take an environmental step
	obs, reward, done, err := env.Step(mat.NewVecDense(1, []float64{0.0}))
	if err != nil {
		panic(err)
	}
	fmt.Println("Step [Obs, reward, done]:", obs, reward, done)

	// Close the environment and package resources
	Close()
}

// F64SliceFromIter converts a Python iterable to a []float64. Borrows
// python.PyObject reference.
//
// Note: this is used to convert NumPy vectors to []float64. Since the
// NumPy C API is currently not supported by this library, no error
// checking is done.
func F64SliceFromIter(obj *python.PyObject) ([]float64, error) {
	seq := obj.GetIter()
	defer seq.DecRef()
	next := seq.GetAttrString("__next__")
	defer next.DecRef()

	data := make([]float64, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		item := next.CallObject(nil)
		if item == nil {
			return nil, fmt.Errorf("f64SliceFromIter: nil item at index %v", i)
		}

		// No error checking: we need to use the NumPy C API for this

		data[i] = python.PyFloat_AsDouble(item)
		item.DecRef()
	}

	return data, nil
}

// StringSliceFromIter converts a Python iterable to a []string. Borrows
// python.PyObject reference.
func StringSliceFromIter(obj *python.PyObject) ([]string, error) {
	seq := obj.GetIter()
	defer seq.DecRef()
	next := seq.GetAttrString("__next__")
	defer next.DecRef()

	data := make([]string, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		item := next.CallObject(nil)
		if item == nil {
			return nil, fmt.Errorf("stringSliceFromIter: nil item at index %v", i)
		}

		if !python.PyUnicode_Check(item) {
			return nil, fmt.Errorf("stringSliceFromIter: item at index %v is "+
				"not a string", i)
		}

		data[i] = python.PyUnicode_AsUTF8(item)
		item.DecRef()
	}

	return data, nil
}

// IntSliceFromIter converts a Python iterable to a []int. Borrows
// python.PyObject reference.
func IntSliceFromIter(obj *python.PyObject) ([]int, error) {
	seq := obj.GetIter()
	defer seq.DecRef()
	next := seq.GetAttrString("__next__")
	defer next.DecRef()

	data := make([]int, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		item := next.CallObject(nil)
		if item == nil {
			return nil, fmt.Errorf("intSliceFromIter: nil item at index %v", i)
		}

		if !python.PyLong_Check(item) {
			return nil, fmt.Errorf("intSliceFromIter: item at index %v is "+
				"not an int", i)
		}

		data[i] = python.PyLong_AsLong(item)
		item.DecRef()
	}

	return data, nil
}

// F64ToList converts a []flaot64 to a Python List. Creates a new
// python.PyObject reference.
func F64ToList(slice []float64) (*python.PyObject, error) {
	list := python.PyList_New(len(slice))
	for i, elem := range slice {
		float := python.PyFloat_FromDouble(elem)
		n := python.PyList_SetItem(list, i, float)
		if n != 0 {
			if python.PyErr_Occurred() != nil {
				python.PyErr_Print()
			}
			float.DecRef()
			list.DecRef()
			return nil, fmt.Errorf("f64ToList: could not set Python list item")
		}
	}
	return list, nil
}

// Print prints a *python.PyObject in a similar way to calling print()
// in Python.
func Print(obj *python.PyObject) {
	fmt.Println(python.PyUnicode_AsUTF8(obj.Str()))
}
