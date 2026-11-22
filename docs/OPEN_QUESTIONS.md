# 未決事項（OPEN_QUESTIONS）

依憲章 v2 第 8、9 節：專家在回報單只寫編號；幕僚長在關卡時把這裡的事項合併、去重後，一次問老闆。

| 編號 | 提出者 | 日期 | 標題 | 待決對象 | 何時問 | 狀態 |
|---|---|---|---|---|---|---|
| OQ-001 | TechLead | 2026-11-22 | `.golangci.yml` 是否在 `test/**` 排除 gosec G702 | Architect 審查 → 老闆 | G3（PR #1） | 已結案（2026-11-22） |

---

## OQ-001：`.golangci.yml` 是否在 `test/**` 排除 gosec G702

> 名詞說明：**gosec** 是 golangci-lint 內含的安全檢查工具。**G702** 是 golangci-lint v2.15 版 gosec 新增的規則，會追蹤「不可信的輸入」（例如 `os.Args`）有沒有流進執行外部指令的程式碼，以此判斷「指令注入」風險。

**背景**
- T-001 的 `.golangci.yml` 只在 `internal/gitlog/`、`internal/testutil/`、`test/` 與 `_test.go` 排除 **G204**（用變數執行子程序）。這是依 ARCHITECTURE §11 的設計：產品程式只有 gitlog 會執行 git。
- T-001 Review 時，TechLead 用 golangci-lint v2.15.0 實測：`test/e2e/fakegit` 若把 `os.Args` 當成指令參數去執行，會被 **G702** 報錯。G702 沒有排除規則，ARCHITECTURE §11 寫的時候這條規則還不存在。
- 目前 repo 還沒有 fakegit（T-013 起才會加），所以 `golangci-lint run` 現在是 `0 issues.`，**不影響 PR #1 合併**。等 fakegit 或 testutil 真的用外部輸入執行指令時才會擋住開發。
- ARCHITECTURE §11 規定：改 lint 規則前需要 Architect 審查。

**選項**
1. **只在 `test/**` 排除 G702**（`exclusions.rules` 加 `linters: [gosec]`、`text: 'G702'`、`path: '^test/'`）。產品程式（含 `internal/gitlog`）與 `internal/testutil` 仍然受 G702 檢查。
2. 比照 G204，在 `internal/gitlog/`、`internal/testutil/`、`test/`、`_test.go` 全部排除。比較省事，但產品程式 gitlog 少了一層指令注入檢查。
3. 不排除。遇到 G702 時逐行加 `//nolint:gosec // G702: <理由>`，並在報告中列出。比較繁瑣，Review 的負擔也比較重。

**TechLead 建議：選項 1**
- 理由和現有的 G204 測試排除一樣：fakegit 只在測試中使用，本來就是故意要執行指令或模擬 git。
- 產品程式繼續保留 G702，符合 ARCHITECTURE §9 對安全的要求。
- 做法：Architect 同意後，由 Architect 更新 ARCHITECTURE §11 的規則表，TechLead 再開任務單交給 Claude Code 修改 `.golangci.yml`。不在 T-001 範圍內改。

**決定（2026-11-22，老闆，已結案）**：採用選項 1。在 `.golangci.yml` 只對 `test/**` 排除 gosec G702，產品程式（含 `internal/gitlog`）與 `internal/testutil` 仍保留 G702 檢查。實際修改依上方做法：Architect 先更新 ARCHITECTURE §11，再放進建立 `test/**` 程式碼的任務單，由 Claude Code 修改。
