# Arara Compatibility System

The compatibility system provides a way to define and check system requirements for scripts and other components in Arara.

## Overview

This package allows scripts in the Arara configuration to define compatibility requirements, such as operating system, architecture, shell, or package manager availability. The system evaluates these conditions against the current environment to determine if a script should be run.

## Usage

### Basic Usage

In your `arara.yaml` configuration file, you can define compatibility requirements for scripts:

```yaml
scripts:
  install:
    - name: docker
      path: scripts/install/docker
      compat:
        os: debian
        shell: bash
        pkgmgr: apt
```

### Built-in Validators

The following validators are built in:

| Validator | Description | Example Values |
|-----------|-------------|---------------|
| `os` | Operating system | `debian`, `ubuntu`, `darwin` |
| `arch` | CPU architecture | `amd64`, `arm64` |
| `shell` | Current shell | `bash`, `zsh` |
| `pkgmgr` | Package manager | `apt`, `yum`, `pacman` |
| `kernel` | Linux kernel version | `5.10`, `4.19` |

### Custom Validators

You can define and register custom validators to extend the compatibility system.

## Extending the System

### Creating a Custom Validator

1. Implement the `CustomValidator` interface:

```go
type MyValidator struct{}

func (v *MyValidator) Name() string {
    return "my-validator"
}

func (v *MyValidator) Validate(value interface{}) bool {
    // Implement your validation logic
    return true
}
```

2. Register your validator during application initialization:

```go
func init() {
    myValidator := &MyValidator{}
    compat.RegisterCustomValidator(myValidator)
}
```

3. Use your validator in `arara.yaml`:

```yaml
scripts:
  install:
    - name: my-script
      path: scripts/install/my-script
      compat:
        custom:
          - name: my-validator
            value: some-value
          - another-validator
```

### Plugin System

For a more decoupled approach, you can use the plugin system to allow third-party packages to register validators without modifying the core code.

To create a plugin:

1. Create a Go package that imports the `compat` package
2. Implement the `CustomValidator` interface
3. Register the validator in the package's `init()` function
4. Import the plugin in your application

Example plugin package:

```go
package myplugin

import "github.com/BuddhiLW/arara/internal/app/compat"

type GpuValidator struct{}

func (v *GpuValidator) Name() string {
    return "has-gpu"
}

func (v *GpuValidator) Validate(value interface{}) bool {
    // Check if GPU is available
    return checkGpuAvailability()
}

func init() {
    compat.RegisterCustomValidator(&GpuValidator{})
}
```

Then import the plugin in your main.go:

```go
import (
    _ "github.com/example/arara-gpu-plugin" // Import for side effects
)
```

## Advanced Usage

### Validator with Configuration

You can create validators that accept configuration:

```go
type MinMemValidator struct{}

func (v *MinMemValidator) Name() string {
    return "min-memory"
}

func (v *MinMemValidator) Validate(value interface{}) bool {
    // Value should be memory size in MB
    requiredMB, ok := value.(float64)
    if !ok {
        return false
    }
    
    // Get available memory
    availableMB := getSystemMemoryMB()
    
    return availableMB >= requiredMB
}
```

And use it in the YAML:

```yaml
compat:
  custom:
    - name: min-memory
      value: 4096  # Require at least 4GB RAM
```

## Internal Architecture

The compatibility system uses two registries:

1. `validatorRegistry` - Stores standard validators for basic fields
2. `customRegistry` - Stores custom validators that implement the `CustomValidator` interface

The `Check` function evaluates all conditions in a `CompatSpec` struct, returning `true` only if all conditions are met.