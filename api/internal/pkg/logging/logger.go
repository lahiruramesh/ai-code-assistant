package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Logger wraps logrus with OpenTelemetry support
type Logger struct {
	*logrus.Logger
}

// Fields type for structured logging
type Fields map[string]interface{}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	log.SetLevel(logrus.InfoLevel)

	return &Logger{Logger: log}
}

// GetContextLogger returns a logger with context information
func (l *Logger) GetContextLogger(ctx context.Context) *logrus.Entry {
	entry := l.WithContext(ctx)

	// Add OpenTelemetry trace information if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		entry = entry.WithFields(logrus.Fields{
			"trace_id": spanCtx.TraceID().String(),
			"span_id":  spanCtx.SpanID().String(),
		})
	}

	return entry
}

// ToolCallEvent represents a tool call event for logging
type ToolCallEvent struct {
	EventType    string                 `json:"event_type"`
	ToolName     string                 `json:"tool_name"`
	ExecutionID  string                 `json:"execution_id"`
	StartTime    time.Time              `json:"start_time,omitempty"`
	EndTime      time.Time              `json:"end_time,omitempty"`
	Duration     time.Duration          `json:"duration,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Arguments    map[string]interface{} `json:"arguments,omitempty"`
	ResultSize   int                    `json:"result_size,omitempty"`
	AgentType    string                 `json:"agent_type,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
}

// LogToolCallStart logs the start of a tool call
func (l *Logger) LogToolCallStart(ctx context.Context, event ToolCallEvent) {
	l.GetContextLogger(ctx).WithFields(logrus.Fields{
		"event_type":    "tool_call_start",
		"tool_name":     event.ToolName,
		"execution_id":  event.ExecutionID,
		"agent_type":    event.AgentType,
		"session_id":    event.SessionID,
		"start_time":    event.StartTime,
		"argument_keys": getArgumentKeys(event.Arguments),
	}).Info("Tool call started")
}

// LogToolCallEnd logs the completion of a tool call
func (l *Logger) LogToolCallEnd(ctx context.Context, event ToolCallEvent) {
	fields := logrus.Fields{
		"event_type":   "tool_call_end",
		"tool_name":    event.ToolName,
		"execution_id": event.ExecutionID,
		"agent_type":   event.AgentType,
		"session_id":   event.SessionID,
		"end_time":     event.EndTime,
		"duration_ms":  event.Duration.Milliseconds(),
		"success":      event.Success,
		"result_size":  event.ResultSize,
	}

	if !event.Success && event.ErrorMessage != "" {
		fields["error_message"] = event.ErrorMessage
		l.GetContextLogger(ctx).WithFields(fields).Error("Tool call failed")
	} else {
		l.GetContextLogger(ctx).WithFields(fields).Info("Tool call completed")
	}
}

// AgentEvent represents an agent event for logging
type AgentEvent struct {
	EventType    string                 `json:"event_type"`
	AgentType    string                 `json:"agent_type"`
	SessionID    string                 `json:"session_id"`
	MessageID    string                 `json:"message_id,omitempty"`
	TaskType     string                 `json:"task_type,omitempty"`
	Success      bool                   `json:"success"`
	Duration     time.Duration          `json:"duration,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LogAgentEvent logs agent-related events
func (l *Logger) LogAgentEvent(ctx context.Context, event AgentEvent) {
	fields := logrus.Fields{
		"event_type": event.EventType,
		"agent_type": event.AgentType,
		"session_id": event.SessionID,
		"success":    event.Success,
	}

	if event.MessageID != "" {
		fields["message_id"] = event.MessageID
	}
	if event.TaskType != "" {
		fields["task_type"] = event.TaskType
	}
	if event.Duration > 0 {
		fields["duration_ms"] = event.Duration.Milliseconds()
	}
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			fields[fmt.Sprintf("meta_%s", k)] = v
		}
	}

	if !event.Success && event.ErrorMessage != "" {
		fields["error_message"] = event.ErrorMessage
		l.GetContextLogger(ctx).WithFields(fields).Error("Agent event failed")
	} else {
		l.GetContextLogger(ctx).WithFields(fields).Info("Agent event")
	}
}

// HTTPEvent represents an HTTP request/response event
type HTTPEvent struct {
	EventType    string        `json:"event_type"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code,omitempty"`
	Duration     time.Duration `json:"duration,omitempty"`
	RequestSize  int64         `json:"request_size,omitempty"`
	ResponseSize int64         `json:"response_size,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	SessionID    string        `json:"session_id,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// LogHTTPEvent logs HTTP request/response events
func (l *Logger) LogHTTPEvent(ctx context.Context, event HTTPEvent) {
	fields := logrus.Fields{
		"event_type": event.EventType,
		"method":     event.Method,
		"path":       event.Path,
	}

	if event.StatusCode > 0 {
		fields["status_code"] = event.StatusCode
	}
	if event.Duration > 0 {
		fields["duration_ms"] = event.Duration.Milliseconds()
	}
	if event.RequestSize > 0 {
		fields["request_size"] = event.RequestSize
	}
	if event.ResponseSize > 0 {
		fields["response_size"] = event.ResponseSize
	}
	if event.UserAgent != "" {
		fields["user_agent"] = event.UserAgent
	}
	if event.SessionID != "" {
		fields["session_id"] = event.SessionID
	}

	if event.StatusCode >= 400 && event.ErrorMessage != "" {
		fields["error_message"] = event.ErrorMessage
		l.GetContextLogger(ctx).WithFields(fields).Error("HTTP request failed")
	} else {
		l.GetContextLogger(ctx).WithFields(fields).Info("HTTP request processed")
	}
}

// AddTraceAttributes adds custom attributes to the current span for enhanced tracing
func AddTraceAttributes(ctx context.Context, keyValues ...string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Log the trace attributes for debugging
	for i := 0; i < len(keyValues)-1; i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		GlobalLogger.WithFields(logrus.Fields{
			"trace_attribute_key":   key,
			"trace_attribute_value": value,
			"span_id":               span.SpanContext().SpanID().String(),
		}).Debug("Added trace attribute")
	}
}

// getArgumentKeys returns only the keys of arguments (no sensitive data)
func getArgumentKeys(args map[string]interface{}) []string {
	if args == nil {
		return nil
	}

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	return keys
}

// Global logger instance
var GlobalLogger = NewLogger()
