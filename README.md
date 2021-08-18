# GoGym: Go Bindings for OpenAI Gym

This `Go` module provides functionality to access [OpenAI Gym](https://github.com/openai/gym) in Go using `cgo` and the C `Python` API. The API has been kept very similar to the API of OpenAI Gym, except that certain aspects have been made more "Go-like". For example, action and observation spaces are unexported, and they must be accessed through their respective getter methods. Additionally, many functions return Go `error`s, when the OpenAI Gym API does not.

This module simply provides `Go` bindings for OpenAI Gym. The module uses an embedded `Python` interpreter in `Go` code, so the actual gym code running under-the-hood is still `Python`. Don't expect `Go`-level performance. If you wanted reinforcement learning environments implemented completely in `Go`, see my [GoLearn: Reinforcement Learning in Go](https://github.com/samuelfneumann/GoLearn) module.

**Current State**: Classic control, MuJoCo, and Atari environments work as returned by `gym.make()` in `Python`. Only the `ClipAction` wrapper is implemented fully. Rendering of environments, either through the `Render()` method or through `PixelObservationWrapper`s does not work. Environments must either have `Box` or `Discrete` observation and action spaces. Other spaces have not been implemented. These environments will still work, you just won't be able to inspect their observation or action spaces with the `ObservationSpace()` and `ActionSpace()` methods respectively.

If all you need is to be able to call the `Python` functions/methods `gym.make()`, `env.step()`, `env.reset()`, and `env.seed()`, then you can consider this module exactly what you need. If you need some of the fancier Open AI Gym tools, like all their wrappers, stay tuned! Those are soon to come!

Currently, any wrappers that deal with multi-dimensional arrays are not supported. This includes `PixelObservationWrapper`s and `FrameStack` wrappers.

# Installation and Dependencies
This package has the following dependencies:
* [OpenAI Gym](https://github.com/openai/gym)
    * Along with all its dependencies
    * A MuJoCo license if you intend to use MuJoCo environments
    * [Atari-Py](https://pypi.org/project/atari-py/) and Atari ROMs if you intend to use Atari environments
    * [PyBox2D](https://pypi.org/project/Box2D/)
* `Python 3.7` (currently tested with [Python 3.7.9](https://www.python.org/downloads/release/python-379/))
* `Python 3.7-dev` (automatically installed when installing `Python3.7` from source)
* [Go-Python](https://github.com/DataDog/go-python3): Go bindings for the C-API of CPython3
    * Along with all its dependencies
* [pkg-config](https://en.wikipedia.org/wiki/Pkg-config#:~:text=pkg%2Dconfig%20is%20a%20computer,of%20detailed%20library%20path%20information)

The pkg-config program looks at the paths described in the `PKG_CONFIG_PATH` environment variable. In one of those paths, you must have have the `python3.pc` package configuration file. On my Ubuntu installation, the `python3.pc` file is in `/usr/local/lib/pkgconfig`. To add this directory to the `PKG_CONFIG_PATH` environment variable, run the following code on the command line:
```
export PKG_CONFIG_PATH="$PKG_CONFIG_PATH":/usr/local/lib/pkgconfig
```
or put that line in your `.zshrc` or `'bashrc` file for the environment variable to be automatically set whenever a terminal is started.

Once all dependencies have been installed, you're ready to start using `GoGym`!

**Warning**: this module **only** works with `Python 3.7`. No other version of `Python` is currently supported by `Go-Python`.

## Installing `Python3.7` from Source
To install `Python 3.7` (along with the `Python3.7-dev` package) from source:

1. ead over to the [Python 3.7 Download Page](https://www.python.org/downloads/release/python-379/) and download one of the compressed archived.
2. Extract the archive.
3. Enter the extracted directory
4. Run `./configure --enable-shared --enable-optimizations`
5. Run `sudo make install`
6. Enjoy your new `Python 3.7` installation! Why not install `gnureadline`?


# Example Usage
```
env, err := Make("Ant-v2")
if err != nil {
	panic(err)
}

_, err = env.Reset()
if err != nil {
	panic(err)
}

for i := 0; i < 10; i++ {
	obs, reward, done, err := env.Step(env.ActionSpace().Sample())
	if err != nil {
		panic(err)
	}
	fmt.Println(obs, reward, done)
}
```


# Known Issues
* The rendering functionality of OpenAI Gym is currently not supported. For some reason the `C Python API` cannot find the `gym.error` package.
* You may need to link the `Python 3.7` library for `cgo`: `#cgo LDFLAGS: -lpython3` or `#cgo LDFLAGS: -lpython3.7`
* If using many environments concurrently in the same process, the dreaded `Python` GIL will ensure that performance decreases. Try to limit the number of environments per-process to 1 to ensure the best performance (in fact, this limitation exists when running OpenAI Gym in `Python` too).
* Since `Go-Python` provides bindings only for the `Python C API` and not the `NumPy C API`, `Python` `List`s are passed as actions to the `gym` environments instead of `NumPy` `ndarray`s. Casting the `Python` `List`s to `NumPy` `ndarray`s would just be an extra unneeded step.
* So far, only Gym environments which satisfy the *regular* Gym interface (having `Step()`, `Reset()`, and `Seed()` methods) can be constructed. Any others (e.g. the *Algorithmic Environments*) will result in a panic. This means that MuJoCo, classic control, and Atari should work.

# Future plans
- [ ] Make all environments work, even those that do not implement the *regular* environment API
- [ ] Add all wrappers
- [ ] Add all spaces

# ToDo
- [ ] Depending on the observation space type, Step() should construct the appropriate observation (vector, dict, tuple) and return a structure of that type. E.g. if the environment is wrapped by a PixelObservation wrapper, then the returned observation is actually a Python dict[string]np.array, so we should also do this. Step() will return an interface{}. Then, for a composite type: for each index we construct an associated value (e.g. if observation["pixels"] is a np.array, we return a []float64 at that index) etc.
