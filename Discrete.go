package gogym

import (
	"fmt"
	"time"

	python "github.com/DataDog/go-python3"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

// DiscreteSpace represents a space of discrete numbers: (0, 1, 2, ..., n-1).
//
// The DiscreteSpace space must be constructed from its Python equivalent.
type DiscreteSpace struct {
	*python.PyObject
	rand.Source
	rng distuv.Categorical
	n   int // Number of actions, actions in (0, 1, ..., n-1)
}

// NewDiscreteSpace takes a Python gym.spaces.DiscreteSpace and converts it into
// its Go counterpart.
func NewDiscreteSpace(space *python.PyObject) (Space, error) {
	if !(space.Type() == discreteSpace) {
		return nil, fmt.Errorf("newDiscreteSpace: space is not a discrete " +
			"space")
	}
	pythonN := space.GetAttrString("n")
	defer pythonN.DecRef()
	if pythonN == nil {
		return nil, fmt.Errorf("newDiscreteSpace: space %v is not a DiscreteSpace",
			space.Type())
	}
	n := python.PyLong_AsLong(pythonN)

	src := rand.NewSource(uint64(time.Now().UnixNano()))
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0
	}
	rng := distuv.NewCategorical(weights, src)

	return &DiscreteSpace{
		PyObject: space,
		Source:   src,
		rng:      rng,
		n:        n,
	}, nil
}

// Sample takes a sample from within the spaces bounds
func (d *DiscreteSpace) Sample() []*mat.VecDense {
	return []*mat.VecDense{
		mat.NewVecDense(1, []float64{
			float64(int(d.rng.Rand()) % d.n),
		}),
	}
}

// Contains returns whether in is in the space. The argument in
// should be a []float64 or *mat.VecDense.
func (d *DiscreteSpace) Contains(in interface{}) bool {
	x, ok := in.([]float64)
	if !ok {
		vec, ok := in.(*mat.VecDense)
		if !ok {
			return false
		}
		x = vec.RawVector().Data
	}
	intX := int(x[0])
	return len(x) == 1 && intX < d.n && intX >= 0
}

// High returns the upper bounds of the space
func (d *DiscreteSpace) High() []*mat.VecDense {
	return []*mat.VecDense{mat.NewVecDense(1, []float64{float64(d.n - 1)})}
}

// Low returns the lower bounds of the space
func (d *DiscreteSpace) Low() []*mat.VecDense {
	return []*mat.VecDense{mat.NewVecDense(1, []float64{1.0})}
}
