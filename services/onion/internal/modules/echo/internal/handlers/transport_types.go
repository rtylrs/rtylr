package handlers

import echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"

// These aliases are deliberately limited to transport DTOs and renderer
// state. Business decisions and their implementations live in application.
type CaretMode = echoapp.CaretMode
type CaretIntentKind = echoapp.CaretIntentKind
type CaretRequestPlan = echoapp.CaretRequestPlan
type CaretResponsePolicy = echoapp.CaretResponsePolicy
type CaretGroundingMode = echoapp.CaretGroundingMode
type ToolDef = echoapp.ToolDef
type AnalystService = echoapp.AnalystService
type AnalystGroundingResult = echoapp.AnalystGroundingResult
type TextCompletionService = echoapp.TextCompletionService
type ToolExecutor = echoapp.ToolExecutor
type toolCallExecutionState = echoapp.ToolCallExecutionState
type cacheWritePlan = echoapp.CacheWritePlan

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

const deferredAnalyticsToolName = "analytics.deferred"

type analystGroundingResult = echoapp.AnalystGroundingResult
