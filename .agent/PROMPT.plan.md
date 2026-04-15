1. Study `docs/specs/` to understand what must be built, then read all Go source files under `src/` to understand what is already implemented.

2. Run `tk list` to see existing tickets. If tickets already exist, scan their tags to identify which epics are already covered. Requirements must be broken down into epics with `<epic-name>, epic`, and then into tasks which belong to the epics. If you find that all epics and tasks already have tickets, output `<promise>COMPLETE</promise>`.

3. From the specs, identify **exactly one** feature that has no existing epic ticket. Use the available go skills to create an implementation plan for that feature. Ultrathink. Be very detailed in the implementation plan, and follow the specs closely when planning.

4. For the chosen feature, create tickets using `tk create`:
   - **One epic** for the feature as a whole: type `epic`, tagged `[<epic-name>, epic]`. The description must include the full implementation plan for this feature, and should summarize the epics's purpose, list all tasks, and reference the relevant spec files.
   - **Multiple tasks per epic**. Each task description must include:
     - Key implementation details derived from the specs (reference the main epic for full details).
     - Acceptance criteria as a checklist.

5. Apply consistent tags on every ticket:
   - Epic: `[<epic-name>, epic]`
   - Tasks: `[<epic-name>, <task-name>, task]`

IMPORTANT: Plan and document only. Do NOT implement any Go code.

If after step 2 you find that all epics already have tickets, output `<promise>COMPLETE</promise>`.
