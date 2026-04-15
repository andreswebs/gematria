1. Run `make build` to initialize the environment. If there are any issues, fix those first.

2. Run `tk ready` to list tickets that are ready to work on (all dependencies resolved). If there are no ready tickets, run `tk blocked` to understand why, then output `<promise>COMPLETE</promise>`.

3. Study @docs/specs to understand the roadmap for the project. Study @docs/learnings.md to absorb previous implementation gotchas.

4. Pick a single ticket to work on — choose the one YOU judge to be highest priority based on ticket type, priority field, and dependencies. Run `tk show <id>` to read the full description and acceptance criteria. Make sure to follow TDD, use the tdd skill. Before implementing anything, if you think it is necessary, span as many subagents as needed to research what needs to be done. When ready, begin to implement the item. Mark the ticket as in_progress with `tk start <id>`. Use the available go skills for expertise.

5. Use the available `make` commands to get feedback about your work during development (see `make help` for the list of available commands). When done, run `make build` to validate your work. If there are any issues, fix them.

6. Run `tk close <id>` to mark the ticket closed.

7. Add a note summarising what was done and any relevant details for the next person: `tk add-note <id> "..."`. Update @docs/learnings.md if there were relevant learnings or problems solved.

ONLY WORK ON A SINGLE TASK.

If after step 2 there are no open tickets at all (i.e. `tk ready` and `tk blocked` both return nothing), output `<promise>COMPLETE</promise>`.
