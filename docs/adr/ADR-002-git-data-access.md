# ADR-002：Git 資料取得方式

- 狀態：提議（待 G2）
- 日期：2026-11-22
- 作者：Architect（交接單 ARCH-001）

## 背景
需要取得「HEAD 可追溯的每個 commit 的作者、email、author date、parent 數、各檔案增刪行數（binary 計 0、改名不計）」（AC-01～AC-12）。PRD 假設 A-01「使用者已裝 Git」，並請 Architect 判斷是否可接受。工具必須唯讀（AC-35）、不連網（NFR-04）。

## 選項比較
| 標準 | A. 呼叫使用者的 `git` CLI（`git log --numstat -z`） | B. 純 Go 函式庫 go-git | C. libgit2 綁定（git2go，需 cgo） |
|---|---|---|---|
| 統計語意與 `git log` 一致（改名偵測、binary 判斷、M3 人工核對） | ◎ 就是 git 本身 | △ diff／改名偵測為獨立實作，行數可能與 git 不同（未驗證） | ○ 接近，但改名偵測設定不同 |
| 使用者需先裝 Git | 需要（A-01） | 不需要 | 不需要 |
| 依賴數量 | ◎ 0 | ✕ 大型依賴樹（數十個遞移依賴，未逐一計數） | ✕ cgo＋C 函式庫，破壞靜態交叉編譯 |
| 效能（10k commits numstat） | ○ git 的 diff 實作成熟（未驗證） | △ 純 Go diff 一般較慢（未驗證） | ○ |
| 受使用者 git 設定影響 | △ 需以明確旗標覆寫（見 ARCHITECTURE 4.2） | ◎ 不受影響 | ○ |
| 支援新 repo 格式（SHA-256 repo、reftable、partial clone） | ◎ 跟著使用者的 git | △ 支援度落後 | △ |
| 跨平台建置 | ◎ | ◎ | ✕ |
| 測試難度 | ○ 需真的 git 建 fixture | ○ | △ |

## 決定（建議）
採用 **A：以子程序呼叫 `git`**，所有呼叫集中在 `internal/gitlog`，只允許唯讀子指令（`--version`、`rev-parse`、`log`）。日期篩選**不**用 git 的 `--since/--until`（它依 committer date，與 AC-12 衝突），改在 Go 端依 `%at` 篩選。完整旗標與輸出格式見 ARCHITECTURE.md 第 4 節（輸出格式已在 git 2.47.3 實測）。

### A-01「使用者已裝 Git」評估：可接受，附條件
1. 目標使用者本來就在用 Git repo；呼叫 git CLI 換來「與 `git log` 完全一致」的語意與零依賴，值得。
2. 最低版本 Git 2.34（Ubuntu 22.04 內建版本），README 寫明；旗標相容性未驗證，由 PoC 或 CI 確認。
3. 找不到 git 或版本過舊時 exit 1，訊息附安裝連結（PRD 尚無此 AC，建議 PM 補）。
4. 不自動繞過 `safe.directory` 檢查。

## 後果
- 好處：零執行期 Git 依賴；結果可直接用 `git log` 人工核對（M3）。
- 壞處：每個使用的旗標都要防範使用者設定（mailmap、diff.renames、showSignature、ext diff、textconv、pager）；git 版本差異需 CI 矩陣覆蓋；無法利用 `--since` 提前停止走訪，效能需實測（未驗證）。
- partial clone repo 可能觸發 git 自動下載物件（違反 NFR-04），以 `GIT_NO_LAZY_FETCH=1` 緩解，舊版 git 是否支援不確定。
- 若未來要支援 `.mailmap`，只需移除 `--no-use-mailmap` 並改用 `%aN`/`%aE`，需新 ADR。
