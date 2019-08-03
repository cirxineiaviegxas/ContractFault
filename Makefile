# contractfault — build, test and demo orchestration.
#
# The Makefile drives both halves of the project: the Go analyzer CLI and the
# TypeScript seismic viewer. Targets are deliberately small and composable so
# CI and humans run the same commands.

GO        ?= go
NPM       ?= npm
BIN       ?= bin/contractfault
VIEWER    ?= viewer
EXAMPLES  ?= examples
OLD       ?= $(EXAMPLES)/contracts/orders-v1.json
