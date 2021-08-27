package wrappers_test

import (
	"testing"

	"github.com/samuelfneumann/gogym"
	"github.com/samuelfneumann/gogym/wrappers"
	"gonum.org/v1/gonum/mat"
)

func TestAlterDefaultTimeLimit(t *testing.T) {
	// Create the environment
	env, err := gogym.Make("MountainCarContinuous-v0")
	if err != nil {
		t.Errorf("make: %v", err)
	}

	cutoff := 1000
	env, err = wrappers.AlterDefaultTimeLimit(env, cutoff)
	if err != nil {
		t.Error(err)
	}

	// Reset the environment
	_, err = env.Reset()
	if err != nil {
		t.Errorf("reset: %v", err)
	}

	// Seed the environment
	_, err = env.Seed(10)
	if err != nil {
		t.Errorf("seed: %v", err)
	}

	// Take some environmental steps
	done := false
	i := 0
	for !done {
		_, _, done, err = env.Step(mat.NewVecDense(1, []float64{0.0}))
		if err != nil {
			t.Errorf("step: %v", err)
		}
		i++
	}
	if i != cutoff {
		t.Errorf("step: expected done == true when i == %v, got i == %v",
			cutoff, i)
	}

	env.Close()
}

func TestNewTimeLimit(t *testing.T) {
	// Create the environment
	env, err := gogym.Make("MountainCarContinuous-v0")
	if err != nil {
		t.Errorf("make: %v", err)
	}

	// Note that the default Mountain Car has a time limit of 200, so
	// this is technically a time limit of a time limit. In this case,
	// OpenAI Gym will always use the lower time limit.
	env, err = wrappers.NewTimeLimit(env, 200)
	if err != nil {
		t.Error(err)
	}

	// Reset the environment
	_, err = env.Reset()
	if err != nil {
		t.Errorf("reset: %v", err)
	}

	// Seed the environment
	_, err = env.Seed(10)
	if err != nil {
		t.Errorf("seed: %v", err)
	}

	// Take an environmental step
	done := false
	i := 0
	for !done {
		_, _, done, err = env.Step(mat.NewVecDense(1, []float64{0.0}))
		if err != nil {
			t.Errorf("step: %v", err)
		}
		i++
	}

	// Ensure the lower time limit is the limiting time limit, as in
	// the Python implementation
	if i != 200 {
		t.Errorf("step: expected done == true when i == %v, got i == %v",
			200, i)
	}

	env.Close()
}
