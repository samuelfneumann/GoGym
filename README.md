# GoGym: Go Bindings for OpenAI Gym

This Go module provides functionality to access [OpenAI Gym](https://github.com/openai/gym) in Go using `cgo` and the C `Python` API. The API has been kept very similar to the API of OpenAI Gym, except that certain aspects have been made more "Go-like". For example, the `make` factory has been changed to be called 	`New`, to keep consistent with the Go convention of constructor names. Another example is that action and observation spaces are unexported, and they must be accessed through their respective getter methods. Additionally, many functions return Go `error`s, when the OpenAI Gym API does not.

This module simply provides `Go` bindings for OpenAI Gym. The module uses an embedded `Python` interpreter in `Go` code, so the actual gym code running under-the-hood is still `Python`. So don't expect `Go`-level performance. If you wanted reinforcement learning environments implemented completely in `Go`, see my [GoLearn: Reinforcement Learning in Go](https://github.com/samuelfneumann/GoLearn) module.

Currently, only the default OpenAI Gym environments are implemented. You cannot yet use the OpenAI Gym wrappers to decorate environments. Coming soon!

# Installation and Dependencies
This package has the following dependencies:
* [OpenAI Gym](https://github.com/openai/gym)
    * Along with all its dependencies
    * A MuJoCo license if you intend to use MuJoCo environments
    * [Atari-Py](https://pypi.org/project/atari-py/) and Atari ROMs if you intend to use Atari environments
    * [Box2D](https://pypi.org/project/Box2D/)
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

# Known Issues
* The rendering functionality of OpenAI Gym is currently not supported. For some reason the `C Python API` cannot find the `gym.error` package.
* You may need to link the `Python 3.7` library for `cgo`: `#cgo LDFLAGS: -lpython3` or `#cgo LDFLAGS: -lpython3.7`
* If using many environments concurrently in the same process, the dreaded `Python` GIL will ensure that performance decreases. Try to limit the number of environments per-process to 1 to ensure the best performance (in fact, this limitation exists when running OpenAI Gym in `Python` too).
* Since `Go-Python` provides bindings only for the `Python C API` and not the `NumPy C API`, `Python` `List`s are passed as actions to the `gym` environments instead of `NumPy` `ndarray`s. Casting the `Python` `List`s to `NumPy` `ndarray`s would just be an extra unneeded step.
* So far, only Gym environments which satisfy the *regular* Gym interface (having `Step()`, `Reset()`, and `Seed()` methods) can be constructed. Any others (e.g. the *Algorithmic Environments*) will result in a panic. This means that MuJoCo, classic control, and Atari should work.
* There is an issue with the reference counts (they're off), so that at one time only 3 environments can be constructed.

# Future plans
- [ ] Make all environments work, even those that do not implement the *regular* environment API
- [ ] Add all wrappers
- [ ] Add all spaces