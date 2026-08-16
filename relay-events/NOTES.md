My understanding of the problem

Relay receives events from different delivery providers. These events can be duplicated, arrive late, or arrive in a different order.

My service needs to receive these events, store them, and provide the correct statistics for each campaign. If the same event is sent again, it should not increase the statistics twice.

Assumptions I made
I assume event_id identifies one event and can be used to detect duplicates.
Campaign statistics should count each unique event only once. For example, if the same opened event is received twice, it should still count as one opened event.
The dashboard mainly needs the number of sent, delivered, opened, and clicked events. I did not calculate conversion rates because they were not specifically required.
If one event in a batch is invalid, I reject that event but continue processing the other valid events in the batch.
I did not add authentication or deployment because the assignment says they are not required.
Things that were unclear to me
The assignment says events can arrive late and out of order. I therefore assume that sent does not necessarily have to arrive before delivered, opened, or clicked.
If the same event_id is received with different data, it is unclear whether we should ignore it or treat it as an error. I chose to return a conflict and keep the original event because silently changing the original event could hide a provider-side problem.
The assignment includes metadata, but it does not say that we need to use it for the campaign statistics. I store it but don't use it for counting.
I chose to make the statistics consistent while requests are running at the same time by protecting the store with a read/write mutex.
Questions I would ask the PM
Do we need statistics for individual contacts, or are campaign-level numbers enough?
If an event arrives several hours later, should it update the current campaign statistics?
What should happen if we receive an event for a campaign that we have never seen before? Should we accept it or reject it?
What I prioritized

Because the assignment has a limited time, I prioritized the following:

Correct duplicate handling and validation.
Implementing POST /events.
Implementing GET /campaigns/{campaign_id}/stats.
Making the store safe when multiple requests arrive at the same time.
Adding tests for the important behavior.
The optional events-list API would be the first thing I would skip if I ran out of time.
What I completed / what I intentionally skipped
Completed
POST /events
Validation of individual events
Duplicate detection
Conflict detection when the same event_id has different data
GET /campaigns/{campaign_id}/stats
Thread-safe in-memory storage
Tests for the important cases
Skipped
GET /campaigns/{campaign_id}/events because it is optional and I wanted to focus on the required APIs.
Part 3, the debugging exercise, because of the time limit in this pass.
SQLite because the assignment allows in-memory storage and the expected traffic is only a few thousand events per day.
I did not try to achieve complete test coverage. Instead, I focused on the cases that could cause serious problems, such as duplicate events, conflicts, invalid events, and concurrent requests.
If I had another day

I would complete the debugging exercise first. I would also add more tests, implement the optional events endpoint with pagination, and try a SQLite version of the store.

Part 4 — Scale memo

The current implementation works for the small amount of traffic described in the assignment, but it would not work well at 100 million events per day.

What would break first?

The biggest problems would be the in-memory map and the single mutex.

All data would be lost if the application restarted, and the amount of memory needed would keep increasing as more events are stored.

Also, all writes use the same lock. At very high traffic, this could become a bottleneck.

What would I change?

I would change the system gradually:

Move the events from memory to a proper database such as PostgreSQL.
Add a unique constraint on event_id so the database also protects us from duplicates.
Add a queue such as Kafka or SQS so the API can accept events quickly and workers can process them.
Instead of scanning every event whenever someone asks for statistics, maintain campaign counters as events are processed.
If the traffic becomes very large, partition the data by campaign or time.
Would the API change?

I would try to keep the provider-facing POST /events contract the same because providers already depend on that format.

At larger scale, I would want the API to acknowledge the request quickly and process events asynchronously.

Part 5 — The angry marketer

At 10:00 the dashboard shows:

Sent:       1,000,000
Delivered:    970,000
Opened:       250,000
Clicked:       20,000

At 10:30 it shows:

Sent:       1,000,000
Delivered:    975,000
Opened:       248,000
Clicked:       20,000

The first thing I would not assume is that the code is broken.

Why did Delivered increase?

This can be completely normal.

The assignment says that events can arrive late. A delivered event for an older message could have arrived between 10:00 and 10:30.

So:

970,000 → 975,000

is not necessarily a problem.

Why did Opened decrease?

This is more suspicious.

Normally, if we are only adding events and never deleting them, the number of opened events should not decrease.

I would investigate several possibilities:

The application restarted and lost its in-memory data.
There are multiple application instances with different data.
There is a problem with duplicate handling.
The statistics calculation has a bug.
Some events were removed or changed.
The provider sent some unexpected/retraction information.
What would I check first?

Because my current implementation uses in-memory storage, I would first check whether the application restarted between 10:00 and 10:30.

If it restarted, all previously stored events would be lost.

Then I would check whether multiple instances of the application are running and whether they all have the same data.

After that, I would look at the raw events for that campaign and compare the event_ids to see whether the opened events are actually missing or whether the statistics calculation is wrong.

Is this a bug or expected behavior?

Delivered increasing: This is expected because events can arrive late.

Opened decreasing: This is suspicious and should be investigated because counts should normally not decrease if events are only being added.

Is the Delivered increase suspicious?

No.

Because late events are explicitly mentioned in the assignment, an increase in Delivered is expected.

I would only become concerned if the number continued increasing unexpectedly long after the campaign had finished or if there was a very large unexplained jump.