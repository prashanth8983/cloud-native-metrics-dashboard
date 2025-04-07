/**
 * Format a number with unit prefix (K, M, G, etc.)
 */
export function formatNumber(value: number): string {
    if (value === 0) return '0';
    
    const units = ['', 'K', 'M', 'G', 'T', 'P'];
    const k = 1000;
    const magnitude = Math.floor(Math.log(Math.abs(value)) / Math.log(k));
    
    if (magnitude === 0) return value.toFixed(2).replace(/\.?0+$/, '');
    
    const scaled = value / Math.pow(k, magnitude);
    return scaled.toFixed(2).replace(/\.?0+$/, '') + units[magnitude];
  }
  
  /**
   * Format bytes with appropriate units (KB, MB, GB, etc.)
   */
  export function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const k = 1024;
    const magnitude = Math.floor(Math.log(bytes) / Math.log(k));
    
    return (bytes / Math.pow(k, magnitude)).toFixed(2).replace(/\.?0+$/, '') + ' ' + units[magnitude];
  }
  
  /**
   * Format a duration in milliseconds to a human-readable string
   */
  export function formatDuration(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    
    const seconds = ms / 1000;
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    
    const minutes = seconds / 60;
    if (minutes < 60) return `${Math.floor(minutes)}m ${Math.floor(seconds % 60)}s`;
    
    const hours = minutes / 60;
    if (hours < 24) return `${Math.floor(hours)}h ${Math.floor(minutes % 60)}m`;
    
    const days = hours / 24;
    return `${Math.floor(days)}d ${Math.floor(hours % 24)}h`;
  }
  
  /**
   * Format a timestamp to a human-readable string
   */
  export function formatTimestamp(timestamp: string | Date): string {
    const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
    return date.toLocaleString();
  }
  
  /**
   * Format a percentage value
   */
  export function formatPercent(value: number): string {
    return `${(value * 100).toFixed(2)}%`;
  }