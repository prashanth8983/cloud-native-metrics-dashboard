import { TimeRangePreset } from '../types/common';

/**
 * Get sensible step size for a time range (in seconds)
 */
export function getStepSize(start: Date, end: Date): number {
  const durationMs = end.getTime() - start.getTime();
  const durationSeconds = durationMs / 1000;
  
  // Aim for ~100-200 data points
  let step = Math.max(60, Math.floor(durationSeconds / 200)); // At least 1 minute
  
  // Round to sensible intervals
  if (step < 300) { // Less than 5 minutes
    step = 60; // 1 minute
  } else if (step < 900) { // Less than 15 minutes
    step = 300; // 5 minutes
  } else if (step < 3600) { // Less than 1 hour
    step = 900; // 15 minutes
  } else if (step < 21600) { // Less than 6 hours
    step = 3600; // 1 hour
  } else if (step < 86400) { // Less than 1 day
    step = 21600; // 6 hours
  } else {
    step = 86400; // 1 day
  }
  
  return step;
}

/**
 * Get display text for a time range preset
 */
export function getTimeRangeText(preset: TimeRangePreset): string {
  switch (preset) {
    case '15m': return 'Last 15 minutes';
    case '1h': return 'Last 1 hour';
    case '3h': return 'Last 3 hours';
    case '6h': return 'Last 6 hours';
    case '12h': return 'Last 12 hours';
    case '1d': return 'Last 24 hours';
    case '3d': return 'Last 3 days';
    case '7d': return 'Last 7 days';
    case '14d': return 'Last 14 days';
    case '30d': return 'Last 30 days';
    case 'custom': return 'Custom range';
    default: return 'Last 1 hour';
  }
}

/**
 * Format a date for display, with different detail levels based on range
 */
export function formatDateForRange(date: Date, start: Date, end: Date): string {
  const durationHours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
  
  if (durationHours <= 24) {
    // For ranges <= 24 hours, show time only
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  } else if (durationHours <= 72) {
    // For ranges <= 3 days, show day and time
    return date.toLocaleString([], {
      weekday: 'short',
      hour: '2-digit',
      minute: '2-digit'
    });
  } else {
    // For larger ranges, show date and time
    return date.toLocaleString([], {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
}