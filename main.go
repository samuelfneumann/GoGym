package main

// #cgo LDFLAGS: -lpython3  // This may not be needed

// Before running, ensure python-3.7.pc is in a directory pointed to
// by PKG_CONFIG_PATH. On Ubuntu:
//
// export PKG_CONFIG_PATH="$PKG_CONFIG_PATH":/usr/local/lib/pkgconfig

// #cgo pkg-config: python-3.7
// #include <Python.h>
import "C"
import (
	"fmt"

	python "github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

var gym *python.PyObject
var numpy *python.PyObject

type GymEnv struct {
	env              *python.PyObject
	envName          string
	continuousAction bool
}

func New(envName string) (*GymEnv, error) {
	// Get the gym.make function
	dict := python.PyModule_GetDict(gym)
	defer dict.DecRef()
	makeEnv := python.PyDict_GetItemString(dict, "make")
	defer makeEnv.DecRef()
	if !(makeEnv != nil && python.PyCallable_Check(makeEnv)) {
		return nil, fmt.Errorf("make: error creating env %v", envName)
	}
	fmt.Println("makeEnv", python.PyUnicode_AsUTF8(makeEnv.Str()))

	// Construct the arguments to the gym.make function
	args := python.PyTuple_New(1)
	python.PyTuple_SetItem(args, 0, python.PyUnicode_FromString(envName))
	fmt.Println("args:", python.PyUnicode_AsUTF8(python.PyTuple_GetItem(args, 0)))

	// Create the gym environment
	env := makeEnv.CallObject(args)
	fmt.Println("Env:", python.PyUnicode_AsUTF8(env.Str()))
	if env == nil {
		return nil, fmt.Errorf("make: could not create call gym.make")
	}

	// Figure out if the environment has continuous actions or not
	spaces := python.PyDict_GetItemString(dict, "spaces")
	defer spaces.DecRef()
	actionSpace := env.GetAttrString("action_space")
	defer actionSpace.DecRef()
	continuousAction := actionSpace.Type() == spaces.GetAttrString("Box")

	return &GymEnv{env, envName, continuousAction}, nil
}

func (g *GymEnv) Seed(seed int) ([]int, error) {
	// Get the seed function
	seedFunc := g.env.GetAttrString("seed")
	defer seedFunc.DecRef()

	// Create the Python arguments
	args := python.PyTuple_New(1)
	defer args.DecRef()
	python.PyTuple_SetItem(args, 0, python.PyLong_FromGoInt(seed))

	// Seed the environment
	retVal := seedFunc.CallObject(args)
	defer retVal.DecRef()
	if retVal == nil {
		return nil, fmt.Errorf("seed: no seed returned from gym")
	}

	// Return the seed
	s, err := IntSliceFromIter(retVal)
	if err != nil {
		return nil, fmt.Errorf("seed: could not convert seed to Go: %v", err)
	}
	return s, nil
}

func (g *GymEnv) Step(a *mat.VecDense) (*mat.VecDense, float64, bool,
	error) {
	// Get the step function
	stepFunc := g.env.GetAttrString("step")
	defer stepFunc.DecRef()
	fmt.Printf("StepFunc: ")
	print(stepFunc)

	// Create the Python arguments
	args := python.PyTuple_New(1)
	defer args.DecRef()
	if g.continuousAction {
		arr, err := F64ToList(a.RawVector().Data)
		if err != nil {
			return nil, 0, false, fmt.Errorf("step: could not convert " +
				"Python list to NumPy array")
		}
		python.PyTuple_SetItem(args, 0, arr)
	} else {
		python.PyTuple_SetItem(args, 0, python.PyLong_FromDouble(a.AtVec(0)))
	}

	retVal := stepFunc.CallObject(args)
	defer retVal.DecRef()
	if retVal == nil {
		return nil, 0, false, fmt.Errorf("step: could not step in " +
			"gym environment")
	}

	obs := python.PyTuple_GetItem(retVal, 0)
	goObsSlice, err := F64SliceFromIter(obs)
	if err != nil {
		return nil, 0, false, fmt.Errorf("step: could not decode observation")
	}
	goObs := mat.NewVecDense(len(goObsSlice), goObsSlice)
	fmt.Println("Observation:", goObs)

	reward := python.PyTuple_GetItem(retVal, 1)
	goReward := python.PyFloat_AsDouble(reward)
	fmt.Println("Reward:", goReward)

	done := python.PyTuple_GetItem(retVal, 2)
	goDone := python.Py_True == done
	fmt.Println("Done:", goDone)

	return goObs, goReward, goDone, nil
}

func F64ToList(slice []float64) (*python.PyObject, error) {
	list := python.PyList_New(len(slice))
	for i, elem := range slice {
		float := python.PyFloat_FromDouble(elem)
		n := python.PyList_SetItem(list, i, float)
		if n != 0 {
			if python.PyErr_Occurred() != nil {
				python.PyErr_Print()
			}
			float.DecRef()
			list.DecRef()
			return nil, fmt.Errorf("f64ToList: could not set Python list item")
		}
	}

	fmt.Println()
	fmt.Printf("LIST: ")
	print(list)
	fmt.Println()
	return list, nil

	// dict := python.PyModule_GetDict(numpy)
	// defer dict.DecRef()
	// toArray := python.PyDict_GetItemString(dict, "array")
	// if toArray == nil {
	// 	return nil, fmt.Errorf("f64ToNumpy: could not get numpy.array function")
	// }

	// args := python.PyTuple_New(len(slice))
	// python.PyTuple_SetItem(args, 0, list)

	// arr := toArray.CallObject(args)
	// if arr == nil {
	// 	return nil, fmt.Errorf("f64ToNumpy: could not convert Python list " +
	// 		"to NumPy array")
	// }
	// return arr, nil
}

func (g *GymEnv) Reset() (*mat.VecDense, error) {
	reset := g.env.GetAttrString("reset")
	defer reset.DecRef()

	out := reset.CallObject(nil)
	defer out.DecRef()
	fmt.Println("out:", python.PyUnicode_AsUTF8(out.Str()))

	data, err := F64SliceFromIter(out)
	if err != nil {
		return nil, fmt.Errorf("reset: could not decode Python iterable: %v",
			err)
	}
	return mat.NewVecDense(len(data), data), nil
}

func (g *GymEnv) Close() {
	g.env.DecRef()
}

func main() {
	python.Py_Initialize()
	defer python.Py_Finalize()

	python.PyRun_SimpleString(`print("Hello, world")`)

	o := python.PyImport_ImportModule("gym")
	if o == nil {
		panic("ah")
	}
	defer o.DecRef()
	gym = python.PyImport_AddModule("gym")

	o2 := python.PyImport_ImportModule("numpy")
	if o2 == nil {
		panic("NO")
	}
	defer o2.DecRef()
	numpy = python.PyImport_AddModule("numpy")

	env, err := New("MountainCarContinuous-v0")
	if err != nil {
		panic(err)
	}
	fmt.Println("ENV:", env)

	data, err := env.Reset()
	if err != nil {
		panic(err)
	}
	fmt.Println("Reset:", data)

	// seed, err := env.Seed(10)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Seed:", seed)

	obs, reward, done, err := env.Step(mat.NewVecDense(1, []float64{0.0}))
	if err != nil {
		panic(err)
	}
	fmt.Println(obs, reward, done)
}

func GoSliceFromPyList(pylist *python.PyObject, itemtype string, strictfail bool) (interface{}, error) {
	seq := pylist.GetIter() //ret val: New reference
	if !(seq != nil && python.PyErr_Occurred() == nil) {
		python.PyErr_Print()
		return nil, fmt.Errorf("error creating iterator for list")
	}
	defer seq.DecRef()
	tNext := seq.GetAttrString("__next__") //ret val: new ref
	if !(tNext != nil && python.PyCallable_Check(tNext)) {
		return nil, fmt.Errorf("iterator has no __next__ function")
	}
	defer tNext.DecRef()

	var golist interface{}
	var compare *python.PyObject
	switch itemtype {
	case "float64":
		golist = []float64{}
		compare = python.PyFloat_FromDouble(0)
	case "int":
		golist = []int{}
		compare = python.PyLong_FromGoInt(0)
	}
	if compare == nil {
		return nil, fmt.Errorf("error creating compare var")
	}
	defer compare.DecRef()

	pytype := compare.Type() //ret val: new ref
	if pytype == nil && python.PyErr_Occurred() != nil {
		python.PyErr_Print()
		return nil, fmt.Errorf("error getting type of compare var")
	}
	defer pytype.DecRef()

	errcnt := 0

	pylistLen := pylist.Length()
	if pylistLen == -1 {
		return nil, fmt.Errorf("error getting list length")
	}

	for i := 1; i <= pylistLen; i++ {
		item := tNext.CallObject(nil) //ret val: new ref
		if item == nil && python.PyErr_Occurred() != nil {
			python.PyErr_Print()
			return nil, fmt.Errorf("error getting next item in sequence")
		}
		itemType := item.Type()
		if itemType == nil && python.PyErr_Occurred() != nil {
			python.PyErr_Print()
			return nil, fmt.Errorf("error getting item type")
		}

		defer itemType.DecRef()

		if itemType != pytype {
			//item has wrong type, skip it
			if item != nil {
				item.DecRef()
			}
			errcnt++
			continue
		}

		switch itemtype {
		case "float64":
			itemGo := python.PyFloat_AsDouble(item)
			if itemGo != -1 && python.PyErr_Occurred() == nil {
				golist = append(golist.([]float64), itemGo)
			} else {
				if item != nil {
					item.DecRef()
				}
				errcnt++
			}
		case "int":
			itemGo := python.PyLong_AsLong(item)
			if itemGo != -1 && python.PyErr_Occurred() == nil {
				golist = append(golist.([]int), itemGo)
			} else {
				if item != nil {
					item.DecRef()
				}
				errcnt++
			}
		}

		if item != nil {
			item.DecRef()
			item = nil
		}
	}
	if errcnt > 0 {
		if strictfail {
			return nil, fmt.Errorf("could not add %d values (wrong type?)", errcnt)
		}
	}

	return golist, nil
}

// Borrow pyobject ref
func F64SliceFromIter(obj *python.PyObject) ([]float64, error) {
	seq := obj.GetIter()
	defer seq.DecRef()
	next := seq.GetAttrString("__next__")
	defer next.DecRef()

	data := make([]float64, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		item := next.CallObject(nil)
		if item == nil {
			return nil, fmt.Errorf("f64SliceFromIter: nil item at index %v", i)
		}

		data[i] = python.PyFloat_AsDouble(item)
		item.DecRef()
	}

	return data, nil
}

// Borrow pyobject ref
func IntSliceFromIter(obj *python.PyObject) ([]int, error) {
	seq := obj.GetIter()
	defer seq.DecRef()
	next := seq.GetAttrString("__next__")
	defer next.DecRef()

	data := make([]int, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		item := next.CallObject(nil)
		if item == nil {
			return nil, fmt.Errorf("f64SliceFromIter: nil item at index %v", i)
		}

		data[i] = python.PyLong_AsLong(item)
		item.DecRef()
	}

	return data, nil
}

func print(obj *python.PyObject) {
	fmt.Println(python.PyUnicode_AsUTF8(obj.Str()))
}
