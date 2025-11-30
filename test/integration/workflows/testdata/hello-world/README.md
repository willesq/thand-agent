# Hello World Workflow Test

This is the simplest integration test to verify the workflow engine is working correctly.

## Components

- **providers.yaml**: Configures AWS (LocalStack) and Email (MailHog) providers
- **roles.yaml**: Defines a simple test role with access to all resources
- **workflow.yaml**: A minimal workflow that takes a name and returns a greeting

## Expected Behavior

1. The workflow receives an input with a `name` field
2. The `greet` task sets a `result` variable with the greeting message
3. The workflow outputs the greeting and a timestamp

## Test Assertions

- Workflow executes successfully
- Output contains `greeting` field with "Hello, {name}!"
- Output contains `timestamp` field
