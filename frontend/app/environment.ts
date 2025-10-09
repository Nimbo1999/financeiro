import zod from "zod";

const envSchema = zod.object({
  API_BASE_URL: zod.url("API_BASE_URL must be a valid URL"),
  LOGGER_LEVEL: zod
    .enum(["debug", "info", "warn", "error"])
    .optional()
    .default("info"),
  SESSION_SECRET: zod
    .string()
    .min(16, "SESSION_SECRET must be at least 16 characters long"),
});

export const env = envSchema.parse(process.env);
