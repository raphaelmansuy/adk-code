# Error Handling Guide

This document describes the recommended error handling patterns across the code agent project.

## Goals

- Use `code_agent/pkg/errors` as a unified error type for application-level errors.
- Reserve standard library `errors` and `fmt.Errorf` for low-level, implementation-specific errors where it makes sense.
- Add contextual information when wrapping errors.
- Keep behavior and error semantics stable.

## When to use `pkg/errors` (AgentError)

- Any error that represents application-level states (validation, API issues, tool or model errors, permissions, path traversal) should be wrapped as an `AgentError` with a meaningful `Code`.
- Use the helpers in `pkg/errors` where possible: `InvalidInputError`, `PermissionDeniedError`, `ExecutionError`, `ModelNotFoundError`, `InternalError`, `Wrap`, etc.
- Prefer `Wrap(code, message, err)` when an underlying error must be preserved in the chain.

Example:

- Returned error from a wrapped external system call:
  return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to open resource", err)

- Input validation error (invalid or missing argument):
  return nil, pkgerrors.InvalidInputError("user_id is required")

## When to use standard `error` or `fmt.Errorf`

- For very local, low-level errors within small helper functions (that don't need a dedicated `AgentError`) prefer to return plain `error`.
- Use `fmt.Errorf` with `%w` when returning an error that should be chainable with `errors.Is` / `errors.As`, and there is no need to attach an AgentError code.

## Choosing Error Codes

Prefer meaningful error codes from `pkg/errors.ErrorCode`:
- Path/File: `CodeFileNotFound`, `CodePermission`, `CodePathTraversal`.
- Execution: `CodeExecution`, `CodeTimeout`.
- Model-related errors: `CodeModelNotFound`, `CodeProviderError`.
- Validation and input: `CodeInvalidInput`, `CodeValidation`.
- General/internal: `CodeInternal`, `CodeNotSupported`.

## Examples and common patterns

- Creating a new resource that fails because of an underlying system error:
  return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create resource", err)

- Request validation:
  if req.Name == "" {
      return nil, pkgerrors.InvalidInputError("name is required")
  }

- Using an explicit permission error:
  if !allowed { return nil, pkgerrors.PermissionDeniedError(path) }

## Tests and Assertions
- Use `pkgerrors.Is(err, pkgerrors.CodeInvalidInput)` where appropriate to assert errors by code.
- For unit tests, compare the `AgentError.Code` and the message prefix where needed.

## Migration Notes
- Migrate incrementally — change one package at a time and add tests.
- For public API that currently returns plain `error`, returning `AgentError` is acceptable because it is still an `error` (no change to the compile-time contract).
- Ensure tests cover both the presence and the `Code` value in behavioral error checks.

## Reference
- See `code_agent/pkg/errors/errors.go` for `AgentError` implementation and helper functions.
