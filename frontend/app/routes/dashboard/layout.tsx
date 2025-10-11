import Box from "@mui/material/Box";
import { Suspense, useState } from "react";
import { Await, Outlet, useLoaderData } from "react-router";
import { getCurrentSession } from "~/session";
import { Sidebar } from "~/components/sidebar";
import { AppBar } from "~/components/app-bar";
import { FinanceiroUserService } from "~/services/user.service";
import { FetchClient } from "~/clients/http";
import { env } from "~/environment";

import type { Route } from "./+types/layout";
import { LayoutError } from "~/components/layout-error";

export async function loader({ request }: Route.LoaderArgs) {
  const { email, accessToken } = await getCurrentSession(
    request.headers.get("Cookie")
  );
  const userService = new FinanceiroUserService(
    FetchClient.new(env.API_BASE_URL, accessToken)
  );
  return { userPromise: userService.getByEmail(email) };
}

export default function Home() {
  const { userPromise } = useLoaderData<typeof loader>();
  const [mobileOpen, setMobileOpen] = useState(false);
  const drawerWidth = 280;

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Suspense fallback={<div>Loading...</div>}>
      <Await resolve={userPromise} errorElement={<LayoutError />}>
        {(user) => (
          <Box sx={{ display: "flex", minHeight: "100vh" }}>
            <AppBar
              drawerWidth={drawerWidth}
              email={user.email}
              fullName={user.full_name}
              handleDrawerToggle={handleDrawerToggle}
            />

            <Sidebar
              drawerWidth={drawerWidth}
              mobileOpen={mobileOpen}
              handleDrawerToggle={handleDrawerToggle}
            />

            <Box
              component="main"
              sx={{
                flexGrow: 1,
                p: 3,
                width: { sm: `calc(100% - ${drawerWidth}px)` },
                mt: 8,
              }}
            >
              <Outlet context={{ user }} />
            </Box>
          </Box>
        )}
      </Await>
    </Suspense>
  );
}

export function ErrorBoundary() {
  return <div>Something went wrong in the layout</div>;
}
