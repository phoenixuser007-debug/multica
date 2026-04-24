# Creating the IoTOps Error Scout autopilot in the multica UI

You only need one autopilot — the UI is the right place to create it. The values below are copy-paste-ready.

## Step 1 — Create the autopilot

**Settings → Autopilots → New Autopilot**

| Field | Value |
|---|---|
| Title | `IoTOps Error Scout` |
| Issue title template | `{{date}} — Scan iotops-client-* errors` |
| Assignee | `IoTOps Scout` agent (ID `17a8d5e1-d102-4402-b717-404b0bf15cb4`) |
| Execution mode | `Create issue` |
| Priority | `Medium` |
| Retry on blocked | **on** |
| Description | *(paste the block below)* |

### Description (copy verbatim)

```markdown
Scan Humio across the aqua, brooke, jedi, and firth clusters for new runtime
errors in iotops-client-* Kubernetes pods and fan out a child Fixer issue per
distinct error.

The IoTOps Scout agent (assigned to this autopilot) runs the `iotops-scout`
skill, which:

1. Determines the polling window since the previous tick (24 h by default).
2. For each cluster (aqua / brooke / jedi / firth), runs the canonical IoTOps
   LSQL filter against the gravity repository.
3. For each error event, fetches ±10 minutes of pod context.
4. Looks up the affected source file (and the ±20 lines around the stack
   frame) from Stash to classify errors that share an exception class but
   originate in different files.
5. Dedupes each fingerprint (`cluster::service::ExceptionClass::file:line`)
   against open multica issues (last 24 h).
6. For each genuinely new fingerprint, ensures a JIRA CNX Bug exists
   (creates one if missing, reuses an existing open ticket otherwise).
7. Creates a child issue assigned to the IoTOps Fixer agent with the full
   context package (error log + ±10-min pod context + source snippet +
   ARCHITECTURE.md + Jira key).

The Fixer then performs strict TDD: propose solution → write a failing JUnit
test that reproduces the exception (red) → minimal fix (green) → `ctl`
quality gate. On success it hands off to the IoTOps Verifier, which opens a
draft PR via `pr-cli`, polls Jenkins CI via `jenkins-skill`, and posts the
final Slack notification.

No action is required from a human unless a fix fails dev-env validation or
Jenkins CI — those cases raise a Slack alert and leave the issue in `blocked`
for review.
```

## Step 2 — Attach a daily schedule trigger

**Triggers → Add schedule**

| Field | Value |
|---|---|
| Kind | `Schedule` |
| Cron expression | `0 6 * * *` (06:00 UTC daily — pick any time you like) |
| Timezone | `UTC` |
| Label | `Daily IoTOps error scan` |

## Step 3 — Host env vars required on the OpenCode runtime

Before the first real run, ensure `/etc/environment` on the daemon host has:

```
AQUA_HUMIO_TOKEN=<…>
BROOKE_HUMIO_TOKEN=<…>
JEDI_HUMIO_TOKEN=<…>
FIRTH_HUMIO_TOKEN=<…>
BITBUCKET_STASH_TOKEN=<…>
JIRA_TOKEN=<…>
JIRA_USERNAME=<…>
SLACK_WEBHOOK_URL=<…>
# Optional, only if jedi/firth don't follow the default cloudops URL pattern:
JEDI_HUMIO_URL=https://jedi.example/logs
FIRTH_HUMIO_URL=https://firth.example/logs
```

Restart the multica daemon after editing `/etc/environment` for the values
to be picked up.

## Step 4 — Force a single run to smoke-test

From the UI: **Autopilots → IoTOps Error Scout → Run now** (or POST
`/api/autopilots/<id>/trigger` with your bearer token).

The first tick produces a Scout issue. Watch the Fixer and Verifier child
issues follow if any real Humio errors were present in the window.

## Why no setup script?

Previous versions of this repo had a `setup-iotops-error-autopilot.sh`. It
mirrored the CVE bulk-provisioning pattern — useful when you need to create
56 autopilots in one shot, pointless for one. Removed in favour of this doc.
