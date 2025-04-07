export type TimeRange = {
    start: Date;
    end: Date;
  };
  
  export type TimeRangePreset = 
    | '15m'
    | '1h'
    | '3h'
    | '6h'
    | '12h'
    | '1d'
    | '3d'
    | '7d'
    | '14d'
    | '30d'
    | 'custom';
  
  export interface HealthStatus {
    status: 'up' | 'degraded' | 'down';
    version: string;
    uptime: string;
    timestamp: string;
    checks: Record<string, string>;
    details?: Record<string, any>;
  }
  
  export interface User {
    id: string;
    username: string;
    email: string;
    avatarUrl?: string;
    role: 'admin' | 'user' | 'viewer';
  }
  
  export interface ThemeOptions {
    mode: 'light' | 'dark';
    primaryColor: string;
  }