# arara

A command-line interface (CLI) to create install scripts. Initially, aiming at dotfile's automated installations.

Using: latest [Bonzai](https://github.com/rwxrob/bonzai) - v0.56.6.

## Go Install

```bash
go install github.com/BuddhiLW/arara/cmd/arara@latest
```

## Tests

``` bash
go test ./...
```


1. config package:
    - Tests for YAML configuration loading
    - Test cases for valid config, invalid path, and invalid YAML
2. dotfiles package:
    - Tests for the constructor function
    - Tests for CreateSymlink functionality
    - Test cases for creating new symlinks, overwriting existing symlinks, and creating parent directories
3. backup command:
    - Tests for backing up directories
    - Verifies directories are moved and backup directory is created
4. link command:
    - Tests for creating symlinks for various dotfiles
    - Verifies symlinks are correctly created and point to the right targets
5. build package:
    - Tests for the list command
    - Mock implementation for the install command (not actively tested)
6. setup command:
    - Tests for the command structure
    - Verifies it has the expected subcommands
7. app root command:
    - Tests for the root command structure
    - Verifies it has the expected subcommands


