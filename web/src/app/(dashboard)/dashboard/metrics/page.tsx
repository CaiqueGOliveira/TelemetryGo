"use client";

import { useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useMetricsStore } from "@/stores/metrics-store";

function statusVariant(status: string): "default" | "secondary" | "destructive" {
  if (status === "ok") return "default";
  if (status === "warn") return "secondary";
  return "destructive";
}

function formatTimestamp(timestamp: string) {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return timestamp;
  return date.toLocaleString("pt-BR");
}

export default function MetricsPage() {
  const metrics = useMetricsStore((s) => s.metrics);
  const isLoading = useMetricsStore((s) => s.isLoading);
  const loadMetrics = useMetricsStore((s) => s.loadMetrics);

  useEffect(() => {
    loadMetrics();
  }, [loadMetrics]);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Métricas</h1>
        <p className="text-sm text-muted-foreground">
          Visualize todas as métricas coletadas dos seus serviços
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Métricas coletadas</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <caption className="sr-only">Métricas coletadas dos serviços</caption>
              <TableHeader>
                <TableRow>
                  <TableHead>Métrica</TableHead>
                  <TableHead>Serviço</TableHead>
                  <TableHead>Valor</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Timestamp</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading
                  ? Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell>
                          <Skeleton className="h-4 w-24" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-32" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-16" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-5 w-16 rounded-full" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-36" />
                        </TableCell>
                      </TableRow>
                    ))
                  : metrics.map((m) => (
                      <TableRow key={m.id}>
                        <TableCell className="font-medium">{m.name}</TableCell>
                        <TableCell>{m.service}</TableCell>
                        <TableCell>
                          {m.value} {m.unit}
                        </TableCell>
                        <TableCell>
                          <Badge variant={statusVariant(m.status)}>{m.status}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatTimestamp(m.timestamp)}
                        </TableCell>
                      </TableRow>
                    ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}