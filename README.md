# grml - A simple build automation tool written in Go

grml is a simple Makefile alternative. Build targets are defined in a `grml.yaml` file located in the project's root directory.
This file uses the [YAML](http://yaml.org/) syntax.

A minimal sample can be found within the [sample](sample/grml.yaml) directory. Enter the directory with a terminal and execute `grml`.

[![asciicast](https://asciinema.org/a/460524.svg)](https://asciinema.org/a/460524)

## Installation
### From Source
    go install github.com/ilizardnetwork/grml@latest

### Prebuild Binaries
https://github.com/ilizardnetwork/grml/releases

## Specification
- Environment variables can either be defined in the **env** or **envs** section.  
  The latter defines paths to dedicated files which use the `key: value` pair syntax.  
  Variables declared in those files will be evaluated first, ordered by their sequence  
  and overwritten by their successor or `env` variable if applicable.  
  These variables are passed to all run target processes.
- Variables are also accessible with the `${}` selector within **help** messages and **import** statements.
- Dependencies can be specified within the command's **deps** section.

### Additonal Environment Variables

The process environment is inherited and following additonal variables are set:

| KEY     | VALUE                                                          |
|:--------|:---------------------------------------------------------------|
| ROOT    | Path to the root build directory containing the grml.yaml file |
| PROJECT | Project name as specified within the grml file                 |
| NUMCPU  | Number of CPU cores                                            |