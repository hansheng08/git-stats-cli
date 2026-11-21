# ADR-001：語言與技術棧

- 狀態：已接受（Accepted，老闆於 2026-11-22 00:44 G2 核准）
- 日期：2026-11-22
- 作者：Architect（交接單 ARCH-001）

## 背景
git-stats-cli 是單一指令、跨平台（Windows／macOS／Linux）的 CLI（PRD 4.1、NFR-03）。老闆已定：發佈＝GitHub tag＋GitHub Release，**不上 PyPI／npm**；安裝要「單一指令」（AC-37）。團隊由 Coding Agent（Claude Code）實作，因此語言要讓錯誤盡量在編譯期或靜態檢查被抓到，且建置／測試／Lint 要快。

## 選項比較
評分：◎ 很好／○ 可以／△ 有缺點／✕ 不符合

| 標準 | Go | Rust | Python | TypeScript（Node.js） |
|---|---|---|---|---|
| 1. 型別嚴格 | ○ 靜態強型別；無泛型以外的進階型別，nil 需靠 lint 補 | ◎ 最嚴格（所有權、無 null） | △ 型別提示需 mypy／pyright，執行期仍可繞過 | ○ strict 模式不錯，但 `any`、執行期無型別 |
| 2. 主流穩定、AI 熟悉度 | ◎ CLI 工具主流（gh、docker CLI、kubectl），標準庫穩定 | ○ 主流，但 AI 常在生命週期／借用檢查上反覆修正 | ◎ | ◎ |
| 3. 回饋迴圈速度 | ◎ 編譯數秒、`go test` 內建 | △ 編譯與 clippy 較慢（數十秒以上，未驗證） | ○ 無編譯，型別檢查快 | ○ tsc＋測試框架，尚可 |
| 4. 讓錯誤寫不出來 | ○ package 邊界＋`internal/` 由編譯器強制；depguard／forbidigo；archtest 可用標準庫寫 | ◎ 型別系統最強；crate 邊界清楚 | △ 邊界只能靠 import-linter 等外部工具 | ○ eslint 規則豐富，但要多個套件 |
| 5. 依賴最少 | ◎ 標準庫涵蓋 exec、csv、測試、fuzz；只需 2 個小套件 | ○ 需 clap、csv、unicode-width 等，遞移依賴多 | ○ 標準庫夠用，但打包工具是大依賴 | △ 測試、參數解析、打包都要套件，node_modules 龐大 |
| 6. 單一執行檔、跨平台、不靠 PyPI／npm | ◎ `CGO_ENABLED=0` 交叉編譯一行出全平台靜態執行檔 | ○ 可，但 macOS／Windows 交叉編譯需額外工具鏈 | ✕ 需使用者先裝 Python，或用 PyInstaller（檔案大、啟動慢、易被防毒誤判） | ✕ 需 Node，或用 bun/pkg/SEA 打包（檔案大、仍不成熟） |
| 啟動時間 | ◎ 毫秒級（未驗證） | ◎ | △ 百毫秒級 | △ 百毫秒級 |

## 決定
採用 **Go**（Go 官方支援中的最新穩定版；確切版號不確定，建立 go.mod 時以 go.dev 為準並固定在 `go.mod` 的 `go`／`toolchain` 指令）。

執行期依賴（每個都要理由）：
| 套件 | 理由 | 不用它的代價 |
|---|---|---|
| `github.com/spf13/pflag` | 支援旗標與位置參數穿插（`git-stats <repo> --since …`）、`--x=y`、`--`；無遞移依賴 | 標準庫 `flag` 遇到第一個非旗標參數就停，需要自己寫參數重排，容易出錯 |
| `github.com/mattn/go-runewidth` | AC-21 依 Unicode East Asian Width 計算顯示寬度 | 自己維護 Unicode 寬度表，易錯 |

不採用：cobra（子指令框架，本工具只有一個指令，依賴過多）、testify（標準庫 `testing` 足夠，少一個依賴也少一種寫法）、任何 Git 函式庫（見 ADR-002）。

開發工具（不進執行期）：golangci-lint（v2 系列，版號由 SRE 固定）、GoReleaser（ADR-003）。

## 後果
- 好處：單一靜態執行檔直接滿足「單一指令安裝、不上 PyPI／npm」；`go build`／`go test`／`go vet` 都是內建指令，Claude Code 每次回饋在數十秒內（未驗證）；`internal/` 由編譯器強制禁止外部 import。
- 壞處：型別嚴格度不如 Rust（nil、未處理錯誤），用 errcheck、staticcheck、nilness 等 lint 補強（ADR-004）；Go 的錯誤處理較冗長。
- 新增任何依賴都要修改本 ADR 與 ARCHITECTURE.md 第 2 節，並通過 archtest 的 go.mod 白名單。
