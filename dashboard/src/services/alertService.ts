import api from './api';
import { Alert, AlertGroup, AlertSummary } from '../types/alerts';

/**
 * Service for interacting with the alerts API
 */
export const alertsService = {
  /**
   * Get all current alerts
   */
  async getAlerts(): Promise<Alert[]> {
    const response = await api.get('/alerts');
    return response.data.alerts;
  },
  
  /**
   * Get a summary of alert status
   */
  async getAlertSummary(): Promise<AlertSummary> {
    const response = await api.get('/alerts/summary');
    return response.data;
  },
  
  /**
   * Get alerts grouped by a specific label
   * @param groupBy The label to group by (default: 'severity')
   */
  async getAlertGroups(groupBy: string = 'severity'): Promise<AlertGroup[]> {
    const response = await api.get('/alerts/groups', {
      params: { by: groupBy }
    });
    return response.data.groups;
  },
  
  /**
   * Get alerts filtered by severity
   * @param severity The severity to filter by
   */
  async getAlertsBySeverity(severity: string): Promise<Alert[]> {
    const alerts = await this.getAlerts();
    return alerts.filter(alert => alert.severity === severity);
  },
  
  /**
   * Get alerts filtered by state
   * @param state The state to filter by
   */
  async getAlertsByState(state: 'firing' | 'pending' | 'resolved'): Promise<Alert[]> {
    const alerts = await this.getAlerts();
    return alerts.filter(alert => alert.state === state);
  },
  
  /**
   * Get the count of alerts by state
   */
  async getAlertCountsByState(): Promise<Record<string, number>> {
    const summary = await this.getAlertSummary();
    
    return {
      firing: summary.firingCount,
      pending: summary.pendingCount,
      resolved: summary.resolvedCount,
      total: summary.totalCount
    };
  },
  
  /**
   * Check if there are any critical alerts
   */
  async hasCriticalAlerts(): Promise<boolean> {
    const alertGroups = await this.getAlertGroups('severity');
    const criticalGroup = alertGroups.find(group => group.name === 'critical');
    
    return criticalGroup ? criticalGroup.count > 0 : false;
  }
};

export default alertsService;