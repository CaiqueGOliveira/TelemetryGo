"use client";

import { useEffect } from "react";
import {
  Activity,
  ArrowDownRight,
  ArrowUpRight,
  Cpu,
  Database,
  Server,
} from "lucide-react";
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
import { useOverviewStore } from "@/stores/overview-store";
import { useMetricsStore } from "@/stores/metrics-store";
import { useEventsStore } from "@/stores/events-store";

const cardIcons: Record<string, typeof Server> = {
  services: Server,
  "active-metrics": Activity,
  "avg-cpu": Cpu,
  storage: Database,
};

function eventBadgeVariant(severity: string) {
  switch (severity) {
    case "critical":
      return "destructive";
    case "warning":
      return "secondary";
    default:
      return "default";
  }
}

const eventBadgeLabel: Record<string, string> = {
  info: "info",
  warning: "alerta",
  critical: "crítico",
};

export default function DashboardPage() {
  const cards = useOverviewStore((s) => s.cards);
  const overviewLoading = useOverviewStore((s) => s.isLoading);
  const loadOverview = useOverviewStore((s) => s.loadOverview);
  const metrics = useMetricsStore((s) => s.metrics);
  const metricsLoading = useMetricsStore((s) => s.isLoading);
  const loadMetrics = useMetricsStore((s) => s.loadMetrics);
  const events = useEventsStore((s) => s.events);
  const eventsLoading = useEventsStore((s) => s.isLoading);
  const loadEvents = useEventsStore((s) => s.loadEvents);

  useEffect(() => {
    loadOverview();
    loadMetrics();
    loadEvents();
  }, [loadOverview, loadMetrics, loadEvents]);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Visão Geral</h1>
        <p className="text-sm text-muted-foreground">
          Acompanhe a saúde da sua infraestrutura em tempo real
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {overviewLoading
          ? Array.from({ length: 4 }).map((_, i) => (
              <Card key={i}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="size-4 rounded" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="mb-2 h-7 w-16" />
                  <Skeleton className="h-3 w-28" />
                </CardContent>
              </Card>
            ))
          : cards.map((card) => {
              const Icon = cardIcons[card.id] ?? Server;
              return (
                <Card key={card.id}>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      {card.title}
                    </CardTitle>
                    <Icon className="size-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{card.value}</div>
                    <p className="flex items-center gap-1 text-xs text-muted-foreground">
                      {card.trend === "up" ? (
                        <ArrowUpRight className="size-3 text-positive" />
                      ) : (
                        <ArrowDownRight className="size-3 text-warning" />
                      )}
                      {card.change} nas últimas 24h
                    </p>
                  </CardContent>
                </Card>
              );
            })}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Métricas recentes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <caption className="sr-only">Métricas recentes dos serviços</caption>
              <TableHeader>
                <TableRow>
                  <TableHead>Métrica</TableHead>
                  <TableHead>Serviço</TableHead>
                  <TableHead>Valor</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {metricsLoading
                  ? Array.from({ length: 4 }).map((_, i) => (
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
                          <Badge
                            variant={
                              m.status === "ok"
                                ? "default"
                                : m.status === "warn"
                                  ? "secondary"
                                  : "destructive"
                            }
                          >
                            {m.status}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Eventos recentes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <caption className="sr-only">Eventos recentes dos serviços</caption>
              <TableHeader>
                <TableRow>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Serviço</TableHead>
                  <TableHead>Mensagem</TableHead>
                  <TableHead>Severidade</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {eventsLoading
                  ? Array.from({ length: 4 }).map((_, i) => (
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
                      </TableRow>
                    ))
                  : events.slice(0, 5).map((e) => (
                      <TableRow key={e.id}>
                        <TableCell className="font-medium">{e.type}</TableCell>
                        <TableCell>{e.service}</TableCell>
                        <TableCell className="max-w-64 truncate text-muted-foreground">
                          {e.message}
                        </TableCell>
                        <TableCell>
                          <Badge variant={eventBadgeVariant(e.severity) as "default" | "secondary" | "destructive"}>
                            {eventBadgeLabel[e.severity] ?? e.severity}
                          </Badge>
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