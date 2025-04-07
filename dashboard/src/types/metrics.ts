export interface Metric {
    name: string;
    labels: string[];
    cardinality: number;
    stats: MetricStats;
    lastUpdated: string;
    samples: MetricSample[];
  }
  
  export interface MetricStats {
    min: number;
    max: number;
    avg: number;
  }
  
  export interface MetricSample {
    labels: Record<string, string>;
    value: number;
    timestamp: string;
  }
  
  export interface TopMetric {
    name: string;
    cardinality: number;
    sampleRate: number;
  }
  
  export interface MetricHealth {
    name: string;
    exists: boolean;
    isStale: boolean;
    hasGaps: boolean;
    lastScraped: string;
    checkedAt: string;
  }