import { createTheme, ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";

const theme = createTheme({
  cssVariables: true,
  colorSchemes: {
    light: true,
    dark: true,
  },
  palette: {
    mode: "light",
    primary: {
      main: "#667eea", // Primary purple-blue
      light: "#8b9ff5", // Lighter shade
      dark: "#4c63d2", // Darker shade
      contrastText: "#ffffff",
    },
    secondary: {
      main: "#764ba2", // Deep purple (from gradient end)
      light: "#9470b8", // Lighter shade
      dark: "#5a3880", // Darker shade
      contrastText: "#ffffff",
    },
    background: {
      default: "#f4f7fa", // Same as email body background
      paper: "#ffffff", // Card/Paper background
    },
    text: {
      primary: "#1f2937", // Dark gray for primary text
      secondary: "#4b5563", // Medium gray for secondary text
      disabled: "#9ca3af", // Light gray for disabled
    },
    info: {
      main: "#3b82f6", // Blue (from security notice)
      light: "#60a5fa",
      dark: "#2563eb",
      contrastText: "#ffffff",
    },
    warning: {
      main: "#f59e0b", // Orange/amber (from expiration box)
      light: "#fbbf24",
      dark: "#d97706",
      contrastText: "#ffffff",
    },
    success: {
      main: "#10b981", // Green
      light: "#34d399",
      dark: "#059669",
      contrastText: "#ffffff",
    },
    error: {
      main: "#ef4444", // Red
      light: "#f87171",
      dark: "#dc2626",
      contrastText: "#ffffff",
    },
    divider: "#e5e7eb", // Light border color
    grey: {
      50: "#f9fafb",
      100: "#f3f4f6",
      200: "#e5e7eb",
      300: "#d1d5db",
      400: "#9ca3af",
      500: "#6b7280",
      600: "#4b5563",
      700: "#374151",
      800: "#1f2937",
      900: "#111827",
    },
  },
  shape: {
    borderRadius: 12, // Rounded corners like email cards
  },
  shadows: [
    "none",
    "0 1px 3px 0 rgb(0 0 0 / 0.1)",
    "0 4px 6px -1px rgb(0 0 0 / 0.1)",
    "0 10px 15px -3px rgb(0 0 0 / 0.1)",
    "0 4px 20px rgba(0, 0, 0, 0.08)", // Email card shadow
    "0 8px 24px rgba(0, 0, 0, 0.15)", // Icon shadow
    "0 10px 40px rgba(0, 0, 0, 0.12)",
    "0 15px 50px rgba(0, 0, 0, 0.15)",
    "0 20px 60px rgba(0, 0, 0, 0.18)",
    "0 25px 70px rgba(0, 0, 0, 0.20)",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
    "none",
  ],
  typography: {
    fontFamily: [
      "-apple-system",
      "BlinkMacSystemFont",
      '"Segoe UI"',
      "Roboto",
      '"Helvetica Neue"',
      "Arial",
      "sans-serif",
    ].join(","),
  },
});

interface AppThemeProps {
  children: React.ReactNode;
}

export default function AppTheme({ children }: AppThemeProps) {
  return (
    <ThemeProvider theme={theme} defaultMode="system">
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
