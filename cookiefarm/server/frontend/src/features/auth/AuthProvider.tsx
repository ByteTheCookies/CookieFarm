import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import useSWR from "swr";
import { Navigate, useLocation } from "react-router";
import {
  authVerifyKey,
  login as loginRequest,
  logout as logoutRequest,
  verifyAuth,
  type AuthSession,
} from "@/api/auth";
import { PageSkeleton } from "@/components/PageSkeleton";

type AuthStatus = "checking" | "authenticated" | "anonymous";

type AuthContextValue = {
  status: AuthStatus;
  user: AuthSession | null;
  login: (password: string, username?: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<boolean>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider(props: Readonly<{ children: ReactNode }>) {
  const [status, setStatus] = useState<AuthStatus>("checking");
  const [user, setUser] = useState<AuthSession | null>(null);
  const { data: authData, mutate: mutateAuth } = useSWR(authVerifyKey, verifyAuth, {
    refreshInterval: status === "authenticated" ? 60_000 : 0,
  });

  const refresh = useCallback(async (): Promise<boolean> => {
    const session = await mutateAuth();
    setUser(session ?? null);
    setStatus(session ? "authenticated" : "anonymous");
    return Boolean(session);
  }, [mutateAuth]);

  useEffect(() => {
    if (authData === undefined) {
      return;
    }
    setUser(authData ?? null);
    setStatus(authData ? "authenticated" : "anonymous");
  }, [authData]);

  const login = useCallback(
    async (password: string, username?: string) => {
      await loginRequest({
        ...(username ? { username } : {}),
        password,
      });
      const session = { username: username || "cookieguest" };
      mutateAuth(session, { revalidate: false });
      setUser(session);
      setStatus("authenticated");
    },
    [mutateAuth],
  );

  const logout = useCallback(async () => {
    try {
      await logoutRequest();
    } finally {
      mutateAuth(null, { revalidate: false });
      setUser(null);
      setStatus("anonymous");
    }
  }, [mutateAuth]);

  const contextValue = useMemo(
    () => ({
      status,
      user,
      login,
      logout,
      refresh,
    }),
    [status, user, login, logout, refresh],
  );

  return (
    <AuthContext.Provider value={contextValue}>
      {props.children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}

export function RequireAuth(props: Readonly<{ children: ReactNode }>) {
  const auth = useAuth();
  const location = useLocation();

  if (auth.status === "checking") {
    return <PageSkeleton />;
  }

  if (auth.status !== "authenticated") {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return props.children;
}
