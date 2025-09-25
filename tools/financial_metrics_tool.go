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

// FinancialMetricsInput 财务指标查询的输入参数
type FinancialMetricsInput struct {
	Symbol string `json:"symbol" description:"股票代码，如 AAPL, TSLA, GOOG"`
	Date   string `json:"date,omitempty" description:"查询日期，格式为 YYYY-MM-DD，如果不提供则使用当前日期"`
	Period string `json:"period,omitempty" description:"财务期间，ttm(过去12个月)、annual(年度)、quarterly(季度)，默认为ttm"`
	Limit  int    `json:"limit,omitempty" description:"返回数据条数，默认为5条，最大10条"`
}

// FinancialMetricsOutput 财务指标查询的输出结果
type FinancialMetricsOutput struct {
	Symbol  string             `json:"symbol"`
	Date    string             `json:"date"`
	Period  string             `json:"period"`
	Metrics []FinancialMetrics `json:"metrics"`
	Count   int                `json:"count"`
	Error   string             `json:"error,omitempty"`
}

// NewFinancialMetricsTool 创建新的财务指标查询工具
func NewFinancialMetricsTool(getMetricsFunc func(symbol, date, period string, limit int) ([]FinancialMetrics, error)) (tool.BaseTool, error) {
	tool, err := utils.InferTool("get_financial_metrics",
		"获取指定股票的财务指标数据，包括估值比率、盈利能力、营运效率、财务健康状况等关键指标。这些数据是进行基本面分析的核心。",
		func(ctx context.Context, req *FinancialMetricsInput) (*FinancialMetricsOutput, error) {
			log.Printf("[FinancialMetricsTool] 接收到请求: Symbol=%s, Date=%s, Period=%s, Limit=%d", req.Symbol, req.Date, req.Period, req.Limit)

			// 验证必需参数
			if req.Symbol == "" {
				log.Printf("[FinancialMetricsTool] 错误: 股票代码为空")
				return &FinancialMetricsOutput{
					Error: "股票代码不能为空",
				}, nil
			}

			// 设置默认值
			date := req.Date
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			period := req.Period
			if period == "" {
				period = "ttm"
			}

			limit := req.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 10 {
				limit = 10
			}

			log.Printf("[FinancialMetricsTool] 准备调用API: Symbol=%s, Date=%s, Period=%s, Limit=%d", req.Symbol, date, period, limit)

			// 调用API获取财务指标
			metrics, err := getMetricsFunc(req.Symbol, date, period, limit)
			if err != nil {
				log.Printf("[FinancialMetricsTool] API调用失败: %v", err)
				return &FinancialMetricsOutput{
					Symbol: req.Symbol,
					Date:   date,
					Period: period,
					Error:  fmt.Sprintf("获取财务指标失败: %v", err),
				}, nil
			}

			log.Printf("[FinancialMetricsTool] API调用成功: 获取到 %d 条财务指标", len(metrics))

			result := &FinancialMetricsOutput{
				Symbol:  req.Symbol,
				Date:    date,
				Period:  period,
				Metrics: metrics,
				Count:   len(metrics),
			}

			// 保存财务指标到本地文件
			if err := saveMetricsToFile(result); err != nil {
				log.Printf("[FinancialMetricsTool] 保存文件失败: %v", err)
				// 不返回错误，继续返回财务指标数据
			}

			log.Printf("[FinancialMetricsTool] 返回响应: Symbol=%s, Date=%s, Period=%s, Count=%d", result.Symbol, result.Date, result.Period, result.Count)
			return result, nil
		})
	if err != nil {
		return nil, err
	}
	return tool, nil
}

// saveMetricsToFile 将财务指标保存到本地文件
func saveMetricsToFile(metricsOutput *FinancialMetricsOutput) error {
	// 创建metrics目录
	dirPath := "output/metrics"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名：metrics_AAPL_ttm_2025-09-25.json
	timeSuffix := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("metrics_%s_%s_%s.json", metricsOutput.Symbol, metricsOutput.Period, timeSuffix)
	filePath := filepath.Join(dirPath, fileName)

	// 将财务指标数据转换为JSON
	data, err := json.MarshalIndent(metricsOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	log.Printf("[FinancialMetricsTool] 财务指标已保存到: %s", filePath)
	return nil
}
