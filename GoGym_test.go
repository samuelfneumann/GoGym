package gogym_test

import (
	"testing"

	"github.com/samuelfneumann/gogym"
	"gonum.org/v1/gonum/mat"
)

func TestMake(t *testing.T) {

	tests := []string{
		// "MountainCar-v0",
		"MountainCarContinuous-v0",
		// "Pendulum-v0",
		// "Acrobot-v1",
		"CartPole-v1",
		// "HalfCheetah-v2",
		"Ant-v2",
	}

	for _, test := range tests {
		// Create the environment
		env, err := gogym.Make(test)
		if err != nil {
			t.Errorf("make: %v", err)
		}

		if env.ObservationSpace() == nil {
			t.Logf("make: observation space type not yet implemented")
		}

		if env.ActionSpace() == nil {
			t.Logf("make: action space type not yet implemented")
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

		env.Close()

	}
	gogym.Close()
}
