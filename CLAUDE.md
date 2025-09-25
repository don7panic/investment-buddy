# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is an investment assistant application built with the Eino framework in Go, providing intelligent stock analysis using AI agents and external financial APIs. The application uses a React Agent architecture to perform multi-step investment analysis on publicly traded companies.

## Project Structure

- `main.go` - Entry point, orchestrates the React Agent and tools
- `api.go` - Financial API client for FinancialDatasets.ai services
- `gemini.go` - Google Gemini AI model configuration
- `types.go` - Basic data structures for price data
- `tools/` - Investment analysis tools implementing the tool interface
  - `market_cap_tool.go` - Market capitalization queries
  - `financial_metrics_tool.go` - Comprehensive financial metrics
  - `company_news_tool.go` - Company news and sentiment analysis  
  - `fundamental_analysis_tool.go` - Buffett-style fundamental scoring

## Dependencies

- **Eino Framework** (github.com/cloudwego/eino) - Core AI orchestration framework
- **Eino Gemini Extension** (github.com/cloudwego/eino-ext/components/model/gemini) - Google Gemini model integration
- **Google GenAI** (google.golang.org/genai) - Google's AI client library

## Development Commands

### Building
```bash
go build -o investment .
```

### Running
```bash
# Analyze Apple stock
./investment AAPL

# Analyze Tesla stock  
./investment TSLA

# Analyze Google stock
./investment GOOG
```

### Testing
Currently no test files exist. To add tests:
```bash
# Create test files following Go conventions
go test ./...
```

## Configuration

### Environment Variables
```bash
# Required for AI functionality
export GEMINI_API_KEY="your-gemini-api-key"
export GEMINI_MODEL_NAME="gemini-2.5-pro"

# Optional for enhanced data (FinancialDatasets.ai)
export FINANCIAL_DATASETS_API_KEY="your-api-key"
```

## Architecture

### React Agent Pattern
The application follows the React Agent architecture from the Eino framework:

1. **Agent Initialization** - Creates a React Agent with investment analysis tools
2. **Tool Chain** - Four specialized tools for investment analysis
3. **Multi-step Reasoning** - The agent plans analysis steps and executes them sequentially
4. **Tool Integration** - Each tool integrates with external financial APIs

### Tool Descriptions

#### 1. Market Cap Tool (`get_market_cap`)
- Fetches company market capitalization
- Uses FinancialDatasets.ai company facts API for real-time data
- Falls back to financial metrics API for historical data

#### 2. Financial Metrics Tool (`get_financial_metrics`)
- Retrieves comprehensive financial indicators:
  - Valuation ratios (P/E, P/B, Enterprise Value multiples)
  - Profitability metrics (margins, ROE, ROA, ROIC)
  - Efficiency ratios (asset turnover, inventory turnover)
  - Financial health (liquidity ratios, debt metrics)
  - Growth indicators

#### 3. Company News Tool (`get_company_news`)
- Fetches recent company news articles
- Includes sentiment analysis for market context
- Provides insights into current events and market sentiment

#### 4. Fundamental Analysis Tool (`analyze_fundamentals`)
- Implements Buffett-style investment criteria
- Scores companies based on key fundamentals:
  - ROE > 15% (2 points)
  - Debt-to-Equity < 0.5 (2 points) 
  - Operating Margin > 15% (2 points)
  - Current Ratio > 1.5 (1 point)
  - P/E ratio < 25 (1 point)
  - P/B ratio < 3 (1 point)

### API Integration

The application integrates with FinancialDatasets.ai API providing:
- Real-time and historical financial data
- Company fundamentals and market data
- News and sentiment analysis
- Comprehensive financial metrics

Key features:
- Retry logic with exponential backoff for rate limiting
- Automatic pagination for large datasets
- Structured error handling with graceful degradation

## Code Style & Conventions

- **Go Standard**: Follows standard Go idioms and error handling patterns
- **Chinese Comments**: Code contains detailed Chinese language comments
- **Logging**: Structured logging with tool-specific prefixes
- **Error Handling**: Graceful error handling with informative messages
- **API Integration**: Clean separation between business logic and external APIs

## Extending the Application

### Adding New Tools
1. Create new tool file in `tools/` directory
2. Implement tool interface using `utils.InferTool`
3. Update `main.go` to include the new tool in the React Agent configuration
4. Modify the system prompt to describe the new tool's capabilities

### Adding New Data Sources
1. Extend `api.go` with new API client methods
2. Create corresponding tool implementations
3. Ensure proper error handling and rate limiting

### Modifying Analysis Logic
- Update tool implementations in the `tools/` directory
- Modify fundamental analysis scoring in `fundamental_analysis_tool.go`
- Update system prompt in `main.go` to reflect new analysis strategies