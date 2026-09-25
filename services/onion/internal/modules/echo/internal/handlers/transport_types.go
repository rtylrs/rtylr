package handlers

import echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"

// These aliases are deliberately limited to transport DTOs and renderer
// state. Business decisions and their implementations live in application.
type (
	CaretMode              = echoapp.CaretMode
	CaretIntentKind        = echoapp.CaretIntentKind
	CaretRequestPlan       = echoapp.CaretRequestPlan
	CaretResponsePolicy    = echoapp.CaretResponsePolicy
	CaretGroundingMode     = echoapp.CaretGroundingMode
	ToolDef                = echoapp.ToolDef
	AnalystService         = echoapp.AnalystService
	AnalystGroundingResult = echoapp.AnalystGroundingResult
	TextCompletionService  = echoapp.TextCompletionService
	ToolExecutor           = echoapp.ToolExecutor
	toolCallExecutionState = echoapp.ToolCallExecutionState
	cacheWritePlan         = echoapp.CacheWritePlan
)

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
