import Container from "@mui/material/Container";
import { useOutletContext } from "react-router";
import { type UseFormSetValue } from "react-hook-form";
import { formAction, SchemaForm } from "remix-forms";
import { clsx } from "clsx";

import { getCurrentSession } from "~/session";
import type { User } from "~/models/user";

import type { Route } from "./+types/upload";
import {
  Alert,
  Box,
  Grid,
  LinearProgress,
  styled,
  Typography,
} from "@mui/material";
import {
  uploadTransactionSchema,
  type UploadTransactionForm,
} from "~/forms/upload-transactions";
import { SubmitButton } from "~/components/submit-button";
import { useState } from "react";

export async function loader({ request }: Route.LoaderArgs) {
  await getCurrentSession(request.headers.get("Cookie"));
  return null;
}

const UploadStyles = {
  Container: styled(Box)`
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 50dvh;
    width: 100%;

    input {
      display: none;
    }

    label {
      width: 100%;
      height: 100%;
      border: 2px dashed ${({ theme }) => theme.palette.primary.main};
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s ease;
      background-color: ${({ theme }) => theme.palette.background.paper};

      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
    }

    &.dragging {
      label {
        > * {
          pointer-events: none;
        }
        border-color: ${({ theme }) => theme.palette.primary.light};
        background-color: #1c234f;
      }
    }
  `,
};

export default function Upload() {
  const { user } = useOutletContext<{ user: User }>();
  const [dragging, setDragging] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  const onDragOver = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    console.log("File is being dragged over");
    if (!dragging) setDragging(true);
  };
  const onDrop =
    (setValueFunction: UseFormSetValue<UploadTransactionForm>) =>
    (event: React.DragEvent<HTMLDivElement>) => {
      event.preventDefault();
      const files = event.dataTransfer.files;
      if (files.length > 0) {
        // Handle the dropped files here
        handleFileChange(setValueFunction)(files);
      }
      if (dragging) setDragging(false);
    };
  const onDragLeave = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    console.log("File is being dragged out");
    if (dragging) setDragging(false);
  };

  const handleFileChange =
    (setValueFunction: UseFormSetValue<UploadTransactionForm>) =>
    (files: FileList): void => {
      console.log("Selected files:", files);
      const formData = new FormData();

      for (const file of files) {
        formData.append("files", file);
      }

      // Validate the file using the schema
      const validation = uploadTransactionSchema.safeParse({ formData });
      if (!validation.success) {
        // Extract the first error message
        const error = validation.error.issues[0];
        setValidationError(error.message);
        return;
      }

      // Clear any previous validation errors
      setValidationError(null);
      // File is valid, proceed with upload logic
      setValueFunction("formData", formData, {
        shouldDirty: true,
        shouldValidate: true,
        shouldTouch: true,
      });
      console.log("File is valid and ready for upload");
    };

  return (
    <Container maxWidth="xl">
      <SchemaForm schema={uploadTransactionSchema} fieldComponent={Grid}>
        {({
          Field,
          register,
          formState: { isSubmitting, isLoading },
          setValue,
          getValues,
        }) => (
          <Grid container spacing={3}>
            <Field name="formData" size={12}>
              {({ name }) => (
                <>
                  <UploadStyles.Container
                    onDragOver={onDragOver}
                    onDrop={onDrop(setValue)}
                    onDragLeave={onDragLeave}
                    className={clsx({ dragging })}
                  >
                    <label htmlFor={name}>
                      <input
                        type="file"
                        {...register(name)}
                        id={name}
                        onChange={(e) => {
                          if (e.target.files && e.target.files.length > 0) {
                            handleFileChange(setValue)(e.target.files);
                          }
                        }}
                      />
                      {dragging ? (
                        <>
                          <Typography variant="h5" color="text.primary">
                            Release your CSV file to upload 🫳
                          </Typography>
                        </>
                      ) : getValues("formData") instanceof FormData ? (
                        <>
                          <Typography variant="h5" color="text.primary">
                            📃{" "}
                            {
                              (
                                getValues("formData").getAll("files") as File[]
                              )[0]?.name
                            }{" "}
                          </Typography>
                        </>
                      ) : (
                        <>
                          <Typography variant="h5" color="text.primary">
                            Drag and Drop your CSV file here
                          </Typography>
                          <Typography variant="h6" color="text.secondary">
                            or
                          </Typography>
                          <Typography variant="h5" color="text.primary">
                            Click to select a file
                          </Typography>
                        </>
                      )}
                    </label>
                  </UploadStyles.Container>

                  {validationError && (
                    <Alert severity="error" sx={{ mt: 2, width: "100%" }}>
                      {validationError}
                    </Alert>
                  )}

                  {isLoading || isSubmitting ? (
                    <Box sx={{ width: "100%", mt: 1 }}>
                      <LinearProgress sx={{ height: 6 }} color="primary" />
                    </Box>
                  ) : null}
                </>
              )}
            </Field>

            <Grid size={12}>
              <SubmitButton disabled={isLoading || isSubmitting}>
                Save
              </SubmitButton>
            </Grid>
          </Grid>
        )}
      </SchemaForm>
    </Container>
  );
}

export function meta() {
  return [
    { title: "Finance tracker - Upload" },
    {
      name: "description",
      content: "This is where you can upload your financial documents",
    },
  ];
}
