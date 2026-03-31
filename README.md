# Scout CLI

A terminal-based implementation of **Scout** (aka *SCOUT!*), the award-winning card game designed by Kei Kajino.

Play against AI opponents right in your terminal with a colorful TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Screenshots

| Menu | In-Game |
|------|---------|
| ![Menu](docs/enter_img.png) | ![Game](docs/play_img.png) |

## Install

**Quick install** (macOS / Linux):

```bash
curl -sL https://raw.githubusercontent.com/daghlny/scout_cli/main/install.sh | sh
```

**With Go**:

```bash
go install github.com/daghlny/scout_cli@latest
```

**Manual download**: grab the binary for your platform from [GitHub Releases](https://github.com/daghlny/scout_cli/releases).

Then run:

```bash
scout_cli
```

## Features

- 3-5 player games (you vs AI bots)
- Full Scout rules: double-sided cards, Show / Scout / Scout & Show actions
- Smart AI with tactical awareness and defensive play
- **LLM AI mode**: powered by DeepSeek, with function calling and chain-of-thought reasoning
- Colorful card rendering in terminal
- Multi-round scoring with final leaderboard

## AI Modes

**Default (Smart AI)**:

```bash
scout_cli
```

**LLM AI** — each bot calls a large language model to decide its move:

```bash
export DEEPSEEK_API_KEY=your_key
scout_cli -ai=llm
```

The LLM receives game rules, current hand, table state, and action history, then reasons through function calling. Falls back to Smart AI automatically on API errors.

## Game Rules

Scout uses 45 unique double-sided cards (each with two different numbers 1-10). Cards in your hand **cannot be rearranged** — you can only insert new cards via scouting. On your turn, choose one action:

- **Show**: Play adjacent cards from your hand that beat the table combo
- **Scout**: Take a card from either end of the table combo and insert it into your hand
- **Scout & Show**: Do both (once per round)

For full rules, see: https://boardgamegeek.com/boardgame/291453/scout

---

# Scout CLI (中文)

**Scout**（马戏星探）的终端命令行版本，基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建。

## 游戏截图

| 主菜单 | 游戏进行中 |
|-------|----------|
| ![主菜单](docs/enter_img.png) | ![游戏进行中](docs/play_img.png) |

## 安装

**一键安装**（macOS / Linux，无需 Go 环境）：

```bash
curl -sL https://raw.githubusercontent.com/daghlny/scout_cli/main/install.sh | sh
```

**有 Go 环境**：

```bash
go install github.com/daghlny/scout_cli@latest
```

**手动下载**：从 [GitHub Releases](https://github.com/daghlny/scout_cli/releases) 下载对应平台的二进制文件（支持 macOS / Linux / Windows）。

运行：

```bash
scout_cli
```

## 功能

- 支持 3-5 人游戏（你 vs AI 对手）
- 完整的 Scout 规则：双面牌、出牌 / 招募 / 双重行动
- 具备攻防意识的智能 AI，懂得保留强牌与战术防守
- **LLM AI 模式**：由 DeepSeek 大模型驱动，通过 function calling 和思维链推理做出决策
- 终端内彩色扑克牌渲染
- 多轮计分与最终排行榜

## AI 模式

**默认（Smart AI）**：

```bash
scout_cli
```

**LLM AI** — 每个 bot 调用大语言模型来决策：

```bash
export DEEPSEEK_API_KEY=your_key
scout_cli -ai=llm
```

LLM 会接收游戏规则、当前手牌、桌面状态和历史行动记录，通过 function calling 进行结构化推理。API 出错时自动回退到 Smart AI。

## 游戏规则

Scout 使用 45 张双面牌（每张牌正反两面各有一个 1-10 的数字）。手牌**不可重新排列**，只能通过招募插入新牌。每回合选择一个行动：

- **出牌（Show）**：打出手中相邻的牌组合，必须强过桌面上的牌
- **招募（Scout）**：从桌面牌组的左端或右端取一张牌插入手中
- **双重行动（Scout & Show）**：先招募再出牌（每轮限一次）

完整规则参见：https://boardgamegeek.com/boardgame/291453/scout
