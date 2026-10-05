package application

import (
	"fmt"
	"sort"
	"strings"
)

type CaretSlashCommand struct {
	Name        string
	Description string
	Usage       string
	Mode        CaretMode
}

var caretSlashCommands = map[string]CaretSlashCommand{
	"goal": {
		Name:        "goal",
		Description: "Turn an outcome into a clear goal and next steps.",
		Usage:       "/goal <outcome>",
		Mode:        ModeAsk,
	},
	"brief": {
		Name:        "brief",
		Description: "Give a concise owner brief for the current business.",
		Usage:       "/brief",
		Mode:        ModeAnalyst,
	},
	"status": {
		Name:        "status",
		Description: "Summarize what is happening and what needs attention.",
		Usage:       "/status",
		Mode:        ModeAnalyst,
	},
	"remember": {
		Name:        "remember",
		Description: "Save a business preference or operating rule.",
		Usage:       "/remember <fact>",
		Mode:        ModeAgent,
	},
	"help": {
		Name:        "help",
		Description: "Show what Caret can do and how to use these commands.",
		Usage:       "/help",
		Mode:        ModeSupport,
	},
}

func ParseCaretSlashCommand(prompt string) (CaretSlashCommand, string, bool) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "/") {
		return CaretSlashCommand{}, "", false
	}

	commandToken := trimmed
	args := ""
	if fields := strings.Fields(trimmed); len(fields) > 0 {
		commandToken = fields[0]
		args = strings.TrimSpace(strings.TrimPrefix(trimmed, commandToken))
	}
	if len(commandToken) < 2 || strings.Contains(commandToken[1:], "/") {
		return CaretSlashCommand{}, "", false
	}

	command, ok := caretSlashCommands[strings.ToLower(commandToken[1:])]
	return command, args, ok
}

// NormalizeCaretSlashCommand expands a recognized command into a server-owned
// intent while leaving the literal command available to the caller for history.
func NormalizeCaretSlashCommand(prompt string) (string, CaretSlashCommand, bool) {
	command, args, ok := ParseCaretSlashCommand(prompt)
	if !ok {
		return prompt, CaretSlashCommand{}, false
	}

	if args == "" && (command.Name == "goal" || command.Name == "remember") {
		return fmt.Sprintf(
			"The user invoked %s without the required details. Explain what this command does, show %s, and ask for the missing input. Do not make up the missing input.",
			command.Usage,
			command.Usage,
		), command, true
	}

	switch command.Name {
	case "goal":
		return "The user invoked /goal. Turn the user-provided outcome below into one clear, measurable goal with success criteria, a short plan, risks, and the next concrete step. Do not treat this as permission to change business data.\n\nUser-provided outcome:\n" + args, command, true
	case "brief":
		return "The user invoked /brief. Prepare a concise owner brief for the current business: current position, strongest evidence, risks or opportunities, and one to three next moves. Use live business evidence when available and say plainly when a metric has no visible data.\n\nBrief request:\n" + args, command, true
	case "status":
		return "The user invoked /status. Give a current business overview with what is happening, what needs attention, and the next useful action. Use live evidence when available; do not invent empty or missing values.\n\nStatus request:\n" + args, command, true
	case "remember":
		return "The user invoked /remember. Treat the user-provided text below as a proposed business preference or operating rule. Propose the memory.remember action and wait for explicit approval before saving it.\n\nUser-provided memory:\n" + args, command, true
	case "help":
		return "The user invoked /help. Explain Caret's slash commands in a short list, including each command's purpose and usage. Do not claim that any command has run.\n\nAvailable commands:\n" + CaretSlashCommandHelp(), command, true
	default:
		return prompt, command, true
	}
}

func CaretSlashCommandHelp() string {
	commands := make([]string, 0, len(caretSlashCommands))
	for _, command := range caretSlashCommands {
		commands = append(commands, fmt.Sprintf("- %s — %s", command.Usage, command.Description))
	}
	sort.Strings(commands)
	return strings.Join(commands, "\n")
}
