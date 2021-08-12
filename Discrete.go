package gogym

import (
	"fmt"
	"time"

	python "github.com/DataDog/go-python3"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

// Discrete represents a space of discrete numbers: (0, 1, 2, ..., n-1).
//
// The Discrete space must be constructed from its Python equivalent.
type Discrete struct {
	*python.PyObject
	rand.Source
	rng distuv.Categorical
	n   int // Number of actions, actions in (0, 1, ..., n-1)
}

// NewDiscrete takes a Python gym.spaces.Discrete and converts it into
// its Go counterpart.
func NewDiscrete(space *python.PyObject) (Space, error) {
	pythonN := space.GetAttrString("n")
	defer pythonN.DecRef()
	if pythonN == nil {
		return nil, fmt.Errorf("newDiscrete: space %v is not a Discrete",
			space.Type())
	}
	n := python.PyLong_AsLong(pythonN)

	src := rand.NewSource(uint64(time.Now().UnixNano()))
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0
	}
	rng := distuv.NewCategorical(weights, src)

	return &Discrete{
		PyObject: space,
		Source:   src,
		rng:      rng,
		n:        n,
	}, nil
}

// Sample takes a sample from within the spaces bounds
func (d *Discrete) Sample() *mat.VecDense {
	return mat.NewVecDense(1, []float64{float64(int(d.rng.Rand()) % d.n)})
}

// Contains returns whether x is in the space
func (d *Discrete) Contains(x *mat.VecDense) bool {
	intX := int(x.AtVec(0))
	return x.Len() == 1 && intX < d.n && intX >= 0
}

// High returns the upper bounds of the space
func (d *Discrete) High() *mat.VecDense {
	return mat.NewVecDense(1, []float64{float64(d.n - 1)})
}

// Low returns the lower bounds of the space
func (d *Discrete) Low() *mat.VecDense {
	return mat.NewVecDense(1, []float64{1.0})
}
