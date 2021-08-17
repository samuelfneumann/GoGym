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

// BoxSpace represents a (possibly unbounded) box in R^n. Specifically, a
// BoxSpace represents the Cartesian product of n closed intervals. Each
// interval has the form of one of [a, b], (-∞, b], [a, ∞), or
// (-∞, ∞) for a, b ϵ R.
//
// The BoxSpace space must be constructed from its Python equivalent.
type BoxSpace struct {
	*python.PyObject // BoxSpace Space
	rng              *distmv.Uniform
	rand.Source
	low, high                  *mat.VecDense
	shape                      []int
	boundedBelow, boundedAbove []bool
}

// NewBoxSpace takes a Python gym.spaces.BoxSpace and converts it into its Go
// counterpart.
func NewBoxSpace(space *python.PyObject) (Space, error) {
	if !(space.Type() == boxSpace) {
		return nil, fmt.Errorf("newBoxSpace: space is not a box space")
	}

	// Shape
	shape := space.GetAttrString("shape")
	defer shape.DecRef()
	if shape == nil {
		return nil, fmt.Errorf("newBoxSpace: space %v is not a BoxSpace",
			space.Type())
	}
	goShape, err := IntSliceFromIter(shape)
	if err != nil {
		return nil, fmt.Errorf("newBoxShape: could not compute shape: %v", err)
	}

	// Lower bounds
	low := space.GetAttrString("low")
	defer low.DecRef()
	if low == nil {
		return nil, fmt.Errorf("newBoxSpace: space %v is not a BoxSpace",
			space.Type())
	}
	goLow, err := F64SliceFromIter(low)
	if err != nil {
		return nil, fmt.Errorf("newBoxSpace: could not compute lower bound: %v",
			err)
	}

	// Upper bounds
	high := space.GetAttrString("high")
	defer high.DecRef()
	if high == nil {
		return nil, fmt.Errorf("newBoxSpace: space %v is not a BoxSpace",
			space.Type())
	}
	goHigh, err := F64SliceFromIter(high)
	if err != nil {
		return nil, fmt.Errorf("newBoxSpace: could not compute upper bound: %v",
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

	return &BoxSpace{
		PyObject:     space,
		low:          mat.NewVecDense(len(goLow), goLow),
		high:         mat.NewVecDense(len(goHigh), goHigh),
		shape:        goShape,
		rng:          rng,
		Source:       src,
		boundedBelow: boundedBelow,
		boundedAbove: boundedAbove,
	}, nil
}

// Sample takes a sample from within the space bounds
func (b *BoxSpace) Sample() []*mat.VecDense {
	sample := b.rng.Rand(nil)
	return []*mat.VecDense{mat.NewVecDense(len(sample), sample)}
}

// Contains returns whether in is in the space. The argument in must
// be either a []float64 or *mat.VecDense
func (b *BoxSpace) Contains(in interface{}) bool {
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
func (b *BoxSpace) High() []*mat.VecDense {
	return []*mat.VecDense{b.high}
}

// Low returns the lower bounds of the space
func (b *BoxSpace) Low() []*mat.VecDense {
	return []*mat.VecDense{b.low}
}

// BoundedAbove returns whether the space is bounded above
func (b *BoxSpace) BoundedAbove() []bool {
	return b.boundedAbove
}

// Bounded below returns whether the space is bounded below
func (b *BoxSpace) BoundedBelow() []bool {
	return b.boundedBelow
}
