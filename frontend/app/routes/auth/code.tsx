import TextField from "@mui/material/TextField";
import Grid from "@mui/material/Grid";
import Box from "@mui/material/Box";
import LinearProgress from "@mui/material/LinearProgress";
import Typography from "@mui/material/Typography";
import { SchemaForm, performMutation } from "remix-forms";
import { data, redirect, useActionData } from "react-router";
import { applySchema } from "composable-functions";
import {
  isRouteErrorResponse,
  Navigate,
  useLoaderData,
  useRouteError,
} from "react-router";
import { loginCodeSchema } from "~/forms/login";
import { logger } from "~/utils/logger";
import { SubmitButton } from "~/components/submit-button";
import { FinanceiroAuthService } from "~/services/auth.service";

import type { Route } from "./+types/code";
import { commitSession, getSession } from "~/session";

const EMAIL_ERROR_MESSAGE = "Email is required to access this page.";

export const loader = async ({ request }: Route.LoaderArgs) => {
  const cookie = await getSession(request.headers.get("Cookie"));
  if (cookie.has("session")) {
    logger.debug("User already logged in, redirecting to home.");
    return redirect("/");
  }
  const url = new URL(request.url);
  const email = url.searchParams.get("email");
  if (!email) {
    logger.warn("No email provided in query parameters");
    throw new Error(EMAIL_ERROR_MESSAGE);
  }
  logger.debug("Rendering login code page for email: %s", email);
  return { email };
};

const mutation = applySchema(loginCodeSchema)(async (values) => {
  const authService = new FinanceiroAuthService();
  logger.debug("Attempting login for email: %s", values.email);
  const data = await authService.login(values.email, values.code);
  logger.debug("Login successful for email: %s", values.email);
  return data;
});

export const action = async ({ request }: Route.ActionArgs) => {
  logger.debug("Processing code submission");
  const result = await performMutation({
    request,
    schema: loginCodeSchema,
    mutation,
  });
  if (!result.success) return data(result, 400);
  const { email, tokens } = result.data;
  const cookie = await getSession(request.headers.get("Cookie"));
  cookie.set("session", {
    accessToken: tokens.access_token,
    email,
    refreshToken: tokens.refresh_token,
  });
  return redirect("/", {
    headers: {
      "Set-Cookie": await commitSession(cookie),
    },
  });
};

export default function LoginCodeComponent() {
  const { email } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const formatServerError = (v: string) => {
    const response = JSON.parse(v);
    return [response.error.message, response.error.details]
      .filter(Boolean)
      .join(": ");
  };
  return (
    <SchemaForm
      schema={loginCodeSchema}
      values={{ email }}
      hiddenFields={["email"]}
      fieldComponent={Grid}
    >
      {({ Field, register, formState: { isLoading, isSubmitting } }) => (
        <Grid container spacing={3}>
          <Field name="email" />
          <Field name="code" label="Código" size={6} offset={3}>
            {({ errors, name, label }) => (
              <TextField
                autoFocus
                type="text"
                variant="outlined"
                label={label}
                required
                error={Boolean(errors?.length)}
                helperText={
                  Boolean(errors?.length) ? errors?.join(", ") : undefined
                }
                {...register(name)}
              />
            )}
          </Field>

          {actionData?.success === false ? (
            <Grid size={12}>
              {Object.entries(actionData.errors).map(([key, value]) =>
                key === "_global" ? (
                  <Typography key={key} color="error">
                    {value.map(formatServerError).join(", ")}
                  </Typography>
                ) : (
                  <Typography key={key} color="error">
                    {value.join(", ")}
                  </Typography>
                )
              )}
            </Grid>
          ) : null}

          <Grid size={12}>
            <SubmitButton disabled={isLoading || isSubmitting}>
              Confirmar
            </SubmitButton>

            {isLoading || isSubmitting ? (
              <Box sx={{ width: "100%", mt: 1 }}>
                <LinearProgress sx={{ height: 6 }} color="primary" />
              </Box>
            ) : null}
          </Grid>
        </Grid>
      )}
    </SchemaForm>
  );
}

export function ErrorBoundary() {
  const error = useRouteError();

  console.log("Route Error:", error);
  if (isRouteErrorResponse(error)) {
    return (
      <div>
        <h1>
          {error.status} {error.statusText}
        </h1>
        <p>{JSON.stringify(error.data)}</p>
      </div>
    );
  }

  if (error instanceof Error && error.message === EMAIL_ERROR_MESSAGE) {
    return <Navigate to="/login" replace />;
  }
  console.log("Unexpected Error:", error);
  return <p>Unknown error!</p>;
}

export function meta() {
  return [
    { title: "Financeiro - Login Code" },
    {
      name: "description",
      content: "Por favor, insira o código enviado para seu email.",
    },
  ];
}
