package consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

type recordingDLQ struct {
	msg     kafka.Message
	handle  error
	publish error
	called  bool
}

func (r *recordingDLQ) PublishDLQ(_ context.Context, msg kafka.Message, handleErr error) error {
	r.called = true
	r.msg = msg
	r.handle = handleErr
	return r.publish
}

func TestMarshalDLQBody_EncodesMetadata(t *testing.T) {
	t.Parallel()

	msg := kafka.Message{
		Topic:     "domain.events.public.accounts",
		Partition: 1,
		Offset:    42,
		Value:     []byte(`{"op":"c"}`),
	}
	handleErr := errors.New("parse failed")
	at := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)

	body, err := marshalDLQBody(msg, handleErr, at)
	if err != nil {
		t.Fatalf("marshalDLQBody: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["originalTopic"] != msg.Topic {
		t.Fatalf("originalTopic = %v", got["originalTopic"])
	}
	if got["partition"] != float64(msg.Partition) {
		t.Fatalf("partition = %v", got["partition"])
	}
	if got["offset"] != float64(msg.Offset) {
		t.Fatalf("offset = %v", got["offset"])
	}
	if got["error"] != handleErr.Error() {
		t.Fatalf("error = %v", got["error"])
	}
	wantPayload := base64.StdEncoding.EncodeToString(msg.Value)
	if got["payloadBase64"] != wantPayload {
		t.Fatalf("payloadBase64 = %v", got["payloadBase64"])
	}
	if got["timestamp"] != at.Format(time.RFC3339Nano) {
		t.Fatalf("timestamp = %v", got["timestamp"])
	}
}

func TestRecordingDLQ_CapturesHandleError(t *testing.T) {
	t.Parallel()

	rec := &recordingDLQ{}
	msg := kafka.Message{Topic: "t", Value: []byte("x")}
	want := errors.New("boom")

	if err := rec.PublishDLQ(context.Background(), msg, want); err != nil {
		t.Fatalf("PublishDLQ: %v", err)
	}
	if !rec.called || rec.handle != want || rec.msg.Topic != "t" {
		t.Fatalf("recording = %+v", rec)
	}
}
