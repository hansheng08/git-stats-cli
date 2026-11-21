# git-stats-cli 部署／發佈規格（DEPLOY_SPEC）

- 文件擁有者：SRE
- 版本：v0.1（初稿，O-001；G4 前會再更新）
- 最後更新：2026-11-22
- 依據：docs/ARCHITECTURE.md v1.1 §9、§11、§12.1.5、§14（commit `5033d3c`）、docs/tasks/T-001.md（共用檔案規則）、docs/PRD.md v1.1（AC-34、AC-37、AC-38、AC-45）、docs/adr/ADR-003、docs/TASKS.md（T-017～T-020）、docs/TEST_PLAN.md §8

> 名詞說明
> - **發佈（Release）**：本專案是命令列工具，沒有伺服器；「部署」指的是在 GitHub 上建立一個版本（tag＋Release＋各平台執行檔）。
> - **tag**：git 裡指向某個 commit 的版本標籤，例如 `v0.1.0`。
> - **GitHub Release**：GitHub 上掛在某個 tag 下的發佈頁面，附上可下載的檔案與說明。
> - **workflow**：GitHub Actions 的自動化流程設定檔。
> - **required check（必過檢查）**：分支保護設定中「沒綠燈就不能合併」的 CI 檢查。
> - **回滾（Rollback）**：出問題時讓使用者回到上一個正常版本。
> - **checksum（SHA-256 校驗值）**：檔案的指紋，用來確認下載的檔案沒被竄改或損壞。

## 1. 發佈架構

```
開發者 ──push──▶ task/* 分支 ──PR──▶ main（分支保護：ci-ok 必過、禁止 force push）
                    │                   │
                    └─ ci.yml（每次 push／PR：build、vet、test×4 平台＋Git 2.34 容器、無網路測試、lint、govulncheck、gitleaks）
                                        │
              老闆 G4 核准後在 main 打 tag vX.Y.Z 並推送
                                        │
                     release.yml（之後的 O-xxx）──▶ GoReleaser ──▶ GitHub Release
                                                              （5 個壓縮檔＋checksums.txt）
                                        │
     使用者：curl … install.sh | sh ／ irm … install.ps1 | iex（從 main 分支取腳本，下載 latest Release、驗 SHA-256）
```

- 產物（ARCHITECTURE §14）：`git-stats_<版本>_linux_amd64.tar.gz`、`linux_arm64.tar.gz`、`darwin_amd64.tar.gz`、`darwin_arm64.tar.gz`、`windows_amd64.zip`、`checksums.txt`；`CGO_ENABLED=0`。含 linux/arm64、不含 windows/arm64（PRD AC-45，G2 定案）。
- 版本號：tag `vX.Y.Z`，GoReleaser 以 ldflags 注入 `internal/version.Version`（值不含 `v`，AC-34）。
- 不上架 PyPI、npm、Homebrew、Scoop（ADR-003）。

## 2. 環境清單

| 環境 | 是否需要 | 說明 |
|---|---|---|
| Dev | 不需要獨立環境 | 開發者本機或雲端副本；用 `go build` 自行建置 |
| CI | GitHub Actions hosted runner | `ubuntu-24.04`（NFR-01 參考機）、`ubuntu-22.04`、`macos-latest`、`windows-latest` |
| 預發佈（Staging） | 不另設 | 以 `goreleaser release --snapshot --clean`（CI 的 `goreleaser-check` job）在不發佈的情況下驗證產物；必要時可發 `vX.Y.Z-rc.N` pre-release（需 G4 核准，見 §6） |
| Prod | GitHub Release（`latest`） | 使用者安裝腳本預設抓 `latest` |

沒有伺服器、資料庫、網域，因此不需要監控告警與執行期指標；改為觀察 §8 的發佈健康訊號。

## 3. 檔案擁有者

| 檔案 | 擁有者 | 任務 |
|---|---|---|
| `.github/workflows/ci.yml` | SRE | O-001 |
| `.github/workflows/release.yml` | SRE | 之後的 O-xxx（T-017 完成後；需 G4 才會實際觸發） |
| `.claude/settings.json` | SRE | O-001 |
| `.goreleaser.yaml` | TechLead | T-017 |
| `scripts/install.sh`、`scripts/install.ps1` | TechLead | T-018、T-019 |
| `go.mod`、`go.sum`、`.golangci.yml`、`.gitattributes` | TechLead | T-001（SRE 的 O-xxx 一律唯讀） |
| `.gitignore` | TechLead | T-001（O-xxx 只能在 T-001 合併後從最新 main 開分支、在結尾追加行） |

## 4. CI/CD 流程

### 4.1 CI（`ci.yml`，O-001）
- 觸發：所有分支的 push、對 main 的 PR、手動 `workflow_dispatch`。
- Jobs（名稱固定，可作為 required check）：

| Job | 內容 | 執行條件 |
|---|---|---|
| `lint` | golangci-lint `v2.15.0`（規則以 `.golangci.yml` 為準） | 有 `go.mod` |
| `build-test (<os>)` | `go build ./...`、`go vet ./...`、`go test ./...`；4 個平台 | 有 `go.mod` |
| `build-test (git-min)` | 在 `ubuntu:22.04` 容器以套件版 git 2.34 跑完整 `go test ./...`（含 E2E），斷言 git 主次版號＝`GIT_MIN`（ARCHITECTURE §12.1.5 S-04、AC-39） | 有 `go.mod` |
| `offline-test` | 在無網路的 network namespace 以非 root 執行 `go test ./...`（NFR-04，ARCHITECTURE §12.1.5） | 有 `go.mod` |
| `vulncheck` | `govulncheck@v1.10.0 ./...`（相依套件漏洞） | 有 `go.mod` |
| `secret-scan` | gitleaks `v8.30.1` 掃完整 git 歷史 | 永遠 |
| `scripts-lint` | shellcheck（install.sh）、PSScriptAnalyzer 1.25.0（install.ps1） | 腳本存在 |
| `goreleaser-check` | `goreleaser check`＋`release --snapshot`（不發佈） | `.goreleaser.yaml` 存在 |
| `ci-ok` | 彙總：上面任一失敗或取消就失敗 | 永遠 |

- Go 版本只從 `go.mod` 讀取；工具版號固定（升級＝改 ci.yml＋RUNBOOK，經 O-xxx 任務單）。
- 第三方 Action 以 commit SHA 釘選。workflow 權限 `contents: read`，CI 不使用任何 Secret。

### 4.2 合併到 main（憲章 §4）
TechLead 開 PR → 影響產品行為的 PR 由 QA 驗證 → `ci-ok` 綠燈 → G3 老闆在 GitHub 按 Merge。

**建議老闆設定 main 分支保護（GitHub → Settings → Branches／Rulesets）**：需要 PR 才能合併、required check `ci-ok`、禁止 force push、禁止刪除分支。這是 `.claude/settings.json` deny 清單之外真正擋住誤推 main 的機制。

### 4.3 發佈（release workflow，之後的 O-xxx；本版只定規格）
- 觸發：推送符合 `v[0-9]+.[0-9]+.[0-9]+*` 的 tag。
- 步驟：checkout（`fetch-depth: 0`）→ setup-go（`go.mod`）→ 再跑一次 `go test ./...` → GoReleaser `release --clean`。
- 權限：只有 release job 給 `contents: write`（建立 Release 與上傳檔案所需）。

## 5. Secrets 清單（只寫名稱與用途，不寫值）

| 名稱 | 來源 | 用途 | 使用位置 |
|---|---|---|---|
| `GITHUB_TOKEN` | GitHub Actions 每次執行自動產生，不需手動建立 | release workflow 建立 GitHub Release、上傳產物 | `release.yml`（之後）；`ci.yml` 不使用 |

- 目前不需要其他 Secret。G2 已決定不做程式碼簽章（ARCHITECTURE v1.0），所以不需要 Apple Developer ID、Windows 憑證；若日後改變決定，屆時新增並更新本表。
- 不使用 `gitleaks/gitleaks-action`（組織帳號需要 `GITLEAKS_LICENSE`），改用 gitleaks CLI，避免多一個 Secret。
- 任何 Secret 都不得寫進 repo；CI 的 `secret-scan` 會擋。

## 6. 發佈步驟（G4 核准後才執行）

前置檢查（SRE 在 G4 前確認並回報幕僚長）：
1. main 最新 commit 的 `ci-ok` 綠燈；QA_REPORT 對應版本為通過。
2. `goreleaser-check` 在 main 上綠燈（snapshot 產物 5 個＋checksums）。
3. `CHANGELOG.md` 已有本版段落（T-020）。
4. 回滾步驟（§7）已在本版演練過或確認可執行（首次發佈建議先發 `v0.1.0-rc.1` pre-release 演練）。
5. 確認前一個正常版本的 tag（首次發佈則無；見 §7 首次發佈注意）。

發佈（老闆核准後，由老闆或指定者執行；打 tag 屬發佈動作）：
```
git checkout main && git pull --ff-only
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```
驗證：
1. GitHub Actions 的 release workflow 綠燈。
2. Release 頁面有 5 個壓縮檔＋`checksums.txt`，說明含三平台安裝指令（AC-38）。
3. 在乾淨的 Linux／macOS／Windows 各執行一次安裝腳本，開新終端機 `git-stats --version` 印出 `X.Y.Z`（AC-34、AC-37；QA 於 G4 驗）。

## 7. 回滾步驟

原則：先止血（讓使用者拿不到壞版本），再修正。**刪除 Release 或 tag 屬刪除動作，必須先取得老闆核准**；優先用不刪除的方式。

| 步驟 | 動作 | 效果 | 需核准 |
|---|---|---|---|
| R1 止血 | GitHub Release 頁面 → 編輯壞版本 → 勾選「Set as a pre-release」（或取消「Set as the latest release」）並把前一版設為 latest | GitHub 的 `releases/latest` 會指向前一個正式版；安裝腳本預設抓 latest，新安裝自動回到前一版 | 是（屬對外發佈狀態變更，G4 範圍內由老闆執行或授權） |
| R2 通知 | 在壞版本 Release 說明最上方加註「已知問題，請改用 vX.Y.(Z-1)」 | 已下載的人看得到 | 是（對外發文） |
| R3 修正 | 以 `fix/Q-xxx` 或 `task/T-xxx` 修正 → PR → G3 → 發佈新版 `vX.Y.(Z+1)`（**不重用**壞版本號） | 正式修復 | 依 G3／G4 |
| R4（選用） | 下架壞版本：刪除 Release（保留 tag 以利追查） | 完全不能下載 | 是（刪除動作） |

- 已安裝壞版本的使用者：用安裝腳本指定版本重裝前一版。**不確定**：T-018／T-019 的腳本是否支援指定版本（例如 `VERSION=v0.1.0`）；建議 TechLead 在腳本加入此參數，否則只能用 README 的手動安裝備案下載舊版。
- 首次發佈（v0.1.0）沒有「前一版」可回：R1 改為把該 Release 設為 pre-release 後，`latest` 不存在，安裝腳本會失敗而不是裝到壞版本；README 需說明。**不確定**：GitHub 在沒有任何正式 Release 時 `releases/latest` 的確切回應（預期 404），需在 rc 演練時確認。
- 不改寫 git 歷史、不 force push、不移動既有 tag。

## 8. 發佈後觀察（取代監控告警）
- GitHub Issues 中新回報的安裝失敗或數字錯誤。
- Release 下載數（GitHub 提供）只做參考。
- 安裝腳本對 latest 的依賴：每次發佈後 24 小時內，SRE 以三平台安裝腳本各跑一次確認（G4 檢查清單的一部分）。

## 9. 未決事項
1. **release workflow 的任務編號與時機**：建議 T-017 完成後由 SRE 開 O-002；需幕僚長排程。
2. **最低 Git 版本**：GitHub hosted `ubuntu-22.04` runner 會裝較新的 git，無法驗證 2.34，因此 O-001 另設 `build-test (git-min)`（`ubuntu:22.04` 容器，套件版 2.34.1，**未驗證**，以 job 的 `git --version` 為準）。PoC A-001 若調高最低版本，SRE 另開 O-xxx 同步調整 `GIT_MIN` 與容器映像。**不確定** ubuntu:22.04 映像與 hosted ubuntu-22.04 runner 會保留到何時。
3. **macOS 前一版、x86_64 macOS、Windows 11**：TEST_PLAN §8 要求，O-001 只放 `macos-latest`、`windows-latest`（Windows Server）；是否加矩陣或改人工驗證，等 QA Q-xxx 決定。
4. **效能測試（NFR-01/02）擁有者**：TASKS 問題 #6，待幕僚長指定；若寫成 CI job，固定跑在 `ubuntu-24.04`。
5. **安裝腳本是否支援指定版本**：影響回滾（§7）。
6. **程式碼簽章**：G2 已決定不做；Windows SmartScreen 警告屬已知限制（A-001 C5 會記錄）。
