package application

import (
	"sort"
	"strings"
)

// ToolSearchName is the meta-tool the model uses to discover the rest of the
// catalog. Only a handful of tools are listed in each prompt; the model asks
// for the others by describing what it needs, which keeps the prompt small
// while every executable tool stays reachable.
const ToolSearchName = "tools.search"

const toolSearchLimit = 8

// toolSearchSynonyms maps everyday words onto the vocabulary the tool
// descriptions use, so "who is on my team" finds the HR employee tools.
var toolSearchSynonyms = map[string]string{
	"team":         "employee staff hr",
	"teammate":     "employee staff",
	"teammates":    "employee staff",
	"colleague":    "employee staff",
	"colleagues":   "employee staff",
	"coworker":     "employee staff",
	"coworkers":    "employee staff",
	"people":       "employee staff customer",
	"person":       "employee customer",
	"member":       "employee member",
	"members":      "employee member",
	"workers":      "employee staff",
	"client":       "customer",
	"clients":      "customer",
	"buyer":        "customer",
	"buyers":       "customer",
	"money":        "revenue cash finance",
	"earned":       "revenue sales",
	"income":       "revenue profit",
	"spent":        "expense",
	"spending":     "expense",
	"salary":       "payroll",
	"salaries":     "payroll",
	"wage":         "payroll",
	"vacation":     "leave",
	"holiday":      "leave",
	"sick":         "leave",
	"absent":       "attendance leave",
	"late":         "attendance",
	"todo":         "task",
	"todos":        "task",
	"assignment":   "task",
	"meeting":      "appointment calendar",
	"meetings":     "appointment calendar",
	"items":        "product",
	"goods":        "product inventory",
	"stock":        "inventory stock",
	"sold":         "sales order",
	"selling":      "sales bestseller",
	"complaint":    "support ticket",
	"earlier":      "chats past",
	"before":       "chats past",
	"discussed":    "chats past",
	"mentioned":    "chats past",
	"decided":      "chats past",
	"conversation": "chats past",
	"forget":       "memory forget",
	"remember":     "memory",
	"know":         "memory list",
	"complaints":   "support ticket",
}

// SearchTools ranks the executable catalog for a free-text need.
func SearchTools(mode CaretMode, query string) []ToolDef {
	words := strings.Fields(normalizePromptToolText(query))
	expanded := append([]string(nil), words...)
	for _, word := range words {
		if extra, ok := toolSearchSynonyms[word]; ok {
			expanded = append(expanded, extra)
		}
	}
	text := strings.Join(expanded, " ")
	if text == "" {
		return nil
	}

	tools := PromptToolsForMode(mode)
	scored := make([]promptToolScore, 0, len(tools))
	for i, tool := range tools {
		if tool.Name == ToolSearchName {
			continue
		}
		if score := promptToolScoreForText(tool, text); score > 0 {
			scored = append(scored, promptToolScore{tool: tool, score: score, index: i})
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].index < scored[j].index
		}
		return scored[i].score > scored[j].score
	})
	if len(scored) > toolSearchLimit {
		scored = scored[:toolSearchLimit]
	}
	out := make([]ToolDef, len(scored))
	for i, item := range scored {
		out[i] = item.tool
	}
	return out
}

// toolSearchOutput is what the model sees: each match's exact call contract.
func toolSearchOutput(mode CaretMode, query string) map[string]any {
	matches := SearchTools(mode, query)
	tools := make([]map[string]string, len(matches))
	for i, tool := range matches {
		tools[i] = map[string]string{
			"name":        tool.Name,
			"contract":    toolContractPrompt(tool),
			"description": tool.Description,
		}
	}
	note := "Call the best matching tool next with the tool marker. These tools are now available for this request."
	if len(tools) == 0 {
		note = "No tool matches. Try a different description, or tell the person this is not available."
	}
	return map[string]any{"query": query, "tools": tools, "note": note}
}
