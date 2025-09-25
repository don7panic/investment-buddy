package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// MarketCapInput 市值查询的输入参数
type MarketCapInput struct {
	Symbol string `json:"symbol" description:"股票代码，如 AAPL, TSLA, GOOG"`
	Date   string `json:"date,omitempty" description:"查询日期，格式为 YYYY-MM-DD，如果不提供则使用当前日期"`
}

// MarketCapOutput 市值查询的输出结果
type MarketCapOutput struct {
	Symbol    string  `json:"symbol"`
	Date      string  `json:"date"`
	MarketCap float64 `json:"market_cap"`
	Currency  string  `json:"currency"`
	Error     string  `json:"error,omitempty"`
}

// NewMarketCapTool 创建新的市值查询工具
func NewMarketCapTool(getMarketCapFunc func(symbol, date string) (float64, error)) (tool.BaseTool, error) {
	tool, err := utils.InferTool("get_market_cap",
		"获取指定股票在指定日期的市值信息。这是投资分析的基础数据，用于评估公司规模。",
		func(ctx context.Context, req *MarketCapInput) (*MarketCapOutput, error) {
			log.Printf("[MarketCapTool] 接收到请求: Symbol=%s, Date=%s", req.Symbol, req.Date)

			// 验证必需参数
			if req.Symbol == "" {
				log.Printf("[MarketCapTool] 错误: 股票代码为空")
				return &MarketCapOutput{
					Error: "股票代码不能为空",
				}, nil
			}

			// 如果没有提供日期，使用当前日期
			date := req.Date
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			log.Printf("[MarketCapTool] 准备调用API: Symbol=%s, Date=%s", req.Symbol, date)

			// 调用API获取市值
			marketCap, err := getMarketCapFunc(req.Symbol, date)
			if err != nil {
				log.Printf("[MarketCapTool] API调用失败: %v", err)
				return &MarketCapOutput{
					Symbol: req.Symbol,
					Date:   date,
					Error:  fmt.Sprintf("获取市值失败: %v", err),
				}, nil
			}

			log.Printf("[MarketCapTool] API调用成功: MarketCap=%.2f", marketCap)

			result := &MarketCapOutput{
				Symbol:    req.Symbol,
				Date:      date,
				MarketCap: marketCap,
				Currency:  "USD",
			}

			log.Printf("[MarketCapTool] 返回响应: Symbol=%s, Date=%s, MarketCap=%.2f, Currency=%s", result.Symbol, result.Date, result.MarketCap, result.Currency)
			return result, nil
		})
	if err != nil {
		return nil, err
	}
	return tool, nil
}
