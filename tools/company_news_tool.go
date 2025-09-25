package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CompanyNews 公司新闻结构体
type CompanyNews struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
	Source   string `json:"source"`
	Category string `json:"category"`
	DateTime string `json:"datetime"`
}

// CompanyNewsInput 公司新闻查询的输入参数
type CompanyNewsInput struct {
	Symbol string `json:"symbol" description:"股票代码，如 AAPL, TSLA, GOOG"`
	Date   string `json:"date,omitempty" description:"查询日期，格式为 YYYY-MM-DD，如果不提供则使用当前日期"`
	Limit  int    `json:"limit,omitempty" description:"返回新闻条数，默认为10条，最大20条"`
}

// CompanyNewsOutput 公司新闻查询的输出结果
type CompanyNewsOutput struct {
	Symbol string        `json:"symbol"`
	Date   string        `json:"date"`
	News   []CompanyNews `json:"news"`
	Count  int           `json:"count"`
	Error  string        `json:"error,omitempty"`
}

// NewCompanyNewsTool 创建新的公司新闻查询工具
func NewCompanyNewsTool(getNewsFunc func(symbol, date string, since *string, limit int) ([]CompanyNews, error)) (tool.BaseTool, error) {
	tool, err := utils.InferTool("get_company_news",
		"获取指定股票公司的最新新闻信息。这些新闻可以帮助分析公司的最新动态、市场情绪和潜在影响因素。",
		func(ctx context.Context, req *CompanyNewsInput) (*CompanyNewsOutput, error) {
			log.Printf("[CompanyNewsTool] 接收到请求: Symbol=%s, Date=%s, Limit=%d", req.Symbol, req.Date, req.Limit)

			// 验证必需参数
			if req.Symbol == "" {
				log.Printf("[CompanyNewsTool] 错误: 股票代码为空")
				return &CompanyNewsOutput{
					Error: "股票代码不能为空",
				}, nil
			}

			// 设置默认值
			date := req.Date
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			limit := req.Limit
			if limit <= 0 {
				limit = 10
			}
			if limit > 20 {
				limit = 20
			}

			log.Printf("[CompanyNewsTool] 准备调用API: Symbol=%s, Date=%s, Limit=%d", req.Symbol, date, limit)

			// 调用API获取新闻
			news, err := getNewsFunc(req.Symbol, date, nil, limit)
			if err != nil {
				log.Printf("[CompanyNewsTool] API调用失败: %v", err)
				return &CompanyNewsOutput{
					Symbol: req.Symbol,
					Date:   date,
					Error:  fmt.Sprintf("获取新闻失败: %v", err),
				}, nil
			}

			log.Printf("[CompanyNewsTool] API调用成功: 获取到 %d 条新闻", len(news))

			result := &CompanyNewsOutput{
				Symbol: req.Symbol,
				Date:   date,
				News:   news,
				Count:  len(news),
			}

			// 保存新闻到本地文件
			if err := saveNewsToFile(result); err != nil {
				log.Printf("[CompanyNewsTool] 保存文件失败: %v", err)
				// 不返回错误，继续返回新闻数据
			}

			log.Printf("[CompanyNewsTool] 返回响应: Symbol=%s, Date=%s, Count=%d", result.Symbol, result.Date, result.Count)
			return result, nil
		})
	if err != nil {
		return nil, err
	}
	return tool, nil
}

// saveNewsToFile 将新闻保存到本地文件
func saveNewsToFile(newsOutput *CompanyNewsOutput) error {
	// 创建news目录
	dirPath := "output/news"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名：news_AAPL_2025-09-25.json
	timeSuffix := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("news_%s_%s.json", newsOutput.Symbol, timeSuffix)
	filePath := filepath.Join(dirPath, fileName)

	// 将新闻数据转换为JSON
	data, err := json.MarshalIndent(newsOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	log.Printf("[CompanyNewsTool] 新闻已保存到: %s", filePath)
	return nil
}
