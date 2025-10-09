import Box from "@mui/material/Box";
import { useState } from "react";
import { Outlet, useLoaderData } from "react-router";
import type { Route } from "./+types/dashboard.layout";
import { getCurrentSession } from "~/session";
import { Sidebar } from "~/components/sidebar";
import { AppBar } from "~/components/app-bar";

export async function loader({ request }: Route.LoaderArgs) {
  const { email } = await getCurrentSession(request.headers.get("Cookie"));
  return { email };
}

export default function Home() {
  const { email } = useLoaderData<typeof loader>();
  const [mobileOpen, setMobileOpen] = useState(false);
  const drawerWidth = 280;

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <AppBar
        drawerWidth={drawerWidth}
        email={email}
        handleDrawerToggle={handleDrawerToggle}
      />

      <Sidebar
        drawerWidth={drawerWidth}
        mobileOpen={mobileOpen}
        handleDrawerToggle={handleDrawerToggle}
      />

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          mt: 8,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}
