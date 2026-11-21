# git-stats-cli 開發任務總表（TASKS）

- 文件擁有者：TechLead
- 版本：v0.1（草案，與 ARCHITECTURE.md 一起送 G2；**G2 通過前不派發任何任務**）
- 最後更新：2026-11-22
- 相關交接單：TASKS-001
- 依據（main `9fbe47b`）：
  - `docs/PRD.md` v1.0（`870d766`，AC-01～AC-38、NFR-01～NFR-08、架構待決 P-01～P-08 見 ARCHITECTURE §18）
  - `docs/ARCHITECTURE.md` v0.1、`docs/adr/ADR-001～004`、`CLAUDE.md`（`56fe859`）
  - `docs/TEST_PLAN.md` v0.1（`9fbe47b`，待釐清 D-01～D-22）

> 名詞說明（第一次出現時的白話解釋）
> - **任務單（T-xxx）**：交給 Claude Code（AI 寫程式工具）一次對話內完成的一小塊工作，格式見 TEAM_CHARTER 第 8 節。
> - **前置任務**：必須先合併進 main，本任務才能開始的任務。
> - **archtest（架構測試）**：自動檢查「哪個 package 可以 import 哪個 package」的測試，違規時 `go test` 直接失敗。
> - **Lint**：程式碼風格與常見錯誤的自動檢查（本專案用 golangci-lint）。
> - **E2E（端對端測試）**：先編譯出執行檔，再像使用者一樣用指令執行，比對 stdout／stderr／exit code。
> - **PoC（概念驗證）**：正式開發前，先寫一段丟棄式程式驗證技術風險（本專案為 A-001，分支 `poc/A-001`，不合併）。

## 1. 使用方式

- **狀態**：`待辦` → `進行中`（任務單已推送、Claude Code 執行中）→ `Review`（TechLead 第一輪 Code Review）→ `QA`（PR 已開、QA 驗證中；不影響產品行為的 PR 跳過）→ `完成`（老闆已在 GitHub Merge，填上 PR 連結）。
- **分支**：依 TEAM_CHARTER 第 4 節，每個 T-xxx 從最新 main 建立 `task/T-xxx`；修正輪次 `T-xxx-fixN` 沿用同一分支。
- **個別任務單** `docs/tasks/T-xxx.md`：**G2 通過後**才撰寫（格式依憲章第 8 節）；本表的「驗證指令」會寫進任務單的「驗證方式」。
- **派工規則**：依「前置任務」順序，一次只派一個；第 4 節標明哪些任務彼此不碰同一批檔案、可以並行。
- **大小估計**：S＝1–2 個檔案、M＝3–4 個檔案、L＝5 個以上（皆為一次 Claude Code 對話、1 小時內可驗證）。
- **CI**：由 SRE 的 O-001 建立（GitHub Actions），不在本表；T-001 只提供 CI 要執行的指令（`go build`／`go vet`／`go test`／`golangci-lint run`）。發佈用 release workflow 亦屬 SRE（O-xxx）。

## 2. 任務總表（依依賴順序排列）

| 編號 | 標題 | 對應驗收標準 | 前置任務 | 大小估計 | 狀態 | 分支 | PR 連結 | 受未決事項影響 |
|---|---|---|---|---|---|---|---|---|
| T-001 | 專案骨架＋archtest＋lint 設定 | —（基礎建設；以靜態規則支撐 AC-35、NFR-04/05/08） | —（CI 由 SRE O-001 配合） | L（6 個小檔） | 待辦 | task/T-001 | | Go |
| T-002 | `model` 共用型別與 `apperr` 錯誤種類 | AC-10（`DateRange.Contains` 半開區間）、AC-32（`ParseFormat`） | T-001 | M（4） | 待辦 | task/T-002 | | Go、D-17 |
| T-003 | `daterange`：`YYYY-MM-DD` 嚴格解析與本地時區區間 | AC-10、AC-11、AC-13、AC-14、AC-15 | T-002 | S（2） | 待辦 | task/T-003 | | Go、D-06 |
| T-004 | `stats`：日期篩選、作者合併、計數、排序（純函式） | AC-01、AC-02、AC-03、AC-04、AC-05、AC-06、AC-10、AC-12、AC-18、AC-19 | T-002 | S（2） | 待辦 | task/T-004 | | Go、D-01、D-07、PoC（R-02 僅在效能超標時） |
| T-005 | `render`：表格輸出（含全形對齊、空資料提示） | AC-17、AC-20、AC-21、AC-30（空資料呈現） | T-002 | L（5，含 go.mod/go.sum 新增 go-runewidth） | 待辦 | task/T-005 | | Go、D-05、D-08、D-21 |
| T-006 | `render`：CSV 輸出（RFC 4180、UTF-8 無 BOM、LF） | AC-22、AC-23、AC-24、AC-25 | T-005（同 package，避免 golden 更新旗標衝突） | M（3–4） | 待辦 | task/T-006 | | Go、P-05、P-06、D-10、D-11 |
| T-007 | `testutil/fixture`：測試用 fixture repo 建立工具 | —（測試基礎建設，支撐 AC-01～AC-09、AC-12、AC-28～AC-30、AC-35） | T-001 | S（2） | 待辦 | task/T-007 | | Go、問題 #1（與 TEST_PLAN fixture 做法重疊） |
| T-008 | `gitlog`：`git log -z --numstat` 串流解析器 | AC-07、AC-08、AC-12（`%at`）、AC-24（UTF-8 名字） | T-002 | M（3） | 待辦 | task/T-008 | | Go |
| T-009 | `gitlog`：Runner、git 版本檢查、`Open`、`Commits`（唯讀白名單） | AC-05、AC-06、AC-09、AC-27、AC-28、AC-29、AC-30、AC-35 | T-008 | M（3–4） | 待辦 | task/T-009 | | Go、**PoC（R-01、R-03）**、P-01、P-03、P-04、D-04、D-09、D-15、D-19 |
| T-010 | `gitlog` 整合測試（真 git＋fixture repo） | AC-05、AC-06、AC-07、AC-08、AC-09、AC-12、AC-28、AC-29、AC-30、AC-35 | T-007、T-009 | S（1–2） | 待辦 | task/T-010 | | Go、**PoC（R-01）**、P-03、P-04、D-03、D-04 |
| T-011 | `cli`：旗標解析、驗證順序、`--help`／`--version`、錯誤訊息 | AC-13、AC-14、AC-15、AC-16、AC-31、AC-32、AC-33、AC-34 | T-003 | L（5，含 go.mod/go.sum 新增 pflag） | 待辦 | task/T-011 | | Go、D-14、D-16、D-17、D-18 |
| T-012 | `cli.Run` 流程串接＋`cmd/git-stats/main.go` | AC-01、AC-05、AC-06、AC-11、AC-20、AC-25、AC-26、AC-27、AC-28、AC-30、AC-36 | T-004、T-006、T-009、T-011 | M（3） | 待辦 | task/T-012 | | Go、PoC（R-04 Windows 時區）、P-01、P-07、D-12、D-13 |
| T-013 | E2E 測試框架＋F1 作者統計 E2E | AC-01、AC-02、AC-03、AC-04、AC-05、AC-06、AC-07、AC-08、AC-09、AC-29 | T-007、T-012 | M（3） | 待辦 | task/T-013 | | Go、問題 #1、D-03、D-09 |
| T-014 | E2E：F2 日期範圍＋F5 錯誤處理 | AC-10、AC-11、AC-12、AC-13、AC-14、AC-15、AC-16、AC-27、AC-28、AC-30、AC-31、AC-32、AC-36 | T-013 | S（2） | 待辦 | task/T-014 | | Go、PoC（R-04）、問題 #1、#5、D-06、D-14 |
| T-015 | E2E：F3 表格／F4 CSV 輸出＋`--help`／`--version`／唯讀 | AC-17、AC-18、AC-19、AC-20、AC-21、AC-22、AC-23、AC-24、AC-25、AC-26、AC-33、AC-34、AC-35 | T-013 | M（3–4） | 待辦 | task/T-015 | | Go、問題 #1、P-05、P-07、D-05、D-10、D-13、D-18、D-19 |
| T-016 | archtest：AC 測試覆蓋檢查（`test/e2e/ac_list.txt`） | AC-01～AC-36（覆蓋強制；AC-37、AC-38 見問題 #2） | T-014、T-015 | S（2） | 待辦 | task/T-016 | | Go、**QA 決定 ac_list 內容**、問題 #2 |
| T-017 | GoReleaser 設定（多平台產物、版本注入、Release 說明） | AC-34、AC-38 | T-012 | S（1–2） | 待辦 | task/T-017 | | Go、P-08、PoC（R-04）、問題 #6（擁有者） |
| T-018 | macOS／Linux 安裝腳本 `scripts/install.sh` | AC-37 | T-017 | S（1–2） | 待辦 | task/T-018 | | P-02、P-08、D-20 |
| T-019 | Windows 安裝腳本 `scripts/install.ps1` | AC-37 | T-017 | S（1–2） | 待辦 | task/T-019 | | PoC（R-04）、P-02、D-20 |
| T-020 | 工程文件：`README.md`＋`CHANGELOG.md` | AC-33（用法說明）、AC-37（安裝說明）、AC-38（變更摘要與三平台安裝方式） | T-012、T-018、T-019 | S（2） | 待辦 | task/T-020 | | P-02、P-03、P-06、D-20、D-21 |

共 **20 個任務**。

## 3. 各任務範圍與驗證指令

> 檔案路徑依 ARCHITECTURE §13；「範圍外」與詳細驗收標準在 G2 後的 `docs/tasks/T-xxx.md` 中寫明。所有任務的完成定義都包含 CLAUDE.md「完成定義」四條指令全過（T-001 之前不適用）。

| 編號 | 主要檔案（預估） | 要點（引用的架構章節） | 驗證指令 |
|---|---|---|---|
| T-001 | `go.mod`、`.gitignore`（`bin/`）、`cmd/git-stats/main.go`（暫時只印版本骨架，T-012 取代）、`internal/version/version.go`、`internal/archtest/imports_test.go`、`.golangci.yml` | §5.1／§5.2 相依白名單、§11 強制規則（depguard、forbidigo、errcheck 等）、go.mod 依賴白名單、`testutil` 只能被測試 import。L 超出 5 檔的指引，但 6 個都是小型設定／樣板檔，拆開反而讓 archtest 與 lint 規則分兩次審，故維持一張 | `go build ./... && go vet ./... && go test ./... && golangci-lint run` |
| T-002 | `internal/model/model.go`、`model_test.go`、`internal/apperr/apperr.go`、`apperr_test.go` | §5.3 型別（`Format` 只能經 `ParseFormat`、`DateRange` 欄位不匯出）、§8 Usage／Runtime | `go test ./internal/model/... ./internal/apperr/...` |
| T-003 | `internal/daterange/parse.go`、`parse_test.go`（含 fuzz） | §5.3、§6-1（`time.Date(y,m,d+1,…)`、DST）；時區由參數傳入，不用 `time.Local` | `go test ./internal/daterange/... && go test -fuzz=FuzzParse -fuzztime=30s ./internal/daterange/` |
| T-004 | `internal/stats/aggregate.go`、`aggregate_test.go` | §6 統計規則 1～5（merge 行數強制 0、小寫 email、最新名字、穩定排序） | `go test ./internal/stats/...` |
| T-005 | `internal/render/table.go`、`table_test.go`、`internal/render/testdata/*.golden`、`go.mod`、`go.sum` | §7.1 版面（英文表頭、欄間 2 空白、控制字元字面顯示、空資料提示行） | `go test ./internal/render/... -run Table` |
| T-006 | `internal/render/csv.go`、`csv_test.go`（含讀回 fuzz）、`testdata/*.golden` | §7.2（`encoding/csv`、LF、無 BOM、不做公式跳脫） | `go test ./internal/render/... -run CSV && go test -fuzz=FuzzCSV -fuzztime=30s ./internal/render/` |
| T-007 | `internal/testutil/fixture/fixture.go`、`fixture_test.go` | §12.1 宣告式 API 與隔離（`GIT_CONFIG_NOSYSTEM`、`HOME`、`core.autocrlf=false`、固定日期） | `go test ./internal/testutil/...` |
| T-008 | `internal/gitlog/parse.go`、`parse_test.go`（含 fuzz）、`testdata/` | §4.3 token 狀態機、`bufio.Reader` 串流；表格驅動測試含改名、binary、merge 無 numstat、名字含 `\t`／逗號／emoji | `go test ./internal/gitlog/... -run Parse && go test -fuzz=FuzzParse -fuzztime=30s ./internal/gitlog/` |
| T-009 | `internal/gitlog/runner.go`、`repo.go`、`repo_test.go`（假 Runner） | §4.1 指令與旗標、§4.2 環境變數、§9 唯讀子指令白名單、git ≥ 2.34 檢查 | `go test ./internal/gitlog/... && golangci-lint run ./internal/gitlog/...` |
| T-010 | `internal/gitlog/repo_integration_test.go` | §12 整合測試：改名、binary、merge、空 repo、子資料夾、未合併分支、執行前後 refs／config 不變 | `go test ./internal/gitlog/... -run Integration` |
| T-011 | `internal/cli/args.go`、`messages.go`、`args_test.go`、`go.mod`、`go.sum` | §3 CLI 規格與「驗證順序」1～6、§8 訊息格式 | `go test ./internal/cli/... -run Args` |
| T-012 | `internal/cli/run.go`、`run_test.go`、`cmd/git-stats/main.go` | §5.4 資料流（先寫 `bytes.Buffer`）、§7.3 `--output`、§8 exit code 對應與 `recover()`、main 傳入 `time.Local` 並 import `time/tzdata` | `go test ./internal/cli/... && go build -o bin/git-stats ./cmd/git-stats` |
| T-013 | `test/e2e/main_test.go`（`TestMain` 編譯執行檔）、`helpers_test.go`、`f1_stats_test.go` | §12 E2E；測試名稱含 AC 編號（`TestE2E_AC01_…`）；fixture 內容依 TEST_PLAN §4.3 FX-01～FX-05 | `go test -run TestE2E ./test/e2e/...` |
| T-014 | `test/e2e/f2_dates_test.go`、`f5_errors_test.go` | 以 `TZ=Asia/Taipei`、`America/Los_Angeles`、`America/New_York` 啟動；錯誤類一律檢查 stdout 0 byte | `go test -run 'TestE2E_AC1[0-6]\|TestE2E_AC2[78]\|TestE2E_AC3[0126]' ./test/e2e/...` |
| T-015 | `test/e2e/f3_table_test.go`、`f4_csv_test.go`、`f5_misc_test.go`、`test/e2e/testdata/*.golden` | 排序連跑 10 次（AC-19）、CSV 以標準解析器讀回（AC-23）、`--output` 逐 byte 相同（AC-26）、執行前後 `status --porcelain`／`for-each-ref`／`.git/config` 比對（AC-35） | `go test -run 'TestE2E_AC(1[7-9]\|2[0-6]\|3[345])' ./test/e2e/...` |
| T-016 | `test/e2e/ac_list.txt`、`internal/archtest/ac_coverage_test.go` | §11「每條 Must AC 至少一個測試」；新檔案，不改 T-001 的 archtest 檔 | `go test ./internal/archtest/...` |
| T-017 | `.goreleaser.yaml` | §14、ADR-003：`CGO_ENABLED=0`、linux／darwin amd64+arm64、windows amd64、`checksums.txt`、ldflags 注入 `internal/version.Version`、Release 說明模板 | `goreleaser check && goreleaser release --snapshot --clean && ./dist/*linux_amd64*/git-stats --version` |
| T-018 | `scripts/install.sh`（必要時加一支測試腳本） | ADR-003：偵測 OS／架構、下載 latest、驗 SHA-256、`INSTALL_DIR`；只用系統內建工具 | `sh -n scripts/install.sh && shellcheck scripts/install.sh`（實際安裝於 G4 由 QA 驗） |
| T-019 | `scripts/install.ps1` | ADR-003：PowerShell 5.1 內建 cmdlet、驗 SHA-256、安裝到 `%LOCALAPPDATA%\Programs\git-stats\` 並加入使用者 PATH | `pwsh -NoProfile -Command "Invoke-ScriptAnalyzer scripts/install.ps1"`（實際安裝於 G4 由 QA 驗） |
| T-020 | `README.md`（repo 根目錄）、`CHANGELOG.md` | 安裝（三平台 1 行指令＋手動備案、macOS quarantine 說明）、用法與範例、Git ≥ 2.34 需求、不套用 `.mailmap`、partial clone 限制、CSV 為 UTF-8 無 BOM／LF、公式注入提醒、exit code 表 | 人工 Review＋`go build -o bin/git-stats ./cmd/git-stats` 後逐一執行 README 範例指令 |

**雲端 Review 環境提醒**：雲端副本目前有 `go`、`git 2.47.3`，**沒有** `golangci-lint`、`goreleaser`、`shellcheck`、`pwsh`；`docs/RUNBOOK.md` 尚未建立。T-001 Review 前需 SRE 在 RUNBOOK 補上安裝方式（含 golangci-lint 固定版號，與 O-001 一致）。

## 4. 依賴圖與可並行的任務

```
T-001 ─┬─ T-002 ─┬─ T-003 ── T-011 ─┐
       │         ├─ T-004 ──────────┤
       │         ├─ T-005 ── T-006 ─┤
       │         └─ T-008 ── T-009 ─┼─ T-012 ─┬─ T-013 ─┬─ T-014 ─┐
       │                     │      │         │         └─ T-015 ─┴─ T-016
       └─ T-007 ─────────────┴ T-010│         └─ T-017 ─┬─ T-018 ─┐
                                    │                   └─ T-019 ─┴─ T-020（另需 T-012）
```

- T-010 前置為 T-007＋T-009；T-012 前置為 T-004、T-006、T-009、T-011（T-010 不擋 T-012，但建議先完成以便早期發現 git 輸出問題）。
- **不碰同一批檔案、可並行派發的組合**：{T-003、T-004、T-005、T-007、T-008}（T-002 完成後）；{T-014、T-015}；{T-018、T-019}。
- **共用檔案、必須依序**：`go.mod`／`go.sum`（T-001 → T-005 → T-011，若 T-005 與 T-011 同時進行，後合併者需 rebase）；`internal/render`（T-005 → T-006）；`internal/gitlog`（T-008 → T-009 → T-010）；`internal/cli`（T-011 → T-012）；`internal/archtest`（T-001 → T-016，T-016 只新增檔案）。

## 5. AC 覆蓋矩陣

| AC | 等級 | 負責任務（實作 → 測試） |
|---|---|---|
| AC-01 | Must | T-004、T-012 → T-013 |
| AC-02 | Must | T-004 → T-013 |
| AC-03 | Must | T-004 → T-013 |
| AC-04 | Must | T-004 → T-013 |
| AC-05 | Must | T-004、T-009、T-012 → T-010、T-013 |
| AC-06 | Must | T-004、T-009、T-012 → T-010、T-013 |
| AC-07 | Must | T-008 → T-010、T-013 |
| AC-08 | Must | T-008 → T-010、T-013 |
| AC-09 | Must | T-009 → T-010、T-013 |
| AC-10 | Must | T-002、T-003、T-004 → T-014 |
| AC-11 | Must | T-003、T-012 → T-014 |
| AC-12 | Must | T-004、T-008 → T-010、T-014 |
| AC-13 | Must | T-003、T-011 → T-014 |
| AC-14 | Must | T-003、T-011 → T-014 |
| AC-15 | Must | T-003、T-011 → T-014 |
| AC-16 | Must | T-011 → T-014 |
| AC-17 | Must | T-005 → T-015 |
| AC-18 | Must | T-004 → T-015 |
| AC-19 | Must | T-004 → T-015 |
| AC-20 | Must | T-005、T-012 → T-015 |
| AC-21 | Should | T-005 → T-015 |
| AC-22 | Must | T-006 → T-015 |
| AC-23 | Must | T-006 → T-015 |
| AC-24 | Must | T-006、T-008 → T-015 |
| AC-25 | Must | T-006、T-012 → T-015 |
| AC-26 | Should | T-012 → T-015 |
| AC-27 | Must | T-009、T-012 → T-014 |
| AC-28 | Must | T-009、T-012 → T-010、T-014 |
| AC-29 | Must | T-009 → T-010、T-013 |
| AC-30 | Must | T-005、T-009、T-012 → T-010、T-014 |
| AC-31 | Must | T-011 → T-014 |
| AC-32 | Must | T-002、T-011 → T-014 |
| AC-33 | Must | T-011、T-020 → T-015 |
| AC-34 | Must | T-011、T-017 → T-015（Release 版號比對於 G4） |
| AC-35 | Must | T-009 → T-010、T-015 |
| AC-36 | Must | T-012 → T-014 |
| AC-37 | Must | T-018、T-019、T-020（實際驗證：G4，QA E2E） |
| AC-38 | Must | T-017、T-020（另需 SRE release workflow；實際驗證：G4） |

**覆蓋檢查結果：AC-01～AC-38 共 38 條，每條至少對應 1 個 T-xxx，無缺漏。** AC-01～AC-36 都有「實作任務＋自動化測試任務」；AC-37、AC-38 的最終驗證屬 G4 人工／E2E，不在 G3 的自動化測試內（見問題 #2）。

NFR 對照（參考）：NFR-01/02 效能測試不在本表（見問題 #6）；NFR-03 → T-017 產物＋SRE CI 矩陣；NFR-04/05/08 → T-001（depguard 禁網路套件）、T-009（`GIT_NO_LAZY_FETCH`、唯讀白名單）、T-012（只在 `--output` 寫檔）；NFR-06 → T-011；NFR-07 → T-005、T-006、T-008。

## 6. 未決事項與受影響任務

### 6.1 技術棧 Go（ADR-001，待老闆在 G2 決定）
- **T-001～T-017 全部以 Go 為前提**（檔案路徑、指令、lint 工具都是 Go 專屬）。若 G2 不採用 Go，本表需整份重拆。
- T-018～T-020（安裝腳本、README）與語言關係較小，但產物檔名與安裝方式依 ADR-003（GoReleaser），仍會受影響。

### 6.2 PoC A-001（是否執行，待老闆決定）
PoC 要驗證：R-01（git 2.34 是否支援所用旗標）、R-02（10k commits 效能）、R-03（`GIT_NO_LAZY_FETCH`）、R-04（Windows 的 `TZ`／`time/tzdata`／非 ASCII 路徑／未簽章警告）。

| 類別 | 任務 | 說明 |
|---|---|---|
| **建議等 PoC 結果（或老闆決定不做 PoC）再派** | T-009、T-010 | git 旗標與最低版本、效能做法（不用 `--since` 全量讀取）都在這兩張；PoC 若推翻，需新 ADR 並重寫 |
| 可先做，PoC 結果可能需小修 | T-004（僅 R-02 超標走備案時）、T-012（`time/tzdata`）、T-014（Windows 時區 E2E）、T-017、T-019（Windows 產物與安裝） | 修改範圍小，以 fix 任務處理 |
| 不受 PoC 影響 | T-001、T-002、T-003、T-005、T-006、T-007、T-008、T-011、T-013、T-015、T-016、T-018、T-020 | |

若老闆決定不做 PoC：R-01 改由 SRE CI 矩陣以 git 2.34 跑 T-010 整合測試確認（ARCHITECTURE R-01 已列此替代方案）。

### 6.3 PRD 問題 P-01～P-08（ARCHITECTURE §18，待 PM 回覆）

| # | 內容摘要 | 受影響任務 | 暫定做法（依 ARCHITECTURE） |
|---|---|---|---|
| P-01 | 找不到 git／版本過舊無 AC | T-009、T-012 | exit 1＋安裝連結；測試寫在單元層，不掛 AC 編號 |
| P-02 | M2「≤5 分鐘」是否含安裝 Git | T-018、T-019、T-020 | README 註明需先裝 Git |
| P-03 | `.mailmap` 是否要明確忽略 | T-009、T-010、T-020 | `--no-use-mailmap`，README 說明 |
| P-04 | bare repo／detached HEAD 行為 | T-009、T-010 | bare → exit 1；detached HEAD 照常統計 |
| P-05 | CSV 換行 LF 或 CRLF | T-006、T-015 | 固定 LF |
| P-06 | CSV 公式注入 vs AC-23 | T-006、T-020 | 不跳脫，README 提醒 |
| P-07 | AC-35 與 `--output` 指向 repo 內 | T-012、T-015 | AC-35 E2E 只測輸出到 repo 外 |
| P-08 | 平台產物範圍（linux/arm64、windows/arm64） | T-017、T-018 | 產出 linux/arm64，不產 windows/arm64 |

### 6.4 QA 待釐清 D-01～D-22（TEST_PLAN §11，待 PM／Architect 回覆）

| # | 受影響任務 | 備註 |
|---|---|---|
| D-01 名字平手 | T-004 | ARCHITECTURE §6-4 已定義（取 git log 輸出順序較先者） |
| D-02 整段歷史／效能量測 | —（效能測試不在本表） | |
| D-03 改名＋小改 | T-010、T-013 | 依 git `-M` 預設相似度，行數以 numstat 為準 |
| D-04 不受個人 git 設定影響 | T-009、T-010 | ARCHITECTURE §4.2 已用明確旗標覆寫 |
| D-05 表格版面 | T-005、T-015 | ARCHITECTURE §7.1 已定義 |
| D-06 日期前後空白 | T-003、T-014 | 依 A-05 嚴格解析 → exit 2 |
| D-07 email 排序規則 | T-004 | 小寫後依字串（位元組）順序 |
| D-08 emoji／組合字元寬度 | T-005 | 依 go-runewidth，不作判定 |
| D-09 `.mailmap` 忽略 | T-009、T-013 | 同 P-03 |
| D-10 CSV 換行 | T-006、T-015 | 同 P-05 |
| D-11 非特殊欄位加引號 | T-006 | `encoding/csv` 預設不加 |
| D-12 `--output` 寫檔失敗 | T-012 | ARCHITECTURE §7.3：exit 1 |
| D-13 `--output`＋表格 | T-012、T-015 | ARCHITECTURE §3：兩種格式都適用 |
| D-14 多重錯誤的 exit code | T-011、T-014 | ARCHITECTURE §3 驗證順序：使用方式錯誤（2）優先 |
| D-15 bare repo／`.git` 當路徑 | T-009 | 同 P-04 |
| D-16 多餘位置參數 | T-011 | ARCHITECTURE §3：exit 2 |
| D-17 `--format CSV` 大寫 | T-002、T-011 | **ARCHITECTURE 未定義**，需決定 |
| D-18 `--help` 與其他參數並用 | T-011、T-015 | 見問題 #4 |
| D-19 `.git/index` 時間戳 | T-009、T-015 | `GIT_OPTIONAL_LOCKS=0`，不呼叫讀 index 的指令 |
| D-20 「1 個步驟」定義 | T-018、T-019、T-020 | |
| D-21 Windows 非 UTF-8 字碼頁 | T-005、T-020 | README 說明建議設定 |
| D-22 NFR 等級 | —（效能／平台測試不在本表） | |

### 6.5 不受任何未決事項影響、G2（採用 Go）後可直接進行的任務
- **完全不受影響**：T-001、T-007、T-008。
- **只受「ARCHITECTURE 已給出明確答案」的 D／P 項影響**（可依架構先做，若 PM／QA 最後裁定不同再開 fix 任務）：T-002（D-17 除外）、T-003、T-004、T-005、T-006、T-011。
- **建議等待**：T-009、T-010（PoC）；T-016（QA 決定 `ac_list.txt` 內容）。

## 7. 發現的問題（需幕僚長轉交）

1. **【Architect＋QA】Fixture 與 AC 測試的做法、位置、負責人重疊**：ARCHITECTURE §12／§12.1 規定以 Go 的 `internal/testutil/fixture` 現場建 repo、E2E 放 `test/e2e/` 並由開發任務撰寫；TEST_PLAN §2／§4.1 規定以 POSIX shell 腳本建 fixture、放 `tests/fixtures/`（`tests` 與 `test` 不同）、Integration 測試由 QA 的 Q-xxx 撰寫且為主要驗收依據，並有 `EXPECTED_SHA` 機制。兩套並存會重工且可能互相矛盾。本表暫依 ARCHITECTURE 排 T-007、T-013～T-015；請確認 QA 的黑箱套件是另外一套（Q-xxx），還是改為沿用 T 任務的 Go E2E。
2. **【Architect＋QA】archtest 的 AC 覆蓋檢查無法涵蓋 AC-37、AC-38**：ARCHITECTURE §11 要求「每條 Must AC 至少一個測試」，但 AC-37、AC-38（及 AC-34 的 Release 版號比對）屬 G4 人工／E2E，無法在 `go test` 中測試。`test/e2e/ac_list.txt` 的內容需 QA 決定排除哪些 AC；T-016 等這個決定。
3. **【QA】TEST_PLAN 早於 ARCHITECTURE 撰寫**，D-01、D-04、D-05、D-07、D-09、D-10、D-12、D-13、D-14、D-15、D-16 已在 ARCHITECTURE 有答案（多數也列為 P-03／P-04／P-05），需 QA 依 TEST_PLAN §12-2 更新；D-17（`--format` 大小寫）ARCHITECTURE 未定義。
4. **【Architect＋QA】D-18 與驗證順序**：QA 暫定「`--help` 與無效參數並用仍 exit 0」；ARCHITECTURE §3 驗證順序第 1 步「旗標解析失敗（未知旗標）→ exit 2」在 `--help` 之前，所以 `git-stats --help --bogus` 會是 exit 2。`--help --format json` 兩者一致（exit 0）。需統一。
5. **【Architect】Windows 時區設定方式不一致**：TEST_PLAN TC-AC11-03 在 Windows 用 `tzutil` 改系統時區；ARCHITECTURE §12.1 E2E 用 `TZ` 環境變數，並寫「Windows 上 Go 也讀 `TZ`」（標為未驗證）。我**不確定** Go 在 Windows 上是否讀 `TZ` 環境變數（印象中 Windows 版 `time.Local` 讀系統時區設定，但未查證），屬 PoC A-001 R-04 範圍，影響 T-012、T-014。
6. **【幕僚長】部分產出的擁有者不明**：(a) `.goreleaser.yaml`：ADR-003 寫「由 SRE／TechLead 以任務單交給 Claude Code」，本表暫列 T-017，release workflow 留給 SRE；(b) 效能測試（NFR-01/02）：ARCHITECTURE §12.1 規劃為 `perf` build tag 的 Go 測試，TEST_PLAN §2／§6 規劃為 QA 定義、寫成 CI job；本表未排入，請指定負責人。
7. **【SRE】雲端 Review 工具與 RUNBOOK**：雲端副本缺 `golangci-lint`、`goreleaser`、`shellcheck`、`pwsh`，`docs/RUNBOOK.md` 尚未建立；golangci-lint 版號由 SRE O-001 固定。T-001 Review 前需補齊。

## 8. 變更紀錄

| 版本 | 日期 | 作者 | 變更內容 |
|---|---|---|---|
| v0.1 | 2026-11-22 | TechLead | 初版（交接單 TASKS-001）：T-001～T-020，AC-01～AC-38 全數對應；送 G2 |
