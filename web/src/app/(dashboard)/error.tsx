"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20 text-center">
      <h1 className="text-xl font-semibold">Algo deu errado</h1>
      <p className="max-w-sm text-sm text-muted-foreground">
        Ocorreu um erro ao carregar o dashboard. Tente novamente.
      </p>
      <Button type="button" variant="outline" onClick={() => reset()}>
        Tentar novamente
      </Button>
    </div>
  );
}