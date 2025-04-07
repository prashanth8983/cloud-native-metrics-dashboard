import React, { createContext, useContext } from 'react';
import { TimeRange, TimeRangePreset } from '../types/common';
import { useTimeRange } from '../hooks/useTimeRange';

interface TimeRangeContextType {
  timeRange: TimeRange;
  preset: TimeRangePreset;
  setTimeRangeFromPreset: (preset: TimeRangePreset) => void;
  setCustomTimeRange: (start: Date, end: Date) => void;
  refreshTimeRange: () => void;
}

const TimeRangeContext = createContext<TimeRangeContextType>({
  timeRange: {
    start: new Date(Date.now() - 3600000), // 1 hour ago
    end: new Date()
  },
  preset: '1h',
  setTimeRangeFromPreset: () => {},
  setCustomTimeRange: () => {},
  refreshTimeRange: () => {}
});

export const useTimeRangeContext = () => useContext(TimeRangeContext);

export const TimeRangeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const timeRangeHook = useTimeRange();
  
  return (
    <TimeRangeContext.Provider value={timeRangeHook}>
      {children}
    </TimeRangeContext.Provider>
  );
};