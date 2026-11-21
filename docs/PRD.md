# git-stats-cli 產品需求規格（PRD）

- 文件擁有者：PM
- 版本：v1.2（G1 已於 2026-11-21 通過；v1.1 補充 G2 決定、v1.2 補充 NFR 優先度與嚴重度，皆 2026-11-22）
- 最後更新：2026-11-22
- 相關交接單：PRD-001、PRD-002、PRD-003

> 名詞說明（第一次出現時的白話解釋）
> - **CLI**（Command Line Interface，命令列介面）：在終端機輸入指令來使用的程式，沒有圖形介面。
> - **commit**：Git 裡的一次「存檔紀錄」，記錄誰在什麼時間改了哪些檔案。
> - **merge commit**（合併 commit）：把兩條開發線合在一起時產生的 commit，通常有兩個以上的「父 commit」。
> - **author date**（作者時間）：commit 內容最初被寫下的時間；和 **committer date**（提交時間，例如 rebase 後會變）不同。
> - **CSV**（Comma-Separated Values）：用逗號分隔欄位的純文字表格，可以直接用 Excel、Google Sheets 開啟。
> - **exit code**（結束碼）：程式結束時回傳給系統的數字，0 代表成功，非 0 代表失敗，腳本可以據此判斷。
> - **stdout / stderr**：程式的「正常輸出」與「錯誤輸出」兩個管道；分開後，把結果導到檔案時不會混入錯誤訊息。
> - **binary 檔**（二進位檔）：圖片、壓縮檔等非純文字檔，沒有「行」的概念。
> - **MVP**（Minimum Viable Product）：能驗證核心價值的最小版本。
> - **AC**（Acceptance Criteria，驗收標準）：用 Given / When / Then（給定情境／執行動作／預期結果）寫成、可以用指令或測試驗證的條件。

---

## 1. 一句話產品定義與要解決的問題

**一句話定義**：`git-stats-cli` 是一個單一指令的跨平台命令列工具，輸入一個 Git repo 路徑和日期範圍，就能輸出每位作者在該期間的 commit 數與增刪行數，格式可選表格或 CSV。

**要解決的問題**：
- 想知道「某段期間誰貢獻了多少」時，現在只能自己組合 `git log`、`git shortlog`、`--numstat` 等指令再手動加總，費時又容易算錯。
- 同一個人常用不同名字（例如換電腦、改了 `user.name`）提交，手動統計時會被拆成好幾個人。
- 結果需要貼進報告或試算表時，沒有現成、格式穩定的輸出。

## 2. 目標使用者與使用情境

**主要使用者**：會用終端機、手上有 Git repo 的開發者、技術主管與專案負責人（含老闆本人）。

| # | 情境 | 誰 | 做什麼 | 期待結果 |
|---|------|----|--------|----------|
| S1 | 月底檢視 | 老闆 | 對自己的專案查「上個月 1 日到月底」的貢獻 | 終端機直接看到每位作者的 commit 數與增刪行數，最多的人排最前面 |
| S2 | 季報匯出 | 技術主管 | 查一季的統計並輸出成 CSV，匯入 Excel／Google Sheets 做圖 | CSV 開啟後欄位正確、中文名字不亂碼、名字有逗號也不會錯欄 |
| S3 | 發版回顧 | 開源專案維護者 | 查兩次發版之間（例如 2026-09-01 到 2026-10-15）的貢獻者 | 同一 email 的不同署名被合併成一列，預設不把 merge commit 算進去 |
| S4 | 自動化腳本 | 開發者 | 在腳本中每週跑一次並把 CSV 存檔 | 成功與失敗有明確的 exit code，錯誤訊息不會混進 CSV 檔 |

## 3. 成功指標（可量測）

| # | 指標 | 目標值 | 量測方式 |
|---|------|--------|----------|
| M1 | 統計正確率 | 100%：在測試用 fixture repo（預先建好、結果已知的測試 repo）上，每位作者的 commit 數、新增行數、刪除行數全部與預期值一致 | 自動化測試（QA 在 TEST_PLAN 定義 fixture） |
| M2 | 安裝到第一次出結果的時間 | Windows、macOS、Linux 三平台各 ≤ 5 分鐘（依 README 操作，從零開始；**不含安裝 Git 的時間**，計時起點為已安裝 Git 的機器，已定案，2026-11-22 老闆於 G2 核准，見 P-02） | G4 前由老闆或 QA 各平台實測一次並記錄時間 |
| M3 | 真實使用驗證 | 發佈後 2 週內，老闆在 ≥ 1 個真實 repo 使用，隨機抽 3 位作者以 `git log` 手動核對，數字 100% 一致 | 老闆回報 |
| M4 | 驗收標準通過率 | 本 PRD 所有 Must 等級 AC 100% 通過 | QA_REPORT |

## 4. MVP 範圍

### 4.1 這版要做（MVP）
1. 輸入一個 Git repo 路徑與日期範圍（起訖日期）。
2. 統計每位作者（以 email 區分）的 commit 數、新增行數、刪除行數。
3. 同一 email、不同名字合併成一列。
4. 預設排除 merge commit，提供選項可納入。
5. 輸出格式：表格（預設）與 CSV。
6. 依 commit 數由多到少排序（同數時有固定的次排序規則）。
7. 清楚的錯誤訊息與 exit code（非 Git 路徑、日期格式錯誤等）。
8. 單一指令、跨平台（Windows／macOS／Linux）可安裝；以 GitHub tag＋Release 發佈。

### 4.2 這版不做，之後再說
- JSON 輸出（老闆已確認不做）。
- 分支篩選、指定統計某條分支或所有分支（老闆已確認不做；MVP 固定統計目前 HEAD，見 AC-09）。
- 檔案類型／路徑篩選（老闆已確認不做）。
- 依增刪行數或其他欄位排序、自訂排序。
- 相對日期（例如「最近 7 天」「last month」）與含時間的輸入（例如 `2026-01-01 13:00`）。
- 指定時區的選項（MVP 固定用執行環境的本地時區，見 AC-11）。
- 支援 `.mailmap`（Git 的作者身分對照檔）、`Co-authored-by` 共同作者統計。
- 總計列、淨增行數、百分比欄位、每日／每週趨勢。
- 同時統計多個 repo。
- 上架 PyPI、npm 或其他套件庫（老闆已確認只用 GitHub Release）。

### 4.3 永遠不做
- 修改 repo 內容（任何寫入、改設定、改 refs）：本工具永遠唯讀（見 AC-35）。
- 網路傳輸：不上傳任何統計資料、不做遙測（telemetry，偷偷回傳使用資料）。
- 圖形介面（GUI）或網頁儀表板：本產品定位就是 CLI。
- 對個人做「績效評分」或排名評語：只呈現客觀數字。

### 4.4 範圍外（Out of Scope）總表
即 4.2 與 4.3 的全部項目。任何不在 4.1 的需求，都需要走「變更路線」新增章節並送 G1，不在開發中途臨時加入。

## 5. 功能清單

> **關於指令與選項名稱**：本文中的指令名稱 `git-stats` 與選項名稱（`--since`、`--until`、`--format`、`--include-merges`、`--output`）僅為**建議名稱**，方便撰寫可驗證的 AC；Architect 可調整，調整後 AC 的語意不變，請在 ARCHITECTURE.md 記錄對照。技術做法（語言、套件、如何呼叫 Git）一律由 Architect 決定。
>
> **標示「（已定案…）」的 AC**：原為 PM 提出的建議預設值（v0.1 標示為「假設」），老闆已於 2026-11-21 G1 核准全部採用建議答案（見第 7.2 節）。之後任何修改都需走變更路線並記入變更紀錄。
>
> **Exit code 約定（已定案，2026-11-21 老闆於 G1 核准，見 Q-07）**：`0` 成功（包含查無資料）；`1` 執行期錯誤（路徑不存在、不是 Git repo、讀取失敗）；`2` 使用方式錯誤（缺參數、日期格式錯誤、起日晚於迄日、未知選項或格式）。發生錯誤時，錯誤訊息寫到 stderr，stdout 不輸出任何內容。

### F1 作者統計（Must）

**User Story**：身為專案負責人，我想看到指定期間內每位作者的 commit 數與增刪行數，以便快速了解誰貢獻了多少。

- **AC-01 基本統計**
  Given 一個 fixture repo，在期間內作者 A（a@example.com）有 3 個 commit、共新增 30 行、刪除 5 行，作者 B（b@example.com）有 1 個 commit、新增 2 行、刪除 0 行
  When 執行 `git-stats <repo> --since <起日> --until <迄日>`
  Then 輸出恰好 2 列資料，A 為 commits=3、added=30、deleted=5，B 為 commits=1、added=2、deleted=0，exit code 為 0。

- **AC-02 同 email 不同名字合併**
  Given 期間內有 2 個 commit 署名「Alice」與 1 個 commit 署名「Alice Wang」，email 都是 alice@example.com
  When 執行統計
  Then 輸出中 alice@example.com 只有 1 列，commits=3，增刪行數為 3 個 commit 的加總。

- **AC-03 合併後顯示的名字（已定案，2026-11-21 老闆於 G1 核准，見 Q-05）**
  Given 同 AC-02，且期間內最新（author date 最晚）的 commit 署名為「Alice Wang」
  When 執行統計
  Then 該列顯示的名字為「Alice Wang」。

- **AC-04 email 大小寫視為同一人（已定案，2026-11-21 老闆於 G1 核准，見 Q-04）**
  Given 期間內有 commit 的 email 分別為 Bob@Example.com 與 bob@example.com
  When 執行統計
  Then 兩者合併為 1 列，顯示的 email 為全小寫 `bob@example.com`。

- **AC-05 預設排除 merge commit**
  Given 期間內作者 A 有 2 個一般 commit 與 1 個 merge commit
  When 不加任何 merge 相關選項執行統計
  Then A 的 commits=2，增刪行數只計算那 2 個一般 commit。

- **AC-06 選項納入 merge commit（行數部分已定案，2026-11-21 老闆於 G1 核准，見 Q-03）**
  Given 同 AC-05
  When 加上 `--include-merges` 執行統計
  Then A 的 commits=3；merge commit 本身的增刪行數計為 0（避免與被合併的 commit 重複計算），因此 A 的增刪行數與 AC-05 相同。

- **AC-07 binary 檔的增刪行數**
  Given 期間內作者 A 有 1 個 commit，只新增了一個 PNG 圖檔與一個 10 行的文字檔
  When 執行統計
  Then A 的 commits=1、added=10、deleted=0；binary 檔的變更不計入增刪行數，但該 commit 仍計入 commit 數。

- **AC-08 檔案改名不重複計算（已定案，2026-11-21 老闆於 G1 核准，見 Q-08）**
  Given 期間內作者 A 有 1 個 commit，只把一個 100 行的文字檔改名、內容完全不變
  When 執行統計
  Then A 的 commits=1、added=0、deleted=0。

- **AC-09 統計範圍為目前 HEAD 可追溯的 commit（已定案，2026-11-21 老闆於 G1 核准，見 Q-02）**
  Given 一個 repo 目前在 main 分支，另有未合併的分支 feature 上有作者 C 在期間內的 2 個 commit
  When 執行統計
  Then 輸出中不包含作者 C 在 feature 分支上的那 2 個 commit。

- **AC-40 明確忽略 `.mailmap`（已定案，2026-11-22 老闆於 G2 核准，見 P-03）**
  Given 一個 fixture repo，期間內有 commit 的 email 為 old@example.com 與 new@example.com，且 repo 根目錄有 `.mailmap` 把 old@example.com 對應到 new@example.com（並改名）
  When 執行統計
  Then 輸出中 old@example.com 與 new@example.com 各自一列，名字與 email 皆取自 commit 原始紀錄，不受 `.mailmap` 影響；刪除 `.mailmap` 後再執行，輸出逐位元組相同。
  > 文件需求（非本 AC 驗收範圍）：README 必須說明本工具忽略 `.mailmap`，因此結果可能與 `git shortlog`（預設會套用 `.mailmap`）不同。README 不由 PM 擁有，請幕僚長轉交負責角色。

- **AC-41 detached HEAD 照常統計（已定案，2026-11-22 老闆於 G2 核准，見 P-04）**
  Given 一個 fixture repo 處於 detached HEAD 狀態（HEAD 直接指向某個 commit X，而不是分支）
  When 執行統計
  Then exit code 為 0，統計範圍為 X 可追溯到的 commit（與 AC-09 規則相同）：結果與「在指向 X 的分支上執行」完全相同。

### F2 日期範圍（Must）

**User Story**：身為使用者，我想指定起訖日期，以便只統計我關心的那段期間。

- **AC-10 起訖日期都包含在內（已定案，2026-11-21 老闆於 G1 核准，見 Q-01）**
  Given 作者 A 的 commit 分別在 2026-01-01 00:00:00、2026-01-31 23:59:59、2026-02-01 00:00:00（皆為執行環境本地時間）
  When 執行 `--since 2026-01-01 --until 2026-01-31`
  Then A 的 commits=2（前兩個計入，第三個不計入）；起日與迄日相同（例如 `--since 2026-01-15 --until 2026-01-15`）時，統計該日整天。

- **AC-11 時區：日期邊界以執行環境的本地時區解讀（已定案，2026-11-21 老闆於 G1 核准，見 Q-01）**
  Given 執行環境時區為 Asia/Taipei（UTC+8），某 commit 的 author date 為 `2025-12-31T23:30:00-08:00`（等於台北時間 2026-01-01 15:30）
  When 執行 `--since 2026-01-01 --until 2026-01-01`
  Then 該 commit 被計入；同一個 commit 在 `--since 2025-12-31 --until 2025-12-31` 時不被計入。（比較依據是「絕對時間點」，不是 commit 上記錄的時區字面日期。）

- **AC-12 以 author date 判斷是否在範圍內（已定案，2026-11-21 老闆於 G1 核准，見 Q-06）**
  Given 某 commit 的 author date 為 2026-01-10、committer date 為 2026-02-10（例如經過 rebase）
  When 執行 `--since 2026-01-01 --until 2026-01-31`
  Then 該 commit 被計入；執行 `--since 2026-02-01 --until 2026-02-28` 時不被計入。

- **AC-13 日期格式錯誤**
  Given 任一 repo
  When 日期參數不是 `YYYY-MM-DD` 格式（例如 `2026/01/01`、`01-31-2026`、`2026-1-1`、`abc`）
  Then stderr 顯示錯誤訊息，內容包含錯誤的值與正確格式 `YYYY-MM-DD`，stdout 無輸出，exit code 為 2。

- **AC-14 不存在的日期**
  Given 任一 repo
  When 日期參數為 `2026-02-30` 或 `2026-13-01`
  Then 依 AC-13 的方式報錯，exit code 為 2（不得自動換算成其他日期）。

- **AC-15 起日晚於迄日**
  Given 任一 repo
  When 執行 `--since 2026-03-01 --until 2026-02-01`
  Then stderr 顯示「起日晚於迄日」類型的錯誤訊息，stdout 無輸出，exit code 為 2。

- **AC-16 缺少日期參數（已定案，2026-11-21 老闆於 G1 核准，見 Q-09）**
  Given 任一 repo
  When 沒有提供起日或沒有提供迄日（任一缺少）
  Then stderr 顯示錯誤訊息與用法說明，stdout 無輸出，exit code 為 2。

### F3 表格輸出與排序（Must，AC-21 為 Should）

**User Story**：身為使用者，我想在終端機直接看到排好序、對齊的表格，以便一眼看出主要貢獻者。

- **AC-17 預設表格輸出與欄位**
  Given 期間內有資料的 repo
  When 不指定輸出格式執行統計
  Then stdout 輸出表格，含表頭，欄位依序為：作者名字、email、commit 數、新增行數、刪除行數；每列一位作者；同一欄位的內容左緣（文字）或右緣（數字）對齊（以純 ASCII 名字驗證）。

- **AC-18 依 commit 數由多到少排序**
  Given 期間內作者 A 有 5 個、B 有 9 個、C 有 1 個 commit
  When 執行統計（表格或 CSV 皆同）
  Then 資料列順序為 B、A、C。

- **AC-19 commit 數相同時的次排序（已定案，2026-11-21 老闆於 G1 核准，見 Q-10）**
  Given 作者 x@example.com 與 m@example.com 各有 4 個 commit
  When 執行統計（表格或 CSV 皆同）
  Then 依 email 字母順序（不分大小寫）由小到大排列，即 m@example.com 在前；連續執行 10 次，順序完全相同。

- **AC-20 查無資料（空範圍）時的表格輸出（已定案，2026-11-21 老闆於 G1 核准，見 Q-11）**
  Given 指定期間內沒有任何符合條件的 commit
  When 以表格格式執行統計
  Then stdout 輸出一行「指定期間內沒有 commit」類型的提示（不輸出空表格），exit code 為 0。

- **AC-21 中日韓全形字元對齊（Should）**
  Given 作者名字含中文（例如「王小明」）與英文名字混在一起
  When 以表格格式執行統計
  Then 在等寬字型終端機中各欄仍對齊（全形字元以 2 個字元寬度計算）。

### F4 CSV 輸出（Must，AC-26 為 Should）

**User Story**：身為技術主管，我想把統計結果輸出成 CSV，以便匯入試算表做報告。

- **AC-22 CSV 格式與欄位**
  Given 期間內有資料的 repo
  When 以 `--format csv` 執行統計
  Then stdout 輸出 CSV：第 1 行為表頭 `name,email,commits,added,deleted`，之後每位作者一行，欄位以半形逗號分隔，排序同 AC-18／AC-19；數字欄位為不含千分位的整數。

- **AC-23 CSV 特殊字元跳脫**
  Given 作者名字為 `Wang, Alice` 或含雙引號 `Bob "The Builder"`
  When 以 CSV 格式執行統計
  Then 依 RFC 4180（CSV 的通行標準）處理：含逗號、雙引號或換行的欄位以雙引號包住，欄位內的雙引號寫成兩個雙引號（例如 `"Bob ""The Builder"""`）；用標準 CSV 解析器讀回後，名字與原始值完全一致。

- **AC-24 CSV 編碼（已定案，2026-11-21 老闆於 G1 核准，見 Q-12）**
  Given 作者名字含中文「王小明」
  When 以 CSV 格式執行統計
  Then 輸出為 UTF-8 編碼、不含 BOM；以 UTF-8 解碼讀回後名字為「王小明」。

- **AC-25 查無資料時的 CSV 輸出**
  Given 指定期間內沒有任何符合條件的 commit
  When 以 CSV 格式執行統計
  Then stdout 只輸出表頭一行 `name,email,commits,added,deleted`，exit code 為 0。

- **AC-26 直接寫入檔案（Should，見 Q-13）**
  Given 任一有資料的 repo
  When 加上 `--output report.csv` 以 CSV 格式執行統計
  Then 產生 `report.csv`，內容與 AC-22～AC-24 規定的 stdout 輸出逐位元組（byte）相同；若檔案已存在則覆寫；stdout 不輸出 CSV 內容；exit code 為 0。（目的：避免 Windows PowerShell 用 `>` 導出時自動轉成 UTF-16 編碼。）

- **AC-43 CSV 換行固定為 LF（已定案，2026-11-22 老闆於 G2 核准，見 P-05）**
  Given 任一有資料的 repo，在 Windows、macOS、Linux 上分別執行
  When 以 CSV 格式執行統計（輸出到 stdout，或依 AC-26 用 `--output` 寫檔）
  Then 每一行（含表頭與最後一行）都以 LF（`\n`，位元組 0x0A）結尾，輸出中不含任何 CR（`\r`，位元組 0x0D）字元（欄位值本身含 CR 者除外）；三平台輸出逐位元組相同。

- **AC-44 CSV 不處理試算表公式字元，原樣輸出（已定案，2026-11-22 老闆於 G2 核准，見 P-06）**
  Given 作者名字為 `=1+1`、`+cmd`、`-x`、`@SUM(A1)`（以 `=`、`+`、`-`、`@` 開頭）
  When 以 CSV 格式執行統計
  Then 這些值原樣輸出，不加前置單引號、空白或其他字元（只依 AC-23 的規則加引號跳脫）；用標準 CSV 解析器讀回後與原始值完全一致。已知風險見 7.3 R-01。

### F5 輸入檢查、錯誤處理與基本指令行為（Must）

**User Story**：身為在腳本中使用本工具的開發者，我想在輸入有誤時得到清楚的訊息和固定的 exit code，以便腳本能正確判斷成功或失敗。

- **AC-27 路徑不存在**
  Given 路徑 `/no/such/dir` 不存在
  When 以該路徑執行統計
  Then stderr 顯示「路徑不存在」類型的錯誤訊息並包含該路徑，stdout 無輸出，exit code 為 1。

- **AC-28 路徑不是 Git repo**
  Given 一個存在但不在任何 Git repo 內的資料夾（或一個一般檔案）
  When 以該路徑執行統計
  Then stderr 顯示「不是 Git repo」類型的錯誤訊息並包含該路徑，stdout 無輸出，exit code 為 1。

- **AC-29 路徑是 repo 內的子資料夾（已定案，2026-11-21 老闆於 G1 核准，見 Q-14）**
  Given repo 根目錄為 `R`，其子資料夾為 `R/src`
  When 以 `R/src` 為路徑執行統計
  Then 結果與以 `R` 執行完全相同（統計整個 repo，不只該子資料夾）。

- **AC-30 沒有任何 commit 的新 repo**
  Given 一個剛 `git init`、尚無任何 commit 的 repo
  When 執行統計
  Then 行為同查無資料（表格依 AC-20、CSV 依 AC-25），exit code 為 0。

- **AC-31 缺少 repo 路徑參數（已定案，2026-11-21 老闆於 G1 核准，見 Q-09）**
  Given 任一工作目錄
  When 執行時沒有提供 repo 路徑
  Then stderr 顯示錯誤訊息與用法說明，stdout 無輸出，exit code 為 2。

- **AC-32 未知的輸出格式**
  Given 任一 repo
  When 執行 `--format json`（或任何非 `table`／`csv` 的值）
  Then stderr 顯示錯誤訊息並列出可用格式 `table`、`csv`，stdout 無輸出，exit code 為 2。

- **AC-33 說明**
  Given 已安裝本工具
  When 執行 `git-stats --help`
  Then stdout 顯示用法，至少包含：repo 路徑、起日、迄日、輸出格式、納入 merge commit 選項及日期格式 `YYYY-MM-DD` 的說明；exit code 為 0。

- **AC-34 版本**
  Given 已安裝某一版本
  When 執行 `git-stats --version`
  Then stdout 顯示版本號（例如 `0.1.0`），與該版本 GitHub Release 的 tag（例如 `v0.1.0`）去掉開頭 `v` 後一致；exit code 為 0。

- **AC-35 唯讀，不修改 repo（v1.1 修改範圍，已定案，2026-11-22 老闆於 G2 核准，見 P-07）**
  Given 一個 fixture repo，執行前記錄 `git status --porcelain` 輸出、所有 refs（`git for-each-ref` 輸出）與 `.git/config` 內容
  When 以任意合法參數執行統計（含成功與查無資料兩種情況），且未使用 `--output`，或 `--output` 指向 repo 以外的路徑
  Then 執行後三者與執行前完全相同。（v1.1 修改，已定案，2026-11-22 老闆於 G2 核准，見 P-07：`--output` 指向 repo 內時，新檔案必然改變 `git status`，不在本 AC 範圍內。）

- **AC-36 錯誤與結果分流**
  Given 任一會報錯的情況（AC-13～AC-16、AC-27、AC-28、AC-31、AC-32）
  When 執行並把 stdout 導向檔案
  Then 該檔案為空，錯誤訊息只出現在 stderr；成功執行時 stderr 為空。

- **AC-39 找不到 Git 或 Git 版本過舊（已定案，2026-11-22 老闆於 G2 核准，見 P-01）**
  Given 執行環境的 PATH 中找不到 `git`，或找到的 `git` 版本低於 2.34
  When 以合法參數執行統計
  Then stderr 顯示錯誤訊息，說明原因（找不到 Git／目前版本與最低需求 2.34）並提示如何安裝或升級 Git；stdout 無輸出；exit code 為 1。

- **AC-42 bare repo 視為不是 Git repo（已定案，2026-11-22 老闆於 G2 核准，見 P-04）**
  Given 一個 bare repo（以 `git init --bare` 或 `git clone --bare` 建立、沒有工作目錄的 repo）
  When 以該路徑執行統計
  Then 行為同 AC-28：stderr 顯示「不是 Git repo」類型的錯誤訊息並包含該路徑，stdout 無輸出，exit code 為 1。

### F6 安裝與發佈（Must）

**User Story**：身為使用者，我想用一行指令安裝，以便在 Windows、macOS、Linux 上都能直接使用 `git-stats`。

- **AC-37 跨平台單一指令安裝**
  Given 一台乾淨的 Windows 11、macOS（最新兩個大版本之一）、Ubuntu 24.04 機器，已安裝 Git（見假設 A-01）
  When 依 README 的安裝說明執行（每個平台安裝步驟為 1 行指令，或下載 Release 檔案後 1 個步驟）
  Then 開新終端機後可直接執行 `git-stats --version` 並得到 AC-34 的結果。

- **AC-38 GitHub Release 發佈**
  Given 某一版本準備發佈
  When 查看 GitHub repo 的 Releases 頁面
  Then 有對應的 tag（格式 `vX.Y.Z`）與 Release，Release 說明包含該版本的變更摘要與三平台安裝方式；不在 PyPI、npm 上架。

- **AC-45 Release 產物平台：含 linux/arm64，不含 windows/arm64（已定案，2026-11-22 老闆於 G2 核准，見 P-08）**
  Given 某一版本的 GitHub Release 已發佈
  When 列出該 Release 的下載檔案
  Then 除 AC-37 三平台所需的產物外，另包含 Linux arm64 的執行檔壓縮檔；不包含 Windows arm64 的產物。

### AC 總覽

| AC | 標題 | 優先度 | 定案依據（已定案） |
|----|------|--------|--------|
| AC-01 | 基本統計 | Must | |
| AC-02 | 同 email 不同名字合併 | Must | |
| AC-03 | 合併後顯示的名字 | Must | Q-05 |
| AC-04 | email 大小寫視為同一人 | Must | Q-04 |
| AC-05 | 預設排除 merge commit | Must | |
| AC-06 | 選項納入 merge commit | Must | Q-03 |
| AC-07 | binary 檔的增刪行數 | Must | |
| AC-08 | 檔案改名不重複計算 | Must | Q-08 |
| AC-09 | 統計範圍為目前 HEAD | Must | Q-02 |
| AC-10 | 起訖日期都包含在內 | Must | Q-01 |
| AC-11 | 時區以本地時區解讀 | Must | Q-01 |
| AC-12 | 以 author date 判斷 | Must | Q-06 |
| AC-13 | 日期格式錯誤 | Must | |
| AC-14 | 不存在的日期 | Must | |
| AC-15 | 起日晚於迄日 | Must | |
| AC-16 | 缺少日期參數 | Must | Q-09 |
| AC-17 | 預設表格輸出與欄位 | Must | |
| AC-18 | 依 commit 數由多到少排序 | Must | |
| AC-19 | commit 數相同時的次排序 | Must | Q-10 |
| AC-20 | 查無資料時的表格輸出 | Must | Q-11 |
| AC-21 | 全形字元對齊 | Should | |
| AC-22 | CSV 格式與欄位 | Must | |
| AC-23 | CSV 特殊字元跳脫 | Must | |
| AC-24 | CSV 編碼 | Must | Q-12 |
| AC-25 | 查無資料時的 CSV 輸出 | Must | |
| AC-26 | 直接寫入檔案 | Should | Q-13 |
| AC-27 | 路徑不存在 | Must | |
| AC-28 | 路徑不是 Git repo | Must | |
| AC-29 | 路徑是 repo 子資料夾 | Must | Q-14 |
| AC-30 | 沒有任何 commit 的新 repo | Must | |
| AC-31 | 缺少 repo 路徑參數 | Must | Q-09 |
| AC-32 | 未知的輸出格式 | Must | |
| AC-33 | 說明 | Must | |
| AC-34 | 版本 | Must | |
| AC-35 | 唯讀，不修改 repo（v1.1 限定範圍） | Must | P-07（G2） |
| AC-36 | 錯誤與結果分流 | Must | |
| AC-37 | 跨平台單一指令安裝 | Must | |
| AC-38 | GitHub Release 發佈 | Must | |
| AC-39 | 找不到 Git 或版本過舊 | Must | P-01（G2） |
| AC-40 | 明確忽略 `.mailmap` | Must | P-03（G2） |
| AC-41 | detached HEAD 照常統計 | Must | P-04（G2） |
| AC-42 | bare repo 視為不是 Git repo | Must | P-04（G2） |
| AC-43 | CSV 換行固定為 LF | Must | P-05（G2） |
| AC-44 | CSV 公式字元原樣輸出 | Must | P-06（G2） |
| AC-45 | Release 含 linux/arm64、不含 windows/arm64 | Must | P-08（G2） |

所有錯誤情況的 exit code 均依第 5 節開頭的約定（已定案，2026-11-21 老闆於 G1 核准，見 Q-07）。

## 6. 非功能需求

| # | 類別 | 需求 | 驗證方式 | 優先度 | 違反時嚴重度 |
|---|------|------|----------|--------|--------------|
| NFR-01 | 效能 | 對 10,000 個 commit 的 repo，統計整段歷史（表格或 CSV）在參考測試機器上 ≤ 10 秒完成；參考機器規格由 Architect／QA 在 TEST_PLAN 定義 | 效能測試量測執行時間 | Should | 高 |
| NFR-02 | 效能 | 對 1,000 個 commit 的 repo，執行時間 ≤ 2 秒 | 同上 | Should | 高 |
| NFR-03 | 平台 | 支援 Windows 11、macOS 最新兩個大版本、Ubuntu 22.04 與 24.04（x86_64；macOS 另含 Apple Silicon） | CI 或人工在各平台跑完整 AC 測試 | Should | 高 |
| NFR-04 | 安全 | 對 repo 只讀不寫（AC-35）；執行期間不發出任何網路連線 | AC-35；在無網路環境執行全部 AC 仍通過 | Must | 致命 |
| NFR-05 | 安全 | 不收集、不上傳任何資料，不含遙測 | Code review | Must | 致命 |
| NFR-06 | 可用性 | 所有錯誤訊息包含「出了什麼錯」與「怎麼修正」（例如正確日期格式）；錯誤訊息語言為英文（已定案，2026-11-21 老闆於 G1 核准，見 Q-15） | AC-13～AC-16、AC-27、AC-28、AC-31、AC-32 | 未定（見 Q-17） | 中 |
| NFR-07 | 相容性 | 名字或 email 含 UTF-8 非 ASCII 字元（中文、日文、emoji）時，統計與輸出不出錯、不亂碼 | fixture 含中文與 emoji 作者 | 未定（見 Q-17） | 高 |
| NFR-08 | 資料保存 | 工具本身不保存任何資料或快取檔案；只在使用者指定 `--output` 時寫入該檔案 | 執行前後比對使用者家目錄與暫存目錄無新增本工具檔案 | 未定（見 Q-17） | 高 |

**NFR 判定條件（已定案，2026-11-22 老闆於 G2 核准）**
- **效能量測條件（D-02）**：NFR-01、NFR-02 等效能需求只以 warm cache（熱快取：同一 repo 已被讀取過、檔案已在作業系統快取中）的量測結果判定是否通過；cold cache（冷快取）結果不列入通過與否的判定。
- **「整段歷史」的量測範圍（D-02，已定案，2026-11-22 01:06 老闆核准）**：因起訖日必填（Q-09），效能測試以 `--since 1970-01-01 --until 2099-12-31` 代表「整段歷史」（效能測試用 fixture 的 commit 日期須全部落在此範圍內）；cold cache 結果只記錄在報告中，不作為通過與否的依據。
- **違反時的嚴重度（D-22）**：安全類 NFR（NFR-04、NFR-05）未通過＝「致命」；效能類（NFR-01、NFR-02）與平台類（NFR-03）未通過＝「高」。
- **NFR 優先度（已定案，2026-11-22 01:06 老闆核准）**：安全類 NFR-04、NFR-05＝Must；效能類 NFR-01、NFR-02 與平台類 NFR-03＝Should（目前平台類只有 NFR-03）。D-22 的嚴重度不變（優先度與嚴重度是兩件事：Should 項目未通過仍依上列嚴重度記錄缺陷）。NFR-06～NFR-08 的優先度尚未決定，見 Q-17。
- **其餘 NFR 的嚴重度（D-22b，已定案，2026-11-22 01:06 老闆核准）**：NFR-06（可用性）未通過＝「中」；NFR-07（相容性）、NFR-08（資料保存）未通過＝「高」。

## 7. 假設與待決問題

### 7.1 假設
- **A-01**：使用者的電腦已安裝 Git（他們本來就在用 Git repo）。是否需要另外安裝 Git、最低 Git 版本，由 Architect 決定並寫入 README。
- **A-02**：使用者是會用終端機的開發者，不需要圖形介面。
- **A-03**：「作者」指 commit 的 author（寫程式的人），不是 committer（實際提交的人）。
- **A-04**：增刪行數以「文字檔的行」為單位，與 `git log --numstat` 的概念一致（此處只用於說明語意，不是指定做法）。
- **A-05**：日期輸入只接受 `YYYY-MM-DD`，時間精度到「日」。

### 7.2 待決問題（全部已定案：2026-11-21 老闆於 G1 核准，Q-01～Q-16 全數採用 PM 建議答案）

| # | 問題 | 採用答案（原 PM 建議） | 影響的 AC | 狀態 |
|---|------|-------------|-----------|------|
| Q-01 | 日期範圍是否包含起訖兩端？時區怎麼算？ | 兩端都包含（`--until 2026-01-31` 會算到 1/31 23:59:59）；以執行環境的本地時區解讀日期邊界。最符合直覺，也最少驚訝 | AC-10、AC-11 | 已定案（2026-11-21 G1） |
| Q-02 | 沒有分支篩選時，要統計哪些 commit？ | 只統計目前 HEAD（目前所在分支）能追溯到的 commit。與 `git log` 預設一致；統計所有分支容易重複計算 | AC-09 | 已定案（2026-11-21 G1） |
| Q-03 | 納入 merge commit 時，它的增刪行數怎麼算？ | 只計入 commit 數，增刪行數計 0，避免與被合併的 commit 重複計算 | AC-06 | 已定案（2026-11-21 G1） |
| Q-04 | email 大小寫不同（Bob@Example.com vs bob@example.com）算同一人嗎？ | 算同一人，顯示為全小寫。實務上 email 不分大小寫 | AC-04 | 已定案（2026-11-21 G1） |
| Q-05 | 同一 email 有多個名字時，顯示哪個？ | 顯示期間內最新一個 commit 的名字（通常是此人目前慣用的名字） | AC-03 | 已定案（2026-11-21 G1） |
| Q-06 | 判斷 commit 是否在日期範圍內，用 author date 還是 committer date？ | 用 author date（程式實際寫下的時間），rebase 不會改變統計結果 | AC-12 | 已定案（2026-11-21 G1） |
| Q-07 | Exit code 怎麼定？ | 0＝成功（含查無資料）；1＝執行期錯誤（路徑不存在、不是 repo）；2＝使用方式錯誤（參數、日期） | 全部錯誤類 AC | 已定案（2026-11-21 G1） |
| Q-08 | 檔案改名（內容不變）算不算增刪行？ | 不算（0/0），只計實際內容變更，避免改名被當成「刪光再新增」灌水 | AC-08 | 已定案（2026-11-21 G1） |
| Q-09 | repo 路徑、起日、迄日是否都必填？ | MVP 三者都必填，最明確、最不會誤用；「省略迄日＝今天」「省略路徑＝目前資料夾」放到下一版再評估 | AC-16、AC-31 | 已定案（2026-11-21 G1） |
| Q-10 | commit 數相同時怎麼排？ | 依 email 字母順序（不分大小寫）由小到大，保證每次結果相同 | AC-19 | 已定案（2026-11-21 G1） |
| Q-11 | 查無資料時表格怎麼顯示？算成功還是失敗？ | 顯示一行提示、exit code 0（沒資料不是錯誤）；CSV 只輸出表頭 | AC-20、AC-25 | 已定案（2026-11-21 G1） |
| Q-12 | CSV 編碼要不要加 BOM（讓舊版 Excel 直接雙擊開啟中文不亂碼）？ | 不加 BOM（UTF-8 標準做法，對腳本與 Google Sheets 最友善）；Excel 使用者用「資料 → 從文字匯入」。若老闆主要用 Excel 雙擊開啟，改為加 BOM（老闆已確認：維持 UTF-8、不加 BOM） | AC-24 | 已定案（2026-11-21 G1） |
| Q-13 | 是否提供 `--output` 直接寫檔（Should）？ | 提供。Windows PowerShell 5.1 用 `>` 導出會變成 UTF-16，中文 CSV 容易出問題；這是跨平台需求的一部分，成本低 | AC-26 | 已定案（2026-11-21 G1） |
| Q-14 | 路徑是 repo 內的子資料夾時，統計整個 repo 還是只統計該資料夾？ | 統計整個 repo（MVP 不做路徑篩選）；只統計子資料夾屬於「檔案／路徑篩選」，下一版再說 | AC-29 | 已定案（2026-11-21 G1） |
| Q-15 | 錯誤訊息與表頭用英文還是中文？ | 英文（CLI 慣例、跨平台終端機不怕亂碼）；CSV 表頭固定英文以利程式讀取 | NFR-06、AC-22 | 已定案（2026-11-21 G1） |
| Q-16 | 指令名稱用什麼？ | `git-stats`（簡短好記；若與現有工具撞名，由 Architect 提替代名稱） | 全部 | 已定案（2026-11-21 G1） |

### 7.3 已知風險
| # | 風險 | 處理 |
|---|------|------|
| R-01 | CSV 公式注入：作者名字若以 `=`、`+`、`-`、`@` 開頭，用 Excel 等試算表開啟 CSV 時可能被當成公式執行 | 已定案，2026-11-22 老闆於 G2 核准（P-06）：不跳脫、原樣輸出（AC-44），以維持 AC-23「讀回與原值完全一致」；使用者應以匯入方式開啟不受信任的 CSV。列為已接受風險 |

### 7.4 待決問題（v1.2 新增）
| # | 問題 | PM 建議答案 | 影響 | 狀態 |
|---|------|-------------|------|------|
| Q-17 | NFR-06（可用性）、NFR-07（相容性）、NFR-08（資料保存）的優先度是 Must 還是 Should？（老闆 2026-11-22 只決定了安全、效能、平台三類的優先度，以及這三項的嚴重度） | NFR-07＝Must（中文名字亂碼會直接讓統計結果錯誤，且嚴重度已定為高）；NFR-08＝Must（不留資料屬隱私承諾，嚴重度高）；NFR-06＝Should（嚴重度中） | 第 6 節 | 待決 |

## 8. 變更紀錄

| 版本 | 日期 | 作者 | 變更內容 |
|------|------|------|----------|
| v0.1 | 2026-11-21 | PM | 初版草稿（交接單 PRD-001）：AC-01～AC-38、NFR-01～NFR-08、待決問題 Q-01～Q-16。待 G1 審查。 |
| v1.0 | 2026-11-21 | PM | 定稿：G1 通過（老闆 2026-11-21 23:59 核准）。Q-01～Q-16 全數採用建議答案（含 Q-12：維持 UTF-8、不加 BOM），相關「假設」標記改為「已定案」。AC 內容未變更。 |
| v1.1 | 2026-11-22 | PM | 交接單 PRD-002：納入 G2 決定（老闆 2026-11-22 00:44 核准 ARCHITECTURE.md §18 P-01～P-08 與 QA D-22、D-02）。新增 AC-39～AC-45；修改 AC-35（限定為未使用 `--output` 或 `--output` 指向 repo 外，P-07）；修改 M2（不含安裝 Git 時間，P-02）；第 6 節新增 NFR 判定條件（D-02 warm cache、D-22 嚴重度）；新增 7.3 已知風險 R-01（CSV 公式注入，P-06）。其他 AC 內容未變更。 |
| v1.2 | 2026-11-22 | PM | 交接單 PRD-003：納入老闆 2026-11-22 01:06 決定。第 6 節 NFR 表新增「優先度」「違反時嚴重度」欄：NFR-04、05＝Must；NFR-01～03（效能、平台類，平台類僅 NFR-03）＝Should；嚴重度 D-22 不變，D-22b：NFR-06＝中、NFR-07／08＝高。NFR 判定條件新增「整段歷史」量測範圍 `--since 1970-01-01 --until 2099-12-31`、cold cache 只記錄（D-02）。新增 7.4 待決問題 Q-17（NFR-06～08 優先度）。AC 內容未變更。 |
