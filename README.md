# Investment Assistant for American Stock Market

基于 Eino 框架的智能股票投资分析助理，使用 **Eino框架 + 多模型支持 + React Agent架构** 实现，提供专业的股票投资分析和自动报告生成服务。

## 功能特性

- 🤖 **React Agent智能分析** - 基于多步推理的智能投资决策流程
- 🎯 **多模型切换支持** - 支持Gemini、OpenAI和DeepSeek模型
- 📈 **全面财务数据** - 市值、财务指标、公司新闻、基本面分析  
- 📊 **自动报告生成** - 分析结果自动保存为markdown格式报告
- 🔍 **巴菲特性投资分析** - 遵循价值投资理念的分析框架
- 🌐 **外部数据集成** - 集成FinancialDatasets.ai金融数据源
- 💬 **中文交互** - 全中文界面和报告输出

## 使用方法

### 环境变量设置

编辑配置文件

```bash
cp .env.example .env
```

```text
# 必需：设置AI模型API密钥以启用智能分析
MODEL_TYPE="deepseek"

# 使用Gemini模型
GEMINI_API_KEY="xxx"
GEMINI_MODEL_NAME="gemini-2.5-pro"

# 使用OpenAI模型
OPENAI_API_KEY="xxx"
OPENAI_MODEL_NAME="gpt-4o"
OPENAI_BASE_URL=""

# 默认使用DeepSeek模型
DEEPSEEK_API_KEY=""
DEEPSEEK_MODEL_NAME="deepseek-reasoner"

# 可选：设置FinancialDatasets.ai API密钥获取更丰富的金融数据
FINANCIAL_DATASETS_API_KEY="your-api-key"
```

### 编译
```bash
go build -o investment .
```

### 运行
```bash
# 分析苹果股票
./investment AAPL

# 分析特斯拉股票
./investment TSLA

# 分析谷歌股票
./investment GOOG

# 分析微软股票
./investment MSFT
```

## React Agent分析流程

应用使用React Agent架构，按照以下标准化流程进行分析：

1. **🔍 市值数据收集** - 获取基础市值信息
2. **📊 财务指标分析** - ROE、利润率、债务率等关键指标
3. **📰 市场动态评估** - 最新新闻和业务动态
4. **💎 基本面评分** - 巴菲特性价值投资评分
5. **📋 投资建议生成** - 综合评估并给出评级

## 输出文件

分析完成后会自动生成markdown格式的详细报告：

```
output/report/
├── AAPL_report.md
├── TSLA_report.md
└── GOOG_report.md
```

报告包含完整的分析过程、财务数据、投资评级、目标价格和风险提示。

## 支持股票

支持主流上市公司股票，包括但不限于：
- **AAPL**: 苹果公司
- **TSLA**: 特斯拉公司
- **GOOG**: 谷歌公司（Alphabet）
- **MSFT**: 微软公司
- **AMZN**: 亚马逊公司
- **其他** 主流上市公司

## 输出示例

```
=== 智能投资助手 - 股票分析：AAPL ===

🔧 创建投资分析工具集...
🤖 启动 React Agent 进行智能分析...
📈 Agent 将自动收集数据、进行分析并生成报告

(React Agent分析过程日志...)

==================================================
📊 投资分析报告
==================================================
【第一步】市值数据收集：获取AAPL当前市值信息...
【第二步】财务指标分析：分析过去5年财务趋势...
【第三步】市场动态评估：查看公司最新动态...
【第四步】基本面评分：进行巴菲特式投资评估...

📋 投资建议：
- 投资评级：推荐
- 目标价位：$180-200
- 主要优势：强劲现金流、高ROE
- 风险提示：市场竞争、估值偏高
==================================================
✅ 分析完成
📄 报告已保存为 markdown 文件: output/report/AAPL_report.md
```

## 技术实现

### 核心架构

- **React Agent架构**: 基于Eino的React Agent多步推理框架
- **Gemini模型集成**: Google Gemini 2.5 Pro模型进行智能分析（其他模型接入中）
- **工具链设计**: 4个专业投资分析工具协同工作
- **自动报告生成**: markdown格式报告自动保存到output目录

### 分析工具

1. **市值查询工具** - 获取公司市值和基本信息
2. **财务指标工具** - 分析ROE、利润率、债务率等关键指标
3. **公司新闻工具** - 获取市场动态和业务新闻
4. **基本面分析工具** - 巴菲特式价值投资评分系统

### 框架特性

- **多步推理**: React Agent自动规划分析步骤和执行
- **数据驱动**: 所有结论基于真实财务数据和市场信息
- **价值投资**: 遵循巴菲特投资理念的分析框架
- **中文优化**: 专门优化的中文提示词和报告输出
- **错误处理**: 优雅的降级机制和错误恢复

## 扩展功能

可扩展的功能包括：
- 集成更多金融数据源（Alpha Vantage、Yahoo Finance等）
- 添加技术分析指标（MACD、RSI等技术指标）
- 支持投资组合分析和比较
- 添加PDF报告导出功能
- 集成更多AI模型（Claude、GPT等）
- 添加Web界面和API服务

## 报告样例

[报告样例](./report_sample.md)
