---
name: Bug Report
about: Report a bug to help us improve
title: "[Bug] "
labels: bug
---

## Problem Description

<!-- Clearly and concisely describe the bug -->

## Steps to Reproduce

1.
2.
3.

## Expected Behavior

<!-- Describe what you expected to happen -->

## Environment Information

- OS:
- Go version (`go version`):
- `ezcapsolver-go` version:
- Task type involved:

## Additional Information

<!--
Logs or other information helpful for locating the issue.

Inject a logger at LevelTrace to capture the request and response bodies; the
SDK redacts clientKey and proxy at any nesting depth:

    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: ezcapsolver.LevelTrace,
    }))
    client, err := ezcapsolver.NewClient(ezcapsolver.WithLogger(logger))

If a solution failed to decode, SolutionDecodeError.Raw holds the exact shape
the worker returned — that is the single most useful thing to attach. Every
solution model also keeps unmodelled keys in its Extra field, and Solved.Raw
keeps the original JSON, so nothing is lost on the way to you.
-->
