package logger

import (
	"time"

	"github.com/influxdata/influxdb/pkg/snowflake"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TraceIDKey          = "trace_id"
	OperationNameKey    = "op_name"
	OperationEventKey   = "op_event"
	OperationElapsedKey = "op_elapsed"
	DBInstanceKey       = "db_instance"
	DBRetentionKey      = "db_rp"
	DBShardGroupKey     = "db_shard_group"
	DBShardIDKey        = "db_shard_id"

	eventStart = "evt.start"
	eventEnd   = "evt.end"
)

var (
	gen = snowflake.New(0)
)

func NextTraceID() string {
	return gen.NextString()
}

// TraceID returns a field for tracking the trace identifier.
func TraceID(id string) zapcore.Field {
	return zap.String(TraceIDKey, id)
}

// OperationName returns a field for tracking the name of an operation.
func OperationName(name string) zapcore.Field {
	return zap.String(OperationNameKey, name)
}

// OperationElapsed returns a field for tracking the duration of an operation.
func OperationElapsed(d time.Duration) zapcore.Field {
	return zap.Duration(OperationElapsedKey, d)
}

// OperationEventStart returns a field for tracking the start of an operation.
func OperationEventStart() zapcore.Field {
	return zap.String(OperationEventKey, eventStart)
}

// OperationEventFinish returns a field for tracking the end of an operation.
func OperationEventEnd() zapcore.Field {
	return zap.String(OperationEventKey, eventEnd)
}

// Database returns a field for tracking the name of a database.
func Database(name string) zapcore.Field {
	return zap.String(DBInstanceKey, name)
}

// Database returns a field for tracking the name of a database.
func RetentionPolicy(name string) zapcore.Field {
	return zap.String(DBRetentionKey, name)
}

// ShardGroup returns a field for tracking the shard group identifier.
func ShardGroup(id uint64) zapcore.Field {
	return zap.Uint64(DBShardGroupKey, id)
}

// Shard returns a field for tracking the shard identifier.
func Shard(id uint64) zapcore.Field {
	return zap.Uint64(DBShardIDKey, id)
}
