package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
)

func createDeepseekChatModel(ctx context.Context) model.ToolCallingChatModel {
	key := os.Getenv("DEEPSEEK_API_KEY")
	modelName := os.Getenv("DEEPSEEK_MODEL_NAME")
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	log.Printf("create deepseek chat model, baseURL=%s, modelName=%s, key=%s", baseURL, modelName, key)
	if err != nil {
		log.Fatalf("create deepseek chat model failed, err=%v", err)
	}
	return chatModel
}
