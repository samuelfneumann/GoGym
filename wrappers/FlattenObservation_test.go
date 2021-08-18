package wrappers_test

import (
	"testing"

	"github.com/samuelfneumann/gogym"
	"github.com/samuelfneumann/gogym/wrappers"
	"gonum.org/v1/gonum/mat"
)

func TestNewFlattenObservation(t *testing.T) {
	// Create the environment
	env, err := gogym.Make("MountainCarContinuous-v0")
	if err != nil {
		t.Errorf("make: %v", err)
	}

	env, err = wrappers.NewFlattenObservation(env)
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
	_, _, _, err = env.Step(mat.NewVecDense(1, []float64{0.0}))
	if err != nil {
		t.Errorf("step: %v", err)
	}

	// Test the action scaling function
	_, err = env.(*wrappers.FlattenObservation).Observation([]float64{
		0.1,
		0.1,
	})
	if err != nil {
		t.Errorf("observation: %v", err)
	}

	env.Close()
}
