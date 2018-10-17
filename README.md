# Pitch [![Build Status](https://travis-ci.org/PM-Connect/pitch.svg?branch=master)](https://travis-ci.org/PM-Connect/pitch) [![Go Report Card](https://goreportcard.com/badge/github.com/pm-connect/pitch)](https://goreportcard.com/report/github.com/pm-connect/pitch)

A simple tool to scaffold/bootstrap files from a single source `yaml` file with flexibility and user variables.

1. [Commands](#commands)

## Commands

```
Usage: pitch [-version] [-help] [-verbose] [-autocomplete-(un)install] <command> [args]

Common commands:
    go      Build the project according to the config.
```

The `-verbose` option may be provided to **ANY** command.

### Go

The go command is responsible for running a given template and setting it up in a given directory.

```
Usage: pitch go <source> <directory>

    Go is used to run the bootstrap/scaffold process.

General Options:

	source:
		The source url/path to the yaml template.

	directory:
		The directory to scaffold/bootstrap into.

    -verbose
        Enables verbose logging.
```

