# Flow User Segmentation Demo

A common problem in the marketing domain is user segmentation. Many companies
maintain segments of their users to better understand behavior and address cohorts.
A company may have many segmentation events coming in continuously, each of which
represents an add or remove of a user to a segment.

These granular events must be transformed into current understandings of:

-   What users are in a given segment?
-   For a given user, what segments are they a member of?
-   Did an event actually change the status of a user & segment? Was it a repeated add or remove?

This demo is derived from a documented [Flow example](https://github.com/estuary/flow/tree/master/examples/segment),
which may provide a more friendly introduction as this repository is designed with a human narrator in mind.

## Narrator Script

Environment:
  - Repo is github.com/estuary/segment-demo
  - Run with [VSCode Remote Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers).
  - This integrates `flowctl` tool with a vanilla PostgreSQL database, used for materializations.

Sources to examine:
  - `1_captures.flow.yaml`: Characterizes upstream data system(s) and collection into which segment events are captured.
  - `schemas/event.schema.yaml`: JSON schema for a segmentation event.
  - `3_materialize.flow.yaml`: Defines endpoint(s) and materializations.

Start Flow:
```console
# Start Flow.
flowctl develop --source 3_materialize.flow.yaml --port 8080

# Start Webhook server stand-in.
node scripts/demo-webhook-api.js
```

Examine created PostgreSQL "segment_profiles" & "segment_memberships" tables.

Start ingesting data.
```console
./scripts/stream-events.sh
```

Observe laking of segment event data:
```console
find flowctl_develop/fragments/examples/segment/events -type f
```
```console
zcat ..pasted path to fragment file .. | jq '.'
```

Live queries:
* `pg_user_active_segments.sql`: Range over users and show active segments (re-query to watch this update).
* `pg_user_point_lookup.sql`: Point lookup of the active and inactive segments of a specific user.
* `pg_segment_membership.sql`: Current user list of segments of a given vendor.

"Toggles" on `demo-webhook-api.js` tab:
 - Demonstrates taking action when an interesting event occurs.
 - Calls webhook on de-duplicated status change ("member" <=> "non-member").
 - Initial adds are filtered, so they are (relatively) rare but still frequent.

Back-pressure:
  - Try changing `delay` parameter in `demo-webhook-api.js`.
  - It models slow APIs / network delay / back-pressure.
  - When increased, Flow does more roll-up per call.
  - When decreased, Flow makes more frequent, smaller calls.

More load:
  - Try updating rate from 1K => 10K (requires a reasonably beefy PC; e.x. 8 cores or so).
  - Watch how "toggles" calls are much bigger.
  - Stop event stream; notice "toggles" is all caught up.

Run tests:
 - `flowctl test --source examples/segment/flow.yaml`

Cleanup:
 - `git clean -f -d` to remove `flowctl_develop` directory.
 - Run `queries/pg_reset_tables.sql` to restore PostgreSQL to pristine state.