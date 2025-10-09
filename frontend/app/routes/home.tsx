import Container from "@mui/material/Container";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import type { Route } from "./+types/home";
import { getCurrentSession } from "~/session";

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getCurrentSession(request.headers.get("Cookie"));
  return { email: session?.email };
}

export default function Home() {
  return (
    <Container maxWidth="xl">
      <Typography variant="h4" fontWeight={700} gutterBottom>
        Welcome back, Matheus! 👋
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Here's what's happening with your finances today.
      </Typography>

      {/* Placeholder for dashboard content */}
      <Box
        sx={{
          mt: 4,
          p: 8,
          textAlign: "center",
          backgroundColor: "background.paper",
          borderRadius: 3,
          border: "2px dashed",
          borderColor: "divider",
        }}
      >
        <Typography variant="h6" color="text.secondary" gutterBottom>
          Dashboard Content
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Your financial widgets and charts will appear here
        </Typography>
      </Box>
    </Container>
  );
}

export function meta() {
  return [
    { title: "Acompanhamento Financeiro" },
    {
      name: "description",
      content: "Bem vindo ao sistema de acompanhamento financeiro",
    },
  ];
}
