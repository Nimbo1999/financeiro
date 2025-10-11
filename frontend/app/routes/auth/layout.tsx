import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Container from "@mui/material/Container";
import IconButton from "@mui/material/IconButton";
import Typography from "@mui/material/Typography";
import { Link, Outlet, useLocation } from "react-router";

import {
  ArrowBack,
  AccountBalanceWallet as WalletIcon,
} from "@mui/icons-material";

// Authentication layout component
export default function LoginPage() {
  const { pathname } = useLocation();
  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)",
        position: "relative",
        overflow: "hidden",
      }}
    >
      {/* Background decoration */}
      <Box
        sx={{
          position: "absolute",
          top: -100,
          right: -100,
          width: 400,
          height: 400,
          borderRadius: "50%",
          background:
            "radial-gradient(circle, rgba(139, 159, 245, 0.15) 0%, transparent 70%)",
          filter: "blur(40px)",
        }}
      />
      <Box
        sx={{
          position: "absolute",
          bottom: -150,
          left: -150,
          width: 500,
          height: 500,
          borderRadius: "50%",
          background:
            "radial-gradient(circle, rgba(148, 112, 184, 0.15) 0%, transparent 70%)",
          filter: "blur(40px)",
        }}
      />

      <Container maxWidth="sm" sx={{ position: "relative", zIndex: 1 }}>
        <Card
          sx={{
            boxShadow: "0 8px 32px rgba(0, 0, 0, 0.4)",
            overflow: "hidden",
          }}
        >
          {/* Header with Gradient */}
          <Box
            sx={{
              background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
              p: 6,
              textAlign: "center",
              position: "relative",
            }}
          >
            {pathname === "/code" ? (
              <IconButton
                sx={{ position: "absolute", top: 16, left: 16 }}
                component={Link}
                to="/login"
              >
                <ArrowBack />
              </IconButton>
            ) : null}
            <Box
              sx={{
                display: "inline-block",
                width: 80,
                height: 80,
                backgroundColor: "rgba(255,255,255,0.95)",
                borderRadius: "20px",
                padding: "18px",
                mb: 2,
                boxShadow: "0 8px 24px rgba(0,0,0,0.15)",
              }}
            >
              <WalletIcon sx={{ fontSize: 44, color: "#667eea" }} />
            </Box>
            <Typography variant="h4" color="white" fontWeight={700}>
              {{
                "/login": "Bem vindo de volta!",
                "/code": "Bem vindo de volta!",
                "/refresh": "Aguarde",
              }[pathname] ?? "Gerencie suas finanças com facilidade"}
            </Typography>
            <Typography
              variant="body1"
              color="rgba(255,255,255,0.9)"
              sx={{ mt: 1 }}
            >
              {{
                "/login": "Faça login para acessar seu painel financeiro",
                "/code": "Por favor, insira o código enviado para seu email.",
                "/refresh": "Renovando sua sessão...",
              }[pathname] ?? "Gerencie suas finanças com facilidade"}
            </Typography>
          </Box>

          {/* Auth forms outlet */}
          <CardContent sx={{ p: 6 }}>
            <Outlet />
          </CardContent>
        </Card>

        <Typography
          variant="body2"
          color="text.secondary"
          align="center"
          sx={{ mt: 3 }}
        >
          © 2025 Finance Tracker. Todos os direitos reservados.
        </Typography>
      </Container>
    </Box>
  );
}
