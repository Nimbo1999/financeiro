import {
  Box,
  Card,
  CardContent,
  CardHeader,
  Container,
  Grid,
  IconButton,
  LinearProgress,
  TextField,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import type { Route } from "./+types/create";
import { getCurrentSession } from "~/session";
import { formAction, SchemaForm } from "remix-forms";
import { createUserSchema } from "~/forms/create-user";
import { applySchema } from "composable-functions";
import { FinanceiroUserService } from "~/services/user.service";
import { logger } from "~/utils/logger";
import { FetchClient } from "~/clients/http";
import { env } from "~/environment";
import { tokenContextSchema } from "~/forms/common";
import { SubmitButton } from "~/components/submit-button";
import { Link } from "react-router";

export const loader = async ({ request }: Route.LoaderArgs) => {
  await getCurrentSession(request.headers.get("Cookie"));
  return null;
};

const mutation = applySchema(
  createUserSchema,
  tokenContextSchema
)(async (values, { token }) => {
  const service = new FinanceiroUserService(
    FetchClient.new(env.API_BASE_URL, token)
  );
  logger.debug("Attempting to create user with email: %s", values.email);
  const data = await service.create(values);
  return { ...data, email: values.email };
});

export const action = async ({ request }: Route.ActionArgs) => {
  const { accessToken } = await getCurrentSession(
    request.headers.get("Cookie")
  );
  return formAction({
    request,
    schema: createUserSchema,
    mutation,
    context: {
      token: accessToken,
    },
    successPath: (value) => {
      logger.info("User for %s created sucessfully.", value.full_name);
      return "/users";
    },
  });
};

export default function CreateUserComponent() {
  return (
    <Container maxWidth="sm" sx={{ py: 4 }}>
      <Card>
        <CardHeader
          avatar={
            <IconButton component={Link} to="..">
              <ArrowBackIcon />
            </IconButton>
          }
          title="Create User"
          slotProps={{
            title: {
              variant: "h5",
              component: "h5",
            },
          }}
        />
        <CardContent>
          <SchemaForm schema={createUserSchema} fieldComponent={Grid}>
            {({ Field, register, formState: { isSubmitting, isLoading } }) => (
              <Grid container spacing={3}>
                <Field name="full_name" label="Full Name" size={12}>
                  {({ errors, name, label }) => (
                    <TextField
                      fullWidth
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

                <Field name="email" label="Email" size={12}>
                  {({ errors, name, label }) => (
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
                  )}
                </Field>

                <Grid size={12}>
                  <SubmitButton disabled={isLoading || isSubmitting}>
                    Create
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
        </CardContent>
      </Card>
    </Container>
  );
}

export function meta() {
  return [
    { title: "Finance tracker - Create users" },
    {
      name: "description",
      content: "Create a new user for the finance tracker app",
    },
  ];
}
