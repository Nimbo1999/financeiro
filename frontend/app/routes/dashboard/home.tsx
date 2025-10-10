import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import { useOutletContext } from "react-router";

import { getCurrentSession } from "~/session";
import { MaintanceBox } from "~/components/maintance-box";
import type { User } from "~/models/user";

import type { Route } from "./+types/home";

export async function loader({ request }: Route.LoaderArgs) {
  await getCurrentSession(request.headers.get("Cookie"));
  return null;
}

export default function Home() {
  const { user } = useOutletContext<{ user: User }>();
  return (
    <Container maxWidth="xl">
      <Typography variant="h4" fontWeight={700} gutterBottom>
        Welcome back, {user.full_name}! 👋
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Here's what's happening with your finances today.
      </Typography>

      {/* Placeholder for dashboard content */}
      <MaintanceBox description="Your financial widgets and charts will appear here" />
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
