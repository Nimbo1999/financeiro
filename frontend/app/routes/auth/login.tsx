import TextField from "@mui/material/TextField";
import Grid from "@mui/material/Grid";
import LinearProgress from "@mui/material/LinearProgress";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { formAction, SchemaForm } from "remix-forms";
import { applySchema } from "composable-functions";
import { loginSchema } from "~/forms/login";
import { logger } from "~/utils/logger";
import { redirect } from "react-router";

import type { Route } from "./+types/login";
import { SubmitButton } from "~/components/submit-button";
import { FinanceiroAuthService } from "~/services/auth.service";
import { useActionData } from "react-router";
import { getSession } from "~/session";

export const loader = async ({ request }: Route.LoaderArgs) => {
  const cookie = await getSession(request.headers.get("Cookie"));
  return cookie.has("session") ? redirect("/") : null;
};

const mutation = applySchema(loginSchema)(async (values) => {
  const authService = new FinanceiroAuthService();
  logger.debug("Attempting to request login code for email: %s", values.email);
  const data = await authService.requestCode(values.email);
  logger.debug("Login code sent to email: %s", values.email);
  return { ...data, email: values.email };
});

export const action = async ({ request }: Route.ActionArgs) =>
  formAction({
    request,
    schema: loginSchema,
    mutation,
    successPath: (data) => {
      logger.info("Login successful for email: %s", data.email);
      return `/code?email=${encodeURIComponent(data.email)}`;
    },
  });

// Login Page Component
export default function LoginPage(_props: Route.ComponentProps) {
  const actionData = useActionData<typeof action>();

  const formatServerError = (v: string) => {
    const response = JSON.parse(v);
    return [response?.error?.message, response?.error?.details]
      .filter(Boolean)
      .join(": ");
  };

  return (
    <SchemaForm schema={loginSchema} fieldComponent={Grid}>
      {({ Field, register, formState: { isSubmitting, isLoading } }) => (
        <Grid container spacing={3}>
          <Field name="email" label="Email" size={12}>
            {({ errors, name, label }) => (
              <>
                <TextField
                  fullWidth
                  autoFocus
                  type="email"
                  variant="outlined"
                  label={label}
                  required
                  error={Boolean(errors?.length)}
                  helperText={
                    Boolean(errors?.length) ? errors?.join(", ") : undefined
                  }
                  {...register(name)}
                />
              </>
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
              Enviar Código de Login
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

export function meta() {
  return [
    { title: "Financeiro - Login" },
    {
      name: "description",
      content: "Faça login para acessar seu painel financeiro",
    },
  ];
}
