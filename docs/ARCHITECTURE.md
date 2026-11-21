# git-stats-cli 架構文件（ARCHITECTURE）

- 文件擁有者：Architect
- 版本：v0.1（草案，待 G2 審查；技術棧為「建議」，由老闆在 G2 決定）
- 最後更新：2026-11-22
- 相關交接單：ARCH-001
- 依據：docs/PRD.md v1.0（commit 870d766，AC-01～AC-38、NFR-01～NFR-08）
- 相關 ADR：[ADR-001 語言與技術棧](adr/ADR-001-language-and-stack.md)、[ADR-002 Git 資料取得方式](adr/ADR-002-git-data-access.md)、[ADR-003 發佈與安裝](adr/ADR-003-distribution-and-install.md)、[ADR-004 模組邊界與自動強制](adr/ADR-004-module-boundaries-enforcement.md)

> 名詞說明（第一次出現時的白話解釋；PRD 已解釋過的 CLI、commit、AC、exit code、stdout/stderr 等不再重複）
> - **Go**：Google 主導的程式語言，編譯成單一執行檔、不需要另外安裝執行環境。
> - **靜態執行檔（static binary）**：把程式需要的東西全部包進一個檔案，複製到別台電腦就能執行。
> - **子程序（subprocess）**：由本程式啟動的另一個程式（這裡指 `git`），本程式讀取它的輸出。
> - **模組／套件（package）**：Go 程式碼的組織單位，一個資料夾就是一個 package。
> - **架構測試（Architecture Test）**：自動檢查「哪個 package 可以 import 哪個 package」的測試，違規就失敗。
> - **Lint**：程式碼風格與常見錯誤的自動檢查工具；本專案用 golangci-lint（把幾十種 Go 檢查工具整合成一個指令）。
> - **fixture repo**：測試時用程式「現場建出來」、內容與預期結果都已知的 Git repo。
> - **golden file（黃金檔）**：存放「正確輸出」的檔案，測試把實際輸出和它逐字比對。
> - **fuzz test（模糊測試）**：自動產生大量隨機輸入，找出會讓程式當掉的情況。
> - **GoReleaser**：把 Go 程式一次編譯成多平台執行檔並上傳到 GitHub Release 的工具。
> - **GitHub Actions**：GitHub 內建的自動化執行環境，用來跑 CI（每次推送自動建置與測試）。
> - **ldflags**：Go 編譯時注入變數值的參數，用來把版本號寫進執行檔。

---

## 1. 系統脈絡（System Context）

```
            ┌──────────────┐  argv / env(TZ)   ┌──────────────────────────┐
  使用者 ──▶ │  終端機/腳本   │ ────────────────▶ │  git-stats（單一執行檔）   │
            └──────────────┘ ◀── stdout/stderr ─│                          │
                                  exit code     │  ─ 讀 ─▶ 本機 git 執行檔   │──▶ 本機 repo（.git，只讀）
                                                │  ─ 寫 ─▶ --output 指定檔案 │
                                                └──────────────────────────┘
```

- **輸入**：命令列參數、執行環境的本地時區（Go 的 `time.Local`，來源為 OS 設定或 `TZ` 環境變數）。
- **外部依賴**：使用者電腦上的 `git` 執行檔（見第 4 節與 A-01 評估）。
- **輸出**：stdout（結果）、stderr（錯誤）、exit code、（選用）`--output` 檔案。
- **永遠不做**：寫入 repo、網路連線、遙測、快取檔案（PRD 4.3、NFR-04/05/08）。

## 2. 技術選型摘要（建議，G2 由老闆決定）

| 項目 | 建議 | 理由（細節與比較表見 ADR） |
|------|------|------|
| 語言 | Go（Go 官方支援中的最新穩定版；確切版號不確定，建立 go.mod 時以 go.dev 為準並固定） | 強型別、編譯快、產生單一靜態執行檔、跨平台交叉編譯一行完成、AI 熟悉度高（ADR-001） |
| Git 資料 | 以子程序呼叫使用者已安裝的 `git log` | 統計語意（改名偵測、binary 判斷、HEAD 可追溯）與 `git` 完全一致，零額外依賴（ADR-002） |
| 參數解析 | `github.com/spf13/pflag`（唯一的正式執行期依賴之一） | 標準庫 `flag` 遇到第一個非旗標參數就停止解析，`git-stats <repo> --since …` 這種「路徑在前」寫法會失效；pflag 支援穿插、GNU 風格 `--x=y`，且本身無遞移依賴 |
| 全形字寬 | `github.com/mattn/go-runewidth` | AC-21 需要依 Unicode East Asian Width 計算顯示寬度；標準庫 `text/tabwriter` 以字元數計算，中文會錯位 |
| CSV | 標準庫 `encoding/csv` | 已實作 RFC 4180 引號規則（AC-23） |
| 測試 | 標準庫 `testing`（含 fuzz）；不引入測試框架 | 依賴最少；`go test` 一個指令跑完全部 |
| Lint | golangci-lint（開發工具，不進 go.mod 執行期依賴） | 內含 depguard、forbidigo 等規則，可把禁止事項變成失敗（ADR-004） |
| 發佈 | GoReleaser + GitHub Actions，tag `vX.Y.Z` 觸發 → GitHub Release | 符合「GitHub tag＋Release、不上 PyPI／npm」（ADR-003） |
| 安裝 | macOS/Linux：`curl … install.sh \| sh`；Windows：`irm … install.ps1 \| iex` | 每平台 1 行指令（AC-37） |

**執行期第三方依賴總數：2 個**（pflag、go-runewidth；go-runewidth 會帶入 `github.com/rivo/uniseg`，需在 go.sum 中確認，不確定是否仍為其遞移依賴）。新增任何依賴都必須先修改本節與 ADR-001，並由 Architect 審查。

## 3. CLI 規格（定稿）

```
git-stats <repo> --since YYYY-MM-DD --until YYYY-MM-DD [--format table|csv] [--include-merges] [--output FILE]
git-stats --help | -h
git-stats --version
```

| 旗標／參數 | 必填 | 型別／值 | 說明 | 對應 AC |
|---|---|---|---|---|
| `<repo>` | 是 | 路徑（恰好 1 個位置參數） | repo 或其子資料夾；統計整個 repo | AC-27～AC-31 |
| `--since` | 是 | `YYYY-MM-DD` | 起日（含），本地時區當天 00:00:00 | AC-10～AC-16 |
| `--until` | 是 | `YYYY-MM-DD` | 迄日（含），本地時區隔天 00:00:00 之前 | AC-10～AC-16 |
| `--format` | 否 | `table`（預設）\| `csv` | 輸出格式；其他值報錯 exit 2 | AC-17、AC-22、AC-32 |
| `--include-merges` | 否 | 布林 | 納入 merge commit（行數計 0） | AC-05、AC-06 |
| `--output` | 否 | 檔案路徑（相對於目前工作目錄） | 結果寫入檔案（覆寫），stdout 不輸出結果；`table` 與 `csv` 都適用 | AC-26 |
| `--help`, `-h` | 否 | — | 用法印到 **stdout**，exit 0；優先於缺少參數的檢查 | AC-33 |
| `--version` | 否 | — | 只印版本號一行（例如 `0.1.0`），exit 0；優先於缺少參數的檢查 | AC-34 |

**規則**
- 名稱與 PRD 建議名稱完全相同，無調整（PRD 第 5 節要求的對照表因此為恆等對照）。
- 指令名稱沿用 `git-stats`。已知撞名：npm 上有同名的知名套件 `git-stats`（IonicaBizau/git-stats），若使用者同時安裝，PATH 上先找到的會被執行。評估後**建議維持 `git-stats`**（目標使用者裝 npm 版的機率低，且本工具不上 npm）；備選名稱 `git-authorstats`。另外，因執行檔名為 `git-stats`，使用者也能用 `git stats …` 呼叫（Git 的外部子指令機制），屬附帶效果，不另行保證。
- 旗標可放在位置參數前或後；支援 `--since=2026-01-01` 與 `--since 2026-01-01` 兩種寫法；`--` 之後的參數一律視為位置參數。
- 同一旗標重複出現：以最後一次為準（pflag 預設行為）。
- 位置參數 0 個 → AC-31；超過 1 個 → 使用方式錯誤 exit 2。
- 不提供短旗標（除 `-h`）、不提供設定檔、不讀取任何環境變數作為選項（`TZ` 除外，由 Go 標準庫讀取）。

**驗證順序（決定多個錯誤同時發生時回報哪一個）**
1. 旗標解析（未知旗標、缺值）→ exit 2
2. `--help` / `--version` → exit 0
3. 位置參數個數、`--since`/`--until` 是否提供 → exit 2（AC-16、AC-31）
4. `--format` 值 → exit 2（AC-32）
5. 日期格式與是否存在 → exit 2（AC-13、AC-14）
6. 起日 ≤ 迄日 → exit 2（AC-15）
7. 路徑存在 → exit 1（AC-27）
8. 找得到 `git` 且版本 ≥ 最低版本 → exit 1
9. 路徑在 Git repo 內 → exit 1（AC-28）
10. 執行統計、渲染、寫出 → 失敗 exit 1；成功 exit 0

使用方式錯誤（2）一律在執行期錯誤（1）之前檢查，確保「參數寫錯」不受環境影響。

## 4. Git 資料取得方式（細節見 ADR-002）

### 4.1 呼叫的指令
所有 `git` 呼叫只發生在 `internal/gitlog`，使用 `os/exec`（不經過 shell，參數逐一傳入，沒有字串拼接）：

1. `git --version`：解析版本，低於 **最低版本 2.34**（Ubuntu 22.04 內建版本）時報錯 exit 1。最低版本的旗標相容性**未驗證**，由 PoC A-001（若核准）或 CI 矩陣確認。
2. `git -C <repo> rev-parse --is-inside-work-tree --show-toplevel`（失敗 → AC-28「不是 Git repo」）。bare repo 的處理見第 15 節 PRD 問題 P-04。
3. `git -C <repo> rev-parse --verify --quiet HEAD^{commit}`（失敗且第 2 步成功 → 尚無 commit，回傳空結果，AC-30）。
4. 主要資料：
```
git -C <repo> --no-pager -c core.quotepath=off log HEAD \
  --no-color --no-show-signature --no-use-mailmap --encoding=UTF-8 \
  --diff-merges=off --no-ext-diff --no-textconv -M \
  [--no-merges]            # 未加 --include-merges 時
  --numstat -z \
  --format=%x1e%H%x00%P%x00%at%x00%an%x00%ae
```

### 4.2 為什麼每個旗標都要寫（防止使用者的 git 設定改變統計結果）
| 旗標 | 防止什麼 |
|---|---|
| `HEAD` | 只統計 HEAD 可追溯的 commit（AC-09） |
| **不使用** `--since/--until` | git 的 `--since` 依 **committer date** 篩選，與 AC-12（author date）衝突；日期一律在 Go 端用 `%at`（author date 的 Unix 秒數）篩選 |
| `--no-use-mailmap` | `log.mailmap` 預設為 true，會套用 `.mailmap` 改寫作者；PRD 4.2 明定 MVP 不支援 `.mailmap`（見 P-03） |
| `--encoding=UTF-8` | 舊 commit 以其他編碼儲存時轉為 UTF-8（NFR-07） |
| `--diff-merges=off` | 避免 `log.diffMerges` 等設定讓 merge commit 輸出 diff；Go 端另外強制 merge 行數為 0（AC-06，雙重保險） |
| `-M` | 明確開啟改名偵測（使用者可能設 `diff.renames=false`），純改名輸出 `0 0`（AC-08） |
| `--no-ext-diff --no-textconv` | 避免使用者／repo 設定的外部 diff 工具被執行（安全，也避免行數改變） |
| `--no-show-signature` | 避免 `log.showSignature` 把 GPG 輸出混進資料 |
| `--no-pager --no-color` | 避免分頁器與色碼 |
| `-z`、`%x1e`、`%x00` | 以 NUL（`\0`）分隔欄位與檔名，名字、email、檔名含逗號、換行、中文、emoji 都不會解析錯（NFR-07） |

子程序環境變數：繼承使用者環境，另外設定 `GIT_PAGER=cat`、`GIT_TERMINAL_PROMPT=0`、`GIT_OPTIONAL_LOCKS=0`（避免 git 順手更新 index 等鎖檔，支援 AC-35）、`GIT_NO_LAZY_FETCH=1`（partial clone 時禁止自動下載缺少的物件，支援 NFR-04；此變數只在較新的 git 版本有效，**不確定**最低支援版本，見風險 R-03）。不解析 git 的 stderr 文字（會隨語系改變），只依 exit code 判斷；但錯誤時把 git stderr 的第一行附在錯誤訊息中方便排查。

### 4.3 輸出格式與解析（已在 git 2.47.3 實測格式）
每個 commit：`\x1e<hash>\0<parents 以空白分隔>\0<author unix 秒>\0<name>\0<email>\0`，之後若有檔案變更，先一個 `\n`，再接多筆 numstat：
- 一般檔案：`<added>\t<deleted>\t<path>\0`
- binary 檔：`-\t-\t<path>\0` → 行數計 0（AC-07）
- 改名：`<added>\t<deleted>\t\0<old>\0<new>\0`（路徑欄位為空，後面多兩個 token）

解析器是一個逐 token 的狀態機（`internal/gitlog/parse.go`），以 `bufio.Reader` 串流讀取，不把整個輸出讀進記憶體。必須有：表格驅動單元測試（含改名、binary、merge 無 numstat、空 parents、名字含 `\t`/逗號/emoji）、以及 Go 原生 fuzz test（輸入任意 bytes 不得 panic）。

## 5. 模組切分、介面與相依方向

### 5.1 Package 一覽
| Package | 職責 | 可以 import（專案內） | 禁止 |
|---|---|---|---|
| `cmd/git-stats` | 程式進入點：只呼叫 `cli.Run` 並 `os.Exit` | `internal/cli`、`internal/version` | 任何邏輯 |
| `internal/cli` | 解析旗標→驗證→組裝流程→把錯誤對應 exit code→寫 stdout/stderr/檔案 | `daterange`、`gitlog`、`stats`、`render`、`model`、`apperr`、`version` | `os.Exit`、直接 `os/exec` |
| `internal/daterange` | 嚴格解析 `YYYY-MM-DD`、依時區建立半開區間 `[Start, End)` | `model`、`apperr` | I/O、`time.Now`、`time.Local`（時區必須由參數傳入） |
| `internal/gitlog` | 找 git、檢查版本、檢查 repo、執行 `git log`、解析成 `[]model.Commit` | `model`、`apperr` | 統計邏輯、輸出格式 |
| `internal/stats` | 純函式：篩選日期、合併作者、計數、排序 | `model` | 任何 I/O、`os`、`os/exec`、`time.Now`、`time.Local` |
| `internal/render` | 把 `[]model.AuthorStat` 寫成表格或 CSV 到 `io.Writer` | `model` | 檔案系統、`os.Stdout`（只能寫傳入的 `io.Writer`） |
| `internal/model` | 共用資料型別（無邏輯或只有極簡方法） | 無（只用標準庫） | import 任何 internal package |
| `internal/apperr` | 錯誤種類（Usage／Runtime）與建構函式 | 無 | import 任何 internal package |
| `internal/version` | `var Version = "0.0.0-dev"`，發佈時以 ldflags 覆寫 | 無 | — |
| `internal/testutil/fixture` | 測試用：建立 fixture repo | 無（只用標準庫） | 被非 `_test.go` 檔 import |
| `internal/archtest` | 架構測試（只有 `_test.go`） | 標準庫 `go/parser` 等 | — |
| `test/e2e` | 端對端測試：編譯執行檔、以子程序執行、比對 stdout/stderr/exit code | `internal/testutil/fixture` | import 產品 package（只能透過執行檔測） |

### 5.2 相依方向圖（箭頭＝可以 import）
```
cmd/git-stats ──▶ cli ──┬──▶ daterange ─┐
                        ├──▶ gitlog ────┼──▶ model
                        ├──▶ stats ─────┤    apperr（daterange、gitlog 可用）
                        ├──▶ render ────┘
                        └──▶ version、apperr
```
**鐵律**：沒有任何 package 可以 import `cli`；`stats`、`render` 互不 import，也不 import `gitlog`；`model`、`apperr` 不 import 任何專案內 package。這些規則由 `internal/archtest` 與 golangci-lint depguard 強制（第 11 節）。

### 5.3 主要介面（簽名為設計約定，實作時名稱可微調但需 Architecture Review）
```go
// internal/model
type Commit struct {
    Hash       string
    ParentCount int
    AuthorTime time.Time // 來自 %at，絕對時間點
    AuthorName string
    AuthorEmail string   // 原始值，未正規化
    Added, Deleted int   // 已排除 binary；merge 由 stats 強制為 0
}
type AuthorStat struct {
    Name    string // 期間內 AuthorTime 最晚之 commit 的名字（AC-03）
    Email   string // 小寫（AC-04）
    Commits, Added, Deleted int
}
type Format struct{ v string }              // 只能透過 ParseFormat 建立："table" | "csv"
func ParseFormat(s string) (Format, error)
type Options struct { IncludeMerges bool }
type DateRange struct{ start, end time.Time } // 半開區間 [start, end)；欄位不匯出
func NewDateRange(start, end time.Time) (DateRange, error) // 僅供 daterange 呼叫（archtest 檢查）
func (DateRange) Contains(t time.Time) bool
func (DateRange) Since() string; func (DateRange) Until() string // 原始 YYYY-MM-DD，供提示訊息

// internal/daterange
func Parse(since, until string, loc *time.Location) (model.DateRange, error) // 錯誤皆為 apperr.Usage

// internal/gitlog
type Runner interface { // 測試可替換
    Run(ctx context.Context, dir string, args ...string) (stdout io.ReadCloser, wait func() error, err error)
}
type Repo struct { /* 已驗證的 repo 根目錄 */ }
func Open(ctx context.Context, r Runner, path string) (Repo, error)          // AC-27/28/29
func (Repo) Commits(ctx context.Context, opts model.Options) ([]model.Commit, error) // AC-09/30

// internal/stats
func Aggregate(commits []model.Commit, rng model.DateRange, opts model.Options) []model.AuthorStat

// internal/render
func Table(w io.Writer, rows []model.AuthorStat, rng model.DateRange) error // 空資料時輸出提示行（AC-20）
func CSV(w io.Writer, rows []model.AuthorStat) error                        // 空資料時只輸出表頭（AC-25）

// internal/cli
func Run(ctx context.Context, args []string, stdout, stderr io.Writer, loc *time.Location) (exitCode int)
```
> 區間型別放在 `model.DateRange`，讓 `stats` 只依賴 `model`、不依賴 `daterange`。

### 5.4 資料流
`cli.Run` → `daterange.Parse` → `gitlog.Open` → `Repo.Commits` → `stats.Aggregate` → `render.Table/CSV` 寫入 **`bytes.Buffer`** → 全部成功後才一次寫到 stdout 或 `--output` 檔案。先寫緩衝區的目的：任何錯誤發生時 stdout 保證是空的（AC-36）。

## 6. 資料模型與統計規則
本工具沒有資料庫、不保存任何資料（NFR-08），資料模型只存在於記憶體（第 5.3 節型別）。統計規則（全部在 `internal/stats`，純函式，100% 單元測試）：

1. **日期篩選**：`Start ≤ AuthorTime < End`。`Start` = 起日在本地時區的 00:00:00；`End` = 迄日**隔天**在本地時區的 00:00:00（以 `time.Date(y, m, d+1, 0,0,0,0, loc)` 計算，夏令時間日也正確）。比較的是絕對時間點（AC-10、AC-11、AC-12）。
2. **merge 判定**：`ParentCount ≥ 2`。預設由 git 端 `--no-merges` 排除；`IncludeMerges` 時計入 commit 數，`Added/Deleted` 強制為 0（AC-05、AC-06）。
3. **作者鍵**：`strings.ToLower(strings.TrimSpace(email))`（AC-02、AC-04）。只做 ASCII／Unicode 小寫，不做其他正規化。
4. **顯示名字**：同一鍵中 `AuthorTime` 最晚的 commit 名字；時間相同時取 git log 輸出順序中較先出現者（即較新的拓撲順序），保證結果固定（AC-03）。
5. **排序**：`Commits` 遞減 → `Email`（已小寫）字典序遞增（AC-18、AC-19）。使用 `slices.SortStableFunc`，結果與輸入順序無關。
6. **binary／改名**：由 git numstat 決定（`-` 計 0；純改名為 0/0）（AC-07、AC-08）。

## 7. 輸出格式
### 7.1 表格（`--format table`，預設）
- 表頭：`Name  Email  Commits  Added  Deleted`（英文，Q-15）。
- 欄寬依顯示寬度（`go-runewidth`，全形字算 2）計算；文字欄左對齊、數字欄右對齊；欄間 2 個空白；行尾不留空白（AC-17、AC-21）。
- 控制字元（例如名字中的 `\t`、`\n`）在表格中以 `\t`、`\n` 字面顯示，避免破壞排版；CSV 不做此替換。
- 查無資料：只輸出一行 `No commits found between YYYY-MM-DD and YYYY-MM-DD.`，exit 0（AC-20、AC-30）。
### 7.2 CSV（`--format csv`）
- `encoding/csv`，`UseCRLF=false`：**換行固定為 LF（`\n`）**，三平台輸出逐位元組相同（PRD 未規定，見 P-05）。
- 表頭固定 `name,email,commits,added,deleted`；UTF-8、無 BOM；數字為十進位整數（AC-22～AC-25）。
- 公式注入（名字以 `=`、`+`、`-`、`@` 開頭時 Excel 可能當公式執行）：為符合 AC-23「讀回與原始值完全一致」**不做跳脫**，列為風險 R-05 與 PRD 問題 P-06。
### 7.3 `--output`
- 內容與 stdout 版本逐位元組相同；以 `os.WriteFile(path, buf, 0o644)` 覆寫；成功時 stdout、stderr 皆空。寫檔失敗 → exit 1。

## 8. 錯誤處理與結束碼
| 結束碼 | 意義 | 來源 |
|---|---|---|
| 0 | 成功（含查無資料、`--help`、`--version`） | — |
| 1 | 執行期錯誤：路徑不存在、不是 Git repo、找不到 git／版本過舊、git 執行失敗、輸出解析失敗、寫檔失敗、**未預期的 panic** | `apperr.Runtime` |
| 2 | 使用方式錯誤：缺參數、未知旗標、多餘位置參數、日期格式錯誤／不存在、起日晚於迄日、未知格式 | `apperr.Usage` |

- **只有 `internal/cli` 決定結束碼**，只有 `cmd/git-stats/main.go` 呼叫 `os.Exit`。其他 package 回傳 `error`，用 `apperr.Usage(...)` 或 `apperr.Runtime(...)` 包裝；未包裝的 error 一律視為 1。
- **panic 處理**：Go 未攔截的 panic 會以 exit code **2** 結束，會和「使用方式錯誤」混淆。`cli.Run` 必須 `recover()`，輸出 `git-stats: internal error: …` 並回傳 1。
- **訊息格式**（英文，NFR-06）：`git-stats: error: <出了什麼錯>` 換行 `<怎麼修正>`；使用方式錯誤另附一行 `Run 'git-stats --help' for usage.`（AC-16、AC-31 需要用法說明：附精簡 usage 行）。所有訊息字串集中在 `internal/cli/messages.go`，測試以子字串斷言。
  - 例：`git-stats: error: invalid date "2026/01/01" for --since` / `Expected format: YYYY-MM-DD (e.g. 2026-01-31).`
  - 例：`git-stats: error: git executable not found in PATH` / `Install Git 2.34 or later: https://git-scm.com/downloads`
- 錯誤時 stdout 必定為空（第 5.4 節緩衝設計）；成功時 stderr 必定為空（AC-36）。

## 9. 安全設計
| 威脅／要求 | 對策 |
|---|---|
| 修改 repo（AC-35） | 只執行唯讀指令（`--version`、`rev-parse`、`log`）；`GIT_OPTIONAL_LOCKS=0`；允許的 git 子指令白名單寫死在 `gitlog`，並有單元測試檢查；E2E 測試比對執行前後 `status --porcelain`、`for-each-ref`、`.git/config` |
| 指令注入 | `os/exec` 直接傳參數陣列，不經 shell；repo 路徑以 `-C <path>` 單獨一個參數傳入，即使以 `-` 開頭也不會被當成旗標 |
| 惡意 repo 設定執行外部程式 | `--no-ext-diff`、`--no-textconv`、`--no-pager`；不呼叫會讀 index 的指令（`status` 等），避免 `core.fsmonitor`。是否已涵蓋所有可執行外部程式的設定：**未驗證** |
| 網路連線（NFR-04） | 程式本身無任何網路程式碼（forbidigo 禁止 import `net/http`、`net`）；partial clone 以 `GIT_NO_LAZY_FETCH=1` 防止 git 自動抓物件（見 R-03） |
| 遙測、資料保存（NFR-05、NFR-08） | 不寫任何檔案（除 `--output`）、不建快取；lint 禁止 `os.Create`/`os.WriteFile` 出現在 `cli` 的輸出函式以外 |
| Git 的 safe.directory（「dubious ownership」） | 不自動繞過（不加 `-c safe.directory=*`）；以 exit 1 回報並附 git 的原始訊息，讓使用者自行決定 |
| 秘密資訊 | 本工具不需要任何 Token；repo、CI 不得出現秘密。發佈用 `GITHUB_TOKEN` 由 GitHub Actions 自動提供，不寫入檔案 |

## 10. 效能預算
| 項目 | 預算 | 狀態 |
|---|---|---|
| 1,000 commits 完整歷史（NFR-02） | ≤ 2 秒 | **未驗證** |
| 10,000 commits 完整歷史（NFR-01） | ≤ 10 秒 | **未驗證**；主要成本在 git 為每個 commit 計算 diff 與改名偵測，Go 端解析與彙總預估 < 0.5 秒（未驗證） |
| 程式啟動到呼叫 git | < 50 ms | 未驗證（Go 執行檔一般如此） |
| 記憶體 | 與 commit 數線性、10,000 commits < 100 MB | 未驗證（串流解析，只保留 commit 摘要） |

- 參考測試機器規格由 QA 在 TEST_PLAN 定義（PRD NFR-01）；建議以 GitHub Actions 標準 `ubuntu-24.04` runner 為參考機（建議，需 QA 同意）。
- 已知效能風險：因 AC-12 必須用 author date，無法用 git 的 `--since` 提前停止走訪，每次都會讀完整 HEAD 歷史（R-02）。若 PoC 量到超標，備案為「先以 committer date 粗篩放寬一定天數」，但會引入邊界誤差，需新 ADR，不預先實作。

## 11. 讓錯誤寫法寫不出來（自動強制機制）
| 規則 | 強制方式 | 違反時 |
|---|---|---|
| 第 5 節的 package 相依方向 | `internal/archtest`（`go/parser` 讀所有 import，依白名單比對）＋ golangci-lint `depguard` | `go test ./...` 失敗、lint 失敗 |
| 只有 `cmd/git-stats` 能 `os.Exit`、`log.Fatal` | `forbidigo` | lint 失敗 |
| 只有 `internal/gitlog` 能 import `os/exec` | depguard ＋ archtest | 失敗 |
| 禁止 `fmt.Print*`、`os.Stdout`、`os.Stderr` 出現在 `cmd/` 以外（輸出必須寫入傳入的 `io.Writer`） | forbidigo | lint 失敗 |
| 禁止 `time.Now`、`time.Local` 出現在 `cmd/git-stats` 以外（時區一律由 main 傳入） | forbidigo | lint 失敗 |
| 禁止 import `net`、`net/http` 等網路套件 | depguard ＋ archtest | 失敗 |
| 禁止 `panic(` 出現在非測試程式碼 | forbidigo | lint 失敗 |
| 錯誤一定要處理 | `errcheck`、`govet`、`staticcheck`、`errorlint` | lint 失敗 |
| 格式一致 | `gofmt`／`goimports`（golangci-lint formatters） | lint 失敗 |
| 安全基本檢查 | `gosec`（排除已審查的 `G204` 於 gitlog） | lint 失敗 |
| go.mod 新增依賴 | archtest 檢查 go.mod `require` 只含白名單（pflag、go-runewidth 及其遞移依賴） | `go test` 失敗 |
| `internal/testutil` 只能被測試 import | archtest | 失敗 |
| 每條 Must AC 至少一個測試 | 測試函式命名含 `AC\d\d`（例如 `TestE2E_AC13_InvalidDateFormat`）；archtest 掃描並比對 AC 清單（清單 `test/e2e/ac_list.txt` 由 QA 的 TEST_PLAN 決定內容） | `go test` 失敗 |
| 型別防呆 | `model.Format` 只能經 `ParseFormat` 建立；`model.DateRange` 欄位不匯出、只能經建構函式建立 | 編譯失敗 |

golangci-lint 版本固定（v2 系列，確切版號由 SRE 在 O-xxx 任務決定並寫入 CI）。

## 12. 測試策略
| 層級 | 位置 | 內容 | 外部依賴 |
|---|---|---|---|
| 單元測試 | 各 package `*_test.go` | `daterange`（所有格式錯誤與不存在日期、DST 邊界）、`stats`（AC-02～AC-06、AC-10～AC-12、AC-18、AC-19，使用手寫 `[]model.Commit`）、`render`（golden file，含中文、emoji、逗號、引號）、`gitlog` 解析器（假的 git 輸出 bytes） | 無 |
| Fuzz | `internal/gitlog`、`internal/daterange` | 任意輸入不得 panic；CSV 輸出以 `encoding/csv` 讀回必須與輸入相同 | 無 |
| 整合測試 | `internal/gitlog` | 用 fixture repo 驗證真正的 git 輸出被正確解析（改名、binary、merge、空 repo、子資料夾、未合併分支） | 本機 git |
| 端對端（E2E） | `test/e2e` | `go build` 出執行檔後以子程序執行，逐條驗證 AC 的 stdout／stderr／exit code／檔案內容 | 本機 git |
| 架構測試 | `internal/archtest` | 第 11 節 | 無 |
| 安裝／發佈 | CI（SRE）＋ G4 人工 | AC-37、AC-38、M2 | GitHub |

### 12.1 Fixture repo 做法
- **測試當下建立，不把 `.git` 資料夾 commit 進 repo**（二進位、難審查、跨平台換行問題）。
- `internal/testutil/fixture` 提供宣告式 API（示意）：
```go
r := fixture.New(t)                    // t.TempDir() 中 git init -b main，隔離環境
r.Commit(fixture.C{Name: "Alice", Email: "alice@example.com",
    AuthorDate: "2026-01-10T10:00:00+08:00", CommitDate: "2026-02-10T10:00:00+08:00",
    Write: map[string]string{"a.txt": "1\n2\n3\n"}, Binary: map[string][]byte{"i.png": png}, Rename: map[string]string{}, Delete: []string{}})
r.Branch("feature"); r.Checkout("feature"); r.Checkout("main"); r.Merge("feature", fixture.C{...})
```
- **隔離**：子程序環境設 `GIT_CONFIG_NOSYSTEM=1`、`GIT_CONFIG_GLOBAL=<空檔>`、`HOME=<暫存目錄>`、`GIT_AUTHOR_*`／`GIT_COMMITTER_*`（名字、email、日期全部固定）；repo 內設 `core.autocrlf=false`、`commit.gpgsign=false`，確保 Windows runner 上行數不被換行轉換影響。
- **時區**：產品程式透過參數接收 `*time.Location`；E2E 測試以環境變數 `TZ=Asia/Taipei` 啟動執行檔（Windows 上 Go 也讀 `TZ`；需在 `main` import `time/tzdata` 才能在無時區資料庫的 Windows 解析 `Asia/Taipei`——此行為**未驗證**，列入 PoC A-001）。
- **情境內容**（哪些作者、幾個 commit、預期數字）由 QA 在 TEST_PLAN 定義（PRD M1）；Architect 只定義建置機制。
- 大型效能 fixture（1,000／10,000 commits）以 `git fast-import` 產生（數秒內），放在 build tag `perf` 的測試中，不在每次 `go test` 執行。

### 12.2 指令（CLAUDE.md 同步）
`go build ./...`、`go vet ./...`、`go test ./...`、`golangci-lint run`；整體目標 < 2 分鐘（未驗證）。

## 13. 目錄結構
```
.
├── CLAUDE.md                     # Claude Code 規則（Architect）
├── README.md                     # 安裝與使用（含 Git 需求與最低版本）
├── go.mod / go.sum
├── .golangci.yml                 # lint 規則（第 11 節）
├── .goreleaser.yaml              # 發佈設定（ADR-003）
├── .github/workflows/            # CI 與 release（SRE）
├── cmd/git-stats/main.go
├── internal/
│   ├── cli/        apperr/   daterange/   gitlog/   stats/
│   ├── render/     model/    version/
│   ├── testutil/fixture/
│   └── archtest/
├── test/e2e/                     # 端對端測試與 ac_list.txt
├── scripts/install.sh            # macOS/Linux 安裝腳本
├── scripts/install.ps1           # Windows 安裝腳本
├── poc/                          # 不存在於 main；只在 poc/A-xxx 分支
└── docs/                         # PRD、ARCHITECTURE、adr/、tasks/、reports/ 等
```
檔名：Go 檔案 `snake_case.go`；測試 `*_test.go`；golden file 放各 package 的 `testdata/`。

## 14. 安裝與發佈（細節見 ADR-003）
- **觸發**：在 main 上打 tag `vX.Y.Z` 並推送 → GitHub Actions 執行 GoReleaser → 產生 GitHub Release（AC-38）。打 tag 屬 G4 發佈動作，需老闆核准。
- **產物**：`git-stats_<版本>_<os>_<arch>.tar.gz`（linux/darwin，amd64/arm64）、`.zip`（windows/amd64）、`checksums.txt`（SHA-256）。`CGO_ENABLED=0` 靜態編譯。
- **版本號**：`-ldflags "-X <module>/internal/version.Version={{.Version}}"`，GoReleaser 的 `.Version` 為 tag 去掉 `v`（AC-34）。
- **安裝（README 與 Release 說明都要寫，AC-37）**：
  - macOS／Linux：`curl -fsSL https://raw.githubusercontent.com/hansheng08/git-stats-cli/main/scripts/install.sh | sh`
    安裝到 `/usr/local/bin`（需要時用 `sudo`，可用 `INSTALL_DIR` 改位置），下載後驗證 SHA-256。
  - Windows（PowerShell）：`irm https://raw.githubusercontent.com/hansheng08/git-stats-cli/main/scripts/install.ps1 | iex`
    安裝到 `%LOCALAPPDATA%\Programs\git-stats\` 並加入使用者 PATH，驗證 SHA-256。
  - 手動：從 Release 下載對應壓縮檔，解壓縮後把執行檔放進 PATH（「下載後 1 個步驟」的備案）。
- **不上架 PyPI、npm、Homebrew 官方、Scoop 官方**。注意：Go 的公開模組代理（proxy.golang.org）會自動鏡像公開 repo，開發者可用 `go install` 安裝；這不是上架動作，也不寫進使用者安裝說明。
- 回滾：刪除或標記有問題的 Release 為 pre-release、重新發佈修正版（細節由 SRE 寫入 DEPLOY_SPEC／RUNBOOK）。

## 15. PRD 假設 A-01 評估：「使用者已裝 Git」
**結論：可接受，附條件。**
- 理由：目標使用者「手上有 Git repo」（PRD 第 2 節），實務上幾乎都已安裝 Git；ADR-002 選擇呼叫 git CLI，正是為了讓統計結果與使用者熟悉的 `git log` 一致，並避免引入大型 Git 函式庫。
- 條件（已納入本架構）：
  1. 最低版本 **Git 2.34**（Ubuntu 22.04 內建 2.34.1、24.04 內建 2.43；macOS Xcode Command Line Tools 與 Git for Windows 目前版本更新），README 明寫需求與下載連結。旗標相容性**未驗證**。
  2. 找不到 git 或版本過舊 → exit 1，訊息說明怎麼安裝（不在 PRD 的 AC 中，屬 Q-07「執行期錯誤」的延伸，建議 PM／QA 補一條 AC，見 P-01）。
  3. macOS 未裝 Command Line Tools 時，`/usr/bin/git` 是一個會跳出安裝視窗的替身程式，執行會失敗並跳窗；本工具照常回報 exit 1（**未驗證**實際行為）。
  4. AC-37 的乾淨機器前提已寫明「已安裝 Git」，M2「≤ 5 分鐘」不含安裝 Git 時間——此解讀**不確定**，見 P-02。

## 16. AC 與模組對照表
| AC | 主要負責模組 | 輔助／驗證 |
|---|---|---|
| AC-01 基本統計 | `stats` | `gitlog`（numstat）、`render`、E2E |
| AC-02 同 email 合併 | `stats` | 單元＋E2E |
| AC-03 合併後名字 | `stats` | 單元＋E2E |
| AC-04 email 大小寫 | `stats` | 單元＋E2E |
| AC-05 預設排除 merge | `gitlog`（`--no-merges`）、`cli`（預設值） | `stats`、E2E |
| AC-06 納入 merge | `cli`（旗標）、`stats`（行數強制 0） | `gitlog`（`--diff-merges=off`） |
| AC-07 binary | `gitlog`（解析 `-`） | 整合測試 |
| AC-08 改名 | `gitlog`（`-M`、改名 token 解析） | 整合測試 |
| AC-09 HEAD 範圍 | `gitlog`（`log HEAD`） | 整合測試 |
| AC-10 起訖包含 | `daterange`、`stats` | 單元＋E2E |
| AC-11 本地時區 | `daterange`（`loc` 參數）、`cmd/git-stats`（傳入 `time.Local`、import `time/tzdata`） | E2E（`TZ=Asia/Taipei`） |
| AC-12 author date | `gitlog`（`%at`、不用 `--since`）、`stats` | 整合測試 |
| AC-13 日期格式錯誤 | `daterange` | `cli`（exit 2、訊息） |
| AC-14 不存在的日期 | `daterange`（`time.Parse` 嚴格＋回轉比對） | `cli` |
| AC-15 起日晚於迄日 | `daterange` | `cli` |
| AC-16 缺日期 | `cli` | — |
| AC-17 表格欄位對齊 | `render`（Table） | golden |
| AC-18 排序 | `stats` | — |
| AC-19 次排序與穩定 | `stats` | E2E 連跑 10 次 |
| AC-20 空資料表格 | `render`（Table） | `cli` |
| AC-21 全形對齊（Should） | `render`（go-runewidth） | golden |
| AC-22 CSV 格式 | `render`（CSV） | golden |
| AC-23 CSV 跳脫 | `render`（`encoding/csv`） | fuzz 讀回 |
| AC-24 CSV 編碼 | `render`、`gitlog`（`--encoding=UTF-8`） | E2E bytes 比對 |
| AC-25 空資料 CSV | `render`（CSV） | — |
| AC-26 `--output`（Should） | `cli`（緩衝＋寫檔） | E2E |
| AC-27 路徑不存在 | `gitlog.Open`（`os.Stat`）、`cli` | E2E |
| AC-28 不是 repo | `gitlog.Open`（`rev-parse`） | E2E |
| AC-29 子資料夾 | `gitlog`（不加 pathspec） | E2E |
| AC-30 空 repo | `gitlog`（`rev-parse --verify HEAD`） | `render` |
| AC-31 缺路徑 | `cli` | — |
| AC-32 未知格式 | `cli`（`model.ParseFormat`） | — |
| AC-33 `--help` | `cli` | — |
| AC-34 `--version` | `version`、`cli`、GoReleaser ldflags | G4 |
| AC-35 唯讀 | `gitlog`（指令白名單、`GIT_OPTIONAL_LOCKS=0`） | E2E |
| AC-36 stdout/stderr 分流 | `cli`（緩衝、單一輸出出口） | E2E、forbidigo |
| AC-37 單一指令安裝 | `scripts/install.sh`、`scripts/install.ps1`、README | SRE CI／G4 人工 |
| AC-38 GitHub Release | `.goreleaser.yaml`、`.github/workflows/release` | SRE／G4 |

38 條 AC 全部有對應模組。NFR 對應：NFR-01/02 → 第 10 節；NFR-03 → 第 14 節產物＋CI 矩陣；NFR-04/05/08 → 第 9 節；NFR-06 → 第 8 節；NFR-07 → 第 4.2 節 `-z`／`--encoding`、第 7 節。

## 17. 風險與未決事項
| # | 風險 | 處理 |
|---|---|---|
| R-01 | 最低 git 版本 2.34 是否支援所有使用的旗標（`--no-use-mailmap`、`--diff-merges=off` 等） | 未驗證；PoC A-001 或 CI 以 2.34 跑整合測試 |
| R-02 | 無法用 git `--since` 提前停止，10k commits 效能 | 未驗證；PoC A-001 量測 |
| R-03 | partial clone 自動抓物件違反 NFR-04；`GIT_NO_LAZY_FETCH` 舊版 git 不支援 | 不確定；README 註明 partial clone 限制 |
| R-04 | Windows 上 `TZ`、`time/tzdata`、非 ASCII 路徑、Defender/SmartScreen 對未簽章執行檔的警告 | 未驗證；PoC A-001（Windows 部分）＋G4 實測 |
| R-05 | CSV 公式注入 | 依 AC-23 不跳脫；見 P-06 |
| R-06 | 與 npm 套件 `git-stats` 撞名 | 建議維持名稱，見第 3 節 |

## 18. PRD 問題（不修改需求，回報幕僚長轉 PM）
- **P-01**：沒有「找不到 git／git 版本過舊」的 AC。本架構以 exit 1 處理，建議補 AC。
- **P-02**：M2「安裝到第一次出結果 ≤ 5 分鐘」是否包含安裝 Git，不明確。
- **P-03**：PRD 4.2 寫「不支援 `.mailmap`」，但 git 預設會套用 `.mailmap`。本架構以 `--no-use-mailmap` **明確忽略**，讓結果與 AC-02～AC-04 一致；若使用者以 `git shortlog`（會套用 mailmap）核對 M3，數字可能不同，README 需說明。請 PM 確認「忽略」是期望行為。
- **P-04**：未定義 bare repo（無工作目錄）與 detached HEAD 的行為。本架構：detached HEAD 照常統計 HEAD；bare repo 依 `rev-parse --is-inside-work-tree` 結果視為「不是 Git repo」（exit 1）——建議 PM 確認或改為支援。
- **P-05**：CSV 換行字元（LF 或 CRLF）未規定；本架構固定 LF。RFC 4180 寫的是 CRLF，但 AC-23 只引用其引號規則。
- **P-06**：CSV 公式注入風險與 AC-23「完全一致讀回」互相牽制；目前依 AC-23 不跳脫。
- **P-07**：AC-35 寫「以任意合法參數執行」，若 `--output` 指向 repo 內，`git status --porcelain` 必然改變。建議把 AC-35 限定為「不含 `--output` 或輸出到 repo 外」。
- **P-08**：AC-37 只寫 Windows x86_64；NFR-03 未提 Windows on ARM、Linux arm64。本架構額外產出 linux/arm64（零成本），不產 windows/arm64（未驗證，需求未要求）。
