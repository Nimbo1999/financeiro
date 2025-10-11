import zod from "zod";

export const tokenContextSchema = zod.object({
  token: zod.jwt("Invalid token"),
});

export type TokenContext = zod.infer<typeof tokenContextSchema>;
