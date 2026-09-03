export type MetricStatus = "ok" | "warn" | "crit";

export interface Metric {
  id: string;
  name: string;
  service: string;
  value: string;
  unit: string;
  status: MetricStatus;
  timestamp: string;
}

export interface MetricDataPoint {
  timestamp: string;
  value: number;
}

export interface MetricSeries {
  name: string;
  service: string;
  unit: string;
  data: MetricDataPoint[];
}

export interface OverviewCard {
  id: string;
  title: string;
  value: string;
  change: string;
  trend: "up" | "down";
}
