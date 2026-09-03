import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function MetricsPage() {
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
          <CardTitle>Em construção</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Aqui você poderá visualizar gráficos e séries temporais das suas
            métricas de telemetria.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
