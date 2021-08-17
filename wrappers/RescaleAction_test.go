package wrappers_test

import (
	"math"
	"testing"

	"github.com/samuelfneumann/gogym"
	"github.com/samuelfneumann/gogym/wrappers"
	"gonum.org/v1/gonum/mat"
)

func TestNewRescaleAction(t *testing.T) {
	// Create the environment
	env, err := gogym.Make("MountainCarContinuous-v0")
	if err != nil {
		t.Errorf("make: %v", err)
	}

	env, err = wrappers.NewRescaleAction(env, -0.5, 0.5)
	if err != nil {
		t.Errorf("newClipAction: %v", err)
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
	action, err := env.(*wrappers.RescaleAction).Action([]float64{0.1})
	if err != nil {
		t.Errorf("action: %v", err)
	}

	threshold := 0.0000001
	if math.Abs(action[0]-0.2) > threshold {
		t.Errorf("action: got %v expected 0.2", action)
	}

	env.Close()
}
