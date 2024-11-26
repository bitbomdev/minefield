name: Bug Report
description: Report a bug or unexpected behavior
title: "[Bug]: "
labels: ["bug", "triage"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for taking the time to report this bug!

  - type: textarea
    id: description
    attributes:
      label: Bug Description
      description: What happened?
      placeholder: A clear and concise description of the bug
    validations:
      required: true

  - type: textarea
    id: expected
    attributes:
      label: Expected Behavior
      description: What did you expect to happen?
      placeholder: Describe what you expected to happen
    validations:
      required: true

  - type: textarea
    id: reproduction
    attributes:
      label: Steps to Reproduce
      description: How can we reproduce this issue?
      value: |
        1. 
        2. 
        3. 
        4. 
    validations:
      required: true

  - type: textarea
    id: environment
    attributes:
      label: Environment
      description: Please provide relevant environment details
      value: |
        - OS: [e.g., Ubuntu 22.04]
        - Go Version: [e.g., 1.23]
        - Project Version/Commit: [e.g., v1.0.0 or commit hash]
    validations:
      required: true

  - type: textarea
    id: logs
    attributes:
      label: Relevant Logs
      description: Please copy and paste any relevant log output
      render: shell
    validations:
      required: false

  - type: textarea
    id: additional
    attributes:
      label: Additional Context
      description: Add any other context about the problem here
    validations:
      required: false 