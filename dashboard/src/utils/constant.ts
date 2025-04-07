/**
 * Common time range presets
 */
export const TIME_RANGE_PRESETS = [
    { value: '15m', label: 'Last 15 minutes' },
    { value: '1h', label: 'Last 1 hour' },
    { value: '3h', label: 'Last 3 hours' },
    { value: '6h', label: 'Last 6 hours' },
    { value: '12h', label: 'Last 12 hours' },
    { value: '1d', label: 'Last 24 hours' },
    { value: '3d', label: 'Last 3 days' },
    { value: '7d', label: 'Last 7 days' },
    { value: '14d', label: 'Last 14 days' },
    { value: '30d', label: 'Last 30 days' },
  ];
  
  /**
   * Default dashboard metrics to display
   */
  export const DEFAULT_DASHBOARD_METRICS = [
    'node_cpu_seconds_total',
    'node_memory_MemFree_bytes',
    'node_disk_io_time_seconds_total',
    'node_network_receive_bytes_total',
    'node_network_transmit_bytes_total',
    'process_resident_memory_bytes',
    'http_requests_total',
    'http_request_duration_seconds',
  ];
  
  /**
   * Default refresh intervals
   */
  export const REFRESH_INTERVALS = [
    { value: 0, label: 'Off' },
    { value: 10000, label: '10s' },
    { value: 30000, label: '30s' },
    { value: 60000, label: '1m' },
    { value: 300000, label: '5m' },
    { value: 900000, label: '15m' },
  ];
  
  /**
   * Severity levels in order of priority
   */
  export const SEVERITY_LEVELS = [
    'critical',
    'high',
    'warning',
    'medium',
    'low',
    'info',
  ];