package lsp

import (
	"encoding/json"
	"fmt"
)

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentPositionParams represents parameters for position-based requests
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// InitializeParams represents the initialization parameters
type InitializeParams struct {
	ProcessID             *int                 `json:"processId"`
	RootPath              *string              `json:"rootPath,omitempty"`
	RootURI               *string              `json:"rootUri"`
	InitializationOptions interface{}          `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities   `json:"capabilities"`
	Trace                 string               `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder    `json:"workspaceFolders,omitempty"`
}

// ClientCapabilities defines the capabilities provided by the client
type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Experimental interface{}                    `json:"experimental,omitempty"`
}

// WorkspaceClientCapabilities defines workspace capabilities
type WorkspaceClientCapabilities struct {
	ApplyEdit              bool                        `json:"applyEdit,omitempty"`
	WorkspaceEdit          *WorkspaceEditCapabilities  `json:"workspaceEdit,omitempty"`
	DidChangeConfiguration *DidChangeConfigurationCapabilities `json:"didChangeConfiguration,omitempty"`
}

// WorkspaceEditCapabilities defines workspace edit capabilities
type WorkspaceEditCapabilities struct {
	DocumentChanges bool `json:"documentChanges,omitempty"`
}

// DidChangeConfigurationCapabilities defines configuration change capabilities
type DidChangeConfigurationCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// TextDocumentClientCapabilities defines text document capabilities
type TextDocumentClientCapabilities struct {
	Synchronization *TextDocumentSyncClientCapabilities `json:"synchronization,omitempty"`
	Completion      *CompletionCapabilities             `json:"completion,omitempty"`
	Hover           *HoverCapabilities                  `json:"hover,omitempty"`
	Definition      *DefinitionCapabilities             `json:"definition,omitempty"`
	References      *ReferencesCapabilities             `json:"references,omitempty"`
}

// TextDocumentSyncClientCapabilities defines text document sync capabilities
type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

// CompletionCapabilities defines completion capabilities
type CompletionCapabilities struct {
	DynamicRegistration bool                       `json:"dynamicRegistration,omitempty"`
	CompletionItem      *CompletionItemCapabilities `json:"completionItem,omitempty"`
}

// CompletionItemCapabilities defines completion item capabilities
type CompletionItemCapabilities struct {
	SnippetSupport          bool     `json:"snippetSupport,omitempty"`
	CommitCharactersSupport bool     `json:"commitCharactersSupport,omitempty"`
	DocumentationFormat     []string `json:"documentationFormat,omitempty"`
}

// HoverCapabilities defines hover capabilities
type HoverCapabilities struct {
	DynamicRegistration bool     `json:"dynamicRegistration,omitempty"`
	ContentFormat       []string `json:"contentFormat,omitempty"`
}

// DefinitionCapabilities defines definition capabilities
type DefinitionCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         bool `json:"linkSupport,omitempty"`
}

// ReferencesCapabilities defines references capabilities
type ReferencesCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// WorkspaceFolder represents a workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// InitializeResult represents the initialization result
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerCapabilities defines the capabilities provided by the server
type ServerCapabilities struct {
	TextDocumentSync           interface{}                 `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionOptions          `json:"completionProvider,omitempty"`
	HoverProvider              interface{}                 `json:"hoverProvider,omitempty"`
	DefinitionProvider         interface{}                 `json:"definitionProvider,omitempty"`
	ReferencesProvider         interface{}                 `json:"referencesProvider,omitempty"`
	DocumentFormattingProvider interface{}                 `json:"documentFormattingProvider,omitempty"`
}

// CompletionOptions defines completion options
type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// CompletionParams represents completion request parameters
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext represents additional information about completion
type CompletionContext struct {
	TriggerKind      int     `json:"triggerKind"`
	TriggerCharacter *string `json:"triggerCharacter,omitempty"`
}

// CompletionList represents a list of completion items
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label            string                 `json:"label"`
	Kind             *int                   `json:"kind,omitempty"`
	Detail           string                 `json:"detail,omitempty"`
	Documentation    interface{}            `json:"documentation,omitempty"`
	Deprecated       bool                   `json:"deprecated,omitempty"`
	Preselect        bool                   `json:"preselect,omitempty"`
	SortText         string                 `json:"sortText,omitempty"`
	FilterText       string                 `json:"filterText,omitempty"`
	InsertText       string                 `json:"insertText,omitempty"`
	InsertTextFormat *int                   `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit              `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit          `json:"additionalTextEdits,omitempty"`
	CommitCharacters []string               `json:"commitCharacters,omitempty"`
	Command          *Command               `json:"command,omitempty"`
	Data             interface{}            `json:"data,omitempty"`
}

// TextEdit represents a textual edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Command represents a reference to a command
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// HoverParams represents hover request parameters
type HoverParams struct {
	TextDocumentPositionParams
}

// Hover represents hover information
type Hover struct {
	Contents interface{} `json:"contents"`
	Range    *Range      `json:"range,omitempty"`
}

// MarkupContent represents marked up content
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// DefinitionParams represents definition request parameters
type DefinitionParams struct {
	TextDocumentPositionParams
}

// ReferenceParams represents reference request parameters
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext represents the context for reference requests
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// JSONRPCNotification represents a JSON-RPC 2.0 notification
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// CompletionItemKind constants
const (
	CompletionItemKindText          = 1
	CompletionItemKindMethod        = 2
	CompletionItemKindFunction      = 3
	CompletionItemKindConstructor   = 4
	CompletionItemKindField         = 5
	CompletionItemKindVariable      = 6
	CompletionItemKindClass         = 7
	CompletionItemKindInterface     = 8
	CompletionItemKindModule        = 9
	CompletionItemKindProperty      = 10
	CompletionItemKindUnit          = 11
	CompletionItemKindValue         = 12
	CompletionItemKindEnum          = 13
	CompletionItemKindKeyword       = 14
	CompletionItemKindSnippet       = 15
	CompletionItemKindColor         = 16
	CompletionItemKindFile          = 17
	CompletionItemKindReference     = 18
	CompletionItemKindFolder        = 19
	CompletionItemKindEnumMember    = 20
	CompletionItemKindConstant      = 21
	CompletionItemKindStruct        = 22
	CompletionItemKindEvent         = 23
	CompletionItemKindOperator      = 24
	CompletionItemKindTypeParameter = 25
)

// CompletionTriggerKind constants
const (
	CompletionTriggerKindInvoked           = 1
	CompletionTriggerKindTriggerCharacter  = 2
	CompletionTriggerKindTriggerForIncompleteCompletions = 3
)

// InsertTextFormat constants
const (
	InsertTextFormatPlainText = 1
	InsertTextFormatSnippet   = 2
)

// MarkupKind constants
const (
	MarkupKindPlainText = "plaintext"
	MarkupKindMarkdown  = "markdown"
)

// Error codes
const (
	ParseError           = -32700
	InvalidRequest       = -32600
	MethodNotFound       = -32601
	InvalidParams        = -32602
	InternalError        = -32603
	ServerErrorStart     = -32099
	ServerErrorEnd       = -32000
	ServerNotInitialized = -32002
	UnknownErrorCode     = -32001
	RequestCancelled     = -32800
	ContentModified      = -32801
)
