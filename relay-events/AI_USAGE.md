Tools used

I used ChatGPT to help me understand the assignment, review the Go code, explain the API design, and help with the written sections of the assignment.

What I used AI for

I used AI to help me understand the requirements of the assignment and to explain the existing Go implementation.

I also used it to understand the purpose of model.go, handlers.go, store.go, and store_test.go, and to check how the APIs could be tested using Postman.

For the written parts, I used AI to help organize my thoughts for NOTES.md, especially the assumptions, ambiguities, scale considerations, and the angry marketer scenario.

I reviewed the generated suggestions and compared them with the assignment requirements before using them.

One suggestion I changed and why

AI suggested focusing on the optional GET /campaigns/{campaign_id}/events endpoint after the required APIs were working.

I decided not to implement it because the assignment clearly marks this endpoint as optional. I wanted to spend the available time on the required APIs, tests, and other required submission files.

One thing AI helped me understand

AI helped me understand the difference between a duplicate and a conflict.

If the same event_id is received again with the same event data, it is a duplicate and should not increase the campaign statistics.

If the same event_id is received with different event data, I treat it as a conflict and keep the original event instead of overwriting it.

This helped me understand why the store checks event_id before storing an event.

Anything AI generated that I had to debug

I did not rely on AI output without testing it. I ran the Go project myself and verified the implementation.

The server successfully started with:

go run ./cmd/server

and listened on port 8080.

I also ran:

go test ./...

The tests passed successfully:

?    relay-events/cmd/server    [no test files]
ok   relay-events/internal/events

I also tested the API behavior using Postman, including event ingestion and campaign statistics.

I still need to run the required debugging exercise in the debugging/ directory before final submission.