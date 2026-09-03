import { z } from "zod";

const messages = {
  required: (field: string) => `${field} é obrigatório`,
  invalidEmail: "Digite um email válido, ex.: voce@exemplo.com",
  minLength: (field: string, n: number) =>
    `${field} deve ter pelo menos ${n} caracteres`,
  maxLength: (field: string, n: number) =>
    `${field} deve ter no máximo ${n} caracteres`,
  lowercase: "A senha deve conter pelo menos 1 letra minúscula",
  uppercase: "A senha deve conter pelo menos 1 letra maiúscula",
  special: "A senha deve conter pelo menos 1 caractere especial",
  passwordMismatch: "As senhas não coincidem",
};

const passwordSchema = z
  .string()
  .min(8, messages.minLength("A senha", 8))
  .max(64, messages.maxLength("A senha", 64))
  .regex(/[a-z]/, messages.lowercase)
  .regex(/[A-Z]/, messages.uppercase)
  .regex(/[^a-zA-Z0-9]/, messages.special);

export const registerSchema = z
  .object({
    name: z
      .string()
      .min(1, messages.required("O nome"))
      .min(2, messages.minLength("O nome", 2)),
    email: z
      .string()
      .min(1, messages.required("O email"))
      .email(messages.invalidEmail),
    password: passwordSchema,
    confirmPassword: z
      .string()
      .min(1, messages.required("A confirmação de senha")),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: messages.passwordMismatch,
    path: ["confirmPassword"],
  });

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, messages.required("O email"))
    .email(messages.invalidEmail),
  password: z.string().min(1, messages.required("A senha")),
});

export const forgotPasswordSchema = z.object({
  email: z
    .string()
    .min(1, messages.required("O email"))
    .email(messages.invalidEmail),
});

export const resetPasswordSchema = z
  .object({
    password: passwordSchema,
    confirmPassword: z
      .string()
      .min(1, messages.required("A confirmação de senha")),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: messages.passwordMismatch,
    path: ["confirmPassword"],
  });

export type RegisterFormData = z.infer<typeof registerSchema>;
export type LoginFormData = z.infer<typeof loginSchema>;
export type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;
