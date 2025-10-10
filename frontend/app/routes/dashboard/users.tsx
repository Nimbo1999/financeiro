import Container from "@mui/material/Container";
import {
  useOutletContext,
  useSearchParams,
  createSearchParams,
  Await,
} from "react-router";

import { getCurrentSession } from "~/session";
import { MaintanceBox } from "~/components/maintance-box";
import type { User } from "~/models/user";

import type { Route } from "./+types/users";
import { FinanceiroUserService } from "~/services/user.service";
import { FetchClient } from "~/clients/http";
import { env } from "~/environment";
import { Suspense } from "react";

function getSearchParams(searchParams: URLSearchParams): URLSearchParams {
  const params = createSearchParams(searchParams);
  if (params.has("page") === false) {
    params.set("page", "1");
  }
  if (params.has("page_size") === false) {
    params.set("page_size", "10");
  }
  if (params.has("order_by") === false) {
    params.set("order_by", "updated_at");
  }
  if (params.has("sort") === false) {
    params.set("sort", "desc");
  }
  return params;
}

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url);
  const searchParams = getSearchParams(url.searchParams);
  const { accessToken } = await getCurrentSession(
    request.headers.get("Cookie")
  );

  const service = new FinanceiroUserService(
    FetchClient.new(env.API_BASE_URL, accessToken)
  );
  const usersPromise = service.getList(searchParams);

  return { searchParams: Object.fromEntries(searchParams), usersPromise };
}

export default function Users({
  loaderData: { searchParams, usersPromise },
}: Route.ComponentProps) {
  const { user } = useOutletContext<{ user: User }>();
  const [_] = useSearchParams(searchParams);
  return (
    <Container maxWidth="xl">
      <Suspense>
        <Await resolve={usersPromise}>
          {(users) => {
            console.log({ users });
            return (
              <MaintanceBox
                description={`${user.full_name}, your users page is under maintenance`}
              />
            );
          }}
        </Await>
      </Suspense>
    </Container>
  );
}

export function meta() {
  return [
    { title: "Finance tracker - Users" },
    {
      name: "description",
      content: "This is where you can manage your users",
    },
  ];
}
