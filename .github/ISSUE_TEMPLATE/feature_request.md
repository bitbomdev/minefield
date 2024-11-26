name: Feature Request
description: Suggest a new feature or enhancement
title: "[Feature]: "
labels: ["enhancement", "triage"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for taking the time to propose a new feature!

  - type: textarea
    id: summary
    attributes:
      label: Summary
      description: A clear and concise description of the feature you're requesting
      placeholder: Briefly describe your feature request in 2-3 sentences
    validations:
      required: true

  - type: textarea
    id: motivation
    attributes:
      label: Motivation
      description: Why should this feature be implemented?
      value: |
        **Problem**: 
        <!-- What problem does this solve? -->

        **Impact**: 
        <!-- Who will benefit from this feature? -->

        **Context**: 
        <!-- How does this relate to existing features? -->
    validations:
      required: true

  - type: textarea
    id: use-cases
    attributes:
      label: Use Cases
      description: List specific examples of how and when this feature would be used
      placeholder: |
        1. 
        2. 
    validations:
      required: true

  - type: textarea
    id: technical-details
    attributes:
      label: Technical Details
      description: Provide technical specifications or implementation ideas
      value: |
        **Scope**: 
        
        **Dependencies**:
        
        **Implementation ideas**:
    validations:
      required: false

  - type: textarea
    id: challenges
    attributes:
      label: Potential Challenges
      description: List any concerns, technical limitations, or other challenges
      placeholder: |
        - 
        - 
    validations:
      required: false

  - type: textarea
    id: alternatives
    attributes:
      label: Alternatives Considered
      description: What other approaches have you considered?
      value: |
        1. Alternative A
           - Pros:
           - Cons:
        
        2. Alternative B
           - Pros:
           - Cons:
    validations:
      required: false

  - type: textarea
    id: additional-context
    attributes:
      label: Additional Context
      description: Add any other context, mockups, or screenshots about the feature request
    validations:
      required: false
