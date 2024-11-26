name: Question
description: Ask a question about this project
title: "[Question]: "
labels: ["question"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for your interest in this project! Before asking your question, please:
        - Check the documentation
        - Search existing issues
        - Check if your question has already been answered in discussions

  - type: textarea
    id: question
    attributes:
      label: Your Question
      description: What would you like to know?
      placeholder: Please be as specific as possible
    validations:
      required: true

  - type: textarea
    id: context
    attributes:
      label: Context
      description: What are you trying to accomplish? This will help us provide a more helpful answer
      placeholder: Provide any additional context that might help us understand your question better
    validations:
      required: true

  - type: textarea
    id: research
    attributes:
      label: What have you tried?
      description: What documentation have you read or solutions have you attempted?
      placeholder: |
        - Documentation consulted
        - Solutions attempted
        - Related issues/discussions checked
    validations:
      required: false

  - type: textarea
    id: environment
    attributes:
      label: Environment (if relevant)
      description: Please provide relevant environment details if your question is technical
      value: |
        - OS: [e.g., Ubuntu 22.04]
        - Go Version: [e.g., 1.23]
        - Project Version: [e.g., v1.0.0]
    validations:
      required: false 