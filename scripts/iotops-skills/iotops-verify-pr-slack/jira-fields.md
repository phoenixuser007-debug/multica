# JIRA CNX Bug — field set

Ported from the original `pipeline/jira_client.py`. These are the mandatory fields for the CNX project on `jira.arubanetworks.com`.

## Endpoint

```
POST https://jira.arubanetworks.com/rest/api/2/issue
Authorization: Basic <base64 of JIRA_USERNAME:JIRA_TOKEN>
Content-Type: application/json
```

## Payload skeleton

```json
{
  "fields": {
    "project":      {"key": "CNX"},
    "issuetype":    {"name": "Bug"},
    "summary":      "[<service>] <ExceptionClass>: <root cause summary>",
    "description":  "<Jira wiki markup — see template below>",
    "priority":     {"name": "Major"},
    "components":   [{"name": "Unified Client - Non IP"}],
    "versions":     [{"name": "CNX-APR-2026"}],
    "fixVersions":  [{"name": "CNX-APR-2026"}],
    "labels":       ["auto-detected", "iotops-error-monitor"],
    "customfield_10708": {"value": "Automated-Regression"},
    "customfield_<FOUND_BY_ID>": {"value": "Dev or Unit Test"}
  }
}
```

## Discovering custom field IDs

The "Found By" custom field ID is workspace-specific. Discover at runtime:

```bash
curl -sf -u "$JIRA_USERNAME:$JIRA_TOKEN" \
  "https://jira.arubanetworks.com/rest/api/2/issue/createmeta?projectKeys=CNX&issuetypeNames=Bug&expand=projects.issuetypes.fields" \
  | jq '.projects[0].issuetypes[0].fields | to_entries[] | select(.value.name == "Found By") | .key'
```

Cache the result in a comment on this agent's workspace so subsequent runs skip the lookup.

## Component mapping (for reference — all iotops-client-* use the same)

| Repo prefix                      | Component                   |
|----------------------------------|-----------------------------|
| `iotops-client-location-*`       | Edge Location Engine        |
| `iotops-client-*` (all others)   | Unified Client - Non IP     |

The Scout already filtered out `iotops-client-location-*` pods, so you'll only ever use `Unified Client - Non IP`.

## Description template (Jira wiki markup)

```
h3. Root Cause
{{root_cause}}

h3. Affected File
{noformat}
{{affected_repo}}/{{affected_file_path}}
{noformat}

h3. Error Log
{code}
{{error_log_first_3000_chars}}
{code}

h3. Confidence
{{confidence_pct}}%

h3. Humio Deep-Link
[Open in Humio|{{humio_link}}]

h3. Auto-fix
Patch previewed below; PR raised against `devel` branch.
{code:language=diff}
{{patch_first_80_lines}}
{code}

h3. Multica Issue Tree
* Original error: [{{scout_child_issue_url}}|Fixer issue]
* This ticket's verification: [{{verifier_issue_url}}|Verifier issue]
```

## Transitions (for "In Review")

Transition IDs vary per workflow. Discover via:

```bash
curl -sf -u "$JIRA_USERNAME:$JIRA_TOKEN" \
  "https://jira.arubanetworks.com/rest/api/2/issue/$JIRA_KEY/transitions"
```

Find the transition whose `name` is `In Review` (or `Start Review`) and use its `id`:

```bash
curl -sf -u "$JIRA_USERNAME:$JIRA_TOKEN" -X POST -H "Content-Type: application/json" \
  "https://jira.arubanetworks.com/rest/api/2/issue/$JIRA_KEY/transitions" \
  -d "{\"transition\":{\"id\":\"$TRANSITION_ID\"}}"
```

## Dedup check (before creating)

```bash
JQL='project=CNX AND component="Unified Client - Non IP" AND summary ~ "\"'$SERVICE'\" \"'$EXCEPTION_CLASS'\"" AND created >= -1d'
EXISTING=$(curl -sf -u "$JIRA_USERNAME:$JIRA_TOKEN" -G \
  --data-urlencode "jql=$JQL" --data-urlencode "fields=summary,status" \
  "https://jira.arubanetworks.com/rest/api/2/search" \
  | jq -r '.issues[0].key // empty')
```

If `$EXISTING` is non-empty, reuse it instead of creating a new ticket.
