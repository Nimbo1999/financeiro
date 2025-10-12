import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import MUIAppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import IconButton from "@mui/material/IconButton";
import Avatar from "@mui/material/Avatar";
import AccountCircleRoundedIcon from "@mui/icons-material/AccountCircleRounded";

import { Menu as MenuIcon } from "@mui/icons-material";
import { useLocation, useNavigation } from "react-router";
import { useMemo } from "react";
import { LinearProgress } from "@mui/material";

export interface AppBarProps {
  drawerWidth: number;
  /**
   * Function to toggle the sidebar on mobile devices.
   */
  handleDrawerToggle: VoidFunction;
  /**
   * User's email address to display in the AppBar.
   */
  email: string;
  /**
   * User's full name to display in the AppBar (optional).
   */
  fullName?: string;
}

export function AppBar({
  drawerWidth,
  handleDrawerToggle,
  email,
  fullName,
}: Readonly<AppBarProps>) {
  const { pathname } = useLocation();
  const { state } = useNavigation();

  const avatarInitials = useMemo(() => {
    return fullName
      ?.split(" ")
      .map((n, i, arr) => (i === 0 || i === arr.length - 1 ? n[0] : ""))
      .filter(Boolean)
      .join("")
      .toUpperCase();
  }, [fullName]);

  const appBarTitle = useMemo(
    () =>
      ({
        "/": "Dashboard",
        "/transactions": "Transactions",
        "/upload": "Upload CSV",
        "/users": "Users",
        "/settings": "Settings",
      }[pathname] ?? "Finance Tracker"),
    [pathname]
  );

  return (
    <MUIAppBar
      position="fixed"
      sx={{
        width: { sm: `calc(100% - ${drawerWidth}px)` },
        ml: { sm: `${drawerWidth}px` },
        backgroundColor: "background.paper",
        color: "text.primary",
        boxShadow: "0 1px 3px 0 rgb(0 0 0 / 0.1)",
      }}
    >
      <Toolbar>
        <IconButton
          color="inherit"
          edge="start"
          onClick={handleDrawerToggle}
          sx={{ mr: 2, display: { sm: "none" } }}
        >
          <MenuIcon />
        </IconButton>

        <Typography
          variant="h6"
          noWrap
          component="div"
          sx={{ flexGrow: 1, fontWeight: 600 }}
        >
          {appBarTitle}
        </Typography>

        <Stack direction="row" spacing={2} alignItems="center">
          <Box
            sx={{ textAlign: "right", display: { xs: "none", sm: "block" } }}
          >
            {fullName ? (
              <Typography variant="body2" fontWeight={600}>
                {fullName}
              </Typography>
            ) : null}

            <Typography variant="caption" color="text.secondary">
              {email}
            </Typography>
          </Box>
          <Avatar
            sx={{
              background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
              fontWeight: 700,
              color: "white",
            }}
          >
            {avatarInitials ?? <AccountCircleRoundedIcon color="inherit" />}
          </Avatar>
        </Stack>
      </Toolbar>

      {state !== "idle" ? <LinearProgress variant="indeterminate" /> : null}
    </MUIAppBar>
  );
}
