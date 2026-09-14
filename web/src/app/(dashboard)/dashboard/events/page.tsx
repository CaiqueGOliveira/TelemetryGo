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
import { useEventsStore } from "@/stores/events-store";

function severityVariant(
  severity: string
): "default" | "secondary" | "destructive" {
  if (severity === "critical") return "destructive";
  if (severity === "warning") return "secondary";
  return "default";
}

const severityLabel: Record<string, string> = {
  info: "info",
  warning: "alerta",
  critical: "crítico",
};

function formatTimestamp(timestamp: string) {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return timestamp;
  return date.toLocaleString("pt-BR");
}

export default function EventsPage() {
  const events = useEventsStore((s) => s.events);
  const isLoading = useEventsStore((s) => s.isLoading);
  const loadEvents = useEventsStore((s) => s.loadEvents);

  useEffect(() => {
    loadEvents();
  }, [loadEvents]);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Eventos</h1>
        <p className="text-sm text-muted-foreground">
          Acompanhe os eventos enviados pelos serviços
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Eventos recebidos</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <caption className="sr-only">Eventos recebidos dos serviços</caption>
              <TableHeader>
                <TableRow>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Serviço</TableHead>
                  <TableHead>Mensagem</TableHead>
                  <TableHead>Severidade</TableHead>
                  <TableHead>Timestamp</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading
                  ? Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell>
                          <Skeleton className="h-4 w-20" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-32" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-56" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-5 w-16 rounded-full" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-36" />
                        </TableCell>
                      </TableRow>
                    ))
                  : events.map((e) => (
                      <TableRow key={e.id}>
                        <TableCell className="font-medium">{e.type}</TableCell>
                        <TableCell>{e.service}</TableCell>
                        <TableCell className="max-w-96 truncate text-muted-foreground">
                          {e.message}
                        </TableCell>
                        <TableCell>
                          <Badge variant={severityVariant(e.severity)}>
                            {severityLabel[e.severity] ?? e.severity}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatTimestamp(e.timestamp)}
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