import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import {
  declareSessionExpired,
  endSession,
  onSessionExpired,
  resumeSession,
} from "@/services/apiFetch";
import {
  SESSION_CHECK_INTERVAL_MS,
  sessionPhase,
} from "@/services/sessionExpiry";

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  createdAt: string;
  lastLogin: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (user: User, token: string) => void;
  logout: () => void;
  checkAuth: () => Promise<boolean>;
  /**
   * True once a request has been rejected for an expired session. The app stays
   * where it is; a prompt is shown over it rather than redirecting away.
   */
  sessionExpired: boolean;
  /** Log back in without losing the page, replaying whatever was interrupted. */
  resumeSession: (user: User, token: string) => void;
  /**
   * True while the session is close enough to expiry to warn about, so it can
   * be extended before anything fails rather than after.
   */
  sessionExpiring: boolean;
  /** Accept a renewed token and reschedule the next warning. */
  renewSession: (token: string) => void;
  /** Stop warning without renewing; the session runs out on its own. */
  dismissExpiryWarning: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [sessionExpired, setSessionExpired] = useState(false);
  const [sessionExpiring, setSessionExpiring] = useState(false);

  const isAuthenticated = !!user && !!token;

  const login = (userData: User, authToken: string) => {
    setSessionExpiring(false);
    setUser(userData);
    setToken(authToken);
    localStorage.setItem("authToken", authToken);
    localStorage.setItem("user", JSON.stringify(userData));
  };

  const logout = () => {
    // Reject anything still waiting on a retry so callers stop hanging.
    endSession();
    setSessionExpired(false);
    setUser(null);
    setToken(null);
    localStorage.removeItem("authToken");
    localStorage.removeItem("user");
  };

  /**
   * Restore an expired session in place: store the new token, then replay the
   * requests that were parked when it expired.
   */
  const restoreSession = (userData: User, authToken: string) => {
    login(userData, authToken);
    setSessionExpired(false);
    resumeSession();
  };

  const checkAuth = async (): Promise<boolean> => {
    const storedToken = localStorage.getItem("authToken");
    const storedUser = localStorage.getItem("user");

    if (!storedToken || !storedUser) {
      setIsLoading(false);
      return false;
    }

    try {
      // Verify token is still valid
      const response = await fetch("/api/auth/user", {
        headers: {
          Authorization: `Bearer ${storedToken}`,
          "Content-Type": "application/json",
        },
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
        setToken(storedToken);
        setIsLoading(false);
        return true;
      } else {
        // Token is invalid, clear storage
        logout();
        setIsLoading(false);
        return false;
      }
    } catch (error) {
      console.error("Auth check failed:", error);
      logout();
      setIsLoading(false);
      return false;
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  // A token is only validated at mount, so expiry is discovered by a request
  // failing. This is how the rest of the app hears about it.
  useEffect(() => onSessionExpired(() => setSessionExpired(true)), []);

  // The token states its own expiry, so the app watches for it locally rather
  // than waiting for a request to fail.
  //
  // Polled rather than scheduled: a timeout set hours ahead does not survive the
  // machine sleeping. It fires late, after the token has already lapsed, and the
  // renewal prompt would then be offered for a session that can no longer be
  // renewed - the refresh endpoint requires a token that is still valid.
  useEffect(() => {
    if (!token) return;

    const check = () => {
      const phase = sessionPhase(token);

      if (phase === "expired") {
        // Nothing to renew. Hand over to the prompt that can actually recover
        // the session rather than offering a button that cannot work.
        setSessionExpiring(false);
        declareSessionExpired();
        return;
      }

      setSessionExpiring(phase === "expiring");
    };

    check();
    const timer = setInterval(check, SESSION_CHECK_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [token]);

  const renewSession = (authToken: string) => {
    setSessionExpiring(false);
    setToken(authToken);
    localStorage.setItem("authToken", authToken);
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated,
    isLoading,
    login,
    logout,
    checkAuth,
    sessionExpired,
    resumeSession: restoreSession,
    sessionExpiring,
    renewSession,
    dismissExpiryWarning: () => setSessionExpiring(false),
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
