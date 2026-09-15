/**
 * The single door every authenticated request goes through.
 *
 * Before this, six modules read the token from localStorage and called `fetch`
 * themselves, and almost none of them recognised a 401. An expired session
 * therefore produced actions that quietly did nothing, with the redirect to the
 * login page only happening on the next full page load.
 *
 * Here a 401 is not a failure. The request is parked, the app is told the
 * session expired, and the request is replayed once the user logs back in - so
 * the action they asked for still happens instead of being lost.
 */

const TOKEN_KEY = "authToken";

type SessionExpiredListener = () => void;

interface ParkedRequest {
  replay: () => void;
  reject: (reason: Error) => void;
}

const listeners = new Set<SessionExpiredListener>();
let parked: ParkedRequest[] = [];
let expired = false;

/**
 * Subscribe to session expiry. Returns an unsubscribe function.
 */
export function onSessionExpired(listener: SessionExpiredListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function isSessionExpired(): boolean {
  return expired;
}

function withAuth(init: RequestInit): RequestInit {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) return init;

  return {
    ...init,
    headers: { ...(init.headers ?? {}), Authorization: `Bearer ${token}` },
  };
}

function signalExpired(): void {
  // Many requests can fail at once - a page full of cards refreshing. The app
  // should be told once, not once per request.
  if (expired) return;

  expired = true;
  listeners.forEach((listener) => listener());
}

/**
 * Fetch with the current token attached. A 401 parks the request until the
 * session is restored, at which point it is replayed and the original caller's
 * promise resolves as though nothing had happened.
 *
 * Request bodies must be replayable - a string, FormData or URLSearchParams,
 * which is everything this app sends. A single-use stream could not be retried.
 */
export async function apiFetch(
  input: RequestInfo | URL,
  init: RequestInit = {},
): Promise<Response> {
  const response = await fetch(input, withAuth(init));

  if (response.status !== 401) {
    return response;
  }

  return new Promise<Response>((resolve, reject) => {
    parked.push({
      // Replayed with plain fetch, so a still-rejected request surfaces the 401
      // to the caller rather than parking again and looping.
      replay: () => {
        fetch(input, withAuth(init)).then(resolve, reject);
      },
      reject,
    });

    signalExpired();
  });
}

/**
 * Called once the user has logged back in. Replays everything that was parked.
 */
export function resumeSession(): void {
  expired = false;

  const queued = parked;
  parked = [];
  queued.forEach((request) => request.replay());
}

/**
 * Called when the user abandons the session rather than logging back in.
 * Parked requests are rejected so their callers stop waiting.
 */
export function endSession(): void {
  expired = false;

  const queued = parked;
  parked = [];
  queued.forEach((request) =>
    request.reject(new Error("Signed out before the request could be retried")),
  );
}
