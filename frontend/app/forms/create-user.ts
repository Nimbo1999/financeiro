import zod from "zod";

export const createUserSchema = zod.object({
  full_name: zod
    .string()
    .min(5, "Full name must be at least 5 characters long")
    .refine(
      (val) => val.trim().split(" ").length >= 2,
      "Please enter your full name (first and last name)"
    )
    .trim(),
  email: zod.email("You must enter a valid email").trim(),
});

export type CreateUserForm = zod.infer<typeof createUserSchema>;
