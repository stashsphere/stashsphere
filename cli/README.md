# stashctl

a cli for StashSphere

## About

Use `stashctl` to access a `StashSphere` account from your terminal.
The tool supports almost all workflows of the official backend API.

While this is can be used standalone, the eventual goal is to use
it within an harness / agent that you can query in natural language
or using your voice depending on your setup.
Since the commands are rather trival and the output is easily parseable,
small models will work that can be inferred locally, even on older hardware.

## TODO

* add a skill template that can be either installed automatically or simply prints
  related instructions
* potentially add token-based auth to the backend to avoid username+password, cookies
  in the CLI
