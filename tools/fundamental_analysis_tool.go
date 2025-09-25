package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// FinancialMetrics 结构体
type FinancialMetrics struct {
	Ticker                        string   `json:"ticker"`
	ReportPeriod                  string   `json:"report_period"`
	Period                        string   `json:"period"`
	Currency                      string   `json:"currency"`
	MarketCap                     float64  `json:"market_cap"`
	EnterpriseValue               float64  `json:"enterprise_value"`
	PriceToEarningsRatio          float64  `json:"price_to_earnings_ratio"`
	PriceToBookRatio              float64  `json:"price_to_book_ratio"`
	PriceToSalesRatio             float64  `json:"price_to_sales_ratio"`
	EnterpriseValueToEbitdaRatio  float64  `json:"enterprise_value_to_ebitda_ratio"`
	EnterpriseValueToRevenueRatio float64  `json:"enterprise_value_to_revenue_ratio"`
	FreeCashFlowYield             float64  `json:"free_cash_flow_yield"`
	PegRatio                      float64  `json:"peg_ratio"`
	GrossMargin                   float64  `json:"gross_margin"`
	OperatingMargin               *float64 `json:"operating_margin"`
	NetMargin                     *float64 `json:"net_margin"`
	ReturnOnEquity                *float64 `json:"return_on_equity"`
	ReturnOnAssets                *float64 `json:"return_on_assets"`
	ReturnOnInvestedCapital       float64  `json:"return_on_invested_capital"`
	AssetTurnover                 float64  `json:"asset_turnover"`
	InventoryTurnover             float64  `json:"inventory_turnover"`
	ReceivablesTurnover           float64  `json:"receivables_turnover"`
	DaysSalesOutstanding          float64  `json:"days_sales_outstanding"`
	OperatingCycle                float64  `json:"operating_cycle"`
	WorkingCapitalTurnover        float64  `json:"working_capital_turnover"`
	CurrentRatio                  *float64 `json:"current_ratio"`
	QuickRatio                    *float64 `json:"quick_ratio"`
	CashRatio                     *float64 `json:"cash_ratio"`
	OperatingCashFlowRatio        float64  `json:"operating_cash_flow_ratio"`
	DebtToEquity                  *float64 `json:"debt_to_equity"`
	DebtToAssets                  float64  `json:"debt_to_assets"`
	InterestCoverage              *float64 `json:"interest_coverage"`
	RevenueGrowth                 float64  `json:"revenue_growth"`
	EarningsGrowth                float64  `json:"earnings_growth"`
	BookValueGrowth               float64  `json:"book_value_growth"`
	EarningsPerShareGrowth        float64  `json:"earnings_per_share_growth"`
	FreeCashFlowGrowth            float64  `json:"free_cash_flow_growth"`
	OperatingIncomeGrowth         float64  `json:"operating_income_growth"`
	EbitdaGrowth                  float64  `json:"ebitda_growth"`
	PayoutRatio                   float64  `json:"payout_ratio"`
	EarningsPerShare              float64  `json:"earnings_per_share"`
	BookValuePerShare             float64  `json:"book_value_per_share"`
	FreeCashFlowPerShare          float64  `json:"free_cash_flow_per_share"`
}

// FundamentalAnalysisRequest 基本面分析请求
type FundamentalAnalysisRequest struct {
	Metrics []FinancialMetrics `json:"metrics" jsonschema:"description=List of financial metrics for fundamental analysis"`
}

// FundamentalAnalysisResponse 基本面分析响应
type FundamentalAnalysisResponse struct {
	Score   int            `json:"score" jsonschema:"description=Overall fundamental score based on Buffett's criteria"`
	Details string         `json:"details" jsonschema:"description=Detailed reasoning for the analysis"`
	Metrics map[string]any `json:"metrics,omitempty" jsonschema:"description=Latest financial metrics used in analysis"`
	Error   string         `json:"error,omitempty" jsonschema:"description=Error message if analysis fails"`
}

// NewFundamentalAnalysisTool 创建基本面分析工具
func NewFundamentalAnalysisTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("analyze_fundamentals",
		"根据巴菲特的投资标准分析公司基本面，评估ROE、债务比率、营运利润率和流动比率等关键指标",
		func(ctx context.Context, req *FundamentalAnalysisRequest) (*FundamentalAnalysisResponse, error) {
			log.Printf("[FundamentalAnalysisTool] 接收到请求: 财务指标数量=%d", len(req.Metrics))

			if len(req.Metrics) == 0 {
				log.Printf("[FundamentalAnalysisTool] 错误: 未提供财务指标数据")
				return &FundamentalAnalysisResponse{
					Score:   0,
					Details: "基本面数据不足",
					Error:   "未提供财务指标数据",
				}, nil
			}

			// 使用最新的财务指标进行分析
			latestMetrics := req.Metrics[0]
			log.Printf("[FundamentalAnalysisTool] 开始分析: Ticker=%s, ReportPeriod=%s", latestMetrics.Ticker, latestMetrics.ReportPeriod)

			score := 0
			var reasoning []string

			// 检查ROE (股本回报率)
			if latestMetrics.ReturnOnEquity != nil && *latestMetrics.ReturnOnEquity > 0.15 {
				score += 2
				reasoning = append(reasoning, fmt.Sprintf("强劲的ROE为%.1f%%", *latestMetrics.ReturnOnEquity*100))
			} else if latestMetrics.ReturnOnEquity != nil {
				reasoning = append(reasoning, fmt.Sprintf("ROE较弱为%.1f%%", *latestMetrics.ReturnOnEquity*100))
			} else {
				reasoning = append(reasoning, "ROE数据不可用")
			}

			// 检查债务股权比
			if latestMetrics.DebtToEquity != nil && *latestMetrics.DebtToEquity < 0.5 {
				score += 2
				reasoning = append(reasoning, "保守的债务水平")
			} else if latestMetrics.DebtToEquity != nil {
				reasoning = append(reasoning, fmt.Sprintf("较高的债务股权比为%.1f", *latestMetrics.DebtToEquity))
			} else {
				reasoning = append(reasoning, "债务股权比数据不可用")
			}

			// 检查营运利润率
			if latestMetrics.OperatingMargin != nil && *latestMetrics.OperatingMargin > 0.15 {
				score += 2
				reasoning = append(reasoning, "强劲的营运利润率")
			} else if latestMetrics.OperatingMargin != nil {
				reasoning = append(reasoning, fmt.Sprintf("营运利润率较弱为%.1f%%", *latestMetrics.OperatingMargin*100))
			} else {
				reasoning = append(reasoning, "营运利润率数据不可用")
			}

			// 检查流动比率
			if latestMetrics.CurrentRatio != nil && *latestMetrics.CurrentRatio > 1.5 {
				score += 1
				reasoning = append(reasoning, "良好的流动性状况")
			} else if latestMetrics.CurrentRatio != nil {
				reasoning = append(reasoning, fmt.Sprintf("流动性较弱，流动比率为%.1f", *latestMetrics.CurrentRatio))
			} else {
				reasoning = append(reasoning, "流动比率数据不可用")
			}

			// 额外检查：价格收益比 (P/E)
			if latestMetrics.PriceToEarningsRatio > 0 && latestMetrics.PriceToEarningsRatio < 25 {
				score += 1
				reasoning = append(reasoning, fmt.Sprintf("合理的P/E比率为%.1f", latestMetrics.PriceToEarningsRatio))
			} else if latestMetrics.PriceToEarningsRatio > 0 {
				reasoning = append(reasoning, fmt.Sprintf("P/E比率较高为%.1f", latestMetrics.PriceToEarningsRatio))
			}

			// 额外检查：价格净值比 (P/B)
			if latestMetrics.PriceToBookRatio > 0 && latestMetrics.PriceToBookRatio < 3 {
				score += 1
				reasoning = append(reasoning, fmt.Sprintf("合理的P/B比率为%.1f", latestMetrics.PriceToBookRatio))
			} else if latestMetrics.PriceToBookRatio > 0 {
				reasoning = append(reasoning, fmt.Sprintf("P/B比率较高为%.1f", latestMetrics.PriceToBookRatio))
			}

			// 创建指标字典
			metricsMap := map[string]any{
				"ticker":           latestMetrics.Ticker,
				"return_on_equity": latestMetrics.ReturnOnEquity,
				"debt_to_equity":   latestMetrics.DebtToEquity,
				"operating_margin": latestMetrics.OperatingMargin,
				"current_ratio":    latestMetrics.CurrentRatio,
				"pe_ratio":         latestMetrics.PriceToEarningsRatio,
				"pb_ratio":         latestMetrics.PriceToBookRatio,
				"market_cap":       latestMetrics.MarketCap,
				"report_period":    latestMetrics.ReportPeriod,
			}

			result := &FundamentalAnalysisResponse{
				Score:   score,
				Details: strings.Join(reasoning, "; "),
				Metrics: metricsMap,
			}

			// 保存分析结果到本地文件
			if err := saveAnalysisToFile(result, latestMetrics.Ticker); err != nil {
				log.Printf("[FundamentalAnalysisTool] 保存文件失败: %v", err)
				// 不返回错误，继续返回分析结果
			}

			log.Printf("[FundamentalAnalysisTool] 分析完成: Score=%d, Ticker=%s, Details=%s", result.Score, latestMetrics.Ticker, result.Details)
			return result, nil
		})
}

// saveAnalysisToFile 将基本面分析结果保存到本地文件
func saveAnalysisToFile(analysisResult *FundamentalAnalysisResponse, ticker string) error {
	// 创建analysis目录
	dirPath := "output/analysis"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名：analysis_AAPL_2025-09-25.json
	timeSuffix := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("analysis_%s_%s.json", ticker, timeSuffix)
	filePath := filepath.Join(dirPath, fileName)

	// 将分析结果转换为JSON
	data, err := json.MarshalIndent(analysisResult, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	log.Printf("[FundamentalAnalysisTool] 分析结果已保存到: %s", filePath)
	return nil
}
