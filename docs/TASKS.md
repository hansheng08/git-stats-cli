# git-stats-cli 開發任務總表（TASKS）

- 文件擁有者：TechLead
- 版本：v0.3（G2 已於 2026-11-22 00:44 由老闆核准，開始派工）
- 最後更新：2026-11-22
- 相關交接單：TASKS-001、T-001 任務單交接
- 依據（main `915474f` 以後）：
  - `docs/PRD.md` v1.3（`0c55219`；AC 內容與 v1.1 `f5dbf86` 相同，AC-01～AC-45；v1.2／v1.3 只補 NFR 優先度與嚴重度：NFR-04/05＝Must，其餘 NFR＝Should）
  - `docs/ARCHITECTURE.md` v1.1（`5033d3c`，對齊 TEST_PLAN v0.2 的 S-01～S-10）、`docs/adr/ADR-001～004`（已接受）、`CLAUDE.md`、`docs/tasks/A-001.md`
  - `docs/TEST_PLAN.md` v0.3.1（`dd68c80`，§13 決定 `ac_list`＝42 條）
- **G2 決定（老闆，2026-11-22）**：技術棧採用 Go＋呼叫 git CLI＋GoReleaser；PoC A-001 核准執行；指令名稱維持 `git-stats`；MVP 不做程式碼簽章；`.goreleaser.yaml` 由 TechLead 負責（T-017），release workflow 由 SRE 負責（O-xxx）；T-001～T-020 為保留編號，新增任務從 T-021 起編。

> 名詞說明（第一次出現時的白話解釋）
> - **任務單（T-xxx）**：交給 Claude Code（AI 寫程式工具）一次對話內完成的一小塊工作，格式見 TEAM_CHARTER 第 8 節。
> - **前置任務**：必須先合併進 main，本任務才能開始的任務。
> - **archtest（架構測試）**：自動檢查「哪個 package 可以 import 哪個 package」的測試，違規時 `go test` 直接失敗。
> - **Lint**：程式碼風格與常見錯誤的自動檢查（本專案用 golangci-lint）。
> - **E2E（端對端測試）**：先編譯出執行檔，再像使用者一樣用指令執行，比對 stdout／stderr／exit code（＝TEST_PLAN 的 Integration 層級，見 ARCHITECTURE §12.1.1）。
> - **PoC（概念驗證）**：正式開發前，先寫一段丟棄式程式驗證技術風險（本專案為 A-001，分支 `poc/A-001`，不合併）。

## 1. 使用方式

- **狀態**：`待辦` → `進行中`（任務單已推送＝模式 A「已就緒」，到 Claude Code 執行完成為止）→ `Review`（TechLead 第一輪 Code Review）→ `QA`（PR 已開、QA 驗證中；不影響產品行為的 PR 跳過）→ `完成`（老闆已在 GitHub Merge，填上 PR 連結）。
- **分支**：依 TEAM_CHARTER 第 4 節，每個 T-xxx 由 Claude Code 從最新 main 建立 `task/T-xxx`；修正輪次 `T-xxx-fixN` 沿用同一分支。
- **個別任務單** `docs/tasks/T-xxx.md`：依派工順序逐張撰寫（格式依憲章第 8 節）；本表的「驗證指令」會寫進任務單的「驗證方式」。
- **編號**：T-001～T-020 為本表保留編號（任務單尚未寫出也不得挪用）；之後新增的開發任務從 **T-021** 起編。修正輪次仍用 `T-xxx-fixN`。
- **派工規則**：依「前置任務」順序，一次只派一個；第 4 節標明哪些任務彼此不碰同一批檔案、可以並行。
- **大小估計**：S＝1–2 個檔案、M＝3–4 個檔案、L＝5 個以上（皆為一次 Claude Code 對話、1 小時內可驗證）。
- **CI**：由 SRE 的 O-001 建立（GitHub Actions），不在本表；T-001 只提供 CI 要執行的指令（`go build`／`go vet`／`go test`／`golangci-lint run`）。發佈：`.goreleaser.yaml` 屬 TechLead（T-017），release workflow 屬 SRE（O-xxx）。T-001 與 O-001 共用檔案的分工見 `docs/tasks/T-001.md`「與 SRE O-001 可能共用的檔案」。
- **E2E 的分工**（ARCHITECTURE §12.1.2）：測試程式由 T-013～T-015 撰寫；FX-01～FX-12 的 repo 內容、`EXPECTED_SHA`、E2E 黃金檔（`test/e2e/testdata/**`）、`ac_list.txt` 的內容由 **QA** 擁有，Claude Code 不得產生或修改黃金檔來讓測試通過。

## 2. 任務總表（依依賴順序排列）

| 編號 | 標題 | 對應驗收標準 | 前置任務 | 大小估計 | 狀態 | 分支 | PR 連結 | 受未決事項影響 |
|---|---|---|---|---|---|---|---|---|
| T-001 | 專案骨架＋archtest＋lint 設定（[任務單](tasks/T-001.md)） | —（基礎建設；以靜態規則支撐 AC-35、NFR-04/05/08） | —（CI 由 SRE O-001 配合） | L（9 個小檔＋archtest testdata） | 完成 | task/T-001 | https://github.com/hansheng08/git-stats-cli/pull/1 | 問題 #9（CI 先後） |
| T-002 | `model` 共用型別與 `apperr` 錯誤種類（[任務單](tasks/T-002.md)） | AC-10（`DateRange.Contains` 半開區間）、AC-32（`ParseFormat`，區分大小寫） | T-001 | M（4） | 完成 | task/T-002 | https://github.com/hansheng08/git-stats-cli/pull/3 | — |
| T-003 | `daterange`：`YYYY-MM-DD` 嚴格解析與本地時區區間 | AC-10、AC-11、AC-13、AC-14、AC-15 | T-002 | S（2） | 待辦 | task/T-003 | | — |
| T-004 | `stats`：日期篩選、作者合併、計數、排序（純函式） | AC-01、AC-02、AC-03、AC-04、AC-05、AC-06、AC-10、AC-12、AC-18、AC-19 | T-002 | S（2） | 待辦 | task/T-004 | | PoC R-02（僅在效能超標時） |
| T-005 | `render`：表格輸出（含全形對齊、空資料提示） | AC-17、AC-20、AC-21、AC-30（空資料呈現） | T-002 | L（5，含 go.mod/go.sum 新增 go-runewidth） | 待辦 | task/T-005 | | — |
| T-006 | `render`：CSV 輸出（RFC 4180、UTF-8 無 BOM、LF、公式字元原樣） | AC-22、AC-23、AC-24、AC-25、AC-43、AC-44 | T-005（同 package，避免 golden 更新旗標衝突） | M（3–4） | 待辦 | task/T-006 | | — |
| T-007 | `testutil/fixture`：fixture repo 建置工具（ARCHITECTURE §12.1.3、§12.1.4 API，`Generate` 除外） | —（測試基礎建設，支撐 E2E 與整合測試） | T-001 | M（3） | 待辦 | task/T-007 | | PoC R-01（跨平台 SHA 一致性） |
| T-008 | `gitlog`：`git log -z --numstat` 串流解析器 | AC-07、AC-08、AC-12（`%at`）、AC-24（UTF-8 名字） | T-002 | M（3） | 待辦 | task/T-008 | | — |
| T-009 | `gitlog`：Runner、git 版本檢查、`Open`、`Commits`（唯讀白名單） | AC-05、AC-06、AC-09、AC-27、AC-28、AC-29、AC-30、AC-35、AC-39、AC-40、AC-41、AC-42 | T-008 | M（3–4） | 待辦 | task/T-009 | | **等 PoC A-001（R-01、R-03）** |
| T-010 | `gitlog` 整合測試（真 git＋fixture repo） | AC-05、AC-06、AC-07、AC-08、AC-09、AC-12、AC-28、AC-29、AC-30、AC-35、AC-40、AC-41、AC-42 | T-007、T-009 | S（1–2） | 待辦 | task/T-010 | | **等 PoC A-001（R-01）** |
| T-011 | `cli`：旗標解析、驗證順序、`--help`／`-h`／`--version`（明確定義為布林旗標，`--help` 優先）、錯誤訊息 | AC-13、AC-14、AC-15、AC-16、AC-31、AC-32、AC-33、AC-34 | T-003 | L（5，含 go.mod/go.sum 新增 pflag） | 待辦 | task/T-011 | | — |
| T-012 | `cli.Run` 流程串接（沿用 T-011 的布林 help／version 旗標與 `--help` 優先）＋`cmd/git-stats/main.go` | AC-01、AC-05、AC-06、AC-11、AC-20、AC-25、AC-26、AC-27、AC-28、AC-30、AC-36、AC-39 | T-004、T-006、T-009、T-011 | M（3） | 待辦 | task/T-012 | | PoC R-04（Windows 時區 W1／W2，ADR-005） |
| T-013 | E2E 測試框架（`TestMain`、helpers、`fixtures_test.go`）＋F1 作者統計 E2E | AC-01、AC-02、AC-03、AC-04、AC-05、AC-06、AC-07、AC-08、AC-09、AC-29、AC-40、AC-41 | T-007、T-012 | L（4＋QA 提供的黃金檔） | 待辦 | task/T-013 | | QA 黃金檔與 `EXPECTED_SHA`（TEST_PLAN §4.3 已定 FX 內容） |
| T-014 | E2E：F2 日期範圍＋F5 錯誤處理 | AC-10、AC-11、AC-12、AC-13、AC-14、AC-15、AC-16、AC-27、AC-28、AC-30、AC-31、AC-32、AC-36、AC-39、AC-42 | T-013 | S（2） | 待辦 | task/T-014 | | PoC R-04（Windows 的 AC-11 案例先標「待 A-001」） |
| T-015 | E2E：F3 表格／F4 CSV 輸出＋`--help`／`--version`／唯讀 | AC-17、AC-18、AC-19、AC-20、AC-21、AC-22、AC-23、AC-24、AC-25、AC-26、AC-33、AC-34、AC-35、AC-43、AC-44 | T-013 | M（3） | 待辦 | task/T-015 | | QA 黃金檔 |
| T-016 | archtest：AC 測試覆蓋檢查（`test/e2e/ac_list.txt`，內容依 TEST_PLAN §13） | AC-01～AC-36、AC-39～AC-44（共 42 條，覆蓋強制；AC-37、AC-38、AC-45 排除，G4 前驗證；AC-34 列入，只有 Release 比對在 G4） | T-014、T-015 | S（2） | 待辦 | task/T-016 | | — |
| T-017 | GoReleaser 設定（多平台產物、版本注入、Release 說明） | AC-34、AC-38、AC-45 | T-012 | S（1–2） | 待辦 | task/T-017 | | PoC R-04 |
| T-018 | macOS／Linux 安裝腳本 `scripts/install.sh` | AC-37 | T-017 | S（1–2） | 待辦 | task/T-018 | | D-20 |
| T-019 | Windows 安裝腳本 `scripts/install.ps1` | AC-37 | T-017 | S（1–2） | 待辦 | task/T-019 | | PoC R-04（未簽章警告）、D-20 |
| T-020 | 工程文件：`README.md`＋`CHANGELOG.md` | AC-33（用法說明）、AC-37（安裝說明）、AC-38（變更摘要與三平台安裝方式）、AC-40 文件需求、AC-44 風險說明 | T-012、T-018、T-019 | S（2） | 待辦 | task/T-020 | | D-20 |
| T-021 | 效能測試：`fixture.Generate`（perf-1k／perf-10k）＋`perf` build tag 測試 | NFR-01、NFR-02（PRD v1.2 起為 Should，違反＝高；附帶 M1：以產生器的預期總數表驗證正確性）；無 PRD AC | T-007、T-013；**PoC A-001（R-02）** | M（3） | 待辦 | task/T-021 | | **等 PoC A-001（R-02）**；TEST_PLAN §6.2 形狀可能依 A-001 由 QA 調整 |

共 **21 個任務**（T-001～T-021；T-021 為 G2 後新增，由 TechLead 負責、Claude Code 實作、QA 依 TEST_PLAN §6 驗證）。

## 3. 各任務範圍與驗證指令

> 檔案路徑依 ARCHITECTURE §13；「範圍外」與詳細驗收標準寫在 `docs/tasks/T-xxx.md`。所有任務的完成定義都包含 CLAUDE.md「完成定義」四條指令全過。

| 編號 | 主要檔案（預估） | 要點（引用的架構章節） | 驗證指令 |
|---|---|---|---|
| T-001 | `go.mod`、`.gitignore`、`.gitattributes`、`cmd/git-stats/main.go`（空進入點，T-012 取代）、`internal/version/version.go`、`internal/archtest/archtest_test.go`、`internal/archtest/checker_test.go`、`internal/archtest/testdata/**`、`.golangci.yml` | §5.1／§5.2 相依白名單、§11 強制規則、§12.1.7 `.gitattributes`；archtest 自我測試＋反向驗證證明規則真的會失敗。超出 5 檔的指引，理由見任務單 | `go build ./... && go vet ./... && go test ./... && golangci-lint run`（另含任務單驗證 #7、#10 反向驗證） |
| T-002 | `internal/model/model.go`、`model_test.go`、`internal/apperr/apperr.go`、`apperr_test.go` | §5.3 型別（`Format` 只能經 `ParseFormat`、區分大小寫；`DateRange` 欄位不匯出）、§8 Usage／Runtime | `go test ./internal/model/... ./internal/apperr/...` |
| T-003 | `internal/daterange/parse.go`、`parse_test.go`（含 fuzz） | §5.3、§6-1（`time.Date(y,m,d+1,…)`、DST）；時區由參數傳入；前後空白報錯（D-06） | `go test ./internal/daterange/... && go test -fuzz=FuzzParse -fuzztime=30s ./internal/daterange/` |
| T-004 | `internal/stats/aggregate.go`、`aggregate_test.go` | §6 統計規則 1～5（merge 行數強制 0、小寫 email、最新名字與平手規則、穩定排序） | `go test ./internal/stats/...` |
| T-005 | `internal/render/table.go`、`table_test.go`、`internal/render/testdata/*.golden`、`go.mod`、`go.sum` | §7.1 版面（英文表頭、欄間 2 空白、無框線、無千分位、控制字元字面顯示、空資料提示行） | `go test ./internal/render/... -run Table` |
| T-006 | `internal/render/csv.go`、`csv_test.go`（含讀回 fuzz）、`testdata/*.golden` | §7.2（`encoding/csv`、LF 含最後一行、無 BOM、`=`／`+`／`-`／`@` 開頭原樣輸出） | `go test ./internal/render/... -run CSV && go test -fuzz=FuzzCSV -fuzztime=30s ./internal/render/` |
| T-007 | `internal/testutil/fixture/fixture.go`、`env.go`、`fixture_test.go` | §12.1.3 API（含 `Detach`、`Merge` octopus、`Snapshot`、`SnapshotGitDir`／`DiffGitDir`；`Generate` 由 T-021 實作）、§12.1.4 隔離環境（含 TMPDIR／TEMP／TMP／APPDATA／LOCALAPPDATA）、`EXPECTED_SHA` 機制 | `go test ./internal/testutil/...` |
| T-008 | `internal/gitlog/parse.go`、`parse_test.go`（含 fuzz）、`testdata/` | §4.3 token 狀態機、`bufio.Reader` 串流；表格驅動測試含改名、binary、merge 無 numstat、名字含 `\t`／逗號／emoji | `go test ./internal/gitlog/... -run Parse && go test -fuzz=FuzzParse -fuzztime=30s ./internal/gitlog/` |
| T-009 | `internal/gitlog/runner.go`、`repo.go`、`repo_test.go`（假 Runner） | §4.1 指令與旗標（`--no-use-mailmap`、`log HEAD` 支援 detached HEAD）、§4.2 環境變數、§9 唯讀白名單、git ≥ 2.34 檢查（AC-39）、bare repo → 不是 repo（AC-42） | `go test ./internal/gitlog/... && golangci-lint run ./internal/gitlog/...` |
| T-010 | `internal/gitlog/repo_integration_test.go` | §12 整合測試：改名、binary、merge、空 repo、子資料夾、未合併分支、`.mailmap`、detached HEAD、bare repo、`Snapshot` 前後相同 | `go test ./internal/gitlog/... -run Integration` |
| T-011 | `internal/cli/args.go`、`messages.go`、`args_test.go`、`go.mod`、`go.sum` | §3 CLI 規格與驗證順序 1～6；**`--help`、`-h`、`--version` 必須在 pflag 明確定義為布林旗標**（否則 pflag 遇到 `--help` 會停止解析，結果隨順序改變）；**`--help` 優先**：與 `--version` 同時出現時不論順序只印用法、exit 0；`--help --bogus` 與 `--bogus --help` 都 → 2；`--help --format json` → 0；`--since --help` → `--help` 被當成起日的值 → 2；`--format CSV` → 2；§8 訊息格式。測試需涵蓋上述每種順序 | `go test ./internal/cli/... -run Args` |
| T-012 | `internal/cli/run.go`、`run_test.go`、`cmd/git-stats/main.go` | 不得改變 T-011 的旗標定義：`--help`／`-h`／`--version` 維持布林旗標、`--help` 優先於 `--version` 與所有缺參數檢查（`cli.Run` 層級的測試再驗一次 `--version --help` → 只印用法、exit 0）；§5.4 資料流（先寫 `bytes.Buffer`）、§7.3 `--output`、§8 exit code 與 `recover()`、git 缺少／過舊 → exit 1（AC-39）、時區決定方式依 §12.1.6（Windows 待 ADR-005） | `go test ./internal/cli/... && go build -o bin/git-stats ./cmd/git-stats` |
| T-013 | `test/e2e/main_test.go`、`helpers_test.go`、`fixtures_test.go`（依 QA 的 FX 定義）、`f1_stats_test.go`、`test/e2e/fakegit/main.go` | §12.1.2、§12.1.5：`TestMain` 以 ldflags 注入 `9.8.7` 編譯一次，並編譯假 git（AC-39 用）；黃金檔逐位元組比對；每案例前後 `Snapshot()`；子測試名＝TC ID | `go test -run TestE2E ./test/e2e/...` |
| T-014 | `test/e2e/f2_dates_test.go`、`f5_errors_test.go` | Linux／macOS 用 `TZ`；Windows 的 AC-11 案例待 A-001；AC-39 以 PATH 不含 git／假的舊版 git 驗證；錯誤類一律檢查 stdout 0 byte | `go test -run 'TestE2E_AC(1[0-6]\|2[78]\|3[0126]\|39\|42)' ./test/e2e/...` |
| T-015 | `test/e2e/f3_table_test.go`、`f4_csv_test.go`、`f5_misc_test.go` | 排序連跑 10 次（AC-19）、CSV 以標準解析器讀回（AC-23、AC-44）、無 CR（AC-43）、`--output` 逐 byte 相同且輸出到 repo 外（AC-26、AC-35）、`--version` 輸出 `9.8.7` | `go test -run 'TestE2E_AC(1[7-9]\|2[0-6]\|3[345]\|4[34])' ./test/e2e/...` |
| T-016 | `test/e2e/ac_list.txt`（內容逐條依 TEST_PLAN §13 的 42 條）、`internal/archtest/ac_coverage_test.go` | §11「每條 Must AC 至少一個測試」；掃描 `TestE2E_ACnn` 函式名；新檔案，不改 T-001 的 archtest 檔；附自我測試證明缺一條 AC 時會失敗 | `go test ./internal/archtest/...` |
| T-017 | `.goreleaser.yaml` | §14、ADR-003：`CGO_ENABLED=0`、linux／darwin amd64+arm64、windows amd64（不含 windows/arm64，AC-45）、`checksums.txt`、ldflags 注入版本、Release 說明模板（變更摘要＋三平台安裝） | `goreleaser check && goreleaser release --snapshot --clean && ./dist/*linux_amd64*/git-stats --version` |
| T-018 | `scripts/install.sh`（必要時加一支測試腳本） | ADR-003：偵測 OS／架構、下載 latest、驗 SHA-256、`INSTALL_DIR`；只用系統內建工具 | `sh -n scripts/install.sh && shellcheck scripts/install.sh`（實際安裝於 G4 由 QA 驗） |
| T-019 | `scripts/install.ps1` | ADR-003：PowerShell 5.1 內建 cmdlet、驗 SHA-256、安裝到 `%LOCALAPPDATA%\Programs\git-stats\` 並加入使用者 PATH | `pwsh -NoProfile -Command "Invoke-ScriptAnalyzer scripts/install.ps1"`（實際安裝於 G4 由 QA 驗） |
| T-020 | `README.md`（repo 根目錄）、`CHANGELOG.md` | 見下方「T-020 必寫內容」 | 人工 Review＋`go build -o bin/git-stats ./cmd/git-stats` 後逐一執行 README 範例指令 |
| T-021 | `internal/testutil/fixture/generate.go`、`generate_test.go`、`test/e2e/perf_test.go`（`//go:build perf`） | ARCHITECTURE §12.1.8（`Generate(t, dir, GenSpec) (*Repo, Totals)`、`-perf.out` 保留 repo）＋TEST_PLAN §6.2 形狀（10,000／1,000 commits、50 位作者、2% merge、1% 改名、binary）；固定 seed、HEAD SHA 可重現；產生器自行輸出預期總數表（不得呼叫產品程式）；Go perf 測試只是快速檢查，NFR 判定依 TEST_PLAN §6.3（QA 以 hyperfine 量測）；CI 的 perf job 由 SRE 接 | `go test -tags perf -run TestPerf -count=1 -v ./test/e2e/...`（預期：1k ≤ 2 秒、10k ≤ 10 秒，並與預期總數表一致） |

**T-020 必寫內容**（README 由 Claude Code 撰寫，TechLead Review）：
1. 安裝：三平台 1 行指令＋手動備案；需先安裝 Git ≥ 2.34（M2 計時不含安裝 Git）；未簽章執行檔的 SmartScreen／Gatekeeper 處理方式（`xattr -d com.apple.quarantine`）。
2. 用法、範例、exit code 表（0／1／2）、日期格式 `YYYY-MM-DD`、時區依執行環境。
3. **`.mailmap` 說明（PRD AC-40 的文件需求）**：本工具**刻意忽略** `.mailmap`，名字與 email 一律取自 commit 原始紀錄；而 `git shortlog`（以及 `git log` 的 `%aN`／`%aE`）預設會套用 `.mailmap`，所以兩者的作者分列與數字可能不同。用 `git log` 手動核對（M3）時要加 `--no-use-mailmap`。
4. CSV：UTF-8 無 BOM、換行固定 LF；以 `=`、`+`、`-`、`@` 開頭的名字原樣輸出（AC-44），用試算表開啟不受信任的 CSV 請用「匯入」而非直接雙擊（PRD 7.3 R-01）。
5. 限制：partial clone repo 的行為（依 PoC A-001 結論）、bare repo 不支援（AC-42）、detached HEAD 照常統計（AC-41）。

**雲端 Review 環境提醒**：雲端副本目前有 `go 1.27.3`、`git 2.47.3`，**沒有** `golangci-lint`、`goreleaser`、`shellcheck`、`pwsh`；`docs/RUNBOOK.md` 尚未建立。T-001 Review 前需 SRE 在 RUNBOOK 補上安裝方式（含 golangci-lint 固定版號，與 O-001 一致）。

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
- **因 PoC 而延後的關鍵路徑**：T-009 → T-012 → T-013 之後都在等 A-001；A-001 進行期間可先完成 T-001～T-008、T-011。T-021 前置為 T-007、T-013 與 A-001（R-02）；只新增檔案（`generate.go` 與 `fixture.go` 分開），可與 T-014～T-016 並行。

## 5. AC 覆蓋矩陣（PRD v1.1，AC-01～AC-45）

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
| AC-35 | Must（v1.1 限定 `--output` 不在 repo 內） | T-009 → T-010、T-015 |
| AC-36 | Must | T-012 → T-014 |
| AC-37 | Must | T-018、T-019、T-020（實際驗證：G4，QA E2E） |
| AC-38 | Must | T-017、T-020（另需 SRE release workflow；實際驗證：G4） |
| AC-39 | Must | T-009、T-012 → T-014 |
| AC-40 | Must | T-009 → T-010、T-013；README 說明 → T-020 |
| AC-41 | Must | T-009 → T-010、T-013 |
| AC-42 | Must | T-009 → T-010、T-014 |
| AC-43 | Must | T-006 → T-015 |
| AC-44 | Must | T-006 → T-015；README 風險說明 → T-020 |
| AC-45 | Must | T-017（實際驗證：G4） |

**覆蓋檢查結果：PRD v1.1 的 AC-01～AC-45 共 45 條，每條至少對應 1 個 T-xxx，無缺漏。**

**對照 TEST_PLAN §13 `ac_list`（42 條：AC-01～AC-36、AC-39～AC-44）**：這 42 條每條都有實作任務，且都有一個會寫出 `TestE2E_ACnn` 的 E2E 任務（T-013、T-014 或 T-015），T-016 啟用檢查時不會缺。排除的 AC-37、AC-38、AC-45（以及 AC-34 的 Release 比對）由 T-017～T-020 實作，G4 前依 TEST_PLAN 的 REL 案例驗證。

NFR 對照（參考）：NFR-01/02 → T-021（Claude Code 實作 perf 測試，QA 依 TEST_PLAN §6.3 在參考機器上判定）；NFR-03 → T-017 產物＋SRE CI 矩陣；NFR-04/05/08 → T-001（depguard 禁網路套件）、T-009（`GIT_NO_LAZY_FETCH`、唯讀白名單）、T-012（只在 `--output` 寫檔）；NFR-06 → T-011；NFR-07 → T-005、T-006、T-008。

## 6. 決定與未決事項

### 6.1 已定案（G2，2026-11-22）
- 技術棧 Go＋git CLI＋GoReleaser（ADR-001～004 已接受）。
- PoC A-001 核准執行（Architect，`poc/A-001`，不合併；任務單 `docs/tasks/A-001.md`）。
- 指令名稱 `git-stats`；MVP 不做程式碼簽章（T-019、T-020 需寫警告處理方式）。
- `.goreleaser.yaml` → TechLead T-017；release workflow → SRE。
- 編號：T-001～T-020 保留，新任務從 T-021。
- PRD P-01～P-08 全部定案（PRD v1.1）：P-01 → AC-39；P-02 → M2 不含安裝 Git；P-03 → AC-40；P-04 → AC-41、AC-42；P-05 → AC-43；P-06 → AC-44；P-07 → AC-35 修訂；P-08 → AC-45。已全數併入第 2、5 節。
- QA D-01～D-19、D-22 已在 TEST_PLAN v0.2 關閉（依 PRD v1.1 與 ARCHITECTURE §12.1.9）。
- `ac_list` 內容：TEST_PLAN v0.2 §13 決定 42 條（排除 AC-37、AC-38、AC-45），T-016 照此實作。
- 效能測試負責人：新增 T-021（TechLead 負責、Claude Code 實作、QA 驗證）。
- ARCHITECTURE v1.1：`os/exec` 與測試專用路徑規則定案（問題 #8 結案）；D-18b（`--help` 優先、旗標缺值）定案；S-01～S-10 全部回覆。
- PRD v1.2／v1.3：NFR-04、NFR-05＝Must，其餘 NFR＝Should；D-02、D-22b、Q-17 定案。AC 無變動。

### 6.2 PoC A-001（已核准，等待結果）
PoC 驗證：R-01（git 2.34 旗標相容性、fixture SHA 跨平台一致）、R-02（1k／10k commits 效能）、R-03（partial clone 與 `GIT_NO_LAZY_FETCH`）、R-04（Windows 時區 W1／W2、非 ASCII 路徑、未簽章警告）。結論寫入 ADR-002 修訂與 ADR-005。

| 類別 | 任務 | 說明 |
|---|---|---|
| **等 PoC A-001 結果再派（狀態維持「待辦」，任務單在 PoC 回報後才寫）** | T-009、T-010、T-021 | git 旗標與最低版本、partial clone 做法、效能做法與 perf fixture 形狀 |
| 可先做，PoC 結果可能需小修 | T-004（僅 R-02 超標走備案時）、T-007（SHA 不一致時）、T-012（Windows 時區）、T-014（Windows 的 AC-11 案例）、T-017、T-019 | 以 fix 任務處理 |
| 不受 PoC 影響 | T-001、T-002、T-003、T-005、T-006、T-008、T-011、T-013、T-015、T-016、T-018、T-020 | |

### 6.3 仍未決
| # | 內容 | 受影響任務 | 等誰 |
|---|---|---|---|
| D-20 | AC-37「下載後 1 個步驟」如何計算；安全警告處理算不算一步 | T-018、T-019、T-020 | PM／QA |
| D-21 | Windows 非 UTF-8 字碼頁下的顯示 | T-005、T-020 | PoC A-001 附帶檢查 |
| E2E 黃金檔、`EXPECTED_SHA` | QA 擁有的 E2E 內容 | T-013～T-015 | QA |
| Git 2.34 CI job（S-04） | 每個 PR 以最低版本 git 跑完整 `go test` | —（SRE O-xxx） | SRE |

### 6.4 不受未決事項影響、可直接進行的任務
- **立即可派**：T-001（問題 #9 的確認不擋撰寫與實作），之後依序 T-002、T-003、T-004、T-005、T-006、T-007、T-008、T-011。
- **等待中**：T-009、T-010、T-021（PoC A-001）；T-013～T-015（QA 黃金檔）；T-018～T-020（D-20，可先寫但驗收需等）。

## 7. 發現的問題（需幕僚長轉交）

已解決（留存備查）：#1 fixture／E2E 兩套做法 → ARCHITECTURE §12.1 統一為 Go 一套；#2 AC-37／38 的覆蓋檢查 → TEST_PLAN v0.2 §13 決定 42 條；#3 D 項答覆 → ARCHITECTURE §12.1.9（QA 待更新 TEST_PLAN）；#4 D-18 → ARCHITECTURE §3 定案；#5 Windows 時區 → 已查證 Go 在 Windows 不讀 `TZ`，交 PoC A-001／ADR-005；#6(a) `.goreleaser.yaml` 擁有者 → T-017；#6(b) 效能測試負責人 → T-021。

仍待處理：
7. **【SRE】雲端 Review 工具與 RUNBOOK**：雲端副本缺 `golangci-lint`、`goreleaser`、`shellcheck`、`pwsh`，`docs/RUNBOOK.md` 尚未建立；golangci-lint 版號由 SRE O-001 固定。T-001 Review 前需補齊。
8. ~~`os/exec` 規則與 fixture 衝突~~ **已結案**（ARCHITECTURE v1.1 §11：產品只有 `internal/gitlog`；測試專用路徑 `internal/testutil/...`、`test/e2e/...`、任何 `_test.go` 可用，且不得被產品 import，由 archtest 強制；T-001 任務單已同步）。剩餘小事（供 Architect 參考，不擋 T-001）：forbidigo 只能限制到 package 層級，做不到 §9「cli 的輸出函式」；`test/e2e/fakegit` 需要印出與 `os.Exit`，T-001 讓 forbidigo 排除 `test/`。
9. **【幕僚長／老闆】T-001 與 O-001 的先後與 G3「CI 全綠」**：T-001 合併前 main 沒有 CI；若 O-001 先合併，main 上沒有 `go.mod` 會讓 CI 失敗。建議：T-001 先合併（以 TechLead 雲端驗證取代 CI，需老闆同意這一次例外），O-001 再從含 T-001 的 main 開分支，其 CI 首跑即驗證 T-001；或 O-001 的 PR 等 T-001 合併後再合併。
10. **【幕僚長】共用雲端副本的並行提交**：撰寫期間，PM 與 Architect 在同一個 `/workspace/projects/git-stats-cli` 先後留下未推送的本機 commit 與未 staged 的修改；在共用副本 push main 會把別人的 commit 一起推上去。TechLead 已改用自己的 worktree `/workspace/projects/git-stats-cli-techlead`，建議其他 Bot 比照。
11. ~~ac_list 描述不一致~~ **已結案**（ARCHITECTURE v1.1 §12.1.2 改為引用 TEST_PLAN §13）。

## 8. 變更紀錄

| 版本 | 日期 | 作者 | 變更內容 |
|---|---|---|---|
| v0.1 | 2026-11-22 | TechLead | 初版（交接單 TASKS-001）：T-001～T-020，AC-01～AC-38 全數對應；送 G2 |
| v0.2 | 2026-11-22 | TechLead | 依 TEST_PLAN v0.2（`c6879c5`）：T-016 依 §13 的 42 條 ac_list；新增 T-021 效能測試（問題 #6 結案）；仍未決改為 D-18b、D-20、D-21、D-22b、S-01～S-09。另：反映 G2 決定（Go、PoC 核准、goreleaser 歸 T-017、編號保留與 T-021 起編）；依 PRD v1.1 併入 AC-39～AC-45 並重做覆蓋矩陣（45 條全覆蓋）；依 ARCHITECTURE v1.0 §12.1 更新 E2E 分工、T-001 加 `.gitattributes`、T-007 API；T-020 明列 `.mailmap` 與 `git shortlog` 差異說明；T-001 任務單就緒（進行中）；T-009／T-010 標為等 PoC；問題 #1～#5 結案，新增 #8～#10 |
| v0.3 | 2026-11-22 | TechLead | 依 ARCHITECTURE v1.1：T-011／T-012 明定 `--help`／`-h`／`--version` 為布林旗標且 `--help` 優先；問題 #8 結案；T-007／T-013／T-021 對齊 v1.1 API 與假 git；S-01～S-10、D-18b、D-22b 移出未決。PRD v1.2／v1.3 檢查：AC 無變動（仍 45 條），覆蓋矩陣不需修改 |
| v0.4 | 2026-11-22 | TechLead | T-001（含 T-001-fix1）已由老闆合併（PR #1，merge `9212bce`），狀態改為完成並填入 PR 連結 |
| v0.5 | 2026-11-22 | TechLead | **測試專案於 T-002 後收尾**（Hansheng 決定）：T-002 任務單就緒（狀態改為進行中）；T-003 以後不再執行，列保留不刪，因本表狀態沒有「取消」選項，維持原狀態（待辦）；收尾後進行流程檢討。O-001（CI＋`.claude/settings.json`）已由 Hansheng 合併（PR #2：https://github.com/hansheng08/git-stats-cli/pull/2），狀態為完成；本表沒有 O-001 列（CI 不列入本表，見 §1），所以記在這裡。O-002 不在本 repo 執行，改放到新專案範本 |
| v0.6 | 2026-11-22 | TechLead | **測試專案正式收尾**：T-002 已由 Hansheng 合併（PR #3），狀態改為完成並填入 PR 連結 |
