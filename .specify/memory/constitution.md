<!--
SYNC IMPACT REPORT - Constitution Update
=========================================
Version Change: INITIAL → 1.0.0
Rationale: First constitution ratification for air-go backend API project

Modified Principles: N/A (initial creation)
Added Sections:
  - All 5 core principles (API-First, Test-Driven Development, Code Review Mandatory, End-to-End Testing, Observability)
  - Development Workflow section
  - Governance section

Removed Sections: N/A

Template Updates:
  ✅ plan-template.md - Constitution Check section aligned with 5 principles
  ✅ spec-template.md - User scenarios support E2E testing requirements
  ✅ tasks-template.md - Task organization supports TDD and E2E test requirements

Follow-up TODOs: None

Last Updated: 2026-01-24
-->

# air-go Constitution

## Core Principles

### I. API-First

Every feature must expose functionality through well-defined APIs. All endpoints MUST:
- Follow RESTful conventions (GET, POST, PUT, DELETE with appropriate status codes)
- Accept and return JSON by default
- Include comprehensive request/response validation
- Be versioned to support backward compatibility
- Be documented with OpenAPI/Swagger specifications

**Rationale**: API-first design ensures consistent interfaces, enables frontend/backend parallel development, and facilitates integration testing. Clear contracts prevent miscommunication and reduce integration bugs.

### II. Test-Driven Development (NON-NEGOTIABLE)

TDD is mandatory for all feature development. The workflow MUST be:
1. Write tests based on acceptance criteria
2. Get user/reviewer approval of test cases
3. Verify tests fail (red)
4. Implement minimal code to pass (green)
5. Refactor while keeping tests green

**Rationale**: TDD ensures testable code design, provides living documentation, catches regressions early, and builds confidence in refactoring. Non-negotiable because it fundamentally shapes code quality and maintainability.

### III. Code Review Mandatory

All code changes MUST be reviewed before merging. Reviews MUST verify:
- Constitution compliance (all principles)
- Test coverage (unit + integration + E2E where applicable)
- Security considerations (input validation, authentication, authorization)
- Performance implications
- Documentation completeness

**Rationale**: Code review catches bugs early, shares knowledge across the team, maintains code quality standards, and ensures architectural consistency. Multiple perspectives improve decision quality.

### IV. End-to-End Testing

All user-facing features MUST include end-to-end tests that validate complete user journeys. E2E tests MUST:
- Test the full request/response cycle through actual HTTP endpoints
- Use realistic test data and scenarios
- Cover primary user workflows (P1 stories minimum)
- Run in CI/CD pipeline before deployment
- Be independent and idempotent

**Rationale**: E2E tests validate that integrated components work together correctly in production-like environments. They catch integration issues, configuration problems, and deployment bugs that unit tests miss.

### V. Observability

Production systems MUST be observable and debuggable. All services MUST implement:
- Structured logging (JSON format with context: request ID, user ID, timestamps)
- Metrics collection (latency, error rates, throughput)
- Distributed tracing for request flows
- Health check endpoints
- Error tracking and alerting

**Rationale**: Observability enables rapid diagnosis of production issues, supports data-driven optimization, and provides visibility into system behavior. Debugging without observability is guesswork.

## Development Workflow

All development follows this workflow:

1. **Specification**: Create feature spec with user stories and acceptance criteria
2. **Planning**: Design technical approach and create implementation plan
3. **Test Writing**: Write tests for acceptance criteria (TDD - principle II)
4. **Review Tests**: Get approval on test cases before implementation
5. **Implementation**: Write minimal code to pass tests
6. **Code Review**: Submit PR for review (principle III)
7. **E2E Validation**: Run end-to-end tests (principle IV)
8. **Deploy**: Merge and deploy with observability enabled (principle V)

## Governance

This constitution supersedes all other development practices and guidelines. All code, PRs, and architectural decisions MUST comply with the principles above.

**Amendment Process**:
- Amendments require documented justification and team consensus
- Version bumping follows semantic versioning:
  - MAJOR: Backward-incompatible principle changes or removals
  - MINOR: New principles or material expansions
  - PATCH: Clarifications, wording fixes, non-semantic refinements
- All amendments MUST include migration plan for existing code

**Compliance**:
- All pull requests MUST pass constitution compliance check
- Complexity and exceptions MUST be justified in implementation plans
- Regular audits ensure ongoing compliance

**Development Guidance**: For day-to-day development guidance and best practices, refer to project README and documentation in `docs/` directory.

**Version**: 1.0.0 | **Ratified**: 2026-01-24 | **Last Amended**: 2026-01-24
