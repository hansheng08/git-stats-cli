# 未決事項（OPEN_QUESTIONS）

依憲章 v2 第 8、9 節：專家在回報單只寫編號；幕僚長在關卡時把這裡的事項合併、去重後，一次問老闆。

| 編號 | 提出者 | 日期 | 標題 | 待決對象 | 何時問 | 狀態 |
|---|---|---|---|---|---|---|
| OQ-001 | TechLead | 2026-11-22 | `.golangci.yml` 是否在 `test/**` 排除 gosec G702 | Architect 審查 → 老闆 | G3（PR #1） | 已結案（2026-11-22） |
| OQ-002 | TechLead | 2026-11-22 | 凍結文件沒有定義 `apperr` API，由 TechLead 補上 | 流程檢討 | 流程檢討 | 已處置（幕僚長決定採用）；留待流程檢討 |
| OQ-003 | TechLead | 2026-11-22 | ARCHITECTURE 說 archtest 會檢查「只有 daterange 可以呼叫 `NewDateRange`」，但 T-001 的 archtest 沒有做 | 流程檢討（若繼續開發：Architect） | 流程檢討 | 待決；留待流程檢討 |

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

---

## OQ-002：凍結文件沒有定義 `apperr` API，由 TechLead 補上

**背景**
- ARCHITECTURE v1.1 §5.3 列出 `model` 等 package 的簽名，但沒有列 `apperr`。§8 和 CLAUDE.md 只寫「用 `apperr.Usage(...)`／`apperr.Runtime(...)` 包裝；未包裝的 error 視為 1」，沒有寫參數型別，也沒有寫 `cli` 要怎麼判斷錯誤種類。
- §5.3 規定「簽名為設計約定，實作時名稱可微調但需 Architecture Review」。
- 為了讓 T-002 能派工，TechLead 在 `docs/tasks/T-002.md` 補上最小 API：`Usage(err error) error`、`Runtime(err error) error`、`IsUsage(err error) bool`。
- **幕僚長在 2026-11-22 決定直接採用，不經 Architect 審查。** T-002 已依此實作（`task/T-002`，`b25654d`）。
- 實作的語意：`IsUsage` 會沿著包裝鏈往內找，所以 `Runtime(Usage(err))` 也回傳 true。任務單的文字允許這種解讀，但 ARCHITECTURE 沒有規定外層和內層種類衝突時以哪一個為準。

**選項（供流程檢討）**
1. 以後 Architect 在 G2 前，對每個 package 都寫出完整的匯出簽名，包括錯誤分類 API。
2. 允許 TechLead 在任務單補最小 API，但要在 OPEN_QUESTIONS 留紀錄，並在下一個關卡一次補進 ARCHITECTURE。
3. 維持現狀：每次缺口都由幕僚長個案決定。

**TechLead 建議**：選項 2。不用為小缺口喚醒 Architect（符合憲章 §0 省用量），同時保留紀錄，關卡時一次補文件。

**狀態**：已處置（採用 TechLead 版本）；留待流程檢討。

---

## OQ-003：archtest 沒有檢查「只有 daterange 可以呼叫 `NewDateRange`」

**背景**
- ARCHITECTURE v1.1 §5.3 在 `func NewDateRange(...)` 旁註明「僅供 daterange 呼叫（archtest 檢查）」。
- T-001 的任務單和 archtest 只檢查 **import**，沒有檢查「哪個 package 呼叫了哪個函式」，所以這條規則目前沒有自動檢查。T-001 Review 時 TechLead 也沒有發現這個落差。
- T-002 的 `internal/model/model.go` 在 doc comment 照抄了「checked by archtest」，和目前的實際情況不符。
- 現在不會造成問題：測試專案在 T-002 後收尾，沒有 `daterange` 以外的呼叫者。

**選項（供流程檢討）**
1. 補一張任務：archtest 用 `go/parser` 掃描非測試檔裡的 `model.NewDateRange` 呼叫，只允許 `internal/daterange`，並附自我測試。
2. 改 ARCHITECTURE：拿掉「archtest 檢查」，改由 Code Review 把關。
3. 檢討流程：Review 任務單時，逐條核對 ARCHITECTURE 寫的「由 archtest／lint 強制」，確認每一條都有對應的檢查或任務。

**TechLead 建議**：選項 3 一定要做（這次是 ARCHITECTURE 和任務單之間的落差，Review 沒有逐條核對）。如果專案繼續開發，再做選項 1。

**狀態**：待決；留待流程檢討。
