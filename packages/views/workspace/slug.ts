export const WORKSPACE_SLUG_REGEX = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const WORKSPACE_SLUG_FORMAT_ERROR =
  "Only lowercase letters, numbers, and hyphens";

export const WORKSPACE_SLUG_CONFLICT_ERROR =
  "That workspace URL is already taken.";

/**
 * Auto-generate a slug from a workspace name.
 *
 * Returns empty string when the name produces no valid slug characters. The
 * form leaves the slug field empty in this case and the user must type one:
 * this is preferable to a hardcoded fallback like "workspace", which silently
 * chooses a useless URL slug and causes 409 conflicts for later workspaces
 * with the same fallback.
 */
export function nameToWorkspaceSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

export function isWorkspaceSlugConflict(error: unknown): boolean {
  return (
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    (error as { status?: unknown }).status === 409
  );
}
