import zod from "zod";

export const uploadTransactionSchema = zod.object({
  formData: zod
    .instanceof(FormData)
    .refine(
      (formData) => {
        // Check if files key exists in FormData
        const files = formData.getAll("files");
        return files.length > 0;
      },
      {
        message: "No files provided. Please upload at least one file.",
      }
    )
    .refine(
      (formData) => {
        // Check if all files are within size limit
        const files = formData.getAll("files");
        for (const file of files) {
          if (file instanceof File && file.size > 10 * 1024 * 1024) {
            return false;
          }
        }
        return true;
      },
      {
        message: "File size must be less than 10MB",
      }
    )
    .refine(
      (formData) => {
        // Check if all files are CSV type
        const files = formData.getAll("files");
        for (const file of files) {
          if (file instanceof File && file.type !== "text/csv") {
            return false;
          }
        }
        return true;
      },
      {
        message: "Unsupported file type. Please upload a CSV file.",
      }
    ),
});

export type UploadTransactionForm = zod.infer<typeof uploadTransactionSchema>;
