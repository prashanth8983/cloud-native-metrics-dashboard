import { useState, useCallback } from 'react';
import { TimeRange, TimeRangePreset } from '../types/common';

export function useTimeRange() {
  const [timeRange, setTimeRange] = useState<TimeRange>({
    start: new Date(Date.now() - 3600000), // 1 hour ago
    end: new Date()
  });
  
  const [preset, setPreset] = useState<TimeRangePreset>('1h');

  const setTimeRangeFromPreset = useCallback((newPreset: TimeRangePreset) => {
    const end = new Date();
    let start: Date;
    
    switch (newPreset) {
      case '15m':
        start = new Date(end.getTime() - 15 * 60 * 1000);
        break;
      case '1h':
        start = new Date(end.getTime() - 60 * 60 * 1000);
        break;
      case '3h':
        start = new Date(end.getTime() - 3 * 60 * 60 * 1000);
        break;
      case '6h':
        start = new Date(end.getTime() - 6 * 60 * 60 * 1000);
        break;
      case '12h':
        start = new Date(end.getTime() - 12 * 60 * 60 * 1000);
        break;
      case '1d':
        start = new Date(end.getTime() - 24 * 60 * 60 * 1000);
        break;
      case '3d':
        start = new Date(end.getTime() - 3 * 24 * 60 * 60 * 1000);
        break;
      case '7d':
        start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '14d':
        start = new Date(end.getTime() - 14 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        start = new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      case 'custom':
        // Don't change the time range for custom
        return;
      default:
        start = new Date(end.getTime() - 60 * 60 * 1000); // Default to 1 hour
    }
    
    setTimeRange({ start, end });
    setPreset(newPreset);
  }, []);

  const setCustomTimeRange = useCallback((start: Date, end: Date) => {
    setTimeRange({ start, end });
    setPreset('custom');
  }, []);

  const refreshTimeRange = useCallback(() => {
    if (preset !== 'custom') {
      setTimeRangeFromPreset(preset);
    }
  }, [preset, setTimeRangeFromPreset]);

  return {
    timeRange,
    preset,
    setTimeRangeFromPreset,
    setCustomTimeRange,
    refreshTimeRange
  };
}