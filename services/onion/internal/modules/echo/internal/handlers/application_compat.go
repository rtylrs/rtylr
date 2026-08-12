package handlers

import (
	"time"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
)

type CaretMode = echoapp.CaretMode
type CaretIntentKind = echoapp.CaretIntentKind
type CaretRequestPlan = echoapp.CaretRequestPlan
type CaretResponsePolicy = echoapp.CaretResponsePolicy
type CaretGroundingMode = echoapp.CaretGroundingMode
type ToolDef = echoapp.ToolDef

const (
	ModeAsk            = echoapp.ModeAsk
	ModeAgent          = echoapp.ModeAgent
	ModeAnalyst        = echoapp.ModeAnalyst
	ModeSupport        = echoapp.ModeSupport
	IntentConversation = echoapp.IntentConversation
	IntentSnapshot     = echoapp.IntentSnapshot
	IntentAnalytics    = echoapp.IntentAnalytics
	IntentLookup       = echoapp.IntentLookup
	IntentAction       = echoapp.IntentAction
	IntentMemory       = echoapp.IntentMemory
	IntentSupport      = echoapp.IntentSupport
	IntentNavigation   = echoapp.IntentNavigation
	GroundingNone      = echoapp.GroundingNone
	GroundingSnapshot  = echoapp.GroundingSnapshot
	GroundingAnalytics = echoapp.GroundingAnalytics
	GroundingDeferred  = echoapp.GroundingDeferred
)

var (
	ParseMode                   = echoapp.ParseMode
	ClassifyCaretRequest        = echoapp.ClassifyCaretRequest
	PromptToolsForRequest       = echoapp.PromptToolsForRequest
	PromptToolsForMode          = echoapp.PromptToolsForMode
	ToolsForMode                = echoapp.ToolsForMode
	ToolByName                  = echoapp.ToolByName
	ValidateToolRegistry        = echoapp.ValidateToolRegistry
	ClarificationHintForRequest = echoapp.ClarificationHintForRequest
)

func toolAllowedForMode(mode CaretMode, tool ToolDef) bool {
	return echoapp.ToolAllowedForMode(mode, tool)
}
func BuildCaretResponsePolicy(mode CaretMode, request, systemContext string, pageContext map[string]any) CaretResponsePolicy {
	return echoapp.BuildCaretResponsePolicy(mode, request, systemContext, pageContext)
}

const (
	maxUserPromptLen    = echoapp.MaxUserPromptLen
	maxSystemContextLen = echoapp.MaxSystemContextLen
)

func sanitizeUserInput(value string, maxBytes int) string {
	return echoapp.SanitizeUserInput(value, maxBytes)
}
func truncateUTF8ToBytes(value string, maxBytes int) string {
	return echoapp.TruncateUTF8ToBytes(value, maxBytes)
}
func buildRequestSystemContext(value string, page map[string]any) string {
	return echoapp.BuildRequestSystemContext(value, page)
}
func renderPageContext(page map[string]any) string { return echoapp.RenderPageContext(page) }
func normalizePageContext(value any, depth int) any {
	return echoapp.NormalizePageContext(value, depth)
}

type caretPromptInput struct {
	Mode               CaretMode
	ConversationPrompt string
	CurrentRequest     string
	SystemContext      string
	Now                time.Time
}

func buildCaretPrompt(in caretPromptInput) string {
	return echoapp.BuildCaretPrompt(echoapp.PromptInput{Mode: in.Mode, ConversationPrompt: in.ConversationPrompt, CurrentRequest: in.CurrentRequest, SystemContext: in.SystemContext, Now: in.Now})
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
