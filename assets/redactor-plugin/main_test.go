package main

import (
	"context"
	"testing"
	"time"

	"github.com/maximhq/bifrost/core/schemas"
)

func newTestContext() *schemas.BifrostContext {
	return schemas.NewBifrostContext(context.Background(), time.Now().Add(time.Hour))
}

func TestGetName(t *testing.T) {
	if got := GetName(); got != pluginName {
		t.Fatalf("GetName() = %q, want %q", got, pluginName)
	}
}

func TestRedactText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"email", "reach me at jane.doe@example.com now", "reach me at [REDACTED:email] now"},
		{"aws key", "key AKIAIOSFODNN7EXAMPLE here", "key [REDACTED:aws-access-key] here"},
		{"provider key", "use sk-abcdefghijklmnop1234 today", "use [REDACTED:provider-api-key] today"},
		{"bearer", "Authorization: Bearer abcdefghijklmnop", "Authorization: [REDACTED:bearer-token]"},
		{"ssn", "ssn 123-45-6789 ok", "ssn [REDACTED:ssn] ok"},
		{"credit card", "card 4111 1111 1111 1111 end", "card [REDACTED:credit-card] end"},
		{"private key", "-----BEGIN RSA PRIVATE KEY-----", "[REDACTED:private-key]"},
		{"clean", "nothing to see here", "nothing to see here"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redactText(tc.in); got != tc.want {
				t.Errorf("redactText(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPreLLMHookRedactsRequest(t *testing.T) {
	req := &schemas.BifrostRequest{
		ChatRequest: &schemas.BifrostChatRequest{
			Input: []schemas.ChatMessage{
				{Content: &schemas.ChatMessageContent{ContentStr: schemas.Ptr("email me at a@b.com")}},
			},
		},
	}

	got, shortCircuit, err := PreLLMHook(newTestContext(), req)
	if err != nil {
		t.Fatalf("PreLLMHook() error = %v, want nil", err)
	}
	if shortCircuit != nil {
		t.Fatalf("PreLLMHook() shortCircuit = %v, want nil", shortCircuit)
	}
	if got != req {
		t.Fatalf("PreLLMHook() returned a different request; want the same pointer")
	}
	if want := "email me at [REDACTED:email]"; *req.ChatRequest.Input[0].Content.ContentStr != want {
		t.Errorf("request content = %q, want %q", *req.ChatRequest.Input[0].Content.ContentStr, want)
	}
}

func TestPreLLMHookToleratesEmptyRequest(t *testing.T) {
	if _, _, err := PreLLMHook(newTestContext(), &schemas.BifrostRequest{}); err != nil {
		t.Fatalf("PreLLMHook() error = %v, want nil", err)
	}
	if _, _, err := PreLLMHook(newTestContext(), nil); err != nil {
		t.Fatalf("PreLLMHook(nil) error = %v, want nil", err)
	}
}

func TestPostLLMHookRedactsResponseAndThreadsError(t *testing.T) {
	resp := &schemas.BifrostResponse{
		ChatResponse: &schemas.BifrostChatResponse{
			Choices: []schemas.BifrostResponseChoice{
				{
					ChatNonStreamResponseChoice: &schemas.ChatNonStreamResponseChoice{
						Message: &schemas.ChatMessage{
							Content: &schemas.ChatMessageContent{ContentStr: schemas.Ptr("ssn 123-45-6789")},
						},
					},
				},
			},
		},
	}
	bifrostErr := &schemas.BifrostError{IsBifrostError: true}

	gotResp, gotErr, err := PostLLMHook(newTestContext(), resp, bifrostErr)
	if err != nil {
		t.Fatalf("PostLLMHook() error = %v, want nil", err)
	}
	if gotResp != resp {
		t.Errorf("PostLLMHook() returned a different response; want the same pointer")
	}
	if gotErr != bifrostErr {
		t.Errorf("PostLLMHook() dropped or replaced the provider error; want the same pointer")
	}
	if want := "ssn [REDACTED:ssn]"; *resp.ChatResponse.Choices[0].ChatNonStreamResponseChoice.Message.Content.ContentStr != want {
		t.Errorf("response content = %q, want %q", *resp.ChatResponse.Choices[0].ChatNonStreamResponseChoice.Message.Content.ContentStr, want)
	}
}

func TestPostLLMHookToleratesNilResponseAndError(t *testing.T) {
	gotResp, gotErr, err := PostLLMHook(newTestContext(), nil, nil)
	if err != nil || gotResp != nil || gotErr != nil {
		t.Fatalf("PostLLMHook(nil, nil) = (%v, %v, %v), want (nil, nil, nil)", gotResp, gotErr, err)
	}
}

func streamChunk(content string, finish *string) *schemas.BifrostStreamChunk {
	return &schemas.BifrostStreamChunk{
		BifrostChatResponse: &schemas.BifrostChatResponse{
			Choices: []schemas.BifrostResponseChoice{
				{
					FinishReason: finish,
					ChatStreamResponseChoice: &schemas.ChatStreamResponseChoice{
						Delta: &schemas.ChatStreamResponseChoiceDelta{Content: &content},
					},
				},
			},
		},
	}
}

func TestStreamChunkRedactsContent(t *testing.T) {
	chunk := streamChunk("card 4111 1111 1111 1111", nil)

	got, err := HTTPTransportStreamChunkHook(newTestContext(), &schemas.HTTPRequest{}, chunk)
	if err != nil {
		t.Fatalf("HTTPTransportStreamChunkHook() error = %v, want nil", err)
	}
	if got != chunk {
		t.Fatalf("returned a different chunk; want the same pointer")
	}
	if want := "card [REDACTED:credit-card]"; *chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content != want {
		t.Errorf("stream content = %q, want %q", *chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content, want)
	}
}

func TestStreamChunkWatermarksFinalChunk(t *testing.T) {
	stop := "stop"
	chunk := streamChunk("bye", &stop)

	if _, err := HTTPTransportStreamChunkHook(newTestContext(), &schemas.HTTPRequest{}, chunk); err != nil {
		t.Fatalf("HTTPTransportStreamChunkHook() error = %v, want nil", err)
	}
	if want := "bye" + streamWatermark; *chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content != want {
		t.Errorf("final stream content = %q, want %q", *chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content, want)
	}
}

func TestStreamChunkToleratesMissingDeltaChain(t *testing.T) {
	cases := map[string]*schemas.BifrostStreamChunk{
		"nil chunk":        nil,
		"no chat response": &schemas.BifrostStreamChunk{},
		"no choices":       {BifrostChatResponse: &schemas.BifrostChatResponse{}},
		"no stream choice": {BifrostChatResponse: &schemas.BifrostChatResponse{Choices: []schemas.BifrostResponseChoice{{}}}},
		"nil delta": {BifrostChatResponse: &schemas.BifrostChatResponse{
			Choices: []schemas.BifrostResponseChoice{{ChatStreamResponseChoice: &schemas.ChatStreamResponseChoice{}}},
		}},
	}
	for name, chunk := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := HTTPTransportStreamChunkHook(newTestContext(), &schemas.HTTPRequest{}, chunk); err != nil {
				t.Fatalf("HTTPTransportStreamChunkHook() error = %v, want nil", err)
			}
		})
	}
}
