import Container from "@mui/material/Container";
import { useOutletContext } from "react-router";

import { getCurrentSession } from "~/session";
import { MaintanceBox } from "~/components/maintance-box";
import type { User } from "~/models/user";

import type { Route } from "./+types/transactions";

export async function loader({ request }: Route.LoaderArgs) {
  await getCurrentSession(request.headers.get("Cookie"));
  return null;
}

export default function Transactions() {
  const { user } = useOutletContext<{ user: User }>();
  return (
    <Container maxWidth="xl">
      <MaintanceBox
        description={`${user.full_name}, your transactions page is under maintenance`}
      />
    </Container>
  );
}

export function meta() {
  return [
    { title: "Finance tracker - Transactions" },
    {
      name: "description",
      content: "This is where you can view and manage your transactions",
    },
  ];
}
