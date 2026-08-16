This directory is where the assignment's Part 3 debugging exercise
(`debugging/README.md`, `events.jsonl`, `expected_output.txt`, and the buggy Go
program) belongs. It was not included in the assignment PDF I was given, only
referenced by it — only you have the actual starter files from the assignment
package.

To complete Part 3:
1. Copy the real `debugging/` folder from the assignment starter package here,
   replacing this file.
2. Run `go run -race . events.jsonl` — the brief flags that at least one bug only
   shows itself under concurrency, so `-race` is the first tool to reach for.
3. Diff actual output against `expected_output.txt` for the other three.
4. Document each bug in `BUGS.md` at the project root (what it is, why it happens,
   your minimal fix, how you verified it).

See NOTES.md ("What I completed / what I intentionally skipped") for why this was
left out of this pass.
