"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";
import { Check, Copy, KeyRound } from "lucide-react";

export default function SettingsPage() {
  const apiKey = useAuthStore((s) => s.apiKey);
  const [copied, setCopied] = useState(false);

  async function copyApiKey() {
    if (!apiKey) return;
    await navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Configurações</h1>
        <p className="text-sm text-muted-foreground">
          Gerencie sua conta e preferências
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <KeyRound className="size-4" />
            API Key
          </CardTitle>
          <CardDescription>
            Use esta chave para autenticar o envio de métricas e eventos para a
            sua conta.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          <Label htmlFor="api_key">Sua API key</Label>
          <div className="flex gap-2">
            <Input
              id="api_key"
              readOnly
              value={apiKey ?? ""}
              placeholder="Log in para ver sua API key"
              className="font-mono"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={copyApiKey}
              disabled={!apiKey}
            >
              {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
