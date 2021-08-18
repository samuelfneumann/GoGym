package gogym

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

// TupleSpace implements a tuple (i.e., product) of simpler spaces
//
// A TupleSpace treats all the spaces it contains in a recursive manner.
// For example, when calling the High() method, the TupleSpace calls
// each of its contained spaces' High() methods and returns a
// []*mat.VecDense resulting from the call to High() on all embedded
// spaces *in recursive order*.
type TupleSpace struct {
	spaces []Space
}

// NewTupleSpace takes a Python TupleSpace and converts it to its Go
// equivalent
func NewTupleSpace(space *python.PyObject) (Space, error) {
	if !(space.Type() == tupleSpace) {
		return nil, fmt.Errorf("newTupleSpace: space is not a tuple space")
	}

	tupleSpace := space.GetAttrString("spaces")
	defer tupleSpace.DecRef()
	if tupleSpace == nil {
		return nil, fmt.Errorf("newTupleSpace: nil tuple of Python spaces")
	}

	spaces := make([]Space, tupleSpace.Length())
	for i := range spaces {
		space := python.PyTuple_GetItem(tupleSpace, i)
		space.IncRef()
		defer space.DecRef()

		value, err := FromPythonSpace(space)
		if err != nil {
			return nil, fmt.Errorf("newTupleSpace: could not convert space: %v",
				err)
		}
		spaces[i] = value
	}

	return &TupleSpace{spaces}, nil
}

// Seed seeds the RNG for all sub-spaces recursively
func (t *TupleSpace) Seed(seed uint64) {
	for _, space := range t.spaces {
		space.Seed(seed)
	}
}

// Low returns the lower bounds of the space. If a composite space
// exists in the TupleSpace, its Low() method is called recursively, and
// all lower bounds are placed in the returned slice sequentially.
func (t *TupleSpace) Low() []*mat.VecDense {
	low := make([]*mat.VecDense, 0, t.Len())

	for _, space := range t.spaces {
		low = append(low, space.Low()...)
	}
	return low
}

// High returns the upper bounds of the space. If a composite space
// exists in the TupleSpace, its High() method is called recursively, and
// all upper bounds are placed in the returned slice sequentially.
func (t *TupleSpace) High() []*mat.VecDense {
	high := make([]*mat.VecDense, 0, t.Len())

	for _, space := range t.spaces {
		high = append(high, space.High()...)
	}
	return high
}

// Contains returns whether in is in the space. The argument in must
// be a []interface{}. Each element of in must be compatible with the
// corresponding element in the tuple space. For example, if the
// tuple space has a box space at index i, then in[i] should be a
// float64, not a map[string]interface{}. If the tuple space has a
// dict space at index i, then in[i] should be a map[string]interface{},
// not a float64.
func (t *TupleSpace) Contains(in interface{}) bool {
	x, ok := in.([]interface{})
	if !ok {
		return false
	}

	if len(x) != t.Len() {
		return false
	}

	for i := range x {
		if !t.spaces[i].Contains(x[i]) {
			return false
		}
	}
	return true
}

// Sample takes a sample from within the space bounds of each space in
// the tuple space. If a composite space exists in the TupleSpace,
// then its Sample() method is (possibly recursively) called, and all
// samples are placed in the returned slice in sequential order.
func (t *TupleSpace) Sample() []*mat.VecDense {
	sample := make([]*mat.VecDense, 0, t.Len())

	for _, space := range t.spaces {
		sample = append(sample, space.Sample()...)
	}
	return sample
}

func (t *TupleSpace) Len() int {
	return len(t.spaces)
}

// At returns the Space in the TupleSpace at index i
func (t *TupleSpace) At(i int) Space {
	return t.spaces[i]
}
