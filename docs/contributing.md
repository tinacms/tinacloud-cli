# Contributing guide

## Pre-commit / PR
Before committing / creating a PR, it's highly advised to run some pre-check
commands:
- Ensure the code is formatted: `make fmt`
- Scan for known security issues: `make security` - while `gosec` can flag
  false positives, please do not disable them without first checking, a) it's
  not a valid concern, and b) an additional comment above the
  disabled rule as to why you feel the scan is a false positive is provided.
- Vet the written code: `make vet`
- Ensure all tests pass: `make test`

If any of these fail, your contribution will not be accepted. Please ensure all
of these are passing before requesting a review.

## PR / Issue etiquette 
Please remember that this project is maintained by humans.

Every discussion, issue, and pull request is read and reviewed by humans (and
sometimes machines, too). It is a boundary point at which people interact with
each other and the work done. It is rude and disrespectful to approach this
boundary with low-effort, unqualified work, since it puts the burden of
validation on the maintainer. 

Please ensure any work you submit has been validated and tested by yourself
before said code has been pushed. Issue / PR descriptions should be succinct
and state the main driving force for the change, any additional background /
context that may be useful for the reviewer, and a linked issue that this code
solves if one exists.

As a project, we welcome meaningful usage of AI as a productive tool.
