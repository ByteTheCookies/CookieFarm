import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { Toasty } from "@cloudflare/kumo/components/toast";
import { AuthProvider } from "@/features/auth/AuthProvider";
import { router } from "@/router";
import "@/styles/main.css";

const root = document.getElementById("root");
if (!root) {
  throw new Error("Root element not found");
}

createRoot(root).render(
  <StrictMode>
    <AuthProvider>
      <Toasty>
        <RouterProvider router={router} />
      </Toasty>
    </AuthProvider>
  </StrictMode>,
);
