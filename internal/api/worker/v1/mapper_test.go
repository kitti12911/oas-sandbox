package workerv1

import (
	"testing"

	workerpb "oas-sandbox/gen/grpc/worker/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitJobToProto(t *testing.T) {
	t.Parallel()
	got, err := submitJobToProto(&SubmitJobInput{Body: SubmitJobRequest{
		ID:      "job-1",
		Type:    "debug.print",
		Payload: map[string]any{"message": "hello"},
	}})
	require.NoError(t, err)
	assert.Equal(t, "job-1", got.GetJob().GetId())
	assert.Equal(t, "debug.print", got.GetJob().GetType())
	assert.Equal(t, "hello", got.GetJob().GetPayload().Fields["message"].GetStringValue())
}

func TestSubmitJobToProtoRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	// structpb.NewStruct rejects values that aren't basic JSON-compatible
	// types (channels, functions, etc.). The mapper converts that to a Huma 400.
	_, err := submitJobToProto(&SubmitJobInput{Body: SubmitJobRequest{
		ID:      "job-1",
		Type:    "debug.print",
		Payload: map[string]any{"bad": make(chan int)},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payload must be a JSON object")
}

func TestPayloadToProtoNil(t *testing.T) {
	t.Parallel()
	got, err := payloadToProto(nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSubmitJobFromProtoNil(t *testing.T) {
	t.Parallel()
	got := submitJobFromProto(nil)
	assert.Equal(t, "", got.Body.ID)
}

func TestSubmitJobFromProto(t *testing.T) {
	t.Parallel()
	got := submitJobFromProto(&workerpb.SubmitJobResponse{Id: "job-1"})
	assert.Equal(t, "job-1", got.Body.ID)
}
