import { redirect } from "react-router";

import type { Route } from "./+types/login";
import { commitSession, destroySession, getSession } from "~/session";
import { Box, CircularProgress } from "@mui/material";
import { FinanceiroAuthService } from "~/services/auth.service";
import { FetchClient } from "~/clients/http";
import { env } from "~/environment";

export const loader = async ({ request }: Route.LoaderArgs) => {
  const cookie = await getSession(request.headers.get("Cookie"));
  const session = cookie.get("session");
  if (!session) {
    return redirect("/login");
  }
  const service = new FinanceiroAuthService(FetchClient.new(env.API_BASE_URL));
  try {
    const { tokens } = await service.refreshToken(session.refreshToken);

    session.accessToken = tokens.access_token;
    session.refreshToken = tokens.refresh_token;
    cookie.set("session", session);

    const url = new URL(request.url);
    const redirectTo = url.searchParams.get("redirectTo") || "/";
    return redirect(redirectTo, {
      headers: {
        "Set-Cookie": await commitSession(cookie),
      },
    });
  } catch (_) {
    return redirect("/login", {
      headers: {
        "Set-Cookie": await destroySession(cookie),
      },
    });
  }
};

// Refresh Page Component
export default function RefreshPage(_props: Route.ComponentProps) {
  return (
    <Box sx={{ display: "flex", justifyContent: "center" }}>
      <CircularProgress />
    </Box>
  );
}

export function meta() {
  return [
    { title: "Financeiro - Renovando Sessão" },
    {
      name: "description",
      content:
        "Renove sua sessão para continuar acessando seu painel financeiro",
    },
  ];
}
