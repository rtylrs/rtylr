package application

import _ "embed"

// caretSystemPromptTemplate serves the mode-aware prompt builder.
//
//go:embed system_prompt.txt
var caretSystemPromptTemplate string

// assistantPromptTemplate serves Caret's unified chat runtime.
//
//go:embed assistant_prompt.txt
var assistantPromptTemplate string
