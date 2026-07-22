// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/usbarmory/kanzashi/internal/tool"
)

var (
	APIKey string
	Model  = anthropic.ModelClaudeOpus4_8
)

const maxTurns = 32

// Addresses and values are passed as strings rather than JSON integers: JSON
// has no hexadecimal literal notation, so an "integer" parameter forces the
// model to convert hex to decimal itself, which is unreliable for large
// values. Keeping the hex literal intact and parsing it here removes that
// conversion step entirely.
const (
	addrDesc  = `Register address as a hex string, e.g. "0xFED40F00". Pass the value verbatim, do not convert it to decimal.`
	msrDesc   = `MSR index as a hex string, e.g. "0x3A". Pass the value verbatim, do not convert it to decimal.`
	val32Desc = `32-bit value to write, as a hex string, e.g. "0x00000001". Pass the value verbatim, do not convert it to decimal.`
	val64Desc = `64-bit value to write, as a hex string, e.g. "0x0000000000000001". Pass the value verbatim, do not convert it to decimal.`
)

var tools = []anthropic.ToolUnionParam{
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_read32",
		Description: anthropic.String("Read a 32-bit value from a memory-mapped I/O register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": addrDesc,
				},
			},
			Required: []string{"address"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_write32",
		Description: anthropic.String("Write a 32-bit value to a memory-mapped I/O register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": addrDesc,
				},
				"value": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": val32Desc,
				},
			},
			Required: []string{"address", "value"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_read64",
		Description: anthropic.String("Read a 64-bit value from a memory-mapped I/O register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": addrDesc,
				},
			},
			Required: []string{"address"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "reg_write64",
		Description: anthropic.String("Write a 64-bit value to a memory-mapped I/O register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": addrDesc,
				},
				"value": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": val64Desc,
				},
			},
			Required: []string{"address", "value"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "msr_read",
		Description: anthropic.String("Read a 64-bit machine-specific register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": msrDesc,
				},
			},
			Required: []string{"address"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "msr_write",
		Description: anthropic.String("Write a 64-bit value to a machine-specific register."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"address": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": msrDesc,
				},
				"value": map[string]interface{}{
					"type":        "string",
					"pattern":     "^0[xX][0-9a-fA-F]+$",
					"description": val64Desc,
				},
			},
			Required: []string{"address", "value"},
		},
		CacheControl: anthropic.CacheControlEphemeralParam{
			TTL: anthropic.CacheControlEphemeralTTLTTL1h,
		},
	}},
}

type RegRead32Args struct {
	Address string `json:"address"`
}

type RegWrite32Args struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

type RegRead64Args struct {
	Address string `json:"address"`
}

type RegWrite64Args struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

// parseUint parses an unsigned integer literal of at most the given bit size,
// accepting the "0x" prefix (as well as "0o" and "0b") via strconv base 0.
func parseUint(s string, bits int) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), 0, bits)
}

func executeTool(name string, input json.RawMessage) string {
	switch name {
	case "reg_read32":
		var args RegRead32Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 32)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := tool.Read32(uint32(addr))
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%08X", val)
	case "reg_write32":
		var args RegWrite32Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 32)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := parseUint(args.Value, 32)
		if err != nil {
			return fmt.Sprintf("error parsing value %q: %v", args.Value, err)
		}
		if err := tool.Write32(uint32(addr), uint32(val)); err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	case "reg_read64":
		var args RegRead64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 64)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := tool.Read64(addr)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%016X", val)
	case "reg_write64":
		var args RegWrite64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 64)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := parseUint(args.Value, 64)
		if err != nil {
			return fmt.Sprintf("error parsing value %q: %v", args.Value, err)
		}
		if err := tool.Write64(addr, val); err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	case "msr_read":
		var args RegRead64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 64)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := tool.ReadMSR(addr)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%016X", val)
	case "msr_write":
		var args RegWrite64Args
		if err := json.Unmarshal(input, &args); err != nil {
			return fmt.Sprintf("error parsing args: %v", err)
		}
		addr, err := parseUint(args.Address, 64)
		if err != nil {
			return fmt.Sprintf("error parsing address %q: %v", args.Address, err)
		}
		val, err := parseUint(args.Value, 64)
		if err != nil {
			return fmt.Sprintf("error parsing value %q: %v", args.Value, err)
		}
		if err := tool.WriteMSR(addr, val); err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}

func RunAgent(ctx context.Context, system, user string) {
	log.Printf("[kanzashi] initializing claude agent (%s)", Model)

	client := anthropic.NewClient(option.WithAPIKey(APIKey))

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
	}

	for turn := 0; turn < maxTurns; turn++ {
		resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     Model,
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
			log.Printf("api error: %v", err)
			return
		}

		// Append assistant response to history
		var assistantBlocks []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(v.Text))
			case anthropic.ToolUseBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(v.ID, v.Input, v.Name))
			}
		}

		messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

		for _, block := range resp.Content {
			if v, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Println(v.Text)
			}
		}

		if resp.StopReason != anthropic.StopReasonToolUse {
			break
		}

		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			if v, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				result := executeTool(v.Name, v.Input)
				toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, result, false))
			}
		}

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	log.Print("[kanzashi] terminated claude agent")
}
