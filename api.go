package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"investment/tools"
)

// FinancialMetricsResponse 结构体
type FinancialMetricsResponse struct {
	FinancialMetrics []tools.FinancialMetrics `json:"financial_metrics"`
}

// CompanyNewsResponse 结构体
type CompanyNewsResponse struct {
	News []tools.CompanyNews `json:"news"`
}

// LineItem 结构体（支持动态字段）
type LineItem struct {
	Ticker       string         `json:"ticker"`
	ReportPeriod string         `json:"report_period"`
	Period       string         `json:"period"`
	Currency     string         `json:"currency"`
	Data         map[string]any `json:"-"` // 额外字段
}

// UnmarshalJSON 实现自定义 JSON 解析以支持动态字段
func (li *LineItem) UnmarshalJSON(data []byte) error {
	type Alias LineItem
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(li),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析所有字段到 map
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 移除已知字段，其余存储到 Data 中
	delete(raw, "ticker")
	delete(raw, "report_period")
	delete(raw, "period")
	delete(raw, "currency")

	li.Data = raw
	return nil
}

// LineItemResponse 结构体
type LineItemResponse struct {
	SearchResults []LineItem `json:"search_results"`
}

// InsiderTrade 结构体
type InsiderTrade struct {
	Ticker                       string   `json:"ticker"`
	Issuer                       *string  `json:"issuer"`
	Name                         *string  `json:"name"`
	Title                        *string  `json:"title"`
	IsBoardDirector              *bool    `json:"is_board_director"`
	TransactionDate              *string  `json:"transaction_date"`
	TransactionShares            *float64 `json:"transaction_shares"`
	TransactionPricePerShare     *float64 `json:"transaction_price_per_share"`
	TransactionValue             *float64 `json:"transaction_value"`
	SharesOwnedBeforeTransaction *float64 `json:"shares_owned_before_transaction"`
	SharesOwnedAfterTransaction  *float64 `json:"shares_owned_after_transaction"`
	SecurityTitle                *string  `json:"security_title"`
	FilingDate                   string   `json:"filing_date"`
}

// InsiderTradeResponse 结构体
type InsiderTradeResponse struct {
	InsiderTrades []InsiderTrade `json:"insider_trades"`
}

// CompanyNews 结构体
type CompanyNews struct {
	Ticker    string  `json:"ticker"`
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	Source    string  `json:"source"`
	Date      string  `json:"date"`
	URL       string  `json:"url"`
	Sentiment *string `json:"sentiment"`
}

// CompanyFacts 结构体
type CompanyFacts struct {
	Ticker                string  `json:"ticker"`
	Name                  string  `json:"name"`
	CIK                   string  `json:"cik"`
	Industry              string  `json:"industry"`
	Sector                string  `json:"sector"`
	Category              string  `json:"category"`
	Exchange              string  `json:"exchange"`
	IsActive              bool    `json:"is_active"`
	ListingDate           string  `json:"listing_date"`
	Location              string  `json:"location"`
	MarketCap             float64 `json:"market_cap"`
	NumberOfEmployees     int     `json:"number_of_employees"`
	SecFilingsURL         string  `json:"sec_filings_url"`
	SicCode               string  `json:"sic_code"`
	SicIndustry           string  `json:"sic_industry"`
	SicSector             string  `json:"sic_sector"`
	WebsiteURL            string  `json:"website_url"`
	WeightedAverageShares int     `json:"weighted_average_shares"`
}

// CompanyFactsResponse 结构体
type CompanyFactsResponse struct {
	CompanyFacts CompanyFacts `json:"company_facts"`
}

var cli *http.Client

func init() {
	cli = &http.Client{Timeout: 30 * time.Second}
}

// makeAPIRequest 执行 API 请求，带有重试和限流处理
func makeAPIRequest(url string, headers map[string]string, method string, jsonData map[string]any, maxRetries int) (*http.Response, error) {

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var req *http.Request
		var err error

		if method == "POST" && jsonData != nil {
			body, err := json.Marshal(jsonData)
			if err != nil {
				return nil, fmt.Errorf("序列化 JSON 数据失败: %w", err)
			}
			req, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
			if err != nil {
				return nil, fmt.Errorf("创建 POST 请求失败: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest("GET", url, nil)
			if err != nil {
				return nil, fmt.Errorf("创建 GET 请求失败: %w", err)
			}
		}

		// 设置请求头
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := cli.Do(req)
		if err != nil {
			return nil, fmt.Errorf("执行 HTTP 请求失败: %w", err)
		}

		if resp.StatusCode == 429 && attempt < maxRetries {
			// 线性退避：60s, 90s, 120s, 150s...
			delay := 60 + (30 * attempt)
			fmt.Printf("接收到限流响应 (429)。尝试 %d/%d。等待 %ds 后重试...\n", attempt+1, maxRetries+1, delay)
			resp.Body.Close()
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("在 %d 次重试后仍然失败", maxRetries)
}

// GetPrices 获取价格数据
func GetPrices(ticker, startDate, endDate string, apiKey ...string) ([]Price, error) {
	// 准备 API 请求
	headers := make(map[string]string)
	financialAPIKey := ""
	if len(apiKey) > 0 && apiKey[0] != "" {
		financialAPIKey = apiKey[0]
	} else {
		financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
	}

	if financialAPIKey != "" {
		headers["X-API-KEY"] = financialAPIKey
	}

	url := fmt.Sprintf("https://api.financialdatasets.ai/prices/?ticker=%s&interval=day&interval_multiplier=1&start_date=%s&end_date=%s",
		ticker, startDate, endDate)

	resp, err := makeAPIRequest(url, headers, "GET", nil, 3)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取数据错误: %s - %d - %s", ticker, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var priceResponse PriceResponse
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		return nil, fmt.Errorf("解析价格响应失败: %w", err)
	}

	if len(priceResponse.Prices) == 0 {
		return []Price{}, nil
	}
	return priceResponse.Prices, nil
}

// GetFinancialMetrics 获取财务指标数据
func GetFinancialMetrics(ticker, endDate string, period string, limit int, apiKey ...string) ([]tools.FinancialMetrics, error) {
	if period == "" {
		period = "ttm"
	}
	if limit == 0 {
		limit = 10
	}

	// 准备 API 请求
	headers := make(map[string]string)
	financialAPIKey := ""
	if len(apiKey) > 0 && apiKey[0] != "" {
		financialAPIKey = apiKey[0]
	} else {
		financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
	}

	if financialAPIKey != "" {
		headers["X-API-KEY"] = financialAPIKey
	}

	url := fmt.Sprintf("https://api.financialdatasets.ai/financial-metrics/?ticker=%s&report_period_lte=%s&limit=%d&period=%s",
		ticker, endDate, limit, period)

	resp, err := makeAPIRequest(url, headers, "GET", nil, 3)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取数据错误: %s - %d - %s", ticker, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var metricsResponse FinancialMetricsResponse
	if err := json.Unmarshal(body, &metricsResponse); err != nil {
		return nil, fmt.Errorf("解析财务指标响应失败: %w", err)
	}

	if len(metricsResponse.FinancialMetrics) == 0 {
		return []tools.FinancialMetrics{}, nil
	}

	return metricsResponse.FinancialMetrics, nil
}

// SearchLineItems 搜索行项目数据
func SearchLineItems(ticker string, lineItems []string, endDate, period string, limit int, apiKey ...string) ([]LineItem, error) {
	if period == "" {
		period = "ttm"
	}
	if limit == 0 {
		limit = 10
	}

	// 准备 API 请求
	headers := make(map[string]string)
	financialAPIKey := ""
	if len(apiKey) > 0 && apiKey[0] != "" {
		financialAPIKey = apiKey[0]
	} else {
		financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
	}

	if financialAPIKey != "" {
		headers["X-API-KEY"] = financialAPIKey
	}

	url := "https://api.financialdatasets.ai/financials/search/line-items"

	body := map[string]any{
		"tickers":    []string{ticker},
		"line_items": lineItems,
		"end_date":   endDate,
		"period":     period,
		"limit":      limit,
	}

	resp, err := makeAPIRequest(url, headers, "POST", body, 3)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取数据错误: %s - %d - %s", ticker, resp.StatusCode, string(responseBody))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var lineItemResponse LineItemResponse
	if err := json.Unmarshal(responseBody, &lineItemResponse); err != nil {
		return nil, fmt.Errorf("解析行项目响应失败: %w", err)
	}

	if len(lineItemResponse.SearchResults) == 0 {
		return []LineItem{}, nil
	}

	// 限制结果数量
	if len(lineItemResponse.SearchResults) > limit {
		return lineItemResponse.SearchResults[:limit], nil
	}

	return lineItemResponse.SearchResults, nil
}

// GetInsiderTrades 获取内部交易数据
func GetInsiderTrades(ticker, endDate string, startDate *string, limit int, apiKey ...string) ([]InsiderTrade, error) {
	if limit == 0 {
		limit = 1000
	}

	// 创建缓存键
	startDateStr := "none"
	if startDate == nil {
		startDateStr = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}

	// 准备 API 请求
	headers := make(map[string]string)
	financialAPIKey := ""
	if len(apiKey) > 0 && apiKey[0] != "" {
		financialAPIKey = apiKey[0]
	} else {
		financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
	}

	if financialAPIKey != "" {
		headers["X-API-KEY"] = financialAPIKey
	}

	var allTrades []InsiderTrade
	currentEndDate := endDate

	for {
		url := fmt.Sprintf("https://api.financialdatasets.ai/insider-trades/?ticker=%s&filing_date_lte=%s", ticker, currentEndDate)
		if startDate != nil {
			url += fmt.Sprintf("&filing_date_gte=%s", startDateStr)
		}
		url += fmt.Sprintf("&limit=%d", limit)

		resp, err := makeAPIRequest(url, headers, "GET", nil, 3)
		if err != nil {
			return nil, fmt.Errorf("API 请求失败: %w", err)
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("获取数据错误: %s - %d - %s", ticker, resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}

		var tradeResponse InsiderTradeResponse
		if err := json.Unmarshal(body, &tradeResponse); err != nil {
			return nil, fmt.Errorf("解析内部交易响应失败: %w", err)
		}

		if len(tradeResponse.InsiderTrades) == 0 {
			break
		}

		allTrades = append(allTrades, tradeResponse.InsiderTrades...)

		// 只有在设置了开始日期且获得了完整页面时才继续分页
		if startDate == nil || len(tradeResponse.InsiderTrades) < limit {
			break
		}

		// 更新下一次迭代的结束日期
		minDate := tradeResponse.InsiderTrades[0].FilingDate
		for _, trade := range tradeResponse.InsiderTrades {
			if trade.FilingDate < minDate {
				minDate = trade.FilingDate
			}
		}

		// 提取日期部分（去除时间）
		if strings.Contains(minDate, "T") {
			minDate = strings.Split(minDate, "T")[0]
		}
		currentEndDate = minDate

		// 如果已达到或超过开始日期，停止
		if startDate != nil && currentEndDate <= *startDate {
			break
		}
	}

	if len(allTrades) == 0 {
		return []InsiderTrade{}, nil
	}
	return allTrades, nil
}

// GetCompanyNews 获取公司新闻数据
func GetCompanyNews(ticker, endDate string, startDate *string, limit int, apiKey ...string) ([]tools.CompanyNews, error) {
	if limit == 0 {
		limit = 1000
	}

	startDateStr := "none"
	if startDate == nil {
		startDateStr = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	}

	// 准备 API 请求
	headers := make(map[string]string)
	financialAPIKey := ""
	if len(apiKey) > 0 && apiKey[0] != "" {
		financialAPIKey = apiKey[0]
	} else {
		financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
	}

	if financialAPIKey != "" {
		headers["X-API-KEY"] = financialAPIKey
	}

	var allNews []tools.CompanyNews
	currentEndDate := endDate

	for {
		url := fmt.Sprintf("https://api.financialdatasets.ai/news/?ticker=%s&end_date=%s", ticker, currentEndDate)
		if startDate != nil {
			url += fmt.Sprintf("&start_date=%s", startDateStr)
		}
		url += fmt.Sprintf("&limit=%d", limit)

		resp, err := makeAPIRequest(url, headers, "GET", nil, 3)
		if err != nil {
			return nil, fmt.Errorf("API 请求失败: %w", err)
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("获取数据错误: %s - %d - %s", ticker, resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}

		var newsResponse CompanyNewsResponse
		if err := json.Unmarshal(body, &newsResponse); err != nil {
			return nil, fmt.Errorf("解析公司新闻响应失败: %w", err)
		}

		if len(newsResponse.News) == 0 {
			break
		}

		allNews = append(allNews, newsResponse.News...)

		// 只有在设置了开始日期且获得了完整页面时才继续分页
		if startDate == nil || len(newsResponse.News) < limit {
			break
		}

		// 更新下一次迭代的结束日期
		minDate := newsResponse.News[0].DateTime
		for _, news := range newsResponse.News {
			if news.DateTime < minDate {
				minDate = news.DateTime
			}
		}

		// 提取日期部分（去除时间）
		if strings.Contains(minDate, "T") {
			minDate = strings.Split(minDate, "T")[0]
		}
		currentEndDate = minDate

		// 如果已达到或超过开始日期，停止
		if startDate != nil && currentEndDate <= *startDate {
			break
		}
	}

	if len(allNews) == 0 {
		return []tools.CompanyNews{}, nil
	}
	return allNews, nil
}

// GetMarketCap 获取市值数据
func GetMarketCap(ticker, endDate string, apiKey ...string) (float64, error) {
	// 检查是否是今天
	today := time.Now().Format("2006-01-02")
	if endDate == today {
		// 从公司事实 API 获取市值
		headers := make(map[string]string)
		financialAPIKey := ""
		if len(apiKey) > 0 && apiKey[0] != "" {
			financialAPIKey = apiKey[0]
		} else {
			financialAPIKey = os.Getenv("FINANCIAL_DATASETS_API_KEY")
		}

		if financialAPIKey != "" {
			headers["X-API-KEY"] = financialAPIKey
		}

		url := fmt.Sprintf("https://api.financialdatasets.ai/company/facts/?ticker=%s", ticker)
		resp, err := makeAPIRequest(url, headers, "GET", nil, 3)
		if err != nil {
			return 0, fmt.Errorf("API 请求失败: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("获取公司事实错误: %s - %d\n", ticker, resp.StatusCode)
			return 0, nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("读取响应体失败: %w", err)
		}

		var factsResponse CompanyFactsResponse
		if err := json.Unmarshal(body, &factsResponse); err != nil {
			return 0, fmt.Errorf("解析公司事实响应失败: %w", err)
		}

		return factsResponse.CompanyFacts.MarketCap, nil
	}

	// 从财务指标获取市值
	financialMetrics, err := GetFinancialMetrics(ticker, endDate, "ttm", 10, apiKey...)
	if err != nil {
		return 0, err
	}

	if len(financialMetrics) == 0 {
		return 0, nil
	}

	return financialMetrics[0].MarketCap, nil
}

// PriceDataFrame 表示价格数据框架
type PriceDataFrame struct {
	Dates  []time.Time
	Open   []float64
	Close  []float64
	High   []float64
	Low    []float64
	Volume []int64
}

// PricesToDataFrame 将价格转换为数据框架
func PricesToDataFrame(prices []Price) (*PriceDataFrame, error) {
	if len(prices) == 0 {
		return &PriceDataFrame{}, nil
	}

	df := &PriceDataFrame{
		Dates:  make([]time.Time, len(prices)),
		Open:   make([]float64, len(prices)),
		Close:  make([]float64, len(prices)),
		High:   make([]float64, len(prices)),
		Low:    make([]float64, len(prices)),
		Volume: make([]int64, len(prices)),
	}

	for i, price := range prices {
		// 解析时间
		date, err := time.Parse(time.RFC3339, price.Time)
		if err != nil {
			// 尝试其他时间格式
			date, err = time.Parse("2006-01-02", price.Time)
			if err != nil {
				return nil, fmt.Errorf("解析时间失败: %s, %w", price.Time, err)
			}
		}

		df.Dates[i] = date
		df.Open[i] = price.Open
		df.Close[i] = price.Close
		df.High[i] = price.High
		df.Low[i] = price.Low
		df.Volume[i] = price.Volume
	}

	// 按日期排序
	type sortData struct {
		date   time.Time
		open   float64
		close  float64
		high   float64
		low    float64
		volume int64
	}

	data := make([]sortData, len(prices))
	for i := range data {
		data[i] = sortData{
			date:   df.Dates[i],
			open:   df.Open[i],
			close:  df.Close[i],
			high:   df.High[i],
			low:    df.Low[i],
			volume: df.Volume[i],
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].date.Before(data[j].date)
	})

	for i, d := range data {
		df.Dates[i] = d.date
		df.Open[i] = d.open
		df.Close[i] = d.close
		df.High[i] = d.high
		df.Low[i] = d.low
		df.Volume[i] = d.volume
	}

	return df, nil
}

// GetPriceData 获取价格数据并转换为数据框架
func GetPriceData(ticker, startDate, endDate string, apiKey ...string) (*PriceDataFrame, error) {
	prices, err := GetPrices(ticker, startDate, endDate, apiKey...)
	if err != nil {
		return nil, err
	}
	return PricesToDataFrame(prices)
}
