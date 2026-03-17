// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build claude

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/usbarmory/kanzashi/internal/tool"
)

var ClaudeAPIKey string
var ClaudeModel = anthropic.ModelClaudeOpus4_6

var tools = []anthropic.ToolUnionParam{
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_read32",
		Description: anthropic.String("Read a 32-bit value from a memory-mapped I/O register at the given physical address."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to read from (e.g. 0xFEB00000)",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_write32",
		Description: anthropic.String("Write a 32-bit value to a memory-mapped I/O register at the given physical address."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to write to",
				},
				"value": map[string]interface{}{
					"type":        "integer",
					"description": "32-bit value to write",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address", "value"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_read64",
		Description: anthropic.String("Read a 64-bit value from a memory-mapped I/O register at the given physical address."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to read from (e.g. 0xFEB00000)",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_write64",
		Description: anthropic.String("Write a 64-bit value to a memory-mapped I/O register at the given physical address."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to write to",
				},
				"value": map[string]interface{}{
					"type":        "integer",
					"description": "64-bit value to write",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address", "value"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "msr_read",
		Description: anthropic.String("Read a 64-bit machine specific register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to read from",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address", "value"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "msr_write",
		Description: anthropic.String("Write a 64-bit value to a machine specific register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "integer",
					"description": "Physical memory address to write to",
				},
				"value": map[string]interface{}{
					"type":        "integer",
					"description": "64-bit value to write",
				},
				"error": map[string]interface{}{
					"type":        "string",
					"description": "Error value",
				},
			},
			Required: []string{"address", "value"},
		},
	}},
}

type RegRead32Args struct {
	Address uint32 `json:"address"`
	Error   error  `json:"error"`
}

type RegWrite32Args struct {
	Address uint32 `json:"address"`
	Value   uint32 `json:"value"`
	Error   error  `json:"error"`
}

type RegRead64Args struct {
	Address uint64 `json:"address"`
	Error   error  `json:"error"`
}

type RegWrite64Args struct {
	Address uint64 `json:"address"`
	Value   uint64 `json:"value"`
	Error   error  `json:"error"`
}

func executeTool(name string, input json.RawMessage) string {
	switch name {
	case "reg_read32":
		var args RegRead32Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		val, err := tool.Read32(args.Address)
		log.Printf("[mmio] READ32 %#x => %#x (%v)", args.Address, val, err)
		return fmt.Sprintf("0x%08X (%v)", val, err)

	case "reg_write32":
		var args RegWrite32Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		err := tool.Write32(args.Address, args.Value)
		log.Printf("[mmio] WRITE32 %#x <= %#x (%v)", args.Address, args.Value, err)
		return "ok"
	case "reg_read64":
		var args RegRead64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		val, err := tool.Read64(args.Address)
		log.Printf("[mmio] READ64 %#x => %#x (%v)", args.Address, val, err)
		return fmt.Sprintf("0x%16X (%v)", val, err)

	case "reg_write64":
		var args RegWrite64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		err := tool.Write64(args.Address, args.Value)
		log.Printf("[mmio] WRITE64 %#x <= %#x (%v)", args.Address, args.Value, err)
		return "ok"
	case "msr_read":
		var args RegRead64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		val, err := tool.ReadMSR(args.Address)
		log.Printf("[msr] READ %#x => %#x (%v)", args.Address, val, err)
		return fmt.Sprintf("0x%16X (%v)", val, err)
	case "msr_write":
		var args RegWrite64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		err := tool.WriteMSR(args.Address, args.Value)
		log.Printf("[msr] WRITE %#x <= %#x (%v)", args.Address, args.Value, err)
		return "ok"
	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}

func RunAgent(ctx context.Context, system, user string) {
	log.Printf("[kanzashi] initializing claude agent (%s)", ClaudeModel)

	client := anthropic.NewClient(option.WithAPIKey(ClaudeAPIKey))

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
	}

	for {
		resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     ClaudeModel,
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: system, CacheControl: anthropic.CacheControlEphemeralParam{
					TTL: anthropic.CacheControlEphemeralTTLTTL1h,
				}},
			},
			Tools:    tools,
			Messages: messages,
		})
		if err != nil {
			log.Fatalf("api error: %w", err)
		}

		// Append assistant response to history
		var assistantBlocks []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(v.Text))
			case anthropic.ToolUseBlock:
				var input any
				json.Unmarshal(v.Input, &input)
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(v.ID, input, v.Name))
			}

		}
		messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

		for _, block := range resp.Content {
			if v, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Printf("\n[claude] %s\n", v.Text)
			}
		}

		if resp.StopReason != anthropic.StopReasonToolUse {
			break
		}

		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			if v, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				result := executeTool(v.Name, v.Input)
				log.Printf("[tool] %s(%s) => %s", v.Name, string(v.Input), result)
				toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, result, false))
			}
		}
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}
