# Multica CLI authentication reference

The CLI reads credentials from `/root/.multica/config.json` (or the path at
`$HOME/.multica/config.json` on non-root users). No flags needed for auth in
the common case.

## Config file

```json
{
  "server_url":   "http://localhost:8090",
  "app_url":      "http://localhost:3001",
  "workspace_id": "4b3bee90-5306-4950-823f-cc567bad8e97",
  "token":        "mul_f1a0911e008d9af40f3eac9f35d49c3bc5fa7a8c",
  "watched_workspaces": [
    { "id": "4b3bee90-…", "name": "admin's Workspace" }
  ]
}
```

- `token` is a Personal Access Token (PAT) with prefix `mul_`.
- `workspace_id` is sent as the `X-Workspace-ID` header.
- `server_url` points at the backend (localhost:8090 for the bundled dev setup).

## Overriding

Rarely needed inside an agent, but supported:

```bash
multica --server-url http://multica.internal:8090 \
        --workspace-id "$OTHER_WORKSPACE" \
        issue list
```

Env equivalents: `MULTICA_SERVER_URL`, `MULTICA_WORKSPACE_ID`.

**Bearer token override**: there's no `--token` flag by design — prefer
switching profiles (`--profile`) or adjusting the config file. If you *must*
use a different token for one command, set `MULTICA_TOKEN` in the shell
environment for that command.

## Profiles

If your host has multiple multica instances (e.g. prod + staging), profiles
isolate config + daemon state:

```bash
multica --profile staging issue list
```

Profiles live under `~/.multica/<profile>/config.json`. Default profile uses
the plain `~/.multica/config.json`.

## Failure modes

| CLI error | What to do |
|---|---|
| `config file not found` | You're running as the wrong user. The daemon config lives at `/root/.multica/config.json`; agents running under root will see it. If running as a different user, re-run `multica login` or copy the config. |
| `unauthorized` / 401 | The PAT is expired or was revoked. Notify a human — agents cannot re-authenticate. |
| `workspace not found` | `workspace_id` in the config points at a workspace the PAT can't see. Human fix required. |
| `server unreachable` | The daemon likely crashed or the multica backend is down. Tail `/root/.multica/daemon.log` for the real error. |

## Not for use inside dev-env-dev-1

The `dev-env-dev-1` container does not have `/root/.multica/config.json` mounted. If you need multica API access from inside the container (rare), run the CLI on the host first (e.g. fetch an issue body with `multica issue get`), write the result to a tempfile, and `docker cp` it into the container.
