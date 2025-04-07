package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// AlertsService handles alert-related operations
type AlertsService struct {
	client  *prometheus.Client
	logger  logger.Logger
}

// NewAlertsService creates a new alerts service
func NewAlertsService(client *prometheus.Client, logger logger.Logger) *AlertsService {
	return &AlertsService{
		client: client,
		logger: logger,
	}
}

// GetAlerts retrieves all current alerts from Prometheus
func (s *AlertsService) GetAlerts(ctx context.Context) ([]models.Alert, error) {
	s.logger.Info("Retrieving current alerts")
	
	promAlerts, err := s.client.GetAlerts(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get alerts: %v", err)
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	
	alerts := make([]models.Alert, 0, len(promAlerts))
	for _, a := range promAlerts {
		alert := models.Alert{
			Name:        a.Name,
			State:       string(a.State),
			Labels:      a.Labels,
			Annotations: a.Annotations,
			ActiveAt:    a.ActiveAt,
			Value:       a.Value,
		}
		
		// Extract severity from labels if available
		if severity, ok := a.Labels["severity"]; ok {
			alert.Severity = severity
		} else {
			alert.Severity = "unknown"
		}
		
		// Extract summary from annotations if available
		if summary, ok := a.Annotations["summary"]; ok {
			alert.Summary = summary
		} else if desc, ok := a.Annotations["description"]; ok {
			alert.Summary = desc
		}
		
		alerts = append(alerts, alert)
	}
	
	// Sort alerts by state and then by name
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].State != alerts[j].State {
			// Firing alerts first, then pending, then resolved
			if alerts[i].State == "firing" {
				return true
			}
			if alerts[j].State == "firing" {
				return false
			}
			if alerts[i].State == "pending" {
				return true
			}
			return false
		}
		return alerts[i].Name < alerts[j].Name
	})
	
	return alerts, nil
}

// GetAlertGroups retrieves alerts grouped by a specified label
func (s *AlertsService) GetAlertGroups(ctx context.Context, groupBy string) ([]models.AlertGroup, error) {
	if groupBy == "" {
		groupBy = "severity" // Default grouping
	}
	
	s.logger.Infof("Retrieving alerts grouped by %s", groupBy)
	
	// Get all alerts first
	alerts, err := s.GetAlerts(ctx)
	if err != nil {
		return nil, err
	}
	
	// Group alerts by the specified label
	groups := make(map[string][]models.Alert)
	for _, alert := range alerts {
		var groupKey string
		if groupBy == "severity" {
			groupKey = alert.Severity
		} else if val, ok := alert.Labels[groupBy]; ok {
			groupKey = val
		} else {
			groupKey = "unknown"
		}
		
		if _, exists := groups[groupKey]; !exists {
			groups[groupKey] = make([]models.Alert, 0)
		}
		groups[groupKey] = append(groups[groupKey], alert)
	}
	
	// Convert map to slice for response
	result := make([]models.AlertGroup, 0, len(groups))
	for groupKey, groupAlerts := range groups {
		result = append(result, models.AlertGroup{
			Name:   groupKey,
			Alerts: groupAlerts,
			Count:  len(groupAlerts),
		})
	}
	
	// Sort groups by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	
	return result, nil
}

// GetAlertSummary provides a summary of current alert status
func (s *AlertsService) GetAlertSummary(ctx context.Context) (*models.AlertSummary, error) {
	s.logger.Info("Generating alert summary")
	
	// Get all alerts first
	alerts, err := s.GetAlerts(ctx)
	if err != nil {
		return nil, err
	}
	
	// Initialize counters
	var firingCount, pendingCount, resolvedCount int
	severityCounts := make(map[string]int)
	
	// Count alerts by state and severity
	for _, alert := range alerts {
		switch alert.State {
		case "firing":
			firingCount++
		case "pending":
			pendingCount++
		default:
			resolvedCount++
		}
		
		severityCounts[alert.Severity]++
	}
	
	// Get the most recent alert
	var mostRecentAlert *models.Alert
	var mostRecentTime time.Time
	
	for i, alert := range alerts {
		if alert.ActiveAt.After(mostRecentTime) {
			mostRecentTime = alert.ActiveAt
			mostRecentAlert = &alerts[i]
		}
	}
	
	// Calculate time since most recent alert
	var timeSinceLastAlert string
	if mostRecentAlert != nil {
		duration := time.Since(mostRecentTime)
		if duration < time.Minute {
			timeSinceLastAlert = "less than a minute ago"
		} else if duration < time.Hour {
			minutes := int(duration.Minutes())
			timeSinceLastAlert = fmt.Sprintf("%d minute(s) ago", minutes)
		} else if duration < 24*time.Hour {
			hours := int(duration.Hours())
			timeSinceLastAlert = fmt.Sprintf("%d hour(s) ago", hours)
		} else {
			days := int(duration.Hours() / 24)
			timeSinceLastAlert = fmt.Sprintf("%d day(s) ago", days)
		}
	}
	
	// Build severity breakdown
	severityBreakdown := make([]models.SeverityCount, 0, len(severityCounts))
	for severity, count := range severityCounts {
		severityBreakdown = append(severityBreakdown, models.SeverityCount{
			Severity: severity,
			Count:    count,
		})
	}
	
	// Sort by severity (critical, high, medium, low, unknown)
	sort.Slice(severityBreakdown, func(i, j int) bool {
		return getSeverityRank(severityBreakdown[i].Severity) < 
			getSeverityRank(severityBreakdown[j].Severity)
	})
	
	// Create summary
	summary := &models.AlertSummary{
		FiringCount:       firingCount,
		PendingCount:      pendingCount,
		ResolvedCount:     resolvedCount,
		TotalCount:        len(alerts),
		SeverityBreakdown: severityBreakdown,
		LastUpdated:       time.Now(),
	}
	
	if mostRecentAlert != nil {
		summary.MostRecentAlert = *mostRecentAlert
		summary.TimeSinceLastAlert = timeSinceLastAlert
	}
	
	return summary, nil
}

// Helper function to get severity rank for sorting
func getSeverityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "high":
		return 1
	case "warning":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	case "info":
		return 5
	default:
		return 6
	}
}