package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type thandLogger struct {
	// Ring buffer for storing events
	sessionUID  uuid.UUID
	eventBuffer []*models.LogEntry
	maxSize     int
	currentPos  int
	isFull      bool
	mu          sync.RWMutex

	// OpenTelemetry remote logging
	openTelemetryProvider *sdklog.LoggerProvider
	openTelemetryLogger   log.Logger
	remoteEnabled         bool
	logCtx                context.Context
	logCtxCancel          context.CancelFunc
}

func NewThandLogger() *thandLogger {
	return &thandLogger{
		sessionUID:  uuid.New(),
		eventBuffer: make([]*models.LogEntry, 1000),
		maxSize:     1000,
		currentPos:  0,
		isFull:      false,
	}
}

// EnableRemoteLogging sets up OpenTelemetry log export after authentication
func (t *thandLogger) EnableRemoteLogging(ctx context.Context, endpoint model.Endpoint, identifier uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	loggingHeaders := map[string]string{}

	// Extract endpoint URL and auth token
	loggingEndpoint := endpoint.EndpointConfig.URI.String()

	// Get auth token from endpoint configuration
	if endpoint.EndpointConfig.Authentication != nil && endpoint.EndpointConfig.Authentication.AuthenticationPolicy != nil {
		auth := endpoint.EndpointConfig.Authentication.AuthenticationPolicy

		switch {
		case auth.Basic != nil:
			credentials := base64.StdEncoding.EncodeToString([]byte(auth.Basic.Username + ":" + auth.Basic.Password))
			loggingHeaders["Authorization"] = "Basic " + credentials
		case auth.Bearer != nil:
			loggingHeaders["Authorization"] = "Bearer " + auth.Bearer.Token
		case auth.Digest != nil:
			// Digest auth not directly supported in headers; skipping
			fallthrough
		default:
			return fmt.Errorf("unsupported authentication type for OpenTelemetry endpoint")
		}
	}

	// Create OTLP HTTP exporter with authentication
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(loggingEndpoint),
		otlploghttp.WithHeaders(loggingHeaders),
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service info
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("thand-agent"),
			semconv.ServiceInstanceID(identifier.String()),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create log provider with batching
	t.openTelemetryProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exporter,
				sdklog.WithExportInterval(30*time.Second),
				sdklog.WithExportMaxBatchSize(100),
			),
		),
		sdklog.WithResource(res),
	)

	t.openTelemetryLogger = t.openTelemetryProvider.Logger("thand-agent")
	t.remoteEnabled = true

	logrus.Info("Remote logging enabled via OpenTelemetry")

	return nil
}

// DisableRemoteLogging stops the OpenTelemetry exporter
func (t *thandLogger) DisableRemoteLogging() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.remoteEnabled = false
}

// IsRemoteLoggingEnabled returns whether remote logging is active
func (t *thandLogger) IsRemoteLoggingEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.remoteEnabled
}

// Shutdown gracefully shuts down the OpenTelemetry exporter
func (t *thandLogger) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.openTelemetryProvider != nil {
		t.remoteEnabled = false
		return t.openTelemetryProvider.Shutdown(ctx)
	}
	return nil
}

// sendToOpenTelemetry sends a log entry to the OpenTelemetry exporter
func (t *thandLogger) sendToOpenTelemetry(entry *logrus.Entry) {
	if t.openTelemetryLogger == nil {
		return
	}

	var record log.Record
	record.SetTimestamp(entry.Time)
	record.SetBody(log.StringValue(entry.Message))
	record.SetSeverity(logrusToOTelSeverity(entry.Level))

	// Add logrus fields as attributes
	for key, value := range entry.Data {
		record.AddAttributes(log.String(key, fmt.Sprintf("%v", value)))
	}

	t.openTelemetryLogger.Emit(context.Background(), record)
}

// logrusToOTelSeverity converts logrus levels to OpenTelemetry severity
func logrusToOTelSeverity(level logrus.Level) log.Severity {
	switch level {
	case logrus.PanicLevel:
		return log.SeverityFatal4
	case logrus.FatalLevel:
		return log.SeverityFatal
	case logrus.ErrorLevel:
		return log.SeverityError
	case logrus.WarnLevel:
		return log.SeverityWarn
	case logrus.InfoLevel:
		return log.SeverityInfo
	case logrus.DebugLevel:
		return log.SeverityDebug
	case logrus.TraceLevel:
		return log.SeverityTrace
	default:
		return log.SeverityInfo
	}
}

func (t *thandLogger) Fire(entry *logrus.Entry) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Add to ring buffer for local queries
	t.eventBuffer[t.currentPos] = models.NewLogEntry(entry)
	t.currentPos = (t.currentPos + 1) % t.maxSize

	if t.currentPos == 0 {
		t.isFull = true
	}

	// Send to OpenTelemetry if remote logging is enabled
	if t.remoteEnabled {
		t.sendToOpenTelemetry(entry)
	}

	return nil
}

func (t *thandLogger) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		// logrus.DebugLevel,
		// logrus.TraceLevel,
	}
}

func (t *thandLogger) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.eventBuffer = make([]*models.LogEntry, t.maxSize)
	t.currentPos = 0
	t.isFull = false
}

func (t *thandLogger) GetEvents() []*models.LogEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.isFull {
		// Return only filled portion
		result := make([]*models.LogEntry, t.currentPos)
		copy(result, t.eventBuffer[:t.currentPos])
		return result
	}

	// Return in chronological order (oldest first)
	result := make([]*models.LogEntry, t.maxSize)
	copy(result, t.eventBuffer[t.currentPos:])
	copy(result[t.maxSize-t.currentPos:], t.eventBuffer[:t.currentPos])
	return result
}

func (t *thandLogger) GetRecentEvents(count int) []*models.LogEntry {
	events := t.GetEvents()
	if len(events) <= count {
		return events
	}
	return events[len(events)-count:]
}

// LogFilter contains the filtering criteria for log events
type LogFilter struct {
	// Filter by log levels (if empty, all levels are included)
	Levels []logrus.Level `json:"levels,omitempty"`
	// Filter events after this time (if nil, no time filter from start)
	Since *time.Time `json:"since,omitempty"`
	// Filter events before this time (if nil, no time filter to end)
	Until *time.Time `json:"until,omitempty"`
	// Maximum number of events to return (if 0, no limit)
	Limit int `json:"limit,omitempty"`
}

// GetEventsWithFilter returns events that match the specified filter criteria
func (t *thandLogger) GetEventsWithFilter(filter LogFilter) []*models.LogEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	allEvents := t.getEventsInternal()
	var filtered []*models.LogEntry

	// Create a map for quick level lookup if levels are specified
	levelMap := make(map[logrus.Level]bool)
	if len(filter.Levels) > 0 {
		for _, level := range filter.Levels {
			levelMap[level] = true
		}
	}

	for _, entry := range allEvents {
		// Filter by log level
		if len(filter.Levels) > 0 && !levelMap[entry.Level] {
			continue
		}

		// Filter by time range
		if filter.Since != nil && entry.Time.Before(*filter.Since) {
			continue
		}
		if filter.Until != nil && entry.Time.After(*filter.Until) {
			continue
		}

		filtered = append(filtered, entry)

		// Apply limit if specified
		if filter.Limit > 0 && len(filtered) >= filter.Limit {
			break
		}
	}

	return filtered
}

// getEventsInternal returns events without additional locking (assumes caller has lock)
func (t *thandLogger) getEventsInternal() []*models.LogEntry {
	if !t.isFull {
		// Return only filled portion
		result := make([]*models.LogEntry, t.currentPos)
		copy(result, t.eventBuffer[:t.currentPos])
		return result
	}

	// Return in chronological order (oldest first)
	result := make([]*models.LogEntry, t.maxSize)
	copy(result, t.eventBuffer[t.currentPos:])
	copy(result[t.maxSize-t.currentPos:], t.eventBuffer[:t.currentPos])
	return result
}
