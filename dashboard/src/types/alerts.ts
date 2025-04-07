export interface Alert {
    name: string;
    state: 'firing' | 'pending' | 'resolved';
    severity: string;
    labels: Record<string, string>;
    annotations: Record<string, string>;
    summary: string;
    activeAt: string;
    value: number;
  }
  
  export interface AlertGroup {
    name: string;
    count: number;
    alerts: Alert[];
  }
  
  export interface AlertSummary {
    firingCount: number;
    pendingCount: number;
    resolvedCount: number;
    totalCount: number;
    severityBreakdown: SeverityCount[];
    mostRecentAlert?: Alert;
    timeSinceLastAlert?: string;
    lastUpdated: string;
  }
  
  export interface SeverityCount {
    severity: string;
    count: number;
  }