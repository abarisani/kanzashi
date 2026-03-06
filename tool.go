package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

var tools = []anthropic.ToolUnionParam{
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_read",
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
		Name:        "reg_write",
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
}

type RegReadArgs struct {
	Address uint32 `json:"address"`
	Error   error  `json:"error"`
}

type RegWriteArgs struct {
	Address uint32 `json:"address"`
	Value   uint32 `json:"value"`
	Error   error  `json:"error"`
}

func executeTool(name string, input json.RawMessage) string {
	switch name {
	case "reg_read":
		var args RegReadArgs
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		val, err := Read(args.Address)
		log.Printf("[mmio] READ  0x%08X => 0x%08X (%v)", args.Address, val, err)
		return fmt.Sprintf("0x%08X (%v)", val, err)

	case "reg_write":
		var args RegWriteArgs
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		err := Write(args.Address, args.Value)
		log.Printf("[mmio] WRITE 0x%08X <= 0x%08X (%v)", args.Address, args.Value, err)
		return "ok"

	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}

func runAgent(ctx context.Context, client *anthropic.Client, system, user string) error {
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
	}

	for {
		resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeOpus4_6,
			MaxTokens: 4096,
			System:    []anthropic.TextBlockParam{{Text: system}},
			Tools:     tools,
			Messages:  messages,
		})
		if err != nil {
			return fmt.Errorf("api error: %w", err)
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

	return nil
}
