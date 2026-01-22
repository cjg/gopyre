# GoPyre

> **GoPyre is an arcane bridge that lets Go invoke Python scripts without the forbidden magic of CGO.**

GoPyre is a Go library that allows Go programs to execute Python scripts by dynamically loading `libpython` at runtime, without using CGO. It is designed for concurrent Go applications and provides a simple JSON-based data exchange model between Go and Python.

## Overview

GoPyre enables embedding Python execution into Go programs while remaining a **pure Go module**. Instead of relying on CGO, it uses `dlopen`/`dlsym` to interface with CPython directly.

The library is suitable for applications that need to:

- Run small Python scripts from Go
- Exchange structured data using JSON
- Maintain Go-native concurrency with minimal Python-related boilerplate

## Features

- **No CGO dependency**  
  Uses dynamic loading of `libpython` via `dlopen`/`dlsym`.

- **CPython 3.9+ support**  
  Tested with CPython ≥ 3.9.

- **Concurrent execution**  
  Designed to support per-goroutine Python interpreters.

- **Hidden GIL management**  
  GIL handling is fully abstracted from the user.

- **JSON-based data exchange**  
  Pass Go values as JSON-serializable data and receive results the same way.

- **Clean execution model**  
  - Inline Python scripts only  
  - Fresh global namespace for every execution

- **Idiomatic Go errors**  
  Python exceptions are converted into Go `error` values.

- **Configurable I/O**  
  Python `stdout` and `stderr` are forwarded to the host application by default and can be redirected.

## Supported Platforms

- Unix-like systems
  - Linux
  - macOS

## Installation

```bash
go get github.com/cjg/gopyre
```

### Runtime Requirements

- CPython ≥ 3.9 installed on the system
- A dynamically linkable `libpython` available at runtime

No `python3-dev` or compilation toolchain is required.

## Configuration

GoPyre locates `libpython` using environment variables.

Example:

```bash
export GOPYRE_LIBPYTHON=/usr/lib/libpython3.11.so
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/<your-org>/gopyre"
)

func main() {
    result, err := gopyre.Exec(`
x = input["x"]
y = input["y"]
x + y
`, map[string]any{
        "x": 2,
        "y": 3,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result)
}
```

## Data Exchange Model

- Input variables are provided as `map[string]any`
- Available in Python as `input`
- The last expression is returned

| Python | Go |
|------|----|
| `None` | `nil` |
| `dict` | `map[string]any` |
| `list` | `[]any` |

## Error Handling

Python exceptions are converted into Go errors.

## Concurrency

Designed for concurrent goroutine usage with internal GIL handling.

## License

MIT
