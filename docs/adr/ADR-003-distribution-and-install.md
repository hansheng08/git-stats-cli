# ADR-003：發佈與安裝方式

- 狀態：已接受（Accepted，老闆於 2026-11-22 00:44 G2 核准）；實際發佈屬 G4，仍需老闆核准
- 日期：2026-11-22
- 作者：Architect（交接單 ARCH-001）

## 背景
老闆已定：G4 發佈＝GitHub tag＋GitHub Release，**不上 PyPI／npm**；Windows 11、macOS（最新兩個大版本，含 Apple Silicon）、Ubuntu 22.04／24.04 都要「1 行指令，或下載 Release 檔後 1 個步驟」安裝，開新終端機就能執行 `git-stats --version`（AC-34、AC-37、AC-38、NFR-03）。M2 要求從零到第一次出結果 ≤ 5 分鐘。

## 選項比較
| 標準 | A. GoReleaser 產出多平台壓縮檔＋官方安裝腳本（sh／PowerShell） | B. 只提供 Release 壓縮檔，使用者手動解壓放進 PATH | C. 自建 Homebrew tap＋Scoop bucket | D. `go install` |
|---|---|---|---|---|
| 1 行指令安裝（三平台） | ◎ | △ Windows 改 PATH 不只 1 步 | ○ 但使用者需先裝 Homebrew／Scoop | ✕ 需先裝 Go |
| 只用 GitHub tag＋Release | ◎ | ◎ | △ 需額外 tap／bucket repo | △ 走 Go module proxy |
| 維護成本 | ○ 兩支腳本＋一份設定 | ◎ | △ 多兩個 repo | ◎ |
| 安全（完整性驗證） | ○ 腳本驗 SHA-256 | △ 靠使用者自行驗 | ○ | ○ |
| 新終端機立即可用 | ○（見後果） | △ | ◎ | △ 需 `$GOPATH/bin` 在 PATH |

## 決定
採用 **A**，B 作為 README 中的手動備案。
- tag `vX.Y.Z` 推送 → GitHub Actions 執行 GoReleaser（`CGO_ENABLED=0`）→ Release 附 `linux_amd64`、`linux_arm64`、`darwin_amd64`、`darwin_arm64`（`.tar.gz`）、`windows_amd64`（`.zip`）、`checksums.txt`，Release 說明含變更摘要與三平台安裝指令（AC-38）。
- 版本以 ldflags 注入 `internal/version.Version`，值為 tag 去掉 `v`（AC-34）。
- macOS／Linux：`curl -fsSL https://raw.githubusercontent.com/hansheng08/git-stats-cli/main/scripts/install.sh | sh` → 偵測 OS／架構、下載最新 Release、驗證 SHA-256、安裝到 `/usr/local/bin`（不可寫時用 `sudo`；`INSTALL_DIR` 可覆寫）。
- Windows：`irm https://raw.githubusercontent.com/hansheng08/git-stats-cli/main/scripts/install.ps1 | iex` → 下載、驗證 SHA-256、解壓到 `%LOCALAPPDATA%\Programs\git-stats\`、加入使用者 PATH（不需系統管理員）。
- 兩支腳本只使用系統內建工具（sh、curl、tar、shasum／sha256sum；PowerShell 5.1 內建 cmdlet），不需要額外安裝。
- 腳本、GoReleaser 設定、release workflow 都是程式／CI 檔案，由 SRE／TechLead 以任務單交給 Claude Code 實作，Architect 不直接撰寫。

## 後果
- 好處：完全符合老闆的發佈限制；每平台 1 行指令。
- 壞處與風險（皆**未驗證**，列入 PoC A-001 或 G4 實測）：
  - 執行檔未做程式碼簽章：Windows SmartScreen／Defender 可能警告；macOS 以 curl 下載不會加 quarantine 屬性，一般不觸發 Gatekeeper，但瀏覽器下載的手動備案會觸發，README 需說明 `xattr -d com.apple.quarantine`。程式碼簽章（Apple Developer ID、Windows 憑證）有費用；**老闆 G2 決定不做**。未簽章的實際警告行為由 PoC A-001 與 G4 實測記錄，README 需寫處理方式。
  - `curl | sh`／`irm | iex` 模式需使用者信任腳本來源；以固定到 main 分支、腳本內驗證 checksum 緩解。
  - Windows 改使用者 PATH 後，只有「新開的」終端機生效（符合 AC-37 敘述）。
- 回滾：將有問題的 Release 改為 pre-release 或下架，安裝腳本預設抓 `latest`，即自動回到前一版；細節由 SRE 寫入 DEPLOY_SPEC／RUNBOOK。刪除 Release／tag 屬刪除動作，需老闆核准。
