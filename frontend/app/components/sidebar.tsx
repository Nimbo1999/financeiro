import Box from "@mui/material/Box";
import Divider from "@mui/material/Divider";
import Drawer from "@mui/material/Drawer";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import DashboardIcon from "@mui/icons-material/Dashboard";
import WalletIcon from "@mui/icons-material/AccountBalanceWallet";
import SettingsIcon from "@mui/icons-material/Settings";
import LogoutIcon from "@mui/icons-material/Logout";
import UploadIcon from "@mui/icons-material/Upload";
import PeopleAltIcon from "@mui/icons-material/PeopleAlt";
import { styled } from "@mui/material";
import { NavLink } from "react-router";

export interface SidebarProps {
  drawerWidth: number;
  mobileOpen: boolean;
  handleDrawerToggle: VoidFunction;
}

const Styled = {
  ListItemButton: styled(ListItemButton)`
    border-radius: ${({ theme }) => theme.spacing(2)};

    .nav-text > span {
      font-weight: 400;
    }

    &.active {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;

      &:hover {
        background: linear-gradient(135deg, #5a6fd8 0%, #6a4391 100%);
      }

      .icon {
        color: white;
      }
      .nav-text > span {
        font-weight: 600;
      }
    }
  `,
};

export function Sidebar({
  drawerWidth,
  mobileOpen,
  handleDrawerToggle,
}: Readonly<SidebarProps>) {
  const menuItems = [
    { text: "Dashboard", icon: <DashboardIcon />, path: "/" },
    { text: "Transactions", icon: <WalletIcon />, path: "/transactions" },
    { text: "Upload CSV", icon: <UploadIcon />, path: "/upload" },
    { text: "Users", icon: <PeopleAltIcon />, path: "/users" },
  ];

  const drawer = (
    <Box sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {/* Sidebar Header */}
      <Box
        sx={{
          p: 3,
          background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
          color: "white",
        }}
      >
        <Stack direction="row" spacing={2} alignItems="center">
          <Box
            sx={{
              width: 48,
              height: 48,
              backgroundColor: "rgba(255,255,255,0.95)",
              borderRadius: "12px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <WalletIcon sx={{ fontSize: 28, color: "#667eea" }} />
          </Box>
          <Box>
            <Typography variant="h6" fontWeight={700}>
              Finance Tracker
            </Typography>
            <Typography variant="caption" sx={{ opacity: 0.9 }}>
              Your money, simplified
            </Typography>
          </Box>
        </Stack>
      </Box>

      <Divider />

      {/* Navigation Items */}
      <List sx={{ flexGrow: 1, px: 2, py: 2 }}>
        {menuItems.map((item) => (
          <ListItem key={item.text} disablePadding sx={{ mb: 1 }}>
            {/* @ts-ignore */}
            <Styled.ListItemButton component={NavLink} to={item.path}>
              <ListItemIcon className="icon">{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} className="nav-text" />
            </Styled.ListItemButton>
          </ListItem>
        ))}
      </List>

      <Divider />

      {/* User Section */}
      <Box sx={{ p: 2 }}>
        <ListItem disablePadding>
          {/* @ts-ignore */}
          <Styled.ListItemButton component={NavLink} to="/settings">
            <ListItemIcon>
              <SettingsIcon />
            </ListItemIcon>
            <ListItemText primary="Settings" />
          </Styled.ListItemButton>
        </ListItem>

        <ListItem disablePadding sx={{ mt: 1 }}>
          {/* @ts-ignore */}
          <Styled.ListItemButton onClick={() => {}}>
            <ListItemIcon>
              <LogoutIcon />
            </ListItemIcon>
            <ListItemText primary="Logout" />
          </Styled.ListItemButton>
        </ListItem>
      </Box>
    </Box>
  );

  return (
    <Box
      component="nav"
      sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
    >
      {/* Mobile drawer */}
      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={handleDrawerToggle}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: "block", sm: "none" },
          "& .MuiDrawer-paper": {
            boxSizing: "border-box",
            width: drawerWidth,
            backgroundColor: "background.paper",
          },
        }}
      >
        {drawer}
      </Drawer>

      {/* Desktop drawer */}
      <Drawer
        variant="permanent"
        sx={{
          display: { xs: "none", sm: "block" },
          "& .MuiDrawer-paper": {
            boxSizing: "border-box",
            width: drawerWidth,
            backgroundColor: "background.paper",
            borderRight: "1px solid",
            borderColor: "divider",
          },
        }}
        open
      >
        {drawer}
      </Drawer>
    </Box>
  );
}
