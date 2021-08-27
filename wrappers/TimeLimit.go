package wrappers

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"github.com/samuelfneumann/gogym"
)

// wrappers.rescale_action Python module
var timeLimitModule *python.PyObject

// init performs setup before running
func init() {
	// Create the gym.wrappers.time_limit Python module
	wrappersModule := python.PyImport_ImportModule("gym.wrappers." +
		"time_limit")
	defer wrappersModule.DecRef()
	if wrappersModule == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		panic("init: could not import gym.wrappers.time_limit")
	}
	timeLimitModule = python.PyImport_AddModule("gym.wrappers." +
		"time_limit")
	wrappersModule.IncRef()
}

// TimeLimit wraps a gogym.Environment and provides for it a limit on
// the time steps. Note that all environments in OpenAI Gym have
// default time limits, and that if a TimeLimit wrapper is used, the
// lower of the two time limits will be the effective one. To get
// around this, you can use the AlterDefaultTimeLimit function.
//
// https://github.com/openai/gym/blob/master/gym/wrappers/time_limit.py
type TimeLimit struct {
	gogym.Environment
	wrapped gogym.Environment

	maxEpisodeSteps int
}

// AlterDefaultTimeLimit alters the existing, default time limit of a
// gogym Environment imposed by OpenAI Gym. If using this function,
// to create a TimeLimit wrapper, then this function must create the
// first wrapper for the gogym Environment. No other wrappers can be
// used before this one.
//
// Don't be tempted to call this function on the Environment field of
// a previously wrapped Environment. Although the function will seem
// to succeed, this will in effect be the same as creating a time
// limit over a time limit, in which case the lower time limit will
// be the effective one.
func AlterDefaultTimeLimit(env gogym.Environment,
	maxEpisodeSteps int) (gogym.Environment, error) {
	_, ok := env.(*gogym.GymEnv)
	if !ok {
		return nil, fmt.Errorf("alterDefaultTimeLimit: cannot alter the " +
			"default time limit after *gogym.GymEnv has been wrapped - use " +
			"this wrapper before any others")
	}
	embedded := env.Env()
	pyEnv := embedded.GetAttrString("env")
	defer pyEnv.DecRef()
	if pyEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		return nil, fmt.Errorf("timeLimitOfEmbedded: no embedded environment "+
			"named 'env' in Python object for %v environment", env.Name())
	}

	newEnv := gogym.New(pyEnv, env.Name(), env.ContinuousAction(),
		env.ActionSpace(), env.ObservationSpace())

	return NewTimeLimit(newEnv, maxEpisodeSteps)
}

// NewTimeLimit create a new TimeLimit wrapper on a gogym Environment.
// All gogym Environments have default time limits imposed by OpenAI
// Gym. If using this function to create a time limit, then the lower
// of the two time limits between the created one and the default on
// will be the effective time limit.
//
// To adjust the default time limit, see AlterDefaultTimeLimit().
func NewTimeLimit(env gogym.Environment,
	maxEpisodeSteps int) (gogym.Environment, error) {
	if maxEpisodeSteps <= 0 {
		return nil, fmt.Errorf("newTimeLimit: maxEpisodeSteps must be positive")
	}
	// Call the TimeLimit constructor with the argument environment
	pyCutoff := python.PyLong_FromGoInt(int(maxEpisodeSteps))
	newEnv := timeLimitModule.CallMethodArgs("TimeLimit", env.Env(), pyCutoff)
	defer newEnv.DecRef()

	if newEnv == nil {
		if python.PyErr_Occurred() != nil {
			fmt.Println()
			fmt.Println("========== Python Error ==========")
			python.PyErr_Print()
			fmt.Println("==================================")
			fmt.Println()
		}
		return nil, fmt.Errorf("newTimeLimit: could not wrap environment")
	}

	// Create the new gogym Environment
	newGymEnv := gogym.New(
		newEnv,
		fmt.Sprintf("TimeLimit(steps: %v)(%v)", maxEpisodeSteps, env.Name()),
		env.ContinuousAction(),
		env.ActionSpace(),
		env.ObservationSpace(),
	)

	return &TimeLimit{
		Environment:     newGymEnv,
		wrapped:         env,
		maxEpisodeSteps: maxEpisodeSteps,
	}, nil
}

// Close performs cleanup of environment resources
func (t *TimeLimit) Close() {
	// Close the wrapped environment
	t.wrapped.Close()

	// Close this environment
	t.Environment.Close()
}
