import { useState, useEffect, useCallback } from 'react';
import alertsService from '../services/alertService';
import { Alert, AlertSummary, AlertGroup } from '../types/alerts';

export function useAlerts() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [summary, setSummary] = useState<AlertSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchAlerts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await alertsService.getAlerts();
      setAlerts(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch alerts'));
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchAlertSummary = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await alertsService.getAlertSummary();
      setSummary(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch alert summary'));
    } finally {
      setLoading(false);
    }
  }, []);

  const getAlertGroups = useCallback(async (groupBy: string = 'severity'): Promise<AlertGroup[]> => {
    try {
      return await alertsService.getAlertGroups(groupBy);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch alert groups'));
      return [];
    }
  }, []);

  useEffect(() => {
    fetchAlerts();
    fetchAlertSummary();
  }, [fetchAlerts, fetchAlertSummary]);

  const hasCriticalAlerts = useCallback(async (): Promise<boolean> => {
    try {
      return await alertsService.hasCriticalAlerts();
    } catch (err) {
      return false;
    }
  }, []);

  return {
    alerts,
    summary,
    loading,
    error,
    fetchAlerts,
    fetchAlertSummary,
    getAlertGroups,
    hasCriticalAlerts
  };
}