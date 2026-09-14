"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  forgotPasswordSchema,
  type ForgotPasswordFormData,
} from "@/schemas/auth-schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [resetToken, setResetToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  async function onSubmit(data: ForgotPasswordFormData) {
    setError(null);
    setSuccess(null);
    setResetToken(null);
    setIsLoading(true);
    try {
      const res = await api.post("/v1/auth/forgot-password", {
        email: data.email,
      });
      const body = res.data as {
        message?: string;
        reset_token?: string;
      };
      setSuccess(
        body.message ?? "Se o email existir, as instruções foram enviadas.",
      );
      if (body.reset_token) {
        setResetToken(body.reset_token);
      }
    } catch {
      setError("Não foi possível solicitar a redefinição. Tente novamente.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="w-full max-w-sm space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">
          Esqueci minha senha
        </h1>
        <p className="text-sm text-muted-foreground">
          Digite seu email e enviaremos instruções para redefinir a senha
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            placeholder="voce@exemplo.com"
            aria-label="Email"
            autoComplete="email"
            {...register("email")}
          />
          {errors.email && (
            <p className="text-sm text-destructive">{errors.email.message}</p>
          )}
        </div>

        {error && (
          <Alert variant="destructive" role="alert">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {success && (
          <Alert role="status">
            <AlertDescription>{success}</AlertDescription>
          </Alert>
        )}

        {resetToken && (
          <p className="text-center text-sm text-muted-foreground">
            <Link
              href={`/reset-password?token=${resetToken}`}
              className="text-primary hover:underline"
            >
              Redefinir senha agora (demo)
            </Link>
          </p>
        )}

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? "Enviando..." : "Enviar instruções"}
        </Button>
      </form>

      <p className="text-center text-sm text-muted-foreground">
        Lembrou a senha?{" "}
        <Link href="/login" className="text-primary hover:underline">
          Entrar
        </Link>
      </p>
    </div>
  );
}
