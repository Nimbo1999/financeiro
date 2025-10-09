import zod from "zod";

export const SessionSchema = zod.object({
  email: zod.email({ error: "Invalid email address" }),
  accessToken: zod.jwt({ alg: "RS256" }),
  refreshToken: zod.jwt({ alg: "RS256" }),
});

export type Session = zod.infer<typeof SessionSchema>;
