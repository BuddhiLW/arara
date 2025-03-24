# Arara Development Guide

## Overview

Arara is a CLI tool for managing dotfiles installation and configuration, built with Bonzai. It provides two main workflows:

1. `build` - Initial dotfiles setup (equivalent to main.sh)
2. `install` - Additional tool installation (from scripts/install/)

## Build & Run Commands

```bash
# Build the binary
go build ./cmd/arara

# Install globally
go install ./cmd/arara

# Run locally
go run ./cmd/arara/main.go

# Run tests
go test ./...
```

## Core Commands

```go
var Cmd = &bonzai.Cmd{
    Name:  `arara`,
    Alias: `ar`,
    Commands: []*bonzai.Cmd{
        buildCmd,    // Initial dotfiles setup
        createCmd,   // Create resources (install scripts, build steps)
        installCmd,  // Install additional tools
        setupCmd,    // Core setup operations
        listCmd,     // List available scripts
        initCmd,     // Initialize new arara.yaml
        helpCmd,     // Show help
    },
}

// Setup subcommands handle core dotfiles operations
var setupCmd = &bonzai.Cmd{
    Name:  `setup`,
    Alias: `s`,
    Commands: []*bonzai.Cmd{
        backupCmd,   // Backup existing dotfiles
        linkCmd,     // Create symlinks
        restoreCmd,  // Restore from backup
    },
}

// Create subcommands for adding new resources
var createCmd = &bonzai.Cmd{
    Name:  `create`,
    Alias: `c`,
    Commands: []*bonzai.Cmd{
        installBinCmd,  // Create installation scripts
        buildStepCmd,   // Add build steps to config
    },
}
```

## Usage Examples

```bash
# Core setup operations
arara setup backup                  # Backup existing dotfiles
arara setup link                    # Create symlinks
arara setup restore <backup-name>   # Restore from backup

# Build operations
arara build list                    # Show build steps
arara build install                 # Execute fresh install

# Install additional tools
arara install docker doom

# Create new resources
arara create install docker         # Create a new installation script
arara create build-step docker "Install Docker" "arara install docker"  # Add a new build step

# List available installation scripts
arara list

# Show help
arara help
```

## Configuration

### arara.yaml Schema

```yaml
name: "dotfiles"
description: "Personal dotfiles configuration"

# Core setup configuration
setup:
  # Directories to backup
  backup_dirs:
    - "$HOME/.config"
    - "$HOME/.local"
  
  # Core symlinks
  core_links:
    - source: "$DOTFILES/.config"
      target: "$HOME/.config"
    - source: "$DOTFILES/.local"
      target: "$HOME/.local"
    - source: "$DOTFILES/.xprofile"
      target: "$HOME/.xprofile"

  # Config file symlinks  
  config_links:
    - source: "$DOTFILES/.bashrc"
      target: "$HOME/.bashrc"
    - source: "$DOTFILES/.vim"
      target: "$HOME/.vim"
    - source: "$DOTFILES/.doom.d"
      target: "$HOME/.doom.d"
    - source: "$DOTFILES/.config/tmux/.tmux.conf"
      target: "$HOME/.tmux.conf"
    - source: "$DOTFILES/.config/vim/.vimrc" 
      target: "$HOME/.vimrc"
    - source: "$DOTFILES/.config/X11/xinitrc"
      target: "$HOME/.xinitrc"

# Build configuration
build:
  # Core system setup steps
  steps:
    - name: "backup"
      description: "Backup existing dotfiles"
      command: "arara setup backup"

    - name: "link"
      description: "Create symlinks"
      command: "arara setup link"

    - name: "xmonad"
      description: "Setup window manager"
      commands:
        - "cd $HOME/.config/xmonad"
        - "rm -rf xmonad xmonad-contrib"
        - "git clone https://github.com/xmonad/xmonad"
        - "git clone https://github.com/xmonad/xmonad-contrib"
        - "curl -sSL https://get.haskellstack.org/ | sh -s - -f"

# Additional installation scripts 
scripts:
  install:
    - name: doom
      description: "Install Doom Emacs"
      path: "scripts/install/doom"
      
    - name: docker
      description: "Install Docker and Docker Desktop"
      path: "scripts/install/docker"

# Environment variables
env:
  DOTFILES: "$HOME/dotfiles"
  SCRIPTS: "$DOTFILES/scripts"
```

## Directory Structure

```
dotfiles/
├── arara.yaml           # Main configuration
├── scripts/
│   ├── setup/          # Core setup scripts (used by build)
│   │   ├── bk-dots     # Backup existing dotfiles
│   │   ├── init        # Initialize environment
│   │   ├── link-config # Create symlinks
│   │   └── xmonad      # Window manager setup
│   │
│   └── install/        # Additional tool installers
│       ├── R
│       ├── anaconda
│       ├── android
│       └── ...
│
├── .config/            # Configuration files
├── doom/              # Doom Emacs config
├── vim/               # Vim config
└── ...
```

## Implementation Details

### Setup Command

The `setup` command provides core dotfiles operations:

1. `backup`:
   - Creates timestamped backup directory
   - Moves existing .config and .local
   - Preserves user's existing configuration

2. `link`:
   - Creates core directory symlinks (.config, .local)
   - Creates config file symlinks (.bashrc, .vimrc, etc)
   - Handles both direct and nested config files

3. `restore`:
   - Lists available backups
   - Restores selected backup
   - Removes current symlinks first

### Build Command

The `build` command now:

1. Uses setup commands for core operations
2. Integrates window manager setup
3. Provides atomic operations
4. Allows resuming failed builds

### Install Command

The `install` command:

1. Manages optional tool installation
2. Handles dependencies between tools
3. Preserves existing configuration
4. Provides installation status

### Create Command

The `create` command:

1. `install`:
   - Creates new installation scripts in scripts/install/
   - Initializes with a shebang line and makes executable
   - Opens the created script in an editor
   - Enables rapid prototyping of new installers

2. `build-step`:
   - Adds new build steps to arara.yaml configuration
   - Properly formats the YAML with correct indentation
   - Can specify name, description, and command(s)
   - Simplifies adding new functionality to the build process

## Code Style Guidelines

- Follow Go standard formatting (gofmt)
- Use descriptive error messages
- Document exported items
- Follow Bonzai patterns
- Use strong typing

## Next Steps

1. ✅ Implement core command structure
2. ✅ Add YAML config parsing
3. ✅ Create script execution engine
4. ✅ Add resource creation commands
5. Add dependency management between install scripts
6. Enhance error handling and recovery
7. Add configuration validation
8. Create comprehensive user documentation
9. Implement plugin system for custom commands

## Project Layout

Following the [standard Go project layout](https://github.com/golang-standards/project-layout), arara is structured as:

```
arara/
├── cmd/
│   └── arara/
│       └── main.go      # Main application entry point
├── internal/
│   ├── app/            # Application core logic
│   │   ├── backup/     # Backup functionality
│   │   ├── link/       # Symlink management
│   │   └── setup/      # Setup operations
│   └── pkg/            # Private shared code
│       ├── config/     # Configuration handling
│       └── utils/      # Common utilities
├── pkg/                # Public libraries
│   └── dotfiles/       # Core dotfiles operations
├── build/             
│   ├── ci/            # CI configuration
│   └── package/       # Build/package configs
├── test/              
│   └── testdata/      # Test fixtures
├── docs/              # Documentation
├── examples/          # Example configurations
└── scripts/           # Build and dev scripts
```

### Key Directories

- `/cmd/arara`: Main application entry point
  - Minimal main.go that imports and executes core logic
  - Command-line interface implementation using Bonzai

- `/internal`: Private application code
  - `app/`: Core application logic
  - `pkg/`: Shared internal utilities
  - Not importable by external projects

- `/pkg`: Public libraries
  - Reusable dotfiles management code
  - Can be imported by other projects
  - Stable API with semantic versioning

- `/build`: Build and packaging
  - CI/CD configurations
  - Package build scripts
  - Release management

### Code Organization

1. Core Logic:
   - Keep cmd/arara/main.go minimal
   - Implement core functionality in internal/app
   - Share common code in internal/pkg
   - Make reusable parts public in pkg/

2. Package Structure:
   - Use meaningful package names
   - Group related functionality
   - Avoid package name collisions
   - Follow Go naming conventions

3. Dependencies:
   - Use Go modules for dependency management
   - Vendor dependencies when needed
   - Keep third-party code separate

4. Testing:
   - Place tests alongside code
   - Use testdata for test fixtures
   - Follow Go testing conventions

### Best Practices

1. Package Organization:
   - Keep packages focused and cohesive
   - Avoid circular dependencies
   - Use internal for private code
   - Make public APIs intentional

2. Code Style:
   - Follow Go formatting conventions
   - Use meaningful variable names
   - Document exported items
   - Write clear error messages

3. Project Management:
   - Use semantic versioning
   - Maintain changelog
   - Document breaking changes
   - Keep dependencies updated

4. Build Process:
   - Automate builds with scripts
   - Use consistent versioning
   - Support multiple platforms
   - Include build documentation


## Bonzai latest version examples

Also shows how to use `yaml` config files.

``` go
package help

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/mark"
	"gopkg.in/yaml.v3"
)

//go:embed style.yaml
var Style []byte

var Cmd = &bonzai.Cmd{
	Name:  `help`,
	Alias: `h|-h|--help|--h|/?`,
	Vers:  `v0.8.0`,
	Short: `display command help`,
	Long: `
		The {{code .Name}} command displays the help information for the
		immediate previous command unless it is passed arguments, in which
		case it resolves the arguments as if they were passed to the
		previous command and the help for the leaf command is displayed
		instead.`,

	Do: func(x *bonzai.Cmd, args ...string) (err error) {

		if len(args) > 0 {
			x, args, err = x.Caller().SeekInit(args...)
		} else {
			x = x.Caller()
		}

		md, err := mark.Bonzai(x)
		if err != nil {
			return err
		}

		// load embedded yaml file and convert to json
		styleMap := map[string]any{}
		if err := yaml.Unmarshal(Style, &styleMap); err != nil {
			return err
		}
		jsonBytes, err := json.Marshal(styleMap)
		if err != nil {
			return err
		}

		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
			glamour.WithStylesFromJSONBytes(jsonBytes),
		)

		if err != nil {
			return fmt.Errorf("developer-error: %v", err)
		}

		rendered, err := renderer.Render(md)
		if err != nil {
			return fmt.Errorf("developer-error: %v", err)
		}

		fmt.Println("\u001b[2J\u001b[H" + rendered)

		return nil
	},
}
```

Extensively use `vars` when needed.

``` go
package kimono

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/comp/completers/git"
	"github.com/rwxrob/bonzai/fn/each"
	"github.com/rwxrob/bonzai/futil"
	"github.com/rwxrob/bonzai/vars"
)

const (
	WorkScopeEnv   = `KIMONO_WORK_SCOPE`
	WorkScopeVar   = `work-scope`
	TagVerPartEnv  = `KIMONO_VERSION_PART`
	TagVerPartVar  = `tag-ver-part`
	TagShortenEnv  = `KIMONO_TAG_SHORTEN`
	TagShortenVar  = `tag-shorten`
	TagRmRemoteEnv = `KIMONO_TAG_RM_REMOTE`
	TagRmRemoteVar = `tag-rm-remote`
	TagPushEnv     = `KIMONO_PUSH_TAG`
	TagPushVar     = `tag-push`
	TidyScopeEnv   = `KIMONO_TIDY_SCOPE`
	TidyScopeVar   = `tidy-scope`
)

var Cmd = &bonzai.Cmd{
	Name:  `kimono`,
	Alias: `kmono|km`,
	Vers:  `v0.7.0`,
	Short: `manage golang monorepos`,
	Long: `
The kimono tool helps manage Go monorepos. It simplifies common monorepo
operations and workflow management.

# Features:
- Toggle go.work files on/off for local development
- Perform coordinated version tagging
- Keep go.mod files tidy across modules
- View dependency graphs and module information
- Track dependent modules and their relationships

# Commands:
- work:     Toggle go.work files for local development
- tidy:     run 'go get -u' and 'go mod tidy' across modules
- tag:      List and coordinate version tagging across modules
- deps:     List and manage module dependencies
- depsonme: List and manage module dependencies
- vars:     View and set configuration variables

Use 'kimono help <command> <subcommand>...' for detailed information
about each command.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		workCmd,
		tidyCmd,
		tagCmd,
		dependenciesCmd,
		dependentsCmd,
		vars.Cmd,
		help.Cmd,
	},
	Def: help.Cmd,
}

var workCmd = &bonzai.Cmd{
	Name:  `work`,
	Alias: `w`,
	Short: `toggle go work files on or off`,
	Long: `
Work command toggles the state of Go workspace files (go.work) between
active (on) and inactive (off) modes. This is useful for managing
monorepo development by toggling Go workspace configurations. The scope
in which to toggle the work files can be configured using either the
'work-scope' variable or the 'KIMONO_WORK_SCOPE' environment variable.

# Arguments
  on  : Renames go.work.off to go.work, enabling the workspace.
  off : Renames go.work to go.work.off, disabling the workspace.

# Environment Variables

- KIMONO_WORK_SCOPE: module|repo|tree (Defaults to "module")
  Configures the scope in which to toggle.
  - module: Toggles the go.work file in the current module.
  - repo: Toggles all go.work files in the monorepo.
  - tree: Toggles go.work files in the directory tree starting from pwd.
`,
	Vars: bonzai.Vars{
		{
			K:     WorkScopeVar,
			V:     `module`,
			Env:   WorkScopeVar,
			Short: `Configures the scope in which to toggle work files`,
		},
	},
	NumArgs:  1,
	RegxArgs: `on|off`,
	Opts:     `on|off`,
	Comp:     comp.CmdsOpts,
	Cmds: []*bonzai.Cmd{
		workInitCmd,
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		root := ``
		var err error
		var from, to string
		invArgsErr := fmt.Errorf("invalid arguments: %s", args[0])
		switch args[0] {
		case `on`:
			from = `go.work.off`
			to = `go.work`
		case `off`:
			from = `go.work`
			to = `go.work.off`
		default:
			return invArgsErr
		}
		// FIXME: the default here could come from Env or Vars.
		scope := vars.Fetch(WorkScopeEnv, WorkScopeVar, `module`)
		switch scope {
		case `module`:
			return WorkToggleModule(from, to)
		case `repo`:
			root, err = getGitRoot()
			if err != nil {
				return err
			}
		case `tree`:
			root, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		return WorkToggleRecursive(root, from, to)
	},
}

var workInitCmd = &bonzai.Cmd{
	Name:  `init`,
	Alias: `i`,
	Short: `new go.work in module using dependencies from monorepo`,
	Long: `
The "init" subcommand initializes a new Go workspace file (go.work) 
for the current module. It helps automate the creation of a workspace
file that includes relevant dependencies, streamlining monorepo
development.

# Arguments
  all:     Automatically generates a go.work file with all module
           dependencies from the monorepo.
  modules: Relative path(s) to modules, same as used with 'go work use'.

Run "work init all" to include all dependencies from the monorepo in a 
new go.work file. Alternatively, provide specific module paths to 
initialize a workspace tailored to those dependencies.
`,
	MinArgs:  1,
	RegxArgs: `all`,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		if args[0] == `all` {
			return WorkGenerate()
		}
		return WorkInit(args...)
	},
}

var tagCmd = &bonzai.Cmd{
	Name:  `tag`,
	Alias: `t`,
	Short: `manage or list tags for the go module`,
	Long: `
The tag command helps with listing, smart tagging of modules in a
monorepo. This ensures that all modules are consistently tagged with the
appropriate module prefix and version numbers, facilitating error-free
version control and release management.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		tagBumpCmd,
		tagListCmd,
		tagDeleteCmd,
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Def: tagListCmd,
}

var tagListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `l`,
	Short: `list the tags for the go module`,
	Long: `
The "list" subcommand displays the list of semantic version (semver)
tags for the current Go module. This is particularly useful for
inspecting version history or understanding the current state of version 
tags in your project.

# Behavior

By default, the command lists all tags that are valid semver tags and 
associated with the current module. The tags can be displayed in their 
full form or shortened by setting the KIMONO_TAG_SHORTEN env var.

# Environment Variables

- KIMONO_TAG_SHORTEN: (Defaults to "true")
  Determines whether to display tags in a shortened format, removing 
  the module prefix. It accepts any truthy value.

# Examples

List tags with the module prefix:

    $ export TAG_SHORTEN=false
    $ tag list

List tags in shortened form (default behavior):

    $ KIMONO_TAG_SHORTEN=1 tag list

The tags are automatically sorted in semantic version order.
`,
	Vars: bonzai.Vars{
		{K: TagShortenVar, V: `true`, Env: TagShortenEnv},
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		shorten := vars.Fetch(
			TagShortenEnv,
			TagShortenVar,
			false,
		)
		each.Println(TagList(shorten))
		return nil
	},
}

var tagDeleteCmd = &bonzai.Cmd{
	Name:  `delete`,
	Alias: `d|del|rm`,
	Short: `delete the given semver tag for the go module`,
	Long: `
The "delete" subcommand removes a specified semantic version (semver) 
tag. This operation is useful for cleaning up incorrect, outdated, or
unnecessary version tags.
By default, the "delete" command only removes the tag locally. To 
delete a tag both locally and remotely, set the TAG_RM_REMOTE 
environment variable or variable to "true". For example:


# Arguments
  tag: The semver tag to be deleted.

# Environment Variables

- TAG_RM_REMOTE: (Defaults to "false")
  Configures whether the semver tag should also be deleted from the 
  remote repository. Set to "true" to enable remote deletion.

# Examples

    $ tag delete v1.2.3
    $ TAG_RM_REMOTE=true tag delete submodule/v1.2.3

This command integrates with Git to manage semver tags effectively.
`,
	Vars: bonzai.Vars{
		{K: TagRmRemoteVar, V: `false`, Env: TagRmRemoteEnv},
	},
	NumArgs: 1,
	Comp:    comp.Combine{git.CompTags},
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		rmRemote := vars.Fetch(
			TagRmRemoteEnv,
			TagRmRemoteVar,
			false,
		)
		return TagDelete(args[0], rmRemote)
	},
}

var tagBumpCmd = &bonzai.Cmd{
	Name:  `bump`,
	Alias: `b|up|i|inc`,
	Short: `bumps semver tags based on given version part.`,
	Long: `
The "bump" subcommand increments the current semantic version (semver) 
tag of the Go module based on the specified version part. This command 
is ideal for managing versioning in a structured manner, following 
semver conventions.

# Arguments
  part: (Defaults to "patch") The version part to increment.
        Accepted values:
          - major (or M): Increments the major version (x.0.0).
          - minor (or m): Increments the minor version (a.x.0).
          - patch (or p): Increments the patch version (a.b.x).

# Environment Variables

- TAG_VER_PART: (Defaults to "patch")
  Specifies the default version part to increment when no argument is 
  passed.

- TAG_PUSH: (Defaults to "false")
  Configures whether the bumped tag should be pushed to the remote 
  repository after being created. Set to "true" to enable automatic 
  pushing. It accepts any truthy value.

# Examples

Increment the version tag locally:

    $ tag bump patch

Automatically push the incremented tag:

    $ TAG_PUSH=true tag bump minor
`,
	Vars: bonzai.Vars{
		{K: TagPushVar, V: `false`, Env: TagPushEnv},
		{K: TagVerPartVar, V: `false`, Env: TagVerPartEnv},
	},
	MaxArgs: 1,
	Opts:    `major|minor|patch|M|m|p`,
	Comp:    comp.CmdsOpts,
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		mustPush := vars.Fetch(TagPushEnv, TagPushVar, false)
		if len(args) == 0 {
			part := vars.Fetch(
				TagVerPartEnv,
				TagVerPartVar,
				`patch`,
			)
			return TagBump(optsToVerPart(part), mustPush)
		}
		part := optsToVerPart(args[0])
		return TagBump(part, mustPush)
	},
}

var tidyCmd = &bonzai.Cmd{
	Name:  `tidy`,
	Alias: `tidy|update`,
	Short: "tidy dependencies on all modules in repo",
	Long: `
The "tidy" command updates and tidies the Go module dependencies
across all modules in a monorepo or within a specific scope. This
is particularly useful for maintaining consistency and ensuring
that dependencies are up-to-date.

# Arguments:
  module|mod:          Tidy the current module only.
  repo:                Tidy all modules in the repository.
  deps|dependencies:   Tidy dependencies of the current module in the 
                       monorepo.
  depsonme|dependents: Tidy modules in the monorepo dependent on the 
                       current module.

# Environment Variables:

- KIMONO_TIDY_SCOPE: (Defaults to "module")
  Defines the scope of the tidy operation. Can be set to "module(mod)",
  "root", "dependencies(deps)", or "dependent(depsonme)".

The scope can also be configured using the "tidy-scope" variable or
the "KIMONO_TIDY_SCOPE" environment variable. If no argument is provided,
the default scope is "module".

# Examples:

    # Tidy all modules in the repository
    $ kimono tidy root

    # Tidy only dependencies of the current module in the monorepo
    $ kimono tidy deps

    # Tidy modules in the monorepo dependent on the current module
    $ kimono tidy depsonme

`,
	Vars: bonzai.Vars{
		{K: TidyScopeVar, V: `module`, Env: TidyScopeEnv},
	},
	MaxArgs: 1,
	Opts:    `module|mod|repo|deps|depsonme|dependencies|dependents`,
	Comp:    comp.Opts,
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		var scope string
		if len(args) == 0 {
			scope = vars.Fetch(
				TidyScopeEnv,
				TidyScopeVar,
				`module`,
			)
		} else {
			scope = args[0]
		}
		switch scope {
		case `module`:
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return TidyAll(pwd)
		case `repo`:
			root, err := futil.HereOrAbove(".git")
			if err != nil {
				return err
			}
			return TidyAll(filepath.Dir(root))
		case `deps`, `dependencies`:
			TidyDependencies()
		case `depsonme`, `dependents`, `deps-on-me`:
			TidyDependents()
		}
		return nil
	},
}

var dependenciesCmd = &bonzai.Cmd{
	Name:  `dependencies`,
	Alias: `deps`,
	Short: `list or update dependencies`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
		dependencyListCmd,
		// dependencyUpdateCmd,
	},
	Def: help.Cmd,
}

var dependencyListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `on`,
	Short: `list the dependencies for the go module`,
	Long: `
The list subcommand provides a list of all dependencies for the Go
module. The scope of dependencies can be customized using the options
provided. By default, it lists all dependencies.
`,
	NoArgs: true,
	Cmds:   []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		deps, err := ListDependencies()
		if err != nil {
			return err
		}
		each.Println(deps)
		return nil
	},
}

var dependentsCmd = &bonzai.Cmd{
	Name:  `dependents`,
	Alias: `depsonme`,
	Short: `list or update dependents`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
		dependentListCmd,
		// dependentUpdateCmd,
	},
	Def: help.Cmd,
}

var dependentListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `onme`,
	Short: `list the dependents of the go module`,
	Long: `
The list subcommand provides a list of all modules or packages that
depend on the current Go module. This is useful to determine the
downstream impact of changes made to the current module.
`,
	Comp: comp.Cmds,
	Do: func(x *bonzai.Cmd, args ...string) error {
		deps, err := ListDependents()
		if err != nil {
			return err
		}
		if len(deps) == 0 {
			fmt.Println(`None`)
			return nil
		}
		each.Println(deps)
		return nil
	},
}

func optsToVerPart(x string) VerPart {
	switch x {
	case `major`, `M`:
		return Major
	case `minor`, `m`:
		return Minor
	case `patch`, `p`:
		return Patch
	}
	return Minor
}

func argIsOr(args []string, is string, fallback bool) bool {
	if len(args) == 0 {
		return fallback
	}
	return args[0] == is
}

func getGitRoot() (string, error) {
	root, err := futil.HereOrAbove(".git")
	if err != nil {
		return "", err
	}
	return filepath.Dir(root), nil
}
```