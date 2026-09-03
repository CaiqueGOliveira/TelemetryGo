"use client";

import { Activity, ArrowDownRight, ArrowUpRight, Cpu, Database, Server } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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

const cardIcons: Record<string, typeof Server> = {
  services: Server,
  "active-metrics": Activity,
  "avg-cpu": Cpu,
  storage: Database,
};

export default function DashboardPage() {
  const cards = useOverviewStore((s) => s.cards);
  const metrics = useMetricsStore((s) => s.metrics);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Visão Geral</h1>
        <p className="text-sm text-muted-foreground">
          Acompanhe a saúde da sua infraestrutura em tempo real
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => {
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
                    <ArrowUpRight className="size-3 text-emerald-500" />
                  ) : (
                    <ArrowDownRight className="size-3 text-amber-500" />
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
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Métrica</TableHead>
                <TableHead>Serviço</TableHead>
                <TableHead>Valor</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {metrics.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-medium">{m.name}</TableCell>
                  <TableCell>{m.service}</TableCell>
                  <TableCell>
                    {m.value}
                    {m.unit}
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
        </CardContent>
      </Card>
    </div>
  );
}
