package main

import (
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

const (
	// pluginName is the plugin's unique identifier. It must match the "name"
	// field in config.json plugins array.
	pluginName = "bifrost-plugin-development"

	// pluginHeader is the HTTP header this demo plugin writes on requests and responses.
	pluginHeader = "x-" + pluginName

	// Demo values stashed in context by the pre-hooks and read by later hooks.
	transportPreHookValue  = "transport-pre-hook-value"
	transportPostHookValue = "transport-post-hook-value"
	preHookStoredValue     = "pre-hook-value"

	// streamChunkSuffix is appended to each streaming chat chunk's content.
	streamChunkSuffix = " - modified by " + pluginName
)

// Context keys as typed constants to avoid collisions. Hooks may only read keys
// written by a hook that runs EARLIER in the documented execution order:
// HTTPTransportPreHook -> PreRequestHook -> PreLLMHook -> provider -> PostLLMHook
// -> HTTPTransportPostHook -> HTTPTransportStreamChunkHook.
const (
	transportPreHookKey  schemas.BifrostContextKey = pluginName + "-transport-pre-hook"
	transportPostHookKey schemas.BifrostContextKey = pluginName + "-transport-post-hook"
	preHookKey           schemas.BifrostContextKey = pluginName + "-pre-hook"
)

// Init is called when the plugin is loaded.
// It is not part of any core interface; the external loader resolves it by name.
// config contains the plugin configuration from config.json.
func Init(config any) error {
	fmt.Println("Init called")
	// Initialize your plugin here (database connections, API clients, etc.)
	return nil
}

// GetName returns the plugin's unique identifier.
// This must match the "name" field in config.json plugins array.
func GetName() string {
	return pluginName
}

// HTTPTransportPreHook intercepts requests BEFORE they enter Bifrost core
// Modify req in-place. Return (*HTTPResponse, nil) to short-circuit.
// Only called when using HTTP transport (bifrost-http)
func HTTPTransportPreHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest) (*schemas.HTTPResponse, error) {
	// Modify request in-place
	req.Headers[pluginHeader] = transportPreHookValue
	// Store value in context for PreLLMHook/PostLLMHook
	ctx.SetValue(transportPreHookKey, transportPreHookValue)
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportPreHook called")
	// Return nil to continue processing, or return &schemas.HTTPResponse{} to short-circuit
	return nil, nil
}

// HTTPTransportPostHook intercepts responses AFTER they exit Bifrost core
// Modify response in-place. Called in REVERSE order of pre-hooks.
// Only called when using HTTP transport (bifrost-http)
// NOT called for streaming responses - use HTTPTransportStreamChunkHook instead
func HTTPTransportPostHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, resp *schemas.HTTPResponse) error {
	// Modify response in-place
	resp.Headers[pluginHeader] = transportPostHookValue
	// Store value in context. Runs after PostLLMHook, so only later phases can read it.
	ctx.SetValue(transportPostHookKey, transportPostHookValue)
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportPostHook called")
	return nil
}

// appendStreamChunkSuffix appends the demo suffix to a streaming chat chunk's
// delta content. It is nil-safe across the whole delta chain.
func appendStreamChunkSuffix(chunk *schemas.BifrostStreamChunk) {
	if chunk == nil || chunk.BifrostChatResponse == nil {
		return
	}
	choices := chunk.BifrostChatResponse.Choices
	if len(choices) == 0 || choices[0].ChatStreamResponseChoice == nil {
		return
	}
	delta := choices[0].ChatStreamResponseChoice.Delta
	if delta == nil || delta.Content == nil {
		return
	}
	*delta.Content += streamChunkSuffix
}

// HTTPTransportStreamChunkHook intercepts streaming response chunks
// Modify chunk in-place or return a modified chunk
// Only called when using HTTP transport (bifrost-http) with streaming
func HTTPTransportStreamChunkHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, chunk *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error) {
	ctx.Log(schemas.LogLevelInfo, "HTTPTransportStreamChunkHook called")
	// Modify streaming content in-place
	appendStreamChunkSuffix(chunk)
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
	// Read a value written by an earlier phase (HTTPTransportPreHook).
	transportPreValue := ctx.Value(transportPreHookKey)
	ctx.SetValue(preHookKey, preHookStoredValue)
	ctx.Log(schemas.LogLevelInfo, fmt.Sprintf("PreLLMHook called (transportPreHook=%v)", transportPreValue))
	return req, nil, nil
}

// PostLLMHook processes responses after provider calls
// Runs per provider attempt, in REVERSE order of PreLLMHook
// It may modify the response and/or error: recover (nil error + response) or
// invalidate (nil response + error). Return both unchanged to pass through.
func PostLLMHook(ctx *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	// Only read keys written by earlier phases; transportPostHookKey is written
	// afterwards by HTTPTransportPostHook, so it is NOT readable here.
	transportPreValue := ctx.Value(transportPreHookKey)
	preHookValue := ctx.Value(preHookKey)
	ctx.Log(schemas.LogLevelInfo, fmt.Sprintf("PostLLMHook called (transportPreHook=%v, preHook=%v)", transportPreValue, preHookValue))
	return resp, bifrostErr, nil
}

// Cleanup is called when Bifrost shuts down
func Cleanup() error {
	fmt.Println("Cleanup called")
	return nil
}

// Compile-time guards: each package-level hook must match the signature of the
// corresponding bifrost/core interface method. A future core upgrade that changes
// a signature fails the build here rather than shipping an unregisterable hook.
// Init has no core interface (the loader resolves it by name), so it is not pinned.
var (
	_ func() string = GetName
	_ func() error  = Cleanup

	_ func(*schemas.BifrostContext, *schemas.HTTPRequest) (*schemas.HTTPResponse, error)                                    = HTTPTransportPreHook
	_ func(*schemas.BifrostContext, *schemas.HTTPRequest, *schemas.HTTPResponse) error                                      = HTTPTransportPostHook
	_ func(*schemas.BifrostContext, *schemas.HTTPRequest, *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error) = HTTPTransportStreamChunkHook

	_ func(*schemas.BifrostContext, *schemas.BifrostRequest) error                                                                            = PreRequestHook
	_ func(*schemas.BifrostContext, *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error)                 = PreLLMHook
	_ func(*schemas.BifrostContext, *schemas.BifrostResponse, *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) = PostLLMHook
)
