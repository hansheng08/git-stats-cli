# CLAUDE.md — git-stats-cli 專案規則（給 Claude Code）

擁有者：Architect。本檔只放每次都需要知道的事；細節以 `docs/ARCHITECTURE.md` 與 `docs/adr/` 為準。
本檔與任務單衝突時，以任務單為準，但「Git 規則」與「禁止事項」任何任務單都不能覆蓋。

## 專案一句話
`git-stats`：單一執行檔的 Go CLI，呼叫使用者已安裝的 `git log`，輸出指定期間每位作者（以小寫 email 區分）的 commit 數與增刪行數，格式為表格或 CSV。需求見 `docs/PRD.md`（AC-01～AC-38）。

## 開始任何任務前
1. `git checkout main && git pull`，再依任務單「分支」一節建立或切換分支。
2. 讀任務單 `docs/tasks/<編號>.md` 列出的背景文件（本檔、PRD 段落、ARCHITECTURE 段落）。
3. 任務單沒寫清楚的事：不要猜，寫進報告的「需要人類決定的問題」。

## 指令（全部要在 repo 根目錄可執行，且三平台通用）
```
go build ./...                       # 建置
go vet ./...                         # 靜態檢查
go test ./...                        # 單元、整合、E2E、架構測試（需要本機 git ≥ 2.34）
golangci-lint run                    # Lint（版本固定於 CI 設定）
go test -run TestE2E ./test/e2e/...  # 只跑端對端測試
go test -fuzz=Fuzz -fuzztime=30s ./internal/gitlog/   # 模糊測試（需要時）
go build -o bin/git-stats ./cmd/git-stats             # 產生本機執行檔（bin/ 不可 commit）
```
若骨架尚未建立（尚無 go.mod），以任務單指示為準。

## 目錄慣例
```
cmd/git-stats/          進入點：只呼叫 cli.Run 並 os.Exit
internal/cli/           旗標解析、驗證、流程、錯誤→exit code、輸出出口
internal/daterange/     YYYY-MM-DD 嚴格解析、建立 model.DateRange
internal/gitlog/        產品程式中唯一可用 os/exec 的地方；解析 git 輸出（測試專用路徑例外見 ARCHITECTURE §11）
internal/stats/         純函式：篩選、合併作者、計數、排序
internal/render/        表格、CSV，只寫入傳入的 io.Writer
internal/model/         共用型別
internal/apperr/        錯誤種類：Usage（exit 2）、Runtime（exit 1）
internal/version/       版本號（發佈時 ldflags 注入）
internal/testutil/fixture/  測試用 fixture repo 建立工具（只能被 _test.go import）
internal/archtest/      架構測試
test/e2e/               端對端測試（編譯執行檔後以子程序執行）
scripts/                install.sh、install.ps1
docs/                   文件（只有任務單明確要求才可以寫 docs/reports/<編號>.md）
```

## 相依方向（違反會讓 go test / lint 失敗）
- `cmd/git-stats` → `cli`；`cli` → `daterange`、`gitlog`、`stats`、`render`、`model`、`apperr`、`version`
- `daterange`、`gitlog` → `model`、`apperr`；`stats`、`render` → `model`
- `model`、`apperr`、`version` 不 import 任何專案內 package
- 沒有任何 package 可以 import `cli`；`stats`／`render` 不可 import `gitlog` 或彼此
- 需要改邊界：停下來，寫進報告，交由 Architect 審查。不要自己改 archtest 或 `.golangci.yml` 來放行。

## 程式規則
- 只有 `cmd/git-stats/main.go` 可呼叫 `os.Exit`；只有 `internal/cli` 決定 exit code。
- `cli.Run` 必須 `recover()` panic 並回傳 1（Go 未攔截的 panic 會以 2 結束，與使用方式錯誤混淆）。
- Exit code：0 成功（含查無資料、--help、--version）；1 執行期錯誤；2 使用方式錯誤。錯誤用 `apperr.Usage(...)`／`apperr.Runtime(...)` 包裝。
- 錯誤訊息英文，格式 `git-stats: error: <出了什麼錯>` + 換行 + `<怎麼修正>`；字串集中在 `internal/cli/messages.go`。
- 輸出先寫入 `bytes.Buffer`，全部成功後才寫到 stdout 或 `--output` 檔案。錯誤時 stdout 必須為空；成功時 stderr 必須為空。
- 不可使用 `fmt.Print*`、`os.Stdout`、`os.Stderr`（`cmd/` 除外），一律寫入傳入的 `io.Writer`。
- 不可在 `cmd/` 以外使用 `time.Now`、`time.Local`；時區由 main 以參數傳入。
- 日期篩選用 author date（`%at`），**不可**使用 git 的 `--since`／`--until`（那是 committer date）。
- git 呼叫只能用 ARCHITECTURE.md 第 4.1 節列出的唯讀指令與旗標；不可新增會寫入 repo 的指令（`status`、`fetch`、`gc`、`config` 等都禁止）。
- 用 `exec.CommandContext(ctx, "git", args...)`，參數逐一傳入，絕不經 shell、絕不拼接字串。
- 不可新增 go.mod 依賴。目前允許：`github.com/spf13/pflag`、`github.com/mattn/go-runewidth`（及其遞移依賴）。需要新依賴：停下來寫進報告。
- 不可有網路程式碼、遙測、快取檔、設定檔。
- `//nolint` 必須寫明規則名與理由（例：`//nolint:gosec // G204: args are fixed allowlist`），並在報告中列出。

## 命名規則
- Package：小寫單字（`daterange`，不用底線或駝峰）。檔名：`snake_case.go`。
- 匯出識別字 `PascalCase`，未匯出 `camelCase`；縮寫全大寫（`CSV`、`URL`、`ID`）。
- 測試函式含 AC 編號：`TestE2E_AC13_InvalidDateFormat`、`TestAggregate_AC19_TieBreakByEmail`。
- 單元測試 golden file：`<package>/testdata/<情境>.golden`，可用 `-update` 重新產生並在報告說明原因。
- E2E 黃金檔 `test/e2e/testdata/**` 與 `EXPECTED_SHA` 由 QA 擁有：不可產生、不可修改來讓測試通過，失敗就回報（ARCHITECTURE §12.1）。
- Commit 訊息：`[ClaudeCode] <任務編號> <摘要>`（例：`[ClaudeCode] T-003 CSV 輸出`）。

## 測試規則
- 每個行為變更都要有測試；修 bug 先有失敗測試。
- fixture repo 一律用 `internal/testutil/fixture` 在 `t.TempDir()` 現場建立；不可把 `.git` 資料夾或二進位 repo commit 進 repo。
- fixture 必須隔離使用者設定（`GIT_CONFIG_NOSYSTEM=1`、`GIT_CONFIG_GLOBAL`、`HOME`、固定作者與日期、`core.autocrlf=false`）。
- 時區相關測試用 `TZ=Asia/Taipei` 等固定時區執行，不可依賴執行機器的時區。
- 測試不可連網、不可寫入 repo 目錄以外（只用 `t.TempDir()`）。
- 不可為了讓測試通過而刪除、跳過（`t.Skip`）或放寬既有測試；平台限制的 skip 必須寫原因並在報告列出。
- 不可修改 `docs/TEST_PLAN.md`（QA 擁有）。

## 禁止事項
- 不可 commit 或 push 到 `main`；不可 merge PR；不可 force push；不可刪除分支或 tag；不可打 tag 或建立 Release。
- 不可刪除檔案，除非任務單明確列出要刪除的路徑。
- 不可修改 `docs/` 底下的文件（PRD、ARCHITECTURE、adr、TASKS、TEST_PLAN 等），唯一例外是寫自己的 `docs/reports/<編號>.md`。
- 不可修改本檔（CLAUDE.md）、`.golangci.yml` 的規則、`internal/archtest`，除非任務單明確要求。
- 不可做任務單「範圍外」的事，包含順手重構、改需求、加旗標或功能。
- 不可在程式碼、測試、文件、commit 訊息中寫入密碼、Token、API Key。
- 不可在 `poc/A-xxx`、`proto/P-xxx` 分支以外寫 PoC／原型程式碼；這些分支永遠不合併到 main。

## Git 規則
- 只在任務單「分支」一節指定的分支工作：
  - 開發 T-xxx：從最新 main 建立 `task/T-xxx`；修正輪次 T-xxx-fixN：沿用 `task/T-xxx`
  - QA Q-xxx：追加在被驗證的 `task/T-xxx`；Bug 修復：`fix/Q-xxx`
  - 維運 O-xxx：`task/O-xxx`；架構 PoC A-xxx：`poc/A-xxx`；原型 P-xxx：`proto/P-xxx`
- 開工前 `git pull`；完成後在該分支 commit 並 `git push -u origin <分支>`。
- 程式碼進 main 的唯一途徑是 TechLead 開 PR → QA → CI 全綠 → 老闆按 Merge。你不開 PR，除非任務單要求。

## 完成定義（Definition of Done）
全部成立才算完成：
1. 任務單每一條驗收標準都已勾選，並引用對應 AC 編號。
2. `go build ./...`、`go vet ./...`、`go test ./...`、`golangci-lint run` 全部通過（無法執行的要在報告說明原因）。
3. 新增或修改的行為有測試；沒有新增依賴；沒有違反相依方向。
4. 只動任務單範圍內的檔案；`git status` 乾淨；已 push 到指定分支。
5. 已寫報告 `docs/reports/<編號>.md`（commit 在同一分支）。

## 回報規則（docs/reports/<編號>.md）
```
# <編號> 報告
## 完成摘要
## 變更檔案清單
## 測試結果（貼上指令與結果摘要）
## 未完成事項
## 需要人類決定的問題
## 使用的 //nolint 與 t.Skip（含理由）
```
