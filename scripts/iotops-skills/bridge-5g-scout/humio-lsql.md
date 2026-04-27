# Humio LSQL filter + query API shape

## Canonical LSQL filter

This is the filter the Bridge 5G team uses in the Humio UI — do not change it without team sign-off. Mirrors the iotops-scout shape with the container glob swapped:

```lsql
"kubernetes.container_name" = "bridge-5g-*"
AND ("ERROR" OR ("Error" AND !"INFO") OR "Exception" OR "exception"
     OR "Traceback" OR "Terminated")
AND !"_error"
```

Notes:
- Glob syntax (`*`), **not** regex.
- `!"_error"` excludes Humio's own internal error field.
- All `bridge-5g-*` containers are in scope — there is no equivalent of the iotops-scout `*location*` exclusion.

## Query API

```
POST https://aqua.cloudops.arubadev.cloud.hpe.com/logs/api/v1/repositories/gravity/query
Authorization: Bearer $HUMIO_TOKEN
Content-Type: application/json

{
  "queryString": "<LSQL filter above>",
  "start": <epoch-ms>,
  "end":   <epoch-ms>,
  "isLive": false,
  "timeZoneOffsetMinutes": 0
}
```

**Epoch-ms is mandatory** on Humio v1.201 — ISO-8601 timestamps are rejected.

Response is a newline-delimited JSON stream; each line has `@rawstring`, `@timestamp`, `kubernetes.pod_name`, `kubernetes.container_name`.

## Context-window query (±30 min around a single event)

Same endpoint, replace `queryString` with:

```
"kubernetes.pod_name" = "<exact-pod-name>"
```

Set `start = event_timestamp_ms - 30*60*1000`, `end = event_timestamp_ms + 30*60*1000`. Cap result at 200 lines client-side.

## Deep-link format (goes into the Fixer issue description)

```
https://aqua.cloudops.arubadev.cloud.hpe.com/logs/gravity/search?query=<url-encoded-LSQL-filter>&start=<ms>&end=<ms>
```

The URL-encoded filter should match the canonical LSQL filter above. Use `python3 -c 'import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1]))'` if `jq @uri` is unavailable.
