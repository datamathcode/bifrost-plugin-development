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

func TestHTTPTransportPreHookSetsHeaderAndContext(t *testing.T) {
	ctx := newTestContext()
	req := &schemas.HTTPRequest{Headers: map[string]string{}}

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook() error = %v, want nil", err)
	}
	if resp != nil {
		t.Fatalf("HTTPTransportPreHook() resp = %v, want nil", resp)
	}
	if got := req.Headers[pluginHeader]; got != transportPreHookValue {
		t.Errorf("request header %q = %q, want %q", pluginHeader, got, transportPreHookValue)
	}
	if got := ctx.Value(transportPreHookKey); got != transportPreHookValue {
		t.Errorf("context value %q = %v, want %q", transportPreHookKey, got, transportPreHookValue)
	}
}

func TestHTTPTransportPostHookSetsHeaderAndContext(t *testing.T) {
	ctx := newTestContext()
	req := &schemas.HTTPRequest{Headers: map[string]string{}}
	resp := &schemas.HTTPResponse{Headers: map[string]string{}}

	if err := HTTPTransportPostHook(ctx, req, resp); err != nil {
		t.Fatalf("HTTPTransportPostHook() error = %v, want nil", err)
	}
	if got := resp.Headers[pluginHeader]; got != transportPostHookValue {
		t.Errorf("response header %q = %q, want %q", pluginHeader, got, transportPostHookValue)
	}
	if got := ctx.Value(transportPostHookKey); got != transportPostHookValue {
		t.Errorf("context value %q = %v, want %q", transportPostHookKey, got, transportPostHookValue)
	}
}

func TestPreLLMHookSetsContextAndReturnsRequestUnchanged(t *testing.T) {
	ctx := newTestContext()
	req := &schemas.BifrostRequest{}

	got, shortCircuit, err := PreLLMHook(ctx, req)
	if err != nil {
		t.Fatalf("PreLLMHook() error = %v, want nil", err)
	}
	if shortCircuit != nil {
		t.Fatalf("PreLLMHook() shortCircuit = %v, want nil", shortCircuit)
	}
	if got != req {
		t.Errorf("PreLLMHook() returned a different request; want the same pointer")
	}
	if got := ctx.Value(preHookKey); got != preHookStoredValue {
		t.Errorf("context value %q = %v, want %q", preHookKey, got, preHookStoredValue)
	}
}

func TestContextValueWrittenByEarlierHookIsReadable(t *testing.T) {
	ctx := newTestContext()
	req := &schemas.HTTPRequest{Headers: map[string]string{}}

	if _, err := HTTPTransportPreHook(ctx, req); err != nil {
		t.Fatalf("HTTPTransportPreHook() error = %v", err)
	}
	if got := ctx.Value(transportPreHookKey); got != transportPreHookValue {
		t.Fatalf("value written by HTTPTransportPreHook not readable later: got %v", got)
	}

	if _, _, err := PreLLMHook(ctx, &schemas.BifrostRequest{}); err != nil {
		t.Fatalf("PreLLMHook() error = %v", err)
	}
	if got := ctx.Value(preHookKey); got != preHookStoredValue {
		t.Fatalf("value written by PreLLMHook not readable later: got %v", got)
	}
}

func TestPostLLMHookReturnsInputsUnchangedAndThreadsError(t *testing.T) {
	ctx := newTestContext()
	resp := &schemas.BifrostResponse{}
	bifrostErr := &schemas.BifrostError{IsBifrostError: true}

	gotResp, gotErr, err := PostLLMHook(ctx, resp, bifrostErr)
	if err != nil {
		t.Fatalf("PostLLMHook() error = %v, want nil", err)
	}
	if gotResp != resp {
		t.Errorf("PostLLMHook() returned a different response; want the same pointer")
	}
	if gotErr != bifrostErr {
		t.Errorf("PostLLMHook() dropped or replaced the provider error; want the same pointer")
	}
}

func TestPostLLMHookToleratesNilResponseAndError(t *testing.T) {
	gotResp, gotErr, err := PostLLMHook(newTestContext(), nil, nil)
	if err != nil {
		t.Fatalf("PostLLMHook() error = %v, want nil", err)
	}
	if gotResp != nil {
		t.Errorf("PostLLMHook() resp = %v, want nil", gotResp)
	}
	if gotErr != nil {
		t.Errorf("PostLLMHook() err = %v, want nil", gotErr)
	}
}

func TestHTTPTransportStreamChunkHookAppendsSuffixOnce(t *testing.T) {
	content := "hello"
	chunk := &schemas.BifrostStreamChunk{
		BifrostChatResponse: &schemas.BifrostChatResponse{
			Choices: []schemas.BifrostResponseChoice{
				{
					ChatStreamResponseChoice: &schemas.ChatStreamResponseChoice{
						Delta: &schemas.ChatStreamResponseChoiceDelta{Content: &content},
					},
				},
			},
		},
	}

	got, err := HTTPTransportStreamChunkHook(newTestContext(), &schemas.HTTPRequest{}, chunk)
	if err != nil {
		t.Fatalf("HTTPTransportStreamChunkHook() error = %v, want nil", err)
	}
	if got != chunk {
		t.Errorf("HTTPTransportStreamChunkHook() returned a different chunk; want the same pointer")
	}
	want := "hello" + streamChunkSuffix
	if content != want {
		t.Errorf("chunk content = %q, want %q (suffix appended exactly once)", content, want)
	}
}

func TestHTTPTransportStreamChunkHookToleratesMissingDeltaChain(t *testing.T) {
	cases := map[string]*schemas.BifrostStreamChunk{
		"nil chunk":        nil,
		"no chat response": &schemas.BifrostStreamChunk{},
		"no choices": &schemas.BifrostStreamChunk{
			BifrostChatResponse: &schemas.BifrostChatResponse{},
		},
		"no stream choice": &schemas.BifrostStreamChunk{
			BifrostChatResponse: &schemas.BifrostChatResponse{
				Choices: []schemas.BifrostResponseChoice{{}},
			},
		},
		"nil delta": &schemas.BifrostStreamChunk{
			BifrostChatResponse: &schemas.BifrostChatResponse{
				Choices: []schemas.BifrostResponseChoice{
					{ChatStreamResponseChoice: &schemas.ChatStreamResponseChoice{}},
				},
			},
		},
		"nil content": &schemas.BifrostStreamChunk{
			BifrostChatResponse: &schemas.BifrostChatResponse{
				Choices: []schemas.BifrostResponseChoice{
					{ChatStreamResponseChoice: &schemas.ChatStreamResponseChoice{
						Delta: &schemas.ChatStreamResponseChoiceDelta{},
					}},
				},
			},
		},
	}

	for name, chunk := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := HTTPTransportStreamChunkHook(newTestContext(), &schemas.HTTPRequest{}, chunk)
			if err != nil {
				t.Fatalf("HTTPTransportStreamChunkHook() error = %v, want nil", err)
			}
			if got != chunk {
				t.Errorf("HTTPTransportStreamChunkHook() returned a different chunk; want the same pointer")
			}
		})
	}
}

func TestLifecycleHooksDoNotError(t *testing.T) {
	if err := Init(nil); err != nil {
		t.Errorf("Init() error = %v, want nil", err)
	}
	if err := Cleanup(); err != nil {
		t.Errorf("Cleanup() error = %v, want nil", err)
	}
	if err := PreRequestHook(newTestContext(), &schemas.BifrostRequest{}); err != nil {
		t.Errorf("PreRequestHook() error = %v, want nil", err)
	}
}
