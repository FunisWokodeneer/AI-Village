# AGS AI Village — Product Roadmap

This roadmap distinguishes implemented foundations from planned capabilities.

## Product principles

- The village world is the primary experience, not a dashboard with decorative avatars.
- Agents are real software actors with explicit capabilities and observable work.
- Visual animation and actual agent execution are separate systems.
- Provider-neutral interfaces should support hosted and local models.
- User control, permissions, and approval gates are required before agents take consequential actions.
- Public development should document limitations honestly and avoid copying another project's distinctive assets or presentation.

## Milestone 0 — Living-world foundation

- [x] Native desktop window and explorable 2D village prototype.
- [x] Five resident identities, roles, personalities, homes, and workplaces.
- [x] Visible movement, simulation clock, pause, and resident inspection.
- [x] Optional OpenAI-compatible model-driven activity selection.
- [ ] Reliable navigation around buildings and obstacles.
- [ ] Settings screen for provider/model/endpoint and secure credential entry.
- [ ] Save and load world state.

## Milestone 1 — Real organization

- [x] Initial organization domain model for departments, agent profiles, projects, tasks, approvals, and activity events.
- [x] Fail-closed policy classification for scoped workspace actions and approval-required actions.
- [ ] Organization model: owner, Chief of Staff, directors, teams, workers, with persistence and validation.
- [ ] Agent profiles with editable role instructions, capabilities, and boundaries.
- [ ] Department and project spaces represented physically in the village.
- [ ] Task lifecycle: proposed, approved, queued, running, blocked, review, completed.
- [ ] Delegation with parent/child task traceability.
- [ ] Live activity feed linked to actual task events.

## Milestone 2 — Agent intelligence

- [ ] Provider adapter interface and model catalog.
- [ ] Per-agent system prompts, context assembly, and structured action schema.
- [ ] Short-term context and persistent, inspectable memory.
- [ ] Agent-to-agent messages, meetings, and collaboration.
- [ ] Schedules, priorities, interruption handling, and idle behavior.
- [ ] Decision budgets, timeouts, retries, and token/cost visibility.

## Milestone 3 — Work engine

- [ ] Permissioned tool registry.
- [ ] Project workspaces and artifact storage.
- [ ] Research, coding, file, and browser tool integrations.
- [ ] Sandboxed execution and explicit approval for external side effects.
- [ ] Review/verification steps before reporting completion.
- [ ] Reproducible task logs and exportable deliverables.

## Milestone 4 — World and experience

- [ ] Distinctive AGS art direction, animation, sound, and environment.
- [ ] Interior spaces and enterable buildings.
- [ ] Character pathfinding, navigation, and richer animation states.
- [ ] Visitor avatar and optional direct interaction.
- [ ] World events, meetings, project milestones, and notifications.
- [ ] Accessibility, scaling, keyboard navigation, and reduced-motion options.

## Milestone 5 — Public release readiness

- [ ] Cross-platform build strategy and Windows installer/portable release.
- [ ] First-run setup wizard with no terminal requirement.
- [ ] Configuration and secret-storage review.
- [ ] Privacy/security documentation and threat model.
- [ ] Automated tests, release workflow, crash reporting opt-in, and upgrade path.
- [ ] Screenshots, demo data, contribution guide, and issue templates.

## Current scope

The current branch starts Milestone 0. The prototype does not yet include persistent memory, true task delegation, tool execution, or a complete organization hierarchy. Those are explicit future milestones rather than simulated claims.
