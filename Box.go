package gogym

import (
	"fmt"
	"math"
	"time"

	"golang.org/x/exp/rand"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/spatial/r1"
	"gonum.org/v1/gonum/stat/distmv"
)

// Space describes a space of actions, observations, etc. It is the Go
// equivalent of a description of the gym.spaces package. Each space
// must be constructed from its Python equivalent. That is, the
// constructor for each space takes in the Python version of the space
// and converts it to a Go version.
type Space interface {
	// Sample takes a sample from within the spaces bounds
	Sample() []*mat.VecDense

	// Contains returns whether x is in the space
	Contains(x interface{}) bool

	// Seed seeds the sampler for the space
	Seed(uint64)

	// Low returns the lower bounds of the space
	Low() []*mat.VecDense

	// High returns the upper bounds of the space
	High() []*mat.VecDense
}

// Box represents a (possibly unbounded) box in R^n. Specifically, a
// Box represents the Cartesian product of n closed intervals. Each
// interval has the form of one of [a, b], (-∞, b], [a, ∞), or
// (-∞, ∞) for a, b ϵ R.
//
// The Box space must be constructed from its Python equivalent.
type Box struct {
	*python.PyObject // Box Space
	rng              *distmv.Uniform
	rand.Source
	low, high                  *mat.VecDense
	boundedBelow, boundedAbove []bool
}

// NewBox takes a Python gym.spaces.Box and converts it into its Go
// counterpart.
func NewBox(boxSpace *python.PyObject) (Space, error) {
	// Lower bounds
	low := boxSpace.GetAttrString("low")
	defer low.DecRef()
	if low == nil {
		return nil, fmt.Errorf("newBox: space %v is not a Box",
			boxSpace.Type())
	}
	goLow, err := F64SliceFromIter(low)
	if err != nil {
		return nil, fmt.Errorf("newBox: could not compute lower bound: %v",
			err)
	}

	// Upper bounds
	high := boxSpace.GetAttrString("high")
	defer high.DecRef()
	if high == nil {
		return nil, fmt.Errorf("newBox: space %v is not a Box",
			boxSpace.Type())
	}
	goHigh, err := F64SliceFromIter(high)
	if err != nil {
		return nil, fmt.Errorf("newBox: could not compute upper bound: %v",
			err)
	}

	boundedBelow := make([]bool, len(goLow))
	for i := range boundedBelow {
		boundedBelow[i] = math.Inf(-1) < goLow[i]
	}

	boundedAbove := make([]bool, len(goHigh))
	for i := range boundedAbove {
		boundedAbove[i] = math.Inf(1) > goHigh[i]
	}

	// Random number generator for sampling from the space
	src := rand.NewSource(uint64(time.Now().UnixNano()))
	bounds := make([]r1.Interval, len(goLow))
	for i := range bounds {
		bounds[i] = r1.Interval{Min: goLow[i], Max: goHigh[i]}
	}
	rng := distmv.NewUniform(bounds, src)

	return &Box{
		PyObject:     boxSpace,
		low:          mat.NewVecDense(len(goLow), goLow),
		high:         mat.NewVecDense(len(goHigh), goHigh),
		rng:          rng,
		Source:       src,
		boundedBelow: boundedBelow,
		boundedAbove: boundedAbove,
	}, nil
}

// Sample takes a sample from within the spaces bounds
func (b *Box) Sample() []*mat.VecDense {
	sample := b.rng.Rand(nil)
	return []*mat.VecDense{mat.NewVecDense(len(sample), sample)}
}

// Contains returns whether in is in the space. The argument in must
// be either a []float64 or *mat.VecDense
func (b *Box) Contains(in interface{}) bool {
	x, ok := in.([]float64)
	if !ok {
		vec, ok := in.(*mat.VecDense)
		if !ok {
			return false
		}
		x = vec.RawVector().Data
	}
	if len(x) != b.Low()[0].Len() {
		return false
	}

	for i := range x {
		if x[i] < b.Low()[0].AtVec(i) || x[i] > b.High()[0].AtVec(i) {
			return false
		}
	}
	return true
}

// High returns the upper bounds of the space
func (b *Box) High() []*mat.VecDense {
	return []*mat.VecDense{b.high}
}

// Low returns the lower bounds of the space
func (b *Box) Low() []*mat.VecDense {
	return []*mat.VecDense{b.low}
}

// BoundedAbove returns whether the space is bounded above
func (b *Box) BoundedAbove() []bool {
	return b.boundedAbove
}

// Bounded below returns whether the space is bounded below
func (b *Box) BoundedBelow() []bool {
	return b.boundedBelow
}
