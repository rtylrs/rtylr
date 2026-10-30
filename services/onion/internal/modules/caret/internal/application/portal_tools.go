package application

import "strings"

// PortalFindPageToolName is offered on every turn: "where do I..." and "how do
// I... in rtylr" are the questions an in-app assistant gets most.
const PortalFindPageToolName = "portal.find_page"

// portalLayoutTriggers are the words people use to rearrange the Portal. The
// business classifier reads "move yourself to the top" as small talk, so the
// layout tools have their own triggers.
var portalLayoutTriggers = []string{
	"sidebar", "side bar", "app bar", "rail", "navigation", "layout", "icons", "menu bar",
	"move yourself", "move you", "move caret", "put caret", "put yourself", "to the top", "to the bottom",
	"reorder", "rearrange", "re-arrange", "pin ", "unpin", "hide ", "unhide", "show again", "bring back",
	"restore", "reset", "default", "customize", "customise", "first in", "last in",
	"الشريط", "القائمة الجانبية", "رتب", "أخفِ", "اخفي", "استعد", "الافتراضي",
	"barre latérale", "barre", "réorganise", "masque", "cache ", "restaure", "par défaut",
}

// PortalToolsForRequest returns the layout tools when the request is about
// arranging the Portal.
func PortalToolsForRequest(request string) []ToolDef {
	text := " " + strings.ToLower(request) + " "
	for _, trigger := range portalLayoutTriggers {
		if strings.Contains(text, trigger) {
			out := []ToolDef{}
			for _, name := range []string{"portal.layout", "portal.customize", "portal.restore_default"} {
				if tool, ok := ToolByName(name); ok {
					out = append(out, tool)
				}
			}
			return out
		}
	}
	return nil
}
