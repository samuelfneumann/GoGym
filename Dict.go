package gogym

import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

// DictSpace implements a dictionary of simpler spaces
type DictSpace map[string]Space

// NewDictSpace takes a Python gym.spaces.Dict and converts it into its Go
// counterpart.
func NewDictSpace(dictSpace *python.PyObject) (Space, error) {
	dictSpaces := dictSpace.GetAttrString("spaces")
	defer spaces.DecRef()
	if spaces == nil || !python.PyDict_Check(dictSpaces) {
		return nil, fmt.Errorf("newDictSpace: space is not a DictSpace")
	}

	// Get the keys in the Python Dict space
	keys := python.PyDict_Items(dictSpaces)
	defer keys.DecRef()
	if keys == nil {
		return nil, fmt.Errorf("newDictSpace: no keys in DictSpace")
	}

	// Convert keys to strings
	goKeys, err := StringSliceFromIter(keys)
	if err != nil {
		return nil, fmt.Errorf("newDictSpace: could not decode keys: %v",
			err)
	}

	// Convert to DictSpace
	goDictSpace := make(DictSpace)
	for _, key := range goKeys {
		spaceAtKey := python.PyDict_GetItemString(dictSpaces, key)

		var value Space
		switch spaceAtKey.Type() {
		case boxSpace:
			value, err = NewBox(spaceAtKey)

		case discreteSpace:
			value, err = NewDiscrete(spaceAtKey)

		case dictSpace:
			value, err = NewDictSpace(spaceAtKey)

		default:
			return nil, fmt.Errorf("newDictSpace: space %v not yet "+
				"implemented", spaceAtKey.Type())
		}
		if err != nil {
			return nil, fmt.Errorf("newDictSpace: could not convert space: %v",
				err)
		}
		goDictSpace[key] = value
	}

	return goDictSpace, nil
}

// Seed seeds the RNG for all sub-spaces recursively
func (d DictSpace) Seed(seed uint64) {
	for _, space := range d {
		space.Seed(seed)
	}
}

// Sample takes a sample from within the space bounds. If a composite
// space exists in the DictSpace, then its Sample() method is
// recursively called, and all samples are placed in the returned
// slice sequentially.
func (d DictSpace) Sample() []*mat.VecDense {
	sample := make([]*mat.VecDense, d.Len())

	i := 0
	for _, space := range d {
		switch sampleSpace := space.(type) {
		case *Box:
			sample[i] = sampleSpace.Sample()[0]

		case *Discrete:
			sample[i] = sampleSpace.Sample()[0]

		case DictSpace:
			sample = append(sample, sampleSpace.Sample()...)

		default:
			panic(fmt.Sprintf("sample: cannot sample space type %T", space))
		}
		i++
	}
	return sample
}

// Contains returns whether in is in the space. The argument in must
// be either a map[string]interface{}
func (d DictSpace) Contains(in interface{}) bool {
	x, ok := in.(map[string]interface{})
	if !ok {
		return false
	}

	if len(x) != d.Len() {
		return false
	}

	for key := range d {
		val, ok := x[key]
		if !ok {
			return false
		}
		if !d[key].Contains(val) {
			return false
		}
	}
	return true
}

// Low returns the lower bounds of the space. If a composite space
// exists in the DictSpace, its Low() method is called recursively, and
// all lower bounds are placed in the returned slice sequentially.
func (d DictSpace) Low() []*mat.VecDense {
	low := make([]*mat.VecDense, d.Len())
	i := 0
	for _, space := range d {
		switch lowSpace := space.(type) {
		case *Box:
			low[i] = lowSpace.Low()[0]

		case *Discrete:
			low[i] = lowSpace.Low()[0]

		case DictSpace:
			low = append(low, lowSpace.Low()...)

		default:
			panic(fmt.Sprintf("low: cannot compute lower bound of space "+
				"type %T", space))
		}
		i++
	}
	return low
}

// High returns the upper bounds of the space. If a composite space
// exists in the DictSpace, its High() method is called recursively, and
// all upper bounds are placed in the returned slice sequentially.
func (d DictSpace) High() []*mat.VecDense {
	high := make([]*mat.VecDense, d.Len())
	i := 0
	for _, space := range d {
		switch highSpace := space.(type) {
		case *Box:
			high[i] = highSpace.High()[0]

		case *Discrete:
			high[i] = highSpace.High()[0]

		case DictSpace:
			high = append(high, highSpace.High()...)

		default:
			panic(fmt.Sprintf("high: cannot compute upper bound of space "+
				"type %T", space))
		}
		i++
	}
	return high
}

// Len returns the number of sub-spaces in the space
func (d *DictSpace) Len() int {
	return len(map[string]Space(*d))
}
