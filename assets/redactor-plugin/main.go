package main

import (
	"fmt"
	"regexp"

	"github.com/maximhq/bifrost/core/schemas"
)

// pluginName must match the "name" field in config.json.
const pluginName = "redactor-plugin"

// streamWatermark is appended to the final streamed chunk so a client can tell
// the stream passed through redaction. (It is a visible marker, not an
// imperceptible provenance watermark.)
const streamWatermark = " [redactor-plugin:stream]"

// redactionRule maps a human name to the pattern that finds it. Order matters:
// higher-specificity secrets are redacted before broader patterns.
type redactionRule struct {
	name string
	re   *regexp.Regexp
}

// redactionRules is a small, illustrative ruleset. Real deployments should widen
// it (names, addresses, account numbers, org-specific tokens) or call a dedicated
// PII/secret detector, and should keep it configurable rather than compiled in.
var redactionRules = []redactionRule{
	{"private-key", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"aws-access-key", regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{"provider-api-key", regexp.MustCompile(`\b(?:sk|pk|rk|api)[-_][A-Za-z0-9]{16,}\b`)},
	{"bearer-token", regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/-]{10,}=*`)},
	{"email", regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`)},
	{"ssn", regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)},
	{"credit-card", regexp.MustCompile(`\b(?:\d[ -]?){13,16}\d\b`)},
}

// redactText replaces every recognised secret/PII span with a typed placeholder.
func redactText(s string) string {
	for _, rule := range redactionRules {
		s = rule.re.ReplaceAllString(s, "[REDACTED:"+rule.name+"]")
	}
	return s
}

// redactMessage rewrites a chat message's text in place. It covers both the
// string form and the content-block form of a message.
func redactMessage(msg *schemas.ChatMessage) {
	if msg == nil || msg.Content == nil {
		return
	}
	if msg.Content.ContentStr != nil {
		redacted := redactText(*msg.Content.ContentStr)
		msg.Content.ContentStr = &redacted
	}
	for i := range msg.Content.ContentBlocks {
		block := &msg.Content.ContentBlocks[i]
		if block.Text != nil {
			redacted := redactText(*block.Text)
			block.Text = &redacted
		}
		if block.Refusal != nil {
			redacted := redactText(*block.Refusal)
			block.Refusal = &redacted
		}
	}
}

// Init is called when the plugin is loaded. It is resolved by name, not through
// an interface. config is the plugin's config value from config.json.
func Init(config any) error {
	fmt.Println(pluginName + " Init called")
	return nil
}

// GetName returns the plugin's unique identifier; it must match config.json.
func GetName() string {
	return pluginName
}

// PreLLMHook redacts the OUTGOING request (ingress to the provider) before it
// leaves the gateway, so secrets never reach the model provider.
func PreLLMHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error) {
	if req != nil && req.ChatRequest != nil {
		for i := range req.ChatRequest.Input {
			redactMessage(&req.ChatRequest.Input[i])
		}
	}
	ctx.Log(schemas.LogLevelInfo, "redactor: request redacted (ingress)")
	return req, nil, nil
}

// PostLLMHook redacts the INCOMING response (egress from the provider) before it
// returns to the caller. It never drops the provider error.
func PostLLMHook(ctx *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	if resp != nil && resp.ChatResponse != nil {
		for i := range resp.ChatResponse.Choices {
			choice := &resp.ChatResponse.Choices[i]
			if ns := choice.ChatNonStreamResponseChoice; ns != nil {
				redactMessage(ns.Message)
			}
			if st := choice.ChatStreamResponseChoice; st != nil && st.Delta != nil && st.Delta.Content != nil {
				redacted := redactText(*st.Delta.Content)
				st.Delta.Content = &redacted
			}
		}
	}
	ctx.Log(schemas.LogLevelInfo, "redactor: response redacted (egress)")
	return resp, bifrostErr, nil
}

// rewriteStreamChunk redacts a streamed chunk's delta content and, on the final
// chunk (one carrying a finish reason), appends the watermark.
//
// LIMITATION: redaction here is per-chunk. A secret split across two chunks
// (e.g. "sk-live" + "abc123") will not match either half and slips through.
// Closing that gap needs a rolling buffer over chunks, which changes streaming
// latency; treat this as illustrative, not production-grade DLP.
func rewriteStreamChunk(chunk *schemas.BifrostStreamChunk) {
	if chunk == nil || chunk.BifrostChatResponse == nil {
		return
	}
	if len(chunk.BifrostChatResponse.Choices) == 0 {
		return
	}
	choice := &chunk.BifrostChatResponse.Choices[0]
	stream := choice.ChatStreamResponseChoice
	if stream == nil {
		return
	}
	if stream.Delta != nil && stream.Delta.Content != nil {
		redacted := redactText(*stream.Delta.Content)
		stream.Delta.Content = &redacted
	}
	if choice.FinishReason != nil {
		content := ""
		if stream.Delta != nil && stream.Delta.Content != nil {
			content = *stream.Delta.Content
		}
		content += streamWatermark
		if stream.Delta == nil {
			stream.Delta = &schemas.ChatStreamResponseChoiceDelta{}
		}
		stream.Delta.Content = &content
	}
}

// HTTPTransportStreamChunkHook redacts streaming chunks before they reach the
// client. HTTP transport only, per chunk, reverse order.
func HTTPTransportStreamChunkHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, chunk *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error) {
	rewriteStreamChunk(chunk)
	ctx.Log(schemas.LogLevelInfo, "redactor: stream chunk redacted")
	return chunk, nil
}

// Cleanup is called when Bifrost shuts down.
func Cleanup() error {
	fmt.Println(pluginName + " Cleanup called")
	return nil
}

// Compile-time guards: each exported hook must match its bifrost/core interface
// method signature, so a future core upgrade fails the build rather than
// shipping an unregisterable hook.
var (
	_ func() string = GetName
	_ func() error  = Cleanup

	_ func(*schemas.BifrostContext, *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error)                 = PreLLMHook
	_ func(*schemas.BifrostContext, *schemas.BifrostResponse, *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) = PostLLMHook
	_ func(*schemas.BifrostContext, *schemas.HTTPRequest, *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error)                   = HTTPTransportStreamChunkHook
)
