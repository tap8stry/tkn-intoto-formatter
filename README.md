# tkn-attest

A common library to convert any tekton resource to intoto attestation format. 

In this library, the idea is to support intoto attestation conversion for all tekton resources including

- Pipeline
- Task
- Pipelinerun
- Taskrun

Currently, the conversion is supported for `Task` and `Pipeline` to intoto layout.

## Usage

Currently, `tkn-attest` is also made available as a stand-alone CLI. You can follow following procedure to install CLI:

```
git clone 
cd tkn-attest
make
```

This should create an executable binary `tkn-attest`. You can add this binary to your local PATH.

### Quick start

1. Get Help

```
% tkn-attest -h
tkn-attest is tool to manage various attestation functions, including
		conversion to intoto format and comparisons.

Usage:
  tkn-attest [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  convert     converts yaml tekton spec to specified format
  help        Help about any command
  version     tkn-attest version

Flags:
      --config string   config file (default is $HOME/.tkn-attest.yaml)
  -h, --help            help for tkn-attest
  -t, --toggle          Help message for toggle

Use "tkn-attest [command] --help" for more information about a command.
```

2. Convert static tekton resources to intoto format

```
%% tkn-attest convert -i sample-pipeline/task-bom.yaml -f ./task-bom-attest.json
```

## WIP

1. Support Pipelinerun and Taskrun 
2. Try to capture attestation flow from event source -> pipeline -> tasks
