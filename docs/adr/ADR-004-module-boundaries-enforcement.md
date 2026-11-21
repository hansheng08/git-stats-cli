# ADR-004：模組邊界與自動強制機制

- 狀態：已接受（Accepted，老闆於 2026-11-22 00:44 G2 核准）
- 日期：2026-11-22
- 作者：Architect（交接單 ARCH-001）

## 背景
專案由 Coding Agent 持續迭代。規範若只寫在文件，遲早會被違反（例如在統計邏輯裡直接 `fmt.Println`、在 render 裡呼叫 git、任意加依賴）。目標是讓違規在 `go test` 或 `golangci-lint run` 時立即失敗。

## 選項比較
| 標準 | A. 只寫在 CLAUDE.md | B. golangci-lint（depguard＋forbidigo） | C. 自寫架構測試（標準庫 `go/parser`） | D. B＋C 並用 |
|---|---|---|---|---|
| 違規會失敗 | ✕ | ◎ | ◎ | ◎ |
| 本機不裝 lint 也能抓到 | ✕ | ✕ | ◎（`go test` 就會跑） | ◎ |
| 能表達「A 不可 import B」 | — | ○ 依路徑設定 | ◎ 白名單最直接 | ◎ |
| 能抓函式呼叫（`os.Exit`、`time.Now`、`fmt.Print`） | — | ◎ forbidigo | △ 要多寫 AST 檢查 | ◎ |
| 額外依賴 | 0 | 0（開發工具） | 0 | 0 |

## 決定
採用 **D**：
- `internal/archtest`（只含 `_test.go`）：讀取所有非測試 Go 檔的 import，依 ARCHITECTURE.md 第 5.1 節白名單比對；檢查 go.mod 依賴白名單；檢查 `internal/testutil` 只被測試 import；掃描 E2E 測試名稱確認每條 Must AC 都有測試。
- golangci-lint：depguard（package 邊界、禁止網路套件、只有 gitlog 可用 `os/exec`）、forbidigo（`os.Exit`、`log.Fatal`、`panic`、`fmt.Print*`、`os.Stdout`、`time.Now`、`time.Local` 的使用範圍）、errcheck、govet、staticcheck、errorlint、gosec、gofmt／goimports。
- 型別防呆：`model.Format`、`model.DateRange` 不匯出欄位，只能經建構函式建立。
- 白名單的唯一來源是 ARCHITECTURE.md 第 5 節；修改邊界＝修改 ARCHITECTURE.md（需 Architecture Review）＋同步修改 archtest 與 `.golangci.yml`。

## 後果
- 好處：Claude Code 寫錯會在本機幾秒內失敗；Review 時不用靠人眼檢查邊界。
- 壞處：新增 package 或依賴時要同步改三處（文件、archtest、lint 設定）；CLAUDE.md 明定此流程並禁止為了通過而放寬規則（`//nolint` 需附理由並經 Review）。
