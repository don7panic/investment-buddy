package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"investment/tools"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: investment_assistant <stock_symbol>")
		fmt.Println("Example: investment_assistant AAPL")
		fmt.Println("Example: investment_assistant TSLA")
		os.Exit(1)
	}

	// load env from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	ctx := context.Background()
	// 创建聊天模型 使用Gemini模型
	modelType := os.Getenv("MODEL_TYPE")
	var chatModel model.ToolCallingChatModel
	switch modelType {
	case "gemini":
		chatModel = createGeminiChatModel(ctx)
	case "openai":
		chatModel = createOpenAIChatModel(ctx)
	case "deepseek":
		chatModel = createDeepseekChatModel(ctx)
	default:
		chatModel = createDeepseekChatModel(ctx)
	}
	log.Printf("Using model: %s", modelType)

	symbol := strings.ToUpper(os.Args[1])
	fmt.Printf("=== 智能投资助手 - 股票分析：%s ===\n", symbol)
	fmt.Printf("正在初始化 React Agent 并准备分析工具...\n")

	// 使用 React Agent 进行分析
	result, err := analyzeWithReactAgent(ctx, chatModel, symbol)
	if err != nil {
		log.Printf("投资分析失败: %v", err)
		return
	}

	// 输出分析结果
	// fmt.Print("\n" + strings.Repeat("=", 50) + "\n")
	// fmt.Printf("📊 投资分析报告\n")
	// fmt.Print(strings.Repeat("=", 50) + "\n")
	// fmt.Printf("%s\n", result)
	fmt.Print(strings.Repeat("=", 50) + "\n")
	fmt.Printf("✅ 分析完成\n")

	// 保存分析结果为 markdown 文件
	if err := saveReportAsMarkdown(symbol, result); err != nil {
		log.Printf("保存报告失败: %v", err)
		return
	}

	fmt.Printf("📄 报告已保存为 markdown 文件: %s_report.md\n", symbol)
}

// 保存分析结果为 markdown 文件
func saveReportAsMarkdown(symbol, result string) error {
	// 生成文件名
	outputDir := "output/report"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	filename := fmt.Sprintf("%s_report.md", symbol)

	// 构建完整的 markdown 内容
	timestamp := fmt.Sprintf("分析时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	reportContent := fmt.Sprintf("# %s 投资分析报告\n\n%s\n\n%s", symbol, timestamp, result)

	// 写入文件
	filePath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(filePath, []byte(reportContent), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// 使用 React Agent 进行分析
func analyzeWithReactAgent(ctx context.Context, chatModel model.ToolCallingChatModel, symbol string) (string, error) {
	fmt.Printf("🔧 创建投资分析工具集...\n")
	// 创建工具集
	var investmentTools []tool.BaseTool

	// 创建市值查询工具
	marketCapToolFunc := func(symbol, date string) (float64, error) {
		return GetMarketCap(symbol, date)
	}
	marketCapTool, err := tools.NewMarketCapTool(marketCapToolFunc)
	if err != nil {
		return "", fmt.Errorf("创建市值工具失败: %v", err)
	}
	investmentTools = append(investmentTools, marketCapTool)

	// 创建财务指标工具
	metricsToolFunc := func(symbol, date, period string, limit int) ([]tools.FinancialMetrics, error) {
		return GetFinancialMetrics(symbol, date, period, limit)
	}
	metricsTool, err := tools.NewFinancialMetricsTool(metricsToolFunc)
	if err != nil {
		return "", fmt.Errorf("创建财务指标工具失败: %v", err)
	}
	investmentTools = append(investmentTools, metricsTool)

	// 创建新闻工具
	newsToolFunc := func(symbol, date string, since *string, limit int) ([]tools.CompanyNews, error) {
		news, err := GetCompanyNews(symbol, date, since, limit)
		if err != nil {
			return nil, err
		}
		return news, nil
	}
	newsTool, err := tools.NewCompanyNewsTool(newsToolFunc)
	if err != nil {
		return "", fmt.Errorf("创建新闻工具失败: %v", err)
	}
	investmentTools = append(investmentTools, newsTool)

	// 创建基本面分析工具
	fundamentalTool, err := tools.NewFundamentalAnalysisTool(ctx)
	if err != nil {
		return "", fmt.Errorf("创建基本面分析工具失败: %v", err)
	}
	investmentTools = append(investmentTools, fundamentalTool)

	toolCallChecker := func(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
		defer sr.Close()
		for {
			msg, err := sr.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// finish
					break
				}

				return false, err
			}

			if len(msg.ToolCalls) > 0 {
				return true, nil
			}
		}

		return false, nil
	}

	// 创建 React Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: investmentTools,
		},
		StreamToolCallChecker: toolCallChecker,
		MaxStep:               10, // 最大推理步数，允许多步骤分析
	})
	if err != nil {
		return "", fmt.Errorf("创建 React Agent 失败: %v", err)
	}

	// 构建系统提示词，指导 Agent 进行投资分析
	systemPrompt := `你是一个专业的股票投资分析师，具有深厚的价值投资理念和丰富的分析经验。你会系统性地收集和分析数据，遵循严格的投资分析流程。

## 你可以使用的工具：

- get_market_cap: 获取股票市值信息
- get_financial_metrics: 获取财务指标数据（ROE、债务比率、营运利润率等）
- get_company_news: 获取公司最新新闻动态
- analyze_fundamentals: 进行巴菲特式基本面分析

## 分析步骤：

- 先思考分析计划，然后获取股票基本信息（市值）
- 获取财务指标数据，重点关注过去5年的趋势
- 获取公司最新新闻，了解业务动态和市场情绪
- 使用基本面分析工具，输入财务指标进行量化评估
- 综合所有信息，形成最终投资建议

## 分析原则：

- 数据驱动：所有结论都要基于具体的财务数据
- 质量优先：重视ROE稳定性、低债务、强现金流
- 长期视角：关注公司的护城河和持续竞争优势
- 估值理性：不追高，寻找价值被低估的机会
- 风险管控：明确指出投资风险和注意事项

## 输出要求：

- 输出格式为 markdown
- 清晰说明每步分析的思路
- 展示关键财务数据和趋势
- 提供明确的投资评级（强烈推荐/推荐/中性/谨慎/避免）
- 给出目标价位和风险提示

请按照以上流程进行分析，确保每个步骤都有充分的数据支撑。`

	userPrompt := fmt.Sprintf("请分析股票 %s 的投资价值。请按照标准的投资分析流程，收集必要的数据并进行综合评估，最后给出投资建议。", symbol)

	// 创建消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: userPrompt,
		},
	}

	fmt.Printf("🤖 启动 React Agent 进行智能分析...\n")
	fmt.Printf("📈 Agent 将自动收集数据、进行分析并生成报告\n\n")

	// 使用 React Agent 的流式输出能力
	opts, future := react.WithMessageFuture()
	stream, err := agent.Stream(ctx, messages, opts)
	if err != nil {
		return "", fmt.Errorf("analyze failed with React Agent stream: %v", err)
	}
	defer stream.Close()

	// Get message streams from future
	sIter := future.GetMessageStreams()
	for {
		s, hasNext, err := sIter.Next()
		if err != nil {
			return "", err
		}
		if !hasNext {
			break
		}

		msg, err := schema.ConcatMessageStream(s)
		if err != nil {
			return "", err
		}
		if msg.Role == schema.Tool {
			fmt.Printf("Tool %s called\n", msg.ToolName)
			continue
		}
		if msg.Content != "" {
			fmt.Println(msg.Content)
		}
		// fmt.Printf("recv msg: role: %v, content: %v\n", msg.Role, msg.Content)
	}
	finalResponse, err := schema.ConcatMessageStream(stream)
	if err != nil {
		return "", err
	}
	return finalResponse.Content, nil
}
