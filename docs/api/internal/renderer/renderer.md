# `internal/renderer`

## Overview

The `renderer` package is deprecated. Its source file contains no renderer implementation, exported declarations, types, functions, or package-level state. New code should use `internal/render` instead.

## API

There is no public API defined in this package. The package declaration is retained only as a deprecated package boundary; this file does not provide rendering behavior.

## Implementation details

`internal/renderer/renderer.go` consists solely of the package comment and package declaration. It has no imports, dependencies, side effects, or runtime behavior.

## Migration

Use the replacement package, `internal/render`, for rendering functionality. Consult that package's documentation and API for available operations.
