package bsonpb

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
	"math"
)

// AsTime converts x to a time.Time.
func AsTime(x *timestamppb.Timestamp) time.Time {
	return time.Unix(int64(x.GetSeconds()), int64(x.GetNanos())).UTC()
}

// NewTimestamp constructs a new Timestamp from the provided time.Time.
func NewTimestamp(t time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{Seconds: int64(t.Unix()), Nanos: int32(t.Nanosecond())}
}

// AsDuration converts x to a time.Duration,
// returning the closest duration value in the event of overflow.
func AsDuration(x *durationpb.Duration) time.Duration {
	secs := x.GetSeconds()
	nanos := x.GetNanos()
	d := time.Duration(secs) * time.Second
	overflow := d/time.Second != time.Duration(secs)
	d += time.Duration(nanos) * time.Nanosecond
	overflow = overflow || (secs < 0 && nanos < 0 && d > 0)
	overflow = overflow || (secs > 0 && nanos > 0 && d < 0)
	if overflow {
		switch {
		case secs < 0:
			return time.Duration(math.MinInt64)
		case secs > 0:
			return time.Duration(math.MaxInt64)
		}
	}
	return d
}

// NewDuration constructs a new Duration from the provided time.Duration.
func NewDuration(d time.Duration) *durationpb.Duration {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &durationpb.Duration{Seconds: int64(secs), Nanos: int32(nanos)}
}