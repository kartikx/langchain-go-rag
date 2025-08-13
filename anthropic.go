package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func anthropic_agent() {
	client := anthropic.NewClient()

	ctx := context.Background()

	tools := getTools()

	take_input := true

	for {
		fmt.Println("Chat (Press Ctrl-C to quit)")

		messages := []anthropic.MessageParam{}

		for {
			if take_input {
				fmt.Print("> ")

				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n;')
				if err != nil {
					break
				}
				
				input = strings.TrimSpace(input)
				
				if input == "" {
					continue
				}
			
				messages = append(messages,
					anthropic.NewUserMessage(anthropic.NewTextBlock(input)))
			}

			message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
				MaxTokens: 1024,
				Messages:  messages,
				Model:     anthropic.ModelClaude3_5HaikuLatest,
				Tools:     tools,
			})

			if err != nil {
				panic(err.Error())
			}

			messages = append(messages, message.ToParam())
			toolResults := []anthropic.ContentBlockParamUnion{}

			for _, block := range message.Content {
				switch variant := block.AsAny().(type) {
				case anthropic.TextBlock:
					fmt.Printf("%s\n", message.Content[0].Text)
				case anthropic.ToolUseBlock:
					// fmt.Printf("Tool: %s\n", block.Name)
					var response interface{}
					switch block.Name {
					case "get_weather":
						var input struct {
							Location string `json:"location"`
						}

						err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
						if err != nil {
							panic(err)
						}
						response = getWeather(input.Location)
					}

					b, err := json.Marshal(response)
					if err != nil {
						panic(err)
					}

					// println(string(b))

					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(block.ID, string(b), false))
				}
			}

			if len(toolResults) == 0 {
				take_input = true
			} else {
				messages = append(messages, anthropic.NewUserMessage(toolResults...))
				take_input = false
			}
		}
	}
}

func getTools() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		{
			OfTool: &anthropic.ToolParam{
				Name:        "get_weather",
				Description: anthropic.String("accepts a place as an address and returns the weather temperature in fahrenheit"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get the weather for",
						},
					},
					Required: []string{"location"},
					Type:     "object",
				},
			},
		},
	}
}

// TODO - create a RAG tool.
func getWeather(location string) string {
	return "60.0"
}
