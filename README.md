# kopl

`kopl` is a command-line tool designed to streamline the development of plugins for KOReader. It helps with project initialization, static code analysis, and dependency management for Lua-based KOReader plugins.

## Features

- **Project Initialization**: Quickly set up a new KOReader plugin project with the necessary file structure and configuration.
- **Static Code Analysis**: Integrate `luacheck` to perform static analysis on your Lua code, helping to identify potential issues and enforce coding standards. `kopl` automatically handles the installation and setup of `luacheck` and `luarocks` if they are not found in your system's PATH.
- **Dependency Management**: Simplifies the management of `luacheck` and `luarocks`, ensuring your development environment is correctly configured.
- **Deployment (Planned)**: Future versions will include functionality to easily deploy your plugins to a KOReader device.

## Installation

**Prerequisites:**

- Go
- Git
- LuaRocks (optional, `kopl` can try to install `luacheck` via `luarocks` if needed)

Example using `go install`:

```bash
go install github.com/Consoleaf/kopl@latest
```

## Usage

### Initialize a new KOReader Plugin Project

To create a new KOReader plugin project, use the `init` command:

```bash
kopl init my-awesome-plugin.koplugin
```

This will create a new directory `my-awesome-plugin.koplugin` with the basic project structure and necessary files like `_meta.lua` and `main.lua`.

### Perform Static Checks

To run `luacheck` on your project and perform static analysis:

```bash
kopl check
```

`kopl` will look for `luacheck` in your PATH. If not found, it will attempt to install it using `luarocks`.

### Deploy Project to a Device

_(This command is currently a placeholder and will be implemented in future versions.)_

```bash
kopl deploy
```

## License

This project is licensed under the MIT License - see the `LICENSE` file for details.
