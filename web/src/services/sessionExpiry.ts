/**
 * Knowing when the session ends, before it does.
 *
 * The token already carries its own expiry, so the app can watch for it locally
 * rather than discovering expiry by having a request fail. That makes the
 * reactive recovery in apiFetch a fallback for the case where nobody was
 * watching the tab, instead of the normal path.
 */

interface JwtPayload {
  exp?: number;
}

/**
 * Milliseconds before expiry to warn. Long enough to notice and act, short
 * enough that the prompt is not a constant companion.
 */
export const EXPIRY_WARNING_LEAD_MS = 5 * 60 * 1000;

/**
 * Reads the expiry out of a JWT without verifying it - the server is the only
 * thing that can trust a token; this only needs to know when to ask.
 * Returns null if the token carries no usable expiry.
 */
export function tokenExpiresAt(token: string): Date | null {
  const payload = token.split(".")[1];
  if (!payload) return null;

  try {
    // JWT uses base64url, which atob does not accept directly.
    const normalised = payload.replace(/-/g, "+").replace(/_/g, "/");
    const claims: JwtPayload = JSON.parse(atob(normalised));

    if (typeof claims.exp !== "number") return null;

    return new Date(claims.exp * 1000);
  } catch {
    return null;
  }
}

/**
 * How often to re-examine the token.
 *
 * The phase is polled rather than scheduled with a single timeout hours ahead:
 * a long timeout does not survive the machine sleeping, and fires late - after
 * expiry - which offered a renewal prompt for a session that had already gone.
 */
export const SESSION_CHECK_INTERVAL_MS = 30 * 1000;

export type SessionPhase = "active" | "expiring" | "expired";

/**
 * Where the token stands right now. "expiring" can be renewed with a click;
 * "expired" cannot, and needs the password prompt instead - the refresh
 * endpoint requires a still-valid token by design.
 *
 * A token carrying no expiry is treated as active: the server is the authority,
 * and guessing would log people out for no reason.
 */
export function sessionPhase(
  token: string,
  now: Date = new Date(),
): SessionPhase {
  const expiry = tokenExpiresAt(token);
  if (!expiry) return "active";

  const remaining = expiry.getTime() - now.getTime();
  if (remaining <= 0) return "expired";
  if (remaining <= EXPIRY_WARNING_LEAD_MS) return "expiring";

  return "active";
}
