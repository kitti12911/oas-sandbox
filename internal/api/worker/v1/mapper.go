package workerv1

import (
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"google.golang.org/protobuf/types/known/structpb"

	workerpb "oas-sandbox/gen/grpc/worker/v1"
)

func submitJobToProto(input *SubmitJobInput) (*workerpb.SubmitJobRequest, error) {
	payload, err := payloadToProto(input.Body.Payload)
	if err != nil {
		return nil, huma.Error400BadRequest("payload must be a JSON object")
	}

	return &workerpb.SubmitJobRequest{
		Job: &workerpb.WorkerJob{
			Id:      input.Body.ID,
			Type:    input.Body.Type,
			Payload: payload,
		},
	}, nil
}

func payloadToProto(payload map[string]any) (*structpb.Struct, error) {
	if payload == nil {
		return nil, nil
	}
	out, err := structpb.NewStruct(payload)
	if err != nil {
		return nil, fmt.Errorf("new protobuf struct: %w", err)
	}
	return out, nil
}

func submitJobFromProto(resp *workerpb.SubmitJobResponse) *SubmitJobOutput {
	if resp == nil {
		return &SubmitJobOutput{}
	}

	return &SubmitJobOutput{
		Body: SubmitJobResult{
			ID: resp.GetId(),
		},
	}
}
