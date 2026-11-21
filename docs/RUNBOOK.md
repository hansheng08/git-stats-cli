# git-stats-cli 維運手冊（RUNBOOK）

- 文件擁有者：SRE
- 版本：v0.1（O-001）
- 最後更新：2026-11-22
- 適用：雲端電腦（Bot 共用的 Linux box）與 GitHub Actions CI

> 名詞說明
> - **雲端電腦（box）**：所有 Bot 共用的一台 Linux 機器，專案副本在 `/workspace/projects/`。
> - **PATH**：系統尋找指令的資料夾清單；工具放進 PATH 裡的資料夾，才能直接打名字執行。
> - **checksum（SHA-256）**：檔案指紋，用來確認下載的檔案沒被竄改。
> - **git worktree**：同一個 repo 開出多個獨立的工作資料夾，各自 checkout 不同分支，互不干擾。
> - **staged（已暫存）**：用 `git add` 放進「下一個 commit 會包含」清單的變更。

## 1. 雲端電腦需要的工具

### 1.1 已安裝版本（2026-11-22 由 SRE 安裝並驗證）
環境：Debian GNU/Linux 13 (trixie)、x86_64、使用者 `box`（可 `sudo`）。全部裝在 `/usr/local/bin`（已在預設 PATH），所有 Bot 共用。

| 工具 | 版本 | 用途 | 位置 | 驗證指令 |
|---|---|---|---|---|
| go | 1.27.4 | 建置、測試 | `/usr/local/go`（`/usr/local/bin/go` → `/usr/local/go/bin/go`） | `go version` |
| git | 2.47.3 | 系統內建（未變更） | `/usr/bin/git` | `git --version` |
| golangci-lint | 2.15.0 | Lint（與 CI 相同版號） | `/usr/local/bin` | `golangci-lint --version` |
| goreleaser | 2.20.3 | 發佈設定檢查、snapshot 建置（T-017 Review） | `/usr/local/bin` | `goreleaser --version` |
| shellcheck | 0.11.0 | `scripts/install.sh` 檢查（T-018） | `/usr/local/bin` | `shellcheck --version` |
| pwsh（PowerShell） | 7.7.0 | `scripts/install.ps1` 檢查（T-019） | `/opt/microsoft/powershell/7`（`/usr/local/bin/pwsh` 為連結） | `pwsh --version` |
| PSScriptAnalyzer | 1.25.0 | PowerShell 靜態檢查（`Invoke-ScriptAnalyzer`） | `/usr/local/share/powershell/Modules`（AllUsers） | `pwsh -NoProfile -Command '(Get-Module -ListAvailable PSScriptAnalyzer).Version.ToString()'` |
| govulncheck | 1.10.0 | 相依套件漏洞掃描（與 CI 相同） | `/usr/local/bin` | `govulncheck -version` |
| gitleaks | 8.30.1 | 機密外洩掃描（與 CI 相同） | `/usr/local/bin` | `gitleaks version` |
| actionlint | 1.7.12 | GitHub Actions workflow 檢查（O-001 Review） | `/usr/local/bin` | `actionlint -version` |

> Go 版本說明：ARCHITECTURE §2 規定「最新穩定版，建立 go.mod 時固定」。2026-11-22 go.dev 最新穩定版為 go1.27.4，box 原有 1.27.3，已升級。go.mod 由 TechLead（T-001）決定，建議 `go 1.27`（或加 `toolchain go1.27.4`）；CI 只讀 go.mod。若 go.mod 指定的版本比 box 新，`go` 會依 `GOTOOLCHAIN=auto` 自動下載對應 toolchain（需要網路）。

### 1.2 重新安裝指令（box 重置後照抄執行）
每段都會先驗 SHA-256，驗證失敗就停止。需要 `sudo`、`curl`、`tar`、`xz-utils`、`python3`。
```bash
set -euo pipefail
mkdir -p /tmp/sre-install && cd /tmp/sre-install

# --- Go 1.27.4（官方 tarball；checksum 取自 go.dev 的 JSON） ---
GO_VER=1.27.4
curl -fsSLO https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz
GO_SHA=$(curl -fsSL 'https://go.dev/dl/?mode=json&include=all' | python3 -c "import sys,json;print([f['sha256'] for r in json.load(sys.stdin) for f in r['files'] if f['filename']=='go${GO_VER}.linux-amd64.tar.gz'][0])")
echo "${GO_SHA}  go${GO_VER}.linux-amd64.tar.gz" | sha256sum -c -
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VER}.linux-amd64.tar.gz
sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go && sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# --- golangci-lint 2.15.0（官方 GitHub Release 執行檔） ---
GCL=2.15.0
curl -fsSLO https://github.com/golangci/golangci-lint/releases/download/v${GCL}/golangci-lint-${GCL}-linux-amd64.tar.gz
curl -fsSLO https://github.com/golangci/golangci-lint/releases/download/v${GCL}/golangci-lint-${GCL}-checksums.txt
grep " golangci-lint-${GCL}-linux-amd64.tar.gz\$" golangci-lint-${GCL}-checksums.txt | sha256sum -c -
tar xzf golangci-lint-${GCL}-linux-amd64.tar.gz
sudo install -m 0755 golangci-lint-${GCL}-linux-amd64/golangci-lint /usr/local/bin/golangci-lint

# --- goreleaser 2.20.3（官方 GitHub Release） ---
GR=2.20.3
curl -fsSLO https://github.com/goreleaser/goreleaser/releases/download/v${GR}/goreleaser_Linux_x86_64.tar.gz
curl -fsSLO https://github.com/goreleaser/goreleaser/releases/download/v${GR}/checksums.txt
grep ' goreleaser_Linux_x86_64.tar.gz$' checksums.txt | sha256sum -c -
mkdir -p gr && tar xzf goreleaser_Linux_x86_64.tar.gz -C gr
sudo install -m 0755 gr/goreleaser /usr/local/bin/goreleaser

# --- shellcheck 0.11.0（官方 GitHub Release；官方未提供 checksum 檔，見 §1.3） ---
SC=0.11.0
curl -fsSLO https://github.com/koalaman/shellcheck/releases/download/v${SC}/shellcheck-v${SC}.linux.x86_64.tar.xz
tar xJf shellcheck-v${SC}.linux.x86_64.tar.xz
sudo install -m 0755 shellcheck-v${SC}/shellcheck /usr/local/bin/shellcheck

# --- PowerShell 7.7.0（官方 GitHub Release tarball；hashes.sha256 為 UTF-16 編碼） ---
PS=7.7.0
curl -fsSLO https://github.com/PowerShell/PowerShell/releases/download/v${PS}/powershell-${PS}-linux-x64.tar.gz
curl -fsSLO https://github.com/PowerShell/PowerShell/releases/download/v${PS}/hashes.sha256
iconv -f UTF-16 -t UTF-8 hashes.sha256 | grep " \*powershell-${PS}-linux-x64.tar.gz" | sed 's/ \*/  /' | sha256sum -c -
sudo mkdir -p /opt/microsoft/powershell/7
sudo tar xzf powershell-${PS}-linux-x64.tar.gz -C /opt/microsoft/powershell/7
sudo chmod +x /opt/microsoft/powershell/7/pwsh
sudo ln -sf /opt/microsoft/powershell/7/pwsh /usr/local/bin/pwsh
sudo pwsh -NoProfile -Command 'Set-PSRepository PSGallery -InstallationPolicy Trusted; Install-Module PSScriptAnalyzer -RequiredVersion 1.25.0 -Scope AllUsers -Force'

# --- govulncheck 1.10.0、actionlint 1.7.12（go install；由 Go module checksum 資料庫驗證） ---
GOBIN=/tmp/sre-install/bin go install golang.org/x/vuln/cmd/govulncheck@v1.10.0
GOBIN=/tmp/sre-install/bin go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
sudo install -m 0755 /tmp/sre-install/bin/govulncheck /tmp/sre-install/bin/actionlint /usr/local/bin/

# --- gitleaks 8.30.1（官方 GitHub Release） ---
GL=8.30.1
curl -fsSLO https://github.com/gitleaks/gitleaks/releases/download/v${GL}/gitleaks_${GL}_linux_x64.tar.gz
curl -fsSLO https://github.com/gitleaks/gitleaks/releases/download/v${GL}/gitleaks_${GL}_checksums.txt
grep " gitleaks_${GL}_linux_x64.tar.gz\$" gitleaks_${GL}_checksums.txt | sha256sum -c -
mkdir -p gl && tar xzf gitleaks_${GL}_linux_x64.tar.gz -C gl
sudo install -m 0755 gl/gitleaks /usr/local/bin/gitleaks
```

### 1.3 安裝注意事項
- shellcheck 官方 Release 沒有附 checksum 檔，只能信任 GitHub Release 來源（**不確定**是否有官方簽章可驗；目前未驗）。
- golangci-lint 的 checksums 檔還列了 `.sbom.json`，所以 `grep` 必須用 `$` 精確比對檔名，否則 `sha256sum -c` 會因找不到 sbom 檔而失敗。
- PowerShell 的 `hashes.sha256` 是 UTF-16 編碼，需先 `iconv` 轉碼。
- `Install-Module -Scope AllUsers` 要 `sudo pwsh`；一般使用者執行會出現「Administrator rights are required」。
- 不要用 `curl … | sh` 安裝工具；一律下載檔案、驗 checksum、再安裝。

### 1.4 一次驗證全部工具
```bash
go version && golangci-lint --version && goreleaser --version | grep GitVersion && shellcheck --version | grep ^version \
 && pwsh --version && pwsh -NoProfile -Command '(Get-Module -ListAvailable PSScriptAnalyzer).Version.ToString()' \
 && govulncheck -version | grep Scanner && gitleaks version && actionlint -version | head -1 && git --version
```
預期依序印出：go1.27.4、2.15.0、2.20.3、0.11.0、PowerShell 7.7.0、1.25.0、govulncheck@v1.10.0、8.30.1、v1.7.12、2.47.3。

### 1.5 版本升級規則
- 工具版號以本表與 `.github/workflows/ci.yml` 為準，兩者必須一致。升級由 SRE 開 O-xxx 任務單同時改 ci.yml，並更新本節。
- Go 版本只看 go.mod（TechLead 擁有）；go.mod 升級後，SRE 依 §1.2 更新 box。

## 2. 共用雲端副本的 Git 使用規則

### 2.1 背景（事故）
曾發生：QA 在共用副本 `/workspace/projects/git-stats-cli` commit 時，把 Architect 已 `git add`（staged）但尚未 commit 的檔案一起帶進 QA 的 commit。原因是所有 Bot 共用同一個工作資料夾與同一個 index（暫存區），再加上用了 `git add -A`／`git commit -a` 或沒先檢查 `git diff --cached`。另外多個 Bot 同時操作同一副本，也可能遇到 `git pull` 暫時失敗（2026-11-22 曾出現一次 `fatal: Cannot fast-forward to multiple branches.`，重試即正常，**不確定**原因，推測為同時操作）。

### 2.2 規則（所有 Bot 都要遵守）
1. **開工前先看狀態**：`git -C <資料夾> status` 與 `git worktree list`。看到不是自己的變更，**不要碰**（不 add、不 stash、不 reset、不 checkout 覆蓋），回報幕僚長。
2. **每個 Bot 用自己的 worktree**，不要在共用的 `/workspace/projects/git-stats-cli` 直接改檔或 commit。共用副本只當「主倉庫」用來 fetch。
   ```bash
   cd /workspace/projects/git-stats-cli
   git fetch origin
   # 新分支（從最新 main 建立；--no-track 避免新分支追蹤 origin/main，誤推到 main）
   git worktree add --no-track -b task/<編號> ../git-stats-cli-<bot> origin/main
   # 既有遠端分支
   git worktree add ../git-stats-cli-<bot> task/<編號>      # 若本地無此分支，git 會自動從 origin/task/<編號> 建立
   ```
   `<bot>` 用角色英文小寫：`sre`、`qa`、`techlead`、`architect`、`pm`、`cos`。同一分支不能同時被兩個 worktree checkout（git 會拒絕），這正好防止兩個 Bot 改同一分支。
3. **只 `git add` 明確路徑**：`git add docs/RUNBOOK.md docs/DEPLOY_SPEC.md`。禁止 `git add -A`、`git add .`、`git add -u`、`git commit -a`。
4. **commit 前檢查暫存區**：`git diff --cached --stat`，確認清單只有自己的檔案，再 commit。
5. **push 明確寫分支**：`git push -u origin task/<編號>`；文件推 main 時（憲章 §4）用 `git push origin HEAD:main`，且推之前 `git diff --stat origin/main..HEAD` 確認只含 `docs/`（與 `CLAUDE.md`）。
6. **絕不** `git stash`、`git reset`、`git checkout -- <檔>`、`git clean` 別人的變更；絕不 force push。
7. **收工**：自己的 worktree 可保留重用（下次先 `git pull --ff-only`）。移除 worktree 屬刪除動作，需老闆核准；只用 `git worktree remove ../git-stats-cli-<bot>`（不要用 `rm -rf`）。
8. 雲端副本的 `git config` 屬共用設定，不要改（例如 `user.name`、`pull.rebase`）。

### 2.3 共用設定檔的擁有權（依 T-001 任務單）
| 檔案 | 擁有者 | 其他角色（含 SRE 的 O-xxx） |
|---|---|---|
| `go.mod`、`go.sum` | TechLead 的開發任務 | 唯讀 |
| `.golangci.yml` | T-001 建立；改規則需 Architect 審查 | 唯讀；CI 需要的設定用 workflow 參數 |
| `.gitattributes` | T-001（ARCHITECTURE §12.1.7） | 唯讀；要加規則先透過幕僚長請 Architect 改 ARCHITECTURE |
| `.gitignore` | T-001 | 只能在 T-001 合併後、從最新 main 開分支、**只在結尾追加行** |
| `.github/workflows/**`、`.claude/settings.json` | SRE（O-xxx） | 其他任務不得建立或修改 |

### 2.4 目前的 worktree
| 路徑 | 擁有者 | 分支 |
|---|---|---|
| `/workspace/projects/git-stats-cli` | 共用（主倉庫） | main |
| `/workspace/projects/git-stats-cli-sre` | SRE | task/O-001 |
| `/workspace/projects/git-stats-cli-sre-main` | SRE | docs/O-001-main（文件推 main 用） |

## 3. 常見 CI 異常（症狀 → 檢查 → 處理）

| 症狀 | 檢查步驟 | 處理方式 |
|---|---|---|
| 所有 Go job 顯示 skipped，警告 `go.mod not found` | repo 根目錄是否有 `go.mod` | 正常（bootstrap）。T-001 合併後自動生效；分支落後 main 時 rebase 到最新 main |
| `setup-go` 失敗：找不到 go 版本 | `go.mod` 的 `go`／`toolchain` 行是否為存在的版本 | 請 TechLead 修正 go.mod（版本以 go.dev 為準） |
| `lint` 失敗但本機通過 | 本機 `golangci-lint --version` 是否為 2.15.0；`.golangci.yml` 是否有未 commit 修改 | 照 §1.2 對齊版本；以 CI 版本為準 |
| 只有 `build-test (windows-latest)` 失敗，錯在行數或 golden file | 失敗訊息是否差在 `\r\n`；是否有 `.gitattributes`；fixture 是否設 `core.autocrlf=false` | ci.yml 已在 checkout 前關閉 autocrlf；fixture repo 另需在測試內設定（ARCHITECTURE §12.1），請 TechLead 修 |
| 只有 Windows 失敗，錯在時區 `Asia/Taipei` | 是否 import `time/tzdata`；是否依賴 `TZ` | 已知風險 R-04（PoC A-001）；回報幕僚長 |
| 只有 macOS 失敗，路徑相關 | `/var` 與 `/private/var` 符號連結（`t.TempDir()`） | 比較路徑前用 `filepath.EvalSymlinks` |
| 只有 `offline-test` 失敗 | 失敗訊息是否為下載模組（`GOPROXY=off`）或連線錯誤；`offline-ok` 檢查步驟是否通過 | 測試或產品程式在無網路時嘗試連線＝可能違反 NFR-04（致命），回報 SRE＋QA；若是模組未預先下載，修正 job 的暖快取步驟 |
| 只有 `build-test (git-min)` 失敗，版本斷言步驟失敗 | job log 中 `git --version` 實際值 vs `GIT_MIN` | ubuntu:22.04 套件版 git 改變或映像下架：回報 SRE 開 O-xxx 調整映像／`GIT_MIN` |
| 只有 `build-test (git-min)` 失敗，測試失敗 | 是否用了 2.34 不支援的 git 旗標（ARCHITECTURE R-01）；是否為 `dubious ownership`（safe.directory）或以 root 執行造成的權限差異 | 旗標問題＝產品缺陷（AC-39 最低版本不成立），回報 TechLead／Architect；環境問題回報 SRE |
| `vulncheck` 失敗 | 看輸出中的 `GO-YYYY-NNNN` 編號與受影響套件 | 標準庫漏洞：升級 go.mod 的 Go patch 版；第三方：升級該依賴（新增或升級依賴需 Architect 審查）；回報 SRE 評估風險 |
| `secret-scan` 失敗 | 輸出（已遮蔽）中的檔案、commit、規則名稱 | **先當作真的外洩**：立刻請老闆撤銷／更換該機密，再移除檔案；已進歷史的需老闆決定是否改寫歷史（force push 需核准）。誤報才加 `.gitleaksignore`（需 SRE 同意） |
| `scripts-lint` 失敗 | shellcheck 的 `SCxxxx` 代碼；PSScriptAnalyzer 規則名 | 修腳本（T-018／T-019）；確定誤報才加行內停用註解並寫理由 |
| `goreleaser-check` 失敗 | `goreleaser check` 輸出；本機 `goreleaser release --snapshot --clean` 重現 | 修 `.goreleaser.yaml`（T-017） |
| `ci-ok` 失敗但其他都綠 | 是否有 job 被 cancelled（新推送取消舊的）| 等最新一次執行結果；必要時手動 Re-run |
| CI 未觸發 | Actions 頁面是否被停用；分支名稱；`.github/workflows/ci.yml` 是否在該分支 | 老闆在 Settings → Actions 啟用 |
| job 卡住或逾時 | 哪個步驟耗時；是否下載依賴失敗 | Re-run；連續失敗回報 SRE |

## 4. 發佈異常
發佈與回滾步驟見 `docs/DEPLOY_SPEC.md` §6、§7。處理順序：先止血（把壞版本設為 pre-release、讓 latest 回到前一版，需老闆核准）→ 收集 release workflow log 與使用者回報 → 以 O-xxx 任務單請 Claude Code 分析 → 回報幕僚長 Root Cause、影響範圍、處理建議。

## 5. 變更紀錄
| 版本 | 日期 | 作者 | 內容 |
|---|---|---|---|
| v0.1 | 2026-11-22 | SRE | 初版（O-001）：雲端工具安裝、共用副本 git 規則、CI 異常處理 |
