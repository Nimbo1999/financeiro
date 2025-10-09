import zod from "zod";

export const loginSchema = zod.object({
  email: zod.email("Você deve inserir um email válido").meta({
    id: "email",
    title: "Email",
    description: "Por favor, insira seu email",
  }),
});

export type LoginForm = zod.infer<typeof loginSchema>;

export const loginCodeSchema = loginSchema.extend({
  code: zod
    .string()
    .min(6, "O código deve ter 6 dígitos")
    .max(6, "O código deve ter 6 dígitos")
    .meta({
      id: "code",
      title: "Código",
      description:
        "Por favor, insira o código de 6 dígitos enviado para seu email",
    }),
});

export type LoginCodeForm = zod.infer<typeof loginCodeSchema>;
