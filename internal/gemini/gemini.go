// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/genai"

	"github.com/usbarmory/kanzashi/internal/tool"
)

var (
	APIKey string
	Model  = "gemini-3-flash-preview"
)

const maxTurns = 64

// safeUint64 converts a function call argument to uint64 without float64
// precision loss. The Gemini SDK unmarshals all JSON numbers as float64,
// which loses precision for integers above 2^53. Re-parsing from the raw
// JSON number preserves the full 64-bit value.
func safeUint64(v interface{}) (uint64, error) {
	switch n := v.(type) {
	case float64:
		// Re-encode to JSON and decode as json.Number to avoid float64 rounding.
		b, err := json.Marshal(n)
		if err != nil {
			return 0, err
		}
		var num json.Number
		if err := json.Unmarshal(b, &num); err != nil {
			return 0, err
		}
		i, err := num.Int64()
		return uint64(i), err
	case json.Number:
		i, err := n.Int64()
		return uint64(i), err
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}

func getTools() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "reg_read32",
				Description: "Read a 32-bit value from a memory-mapped I/O register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Physical memory address (e.g. 0xfeb00000)"},
					},
				},
			},
			{
				Name:        "reg_write32",
				Description: "Write a 32-bit value to a memory-mapped I/O register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address", "value"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Register address (e.g. 0xfeb00000)"},
						"value":   {Type: genai.TypeInteger, Description: "Register value"},
					},
				},
			},
			{
				Name:        "reg_read64",
				Description: "Read a 64-bit value from a memory-mapped I/O register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Physical memory address (e.g. 0xfeb00000)"},
					},
				},
			},
			{
				Name:        "reg_write64",
				Description: "Write a 64-bit value to a memory-mapped I/O register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address", "value"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Register address (e.g. 0xfeb00000)"},
						"value":   {Type: genai.TypeInteger, Description: "Register value"},
					},
				},
			},
			{
				Name:        "msr_read",
				Description: "Read a 64-bit value from a machine-specific register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Register address (e.g. 0xfeb00000)"},
					},
				},
			},
			{
				Name:        "msr_write",
				Description: "Write a 64-bit value to a machine-specific register.",
				Parameters: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"address", "value"},
					Properties: map[string]*genai.Schema{
						"address": {Type: genai.TypeInteger, Description: "Register address (e.g. 0xfeb00000)"},
						"value":   {Type: genai.TypeInteger, Description: "Register value"},
					},
				},
			},
		},
	}
}

func executeTool(call *genai.FunctionCall) interface{} {
	switch call.Name {
	case "reg_read32":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		val, err := tool.Read32(uint32(a))
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%08X", val)
	case "reg_write32":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		v, err := safeUint64(call.Args["value"])
		if err != nil {
			return fmt.Sprintf("error parsing value: %v", err)
		}
		err = tool.Write32(uint32(a), uint32(v))
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	case "reg_read64":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		val, err := tool.Read64(a)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%016X", val)
	case "reg_write64":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		v, err := safeUint64(call.Args["value"])
		if err != nil {
			return fmt.Sprintf("error parsing value: %v", err)
		}
		err = tool.Write64(a, v)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	case "msr_read":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		val, err := tool.ReadMSR(a)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return fmt.Sprintf("0x%016X", val)
	case "msr_write":
		a, err := safeUint64(call.Args["address"])
		if err != nil {
			return fmt.Sprintf("error parsing address: %v", err)
		}
		v, err := safeUint64(call.Args["value"])
		if err != nil {
			return fmt.Sprintf("error parsing value: %v", err)
		}
		err = tool.WriteMSR(a, v)
		if err != nil {
			return fmt.Sprintf("error:%v", err)
		}
		return "ok"
	default:
		return "unknown tool"
	}
}

func RunAgent(ctx context.Context, system, user string) {
	log.Printf("[kanzashi] initializing gemini agent (%s)", Model)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  APIKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Print(err)
		return
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(system, genai.RoleUser),
		Tools:             []*genai.Tool{getTools()},
	}

	session, err := client.Chats.Create(ctx, Model, config, nil)

	if err != nil {
		log.Print(err)
		return
	}

	var prompt []genai.Part
	prompt = append(prompt, genai.Part{Text: user})

	for turn := 0; turn < maxTurns; turn++ {
		resp, err := session.SendMessage(ctx, prompt...)

		if err != nil {
			log.Printf("genai error: %+v", err)
			return
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			log.Printf("[gemini] empty response (finish reason: %v)", resp.Candidates[0].FinishReason)
			break
		}

		var toolCalls []*genai.FunctionCall
		for _, part := range resp.Candidates[0].Content.Parts {
			if len(part.Text) > 0 {
				fmt.Println(part.Text)
			}

			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, part.FunctionCall)
			}
		}

		prompt = nil

		for _, call := range toolCalls {
			result := executeTool(call)
			prompt = append(prompt, genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name:     call.Name,
					Response: map[string]interface{}{"result": result},
				},
			})
		}
	}
}
