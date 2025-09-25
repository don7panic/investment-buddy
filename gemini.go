package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"
)

func createGeminiChatModel(ctx context.Context) model.ToolCallingChatModel {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}
	modelName := os.Getenv("GEMINI_MODEL_NAME")
	if modelName == "" {
		log.Fatalf("GEMINI_MODEL_NAME is not set")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})
	if err != nil {
		log.Fatalf("create gemini client failed, err=%v", err)
	}
	chatModel, err := gemini.NewChatModel(ctx, &gemini.Config{
		Client: client,
		Model:  modelName,
	})
	if err != nil {
		log.Fatalf("create gemini chat model failed, err=%v", err)
	}
	return chatModel
}
