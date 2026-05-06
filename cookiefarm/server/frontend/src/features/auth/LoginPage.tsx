import { useActionState, useEffect } from "react";
import { Banner } from "@cloudflare/kumo/components/banner";
import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { WarningCircleIcon } from "@phosphor-icons/react";
import { useLocation, useNavigate } from "react-router";
import { ApiError } from "@/api/client";
import { useAuth } from "./AuthProvider";

type LoginState = {
  errorMessage: string | null;
  completed: boolean;
};

const initialState: LoginState = {
  errorMessage: null,
  completed: false,
};

function getFormString(formData: FormData, name: string) {
  const value = formData.get(name);

  return typeof value === "string" ? value.trim() : "";
}

function getLoginErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return error.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Login failed.";
}

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const redirectTarget =
    typeof location.state === "object" &&
      location.state &&
      "from" in location.state &&
      typeof location.state.from === "string"
      ? location.state.from
      : "/";

  const [state, submitAction, pending] = useActionState(
    async (_previousState: LoginState, formData: FormData): Promise<LoginState> => {
      const password = getFormString(formData, "password");
      const username = getFormString(formData, "username");

      try {
        await auth.login(password, username || undefined);
        return {
          errorMessage: null,
          completed: true,
        };
      } catch (error) {
        return {
          errorMessage: getLoginErrorMessage(error),
          completed: false,
        };
      }
    },
    initialState,
  );

  useEffect(() => {
    if (auth.status === "authenticated" || state.completed) {
      navigate(redirectTarget, { replace: true });
    }
  }, [auth.status, navigate, redirectTarget, state.completed]);

  return (
    <main
      className="flex min-h-screen items-center justify-center px-4 py-10"
      style={{
        backgroundImage: `url('/images/background-login.png')`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
      }}
    >
      <section className="dashboard-surface w-full max-w-md rounded-3xl border border-kumo-line p-6 shadow-sm">
        <div className="space-y-2">
          <p className="text-sm uppercase tracking-[0.2em] text-kumo-fg-secondary">
            CookieFarm
          </p>
          <h1 className="text-3xl font-semibold text-kumo-fg-primary">Player Login</h1>
          <p className="text-sm text-kumo-fg-secondary">
            Authenticate with the server password to access the dashboard.
          </p>
        </div>

        <form action={submitAction} className="mt-6 space-y-4">
          {state.errorMessage ? (
            <Banner
              variant="error"
              icon={<WarningCircleIcon weight="fill" />}
              title="Authentication failed"
              description={state.errorMessage}
            />
          ) : null}

          <Input
            name="username"
            label="Username (optional)"
            placeholder="cookieguest"
            autoComplete="username"
          />

          <Input
            name="password"
            type="password"
            label="Password"
            placeholder="Enter server password"
            autoComplete="current-password"
            required
          />

          <Button type="submit" variant="primary" className="w-full" loading={pending}>
            Sign in
          </Button>
        </form>
      </section>
    </main>
  );
}
