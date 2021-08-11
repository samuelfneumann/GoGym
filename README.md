# GoGym: OpenAI Gym in Go

This Go module provides functionality to access OpenAI Gym in Go using `cgo` and the C `Python` API. The API has been kept very similar to the API of OpenAI Gym, except that certain aspects have been made more "Go-like". For example, the `make` factory has been changed to be called 	`New`, to keep consistent with the Go convention of constructor names. Additionally, many functions return Go `error`s, when the OpenAI Gym API does not.
