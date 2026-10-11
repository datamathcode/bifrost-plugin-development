package main

import (
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

const (
	// Context keys as constants to avoid collisions
	transportPreHookKey  schemas.BifrostContextKey = "bifrost-plugin-development-transport-pre-hook"
	transportPostHookKey schemas.BifrostContextKey = "bifrost-plugin-development-transport-post-hook"
	preHookKey           schemas.BifrostContextKey = "bifrost-plugin-development-pre-hook"
)

// Init is called when the plugin is loaded
// config contains the plugin configuration from config.json
func Init(config any) error {
	fmt.Println("Init called")
	// Initialize your plugin here (database connections, API clients, etc.)
	return nil
}

// GetName returns the plugin's unique identifier
// This must match the "name" field in config.json plugins array
func GetName() string {
	return "bifrost-plugin-development"
}

// HTTPTransportPreHook intercepts requests BEFORE they enter Bifrost core
// Modify req in-place. Return (*HTTPResponse, nil) to short-circuit.
// Only called when using HTTP transport (bifrost-http)
func HTTPTransportPreHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest) (*schemas.HTTPResponse, error) {
	fmt.Println("HTTPTransportPreHook called")
	// Modify request in-place
	req.Headers["x-bifrost-plugin-development"] = "transport-pre-hook-value"
	// Store value in context for PreLLMHook/PostLLMHook
	ctx.SetValue(transportPreHookKey, "transport-pre-hook-value")
	// Use ctx.Log for structured, scoped logging (recommended)
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportPreHook called")
	// Return nil to continue processing, or return &schemas.HTTPResponse{} to short-circuit
	return nil, nil
}

// HTTPTransportPostHook intercepts responses AFTER they exit Bifrost core
// Modify response in-place. Called in REVERSE order of pre-hooks.
// Only called when using HTTP transport (bifrost-http)
// NOT called for streaming responses - use HTTPTransportStreamChunkHook instead
func HTTPTransportPostHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, resp *schemas.HTTPResponse) error {
	fmt.Println("HTTPTransportPostHook called")
	// Modify response in-place
	resp.Headers["x-bifrost-plugin-development"] = "transport-post-hook-value"
	// Store value in context
	ctx.SetValue(transportPostHookKey, "transport-post-hook-value")
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportPostHook called")
	return nil
}

// HTTPTransportStreamChunkHook intercepts streaming response chunks
// Modify chunk in-place or return a modified chunk
// Only called when using HTTP transport (bifrost-http) with streaming
func HTTPTransportStreamChunkHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, chunk *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error) {
	fmt.Println("HTTPTransportStreamChunkHook called")
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportStreamChunkHook called")
	// Modify streaming content in-place
	if chunk.BifrostChatResponse != nil &&
		chunk.BifrostChatResponse.Choices != nil &&
		len(chunk.BifrostChatResponse.Choices) > 0 &&
		chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice != nil &&
		chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta != nil &&
		chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content != nil {
		*chunk.BifrostChatResponse.Choices[0].ChatStreamResponseChoice.Delta.Content += " - modified by bifrost-plugin-development"
	}
	return chunk, nil
}

// PreRequestHook is called ONCE per request, before any provider call
// Routing decisions (provider/model/fallbacks) happen here and propagate to every attempt
func PreRequestHook(_ *schemas.BifrostContext, _ *schemas.BifrostRequest) error {
	return nil
}

// PreLLMHook intercepts requests before they reach providers
// Runs per provider attempt, in registration order
// Can short-circuit by returning (*BifrostRequest, *LLMPluginShortCircuit, error)
func PreLLMHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error) {
	value1 := ctx.Value(transportPreHookKey)
	fmt.Println("value1:", value1)
	ctx.SetValue(preHookKey, "pre-hook-value")
	ctx.Log(schemas.LogLevelInfo, "PreLLMHook called")
	fmt.Print("PreLLMHook called\n")
	return req, nil, nil
}

// PostLLMHook processes responses after provider calls
// Runs per provider attempt, in REVERSE order of PreLLMHook
func PostLLMHook(ctx *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr error) (*schemas.BifrostResponse, error) {
	value1 := ctx.Value(transportPreHookKey)
	value2 := ctx.Value(transportPostHookKey)
	value3 := ctx.Value(preHookKey)
	fmt.Println("value1:", value1, "value2:", value2, "value3:", value3)
	ctx.Log(schemas.LogLevelInfo, "PostLLMHook called")
	return resp, nil
}

// Cleanup is called when Bifrost shuts down
func Cleanup() error {
	fmt.Println("Cleanup called")
	return nil
}