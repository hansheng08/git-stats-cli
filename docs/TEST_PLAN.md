# git-stats-cli 測試計畫（TEST_PLAN）

- 文件擁有者：QA
- 版本：v0.1（初版，交接單 TP-001）
- 最後更新：2026-11-22
- 依據：`docs/PRD.md` v1.0（commit `870d766`，AC-01～AC-38、NFR-01～NFR-08；AC 內容與 `d7893ca` 相同）
- 參考：撰寫時 `docs/ARCHITECTURE.md` 尚未推上，本計畫不依賴任何實作語言或內部設計

> 名詞說明（第一次出現時的白話解釋）
> - **TC**（Test Case，測試案例）：一組「前置條件 → 步驟 → 預期結果」，可以重複執行並判定通過或失敗。
> - **fixture repo**（測試用 repo）：用固定腳本造出來、內容與預期結果事先就知道的 Git repo。
> - **黑箱測試**：只從外部（命令列參數、stdout、stderr、exit code、產生的檔案）觀察，不看程式內部怎麼寫。
> - **Edge Case**（邊界條件）：空值、最大值、剛好在界線上、重複操作等容易出錯的情況。
> - **Regression Test**（回歸測試）：確認新改動沒有把原本正常的功能弄壞。
> - **hyperfine**：一個命令列效能量測工具，會自動重複執行指令並算出平均、中位數等統計值。
> - **中位數（median）／p95**：把多次量測由小到大排好，中間那個值／第 95 百分位的值。
> - **warm cache／cold cache**：檔案已經被作業系統快取在記憶體裡（第二次以後執行）／快取被清空、必須重新從磁碟讀（第一次執行）。

---

## 1. 測試範圍與不測範圍

### 1.1 測試範圍
1. PRD 第 5 節 F1～F6 全部 AC（AC-01～AC-38，Must 36 條、Should 2 條：AC-21、AC-26）。
2. PRD 第 6 節 NFR-01～NFR-08。
3. PRD 第 3 節成功指標中可由 QA 驗證的部分：M1（統計正確率，由本計畫的 fixture 承擔）、M2（三平台安裝時間，G4 前實測）、M4（Must AC 通過率，記在 QA_REPORT）。
4. PRD 7.1 假設 A-03（作者＝author，不是 committer）、A-04（行數語意同 `git log --numstat`）、A-05（日期只接受 `YYYY-MM-DD`）衍生的邊界條件。
5. 第 5 節開頭的 exit code 約定（0／1／2）與「錯誤時 stdout 無輸出」規則，套用到所有案例。

### 1.2 不測範圍（PRD 4.2／4.3 範圍外，或非 QA 黑箱可驗證）
- JSON 輸出、分支篩選、檔案／路徑篩選、自訂排序、相對日期、含時間的日期輸入、時區選項、`.mailmap`、`Co-authored-by`、總計列／百分比／趨勢、多 repo、PyPI／npm 上架（4.2）。
  - 例外：「不支援」本身會被測（例如 `--format json` 必須報錯＝AC-32；含時間的日期必須報錯＝AC-13 邊界；`.mailmap` 存在時不得被套用＝TC-AC02-04）。
- GUI、網頁儀表板、績效評分（4.3）。
- M3（發佈後老闆在真實 repo 抽查）：由老闆回報，QA 只提供核對步驟範本（見 7.4）。
- 實作內部的單元測試內容：由 TechLead／Claude Code 在開發任務中撰寫；QA 只檢查它們存在並在 CI 通過，**不作為驗收依據**。
- NFR-05 的 Code review 部分：依獨立性原則 QA 不讀實作程式碼；QA 只做行為面驗證（無網路、無對外連線），Code review 由 TechLead 負責並在 PR 記錄（見 TC-NFR05-01）。

### 1.3 名稱約定
- 本計畫用 PRD 的建議名稱：指令 `git-stats`，選項 `--since`、`--until`、`--format`、`--include-merges`、`--output`，repo 路徑為位置參數（positional argument，直接寫在指令後面、沒有選項名稱的參數）。
- 若 Architect 在 `ARCHITECTURE.md` 改名，測試程式依對照表替換名稱，**預期結果不變**。
- 所有案例的指令以 `git-stats <repo> --since <起日> --until <迄日> [其他選項]` 表示；`<FX-xx>` 代表該 fixture repo 的路徑。

---

## 2. 測試層級分工

| 層級 | 誰寫 | 驗什麼 | 本專案的做法 | 是否作為 QA 驗收依據 |
|------|------|--------|--------------|----------------------|
| Unit（單元測試：驗證單一函式） | TechLead 任務中由 Claude Code 撰寫 | 日期解析、彙總、排序、CSV 跳脫、字寬計算等純邏輯 | 用實作語言的測試框架；QA 只確認 CI 有跑且全綠 | 否（輔助） |
| Integration（整合測試：驗證模組之間的配合） | QA 開 Q-xxx 任務單，由 Claude Code 撰寫 | 「本工具＋真實 Git＋fixture repo」這段串起來是否正確 | 在 CI 上建 fixture repo，直接執行建置好的 `git-stats` 執行檔，比對 stdout／stderr／exit code／檔案。**不呼叫內部函式**，因此與實作語言無關 | **是（主要）** |
| E2E（端對端測試：從使用者操作一路驗證到結果） | QA 執行（部分由 Claude Code 寫成 CI 流程） | 從 Release 下載／安裝 → 開新終端機 → 執行 → 得到結果；以及 PRD 情境 S1～S4 | 在乾淨的三平台機器（CI runner 或 VM）依 README 操作 | **是**（AC-33、34、37、38、M2、S1～S4） |
| 效能測試 | QA 定義，Claude Code 寫成 CI job | NFR-01、NFR-02 | 見第 6 節 | **是** |

原則：
- Integration 層一律以「黑箱」方式呼叫已建置的執行檔，測試程式可以用任何語言寫（建議 POSIX shell＋一個標準 CSV 解析器，例如 Python `csv` 模組），但**不得 import 產品程式碼**。
- 所有 Integration 測試要在 NFR-03 的 OS 矩陣上跑（Windows 11／macOS 兩版／Ubuntu 22.04、24.04）；見第 8 節。

---

## 3. 通用檢查（每個案例都要套用）

為避免每條案例重複，以下規則套用到所有 TC，任一條不符即判該 TC 失敗：

| 編號 | 適用 | 檢查內容 | 依據 |
|------|------|----------|------|
| G-1 | 所有預期成功的執行 | exit code = 0，且 **stderr 為空（0 byte）** | AC-36、Q-07 |
| G-2 | 所有預期失敗的執行 | stdout **為 0 byte**；stderr 非空；exit code 依約定（1＝執行期錯誤、2＝使用方式錯誤） | 第 5 節約定、AC-36 |
| G-3 | 所有錯誤訊息 | 英文；同時說明「出了什麼錯」與「怎麼修正」（例如附正確格式或用法），以人工檢查清單判定：①指出錯誤的參數或路徑 ②給出修正方式 | NFR-06、Q-15 |
| G-4 | 所有對 fixture 的執行 | 執行前後 fixture 的 `git status --porcelain`、`git for-each-ref`、`.git/config` 完全相同 | AC-35（在 Integration 套件中全面套用，不只 AC-35 本身） |
| G-5 | 所有執行 | 測試環境無網路（見 NFR-04）時結果不變 | NFR-04 |
| G-6 | 數字欄位 | 十進位整數、不含千分位、不含小數點 | AC-22（CSV 明訂；表格見待釐清 D-05） |

「資料列」定義：表格輸出中除表頭（與可能存在的框線／分隔線）以外、每位作者一行的內容；CSV 中除第 1 行表頭以外的每一筆紀錄。

---

## 4. 測試資料（fixture repo）

### 4.1 建置原則（確保每台機器、每次造出的 repo 完全一樣）
1. 每個 fixture 由一支 **POSIX shell 腳本 + 純 `git` 指令** 建立（Windows 用 Git for Windows 內附的 Git Bash 執行），不依賴產品實作語言。腳本放在 `tests/fixtures/`（由 Q-xxx 任務單交給 Claude Code 撰寫）。
2. 腳本開頭固定環境，隔離使用者自己的 Git 設定：
   ```sh
   export HOME="$TMP_HOME"            # 空的暫存家目錄，避免讀到 ~/.gitconfig
   export GIT_CONFIG_NOSYSTEM=1
   export GIT_CONFIG_GLOBAL=/dev/null # Windows 用 NUL 或空檔
   export LC_ALL=C.UTF-8 TZ=Asia/Taipei
   git -c init.defaultBranch=main init "$DIR"
   git -C "$DIR" config core.autocrlf false
   git -C "$DIR" config commit.gpgsign false
   ```
3. 每個 commit 都明確設定 6 個環境變數，日期一律用含時區偏移的 ISO 8601（例如 `2026-01-10T12:00:00+08:00`）：
   `GIT_AUTHOR_NAME`、`GIT_AUTHOR_EMAIL`、`GIT_AUTHOR_DATE`、`GIT_COMMITTER_NAME`、`GIT_COMMITTER_EMAIL`、`GIT_COMMITTER_DATE`。
   除非案例特別要求（AC-12、A-03），committer 名字／email／日期與 author 相同。
4. 檔案內容由腳本以固定文字產生（例如 `line 1`…`line N`，LF 換行、最後一行有換行），binary 檔用固定 byte 序列（含 NUL byte，例如 PNG 檔頭 `\x89PNG\r\n\x1a\n` + 固定 64 byte），確保 Git 判定為 binary。
5. 建置完成後，腳本輸出 HEAD 的 commit SHA。因為內容與時間都固定，**同一腳本在任何機器上造出的 SHA 必須相同**；首次建置後把 SHA 記在 `tests/fixtures/EXPECTED_SHA`，之後每次建置都比對，不同就代表 fixture 不可信，該輪測試判為「無法驗證」而不是通過。
6. 每個 fixture 附一份「預期結果表」（下方 4.3），由 QA 依 PRD 手算，不是跑工具得到的。
7. 測試時一律設定 `TZ=Asia/Taipei`（Windows 見 TC-AC11-03），除非該案例另有指定。

### 4.2 Fixture 清單

| ID | 名稱 | 內容摘要 | 使用的 TC |
|----|------|----------|-----------|
| FX-01 | basic | A、B 兩位作者；期間內外各有 commit；含 `src/` 子資料夾 | AC-01、17、22、26、29、33～36、Regression |
| FX-02 | identity | 同 email 不同名字、email 大小寫不同、author date 與 commit 順序相反 | AC-02～04 |
| FX-03 | merges | `--no-ff` merge、含衝突解決內容的 merge、octopus merge（3 個父 commit） | AC-05、06 |
| FX-04 | binary-rename | binary 新增／修改／刪除、純改名、跨資料夾改名、改名＋小改 | AC-07、08 |
| FX-05 | branches | main＋未合併的 feature 分支、detached HEAD | AC-09 |
| FX-06 | dates-tz | 日期邊界、跨時區 author date、author≠committer date | AC-10～12 |
| FX-07 | sorting | A=5、B=9、C=1、x@=4、m@=4、大小寫混合 email 平手 | AC-18、19 |
| FX-08 | unicode-csv | 中文、日文、emoji 名字與 email；含逗號、雙引號的名字 | AC-21、23、24、NFR-07 |
| FX-09 | empty | 剛 `git init`，0 個 commit | AC-20、25、30 |
| FX-10 | misc-lines | 空 commit、刪檔、CRLF 檔、無結尾換行、author≠committer、大量行數 | AC-01 邊界、A-03、A-04、G-6 |
| FX-11 | bad-paths | 不存在路徑、一般資料夾、一般檔案、無讀取權限的資料夾、損壞的 repo、含空白與中文的路徑 | AC-27、28、36 |
| FX-12 | perf-10k / perf-1k | 產生器造出 10,000／1,000 個 commit | NFR-01、02 |

### 4.3 各 Fixture 的精確內容與預期值

下表時間全部為 `+08:00`（台北時間）除非另註。「行」＝文字檔的行。

**FX-01 basic**（查詢範圍 `--since 2026-01-01 --until 2026-01-31`）

| # | author（name / email） | author date | 變更 | 在範圍內？ |
|---|------------------------|-------------|------|-----------|
| 1 | A / a@example.com | 2025-12-20T10:00 | 新增 `z.txt` 7 行 | 否 |
| 2 | A / a@example.com | 2026-01-05T10:00 | 新增 `src/a.txt` 20 行 | 是 |
| 3 | B / b@example.com | 2026-01-06T10:00 | 新增 `docs/b.txt` 2 行 | 是 |
| 4 | A / a@example.com | 2026-01-10T10:00 | 修改 `src/a.txt` 5 行（+5/-5） | 是 |
| 5 | A / a@example.com | 2026-01-20T10:00 | 新增 `c.txt` 5 行 | 是 |
| 6 | B / b@example.com | 2026-02-03T10:00 | 新增 `docs/b2.txt` 4 行 | 否 |

預期（範圍內）：A commits=3、added=30、deleted=5；B commits=1、added=2、deleted=0；順序 A、B。
預期（整段歷史 `--since 2025-01-01 --until 2026-12-31`）：A 4／37／5；B 2／6／0。

**FX-02 identity**（範圍 2026-01-01～2026-01-31）

| # | name / email | author date | 變更 |
|---|--------------|-------------|------|
| 1 | Alice / alice@example.com | 2026-01-03T09:00 | +3/-0 |
| 2 | Alice / alice@example.com | 2026-01-04T09:00 | +4/-1 |
| 3 | Alice Wang / alice@example.com | 2026-01-05T09:00 | +5/-2 |
| 4 | Bob / Bob@Example.com | 2026-01-06T09:00 | +1/-0 |
| 5 | Bob / bob@example.com | 2026-01-07T09:00 | +2/-0 |
| 6 | Carol Lee / carol@example.com | **2026-01-20T09:00**（但在歷史中先 commit） | +1/-0 |
| 7 | Carol / carol@example.com | **2026-01-10T09:00**（在歷史中後 commit，author date 較早） | +1/-0 |

預期：alice@example.com 1 列、commits=3、added=12、deleted=3、名字 `Alice Wang`；bob@example.com 1 列、commits=2、added=3、deleted=0、email 顯示全小寫；carol@example.com 名字為 `Carol Lee`（author date 最晚者，不是歷史上最後一個）。

**FX-03 merges**（範圍 2026-01-01～2026-01-31）
- 1：Z / z@example.com，2025-12-01，新增 `base.txt` 10 行（範圍外，root commit）。
- 在 `feature` 分支：2：A，2026-01-05，+10/-0；3：A，2026-01-06，+3/-2。
- 回到 main：4：Z，2026-01-07，修改 `base.txt` 第 1 行（+1/-1），使之後的 merge 產生衝突。
- 5：A，2026-01-08，`git merge --no-ff feature`，手動解衝突並在解衝突時額外加 2 行（evil merge，在 merge commit 本身產生內容差異）。
- 預期（預設）：A commits=2、added=13、deleted=2；Z commits=1、+1/-1。
- 預期（`--include-merges`）：A commits=3、added=13、deleted=2（merge 本身 0/0，含 evil merge 的 2 行也不計）；Z 不變。
- 變體 FX-03b（octopus）：A 在 2026-01-15 對 3 條分支做 octopus merge（`git merge b1 b2 b3`，無衝突），b1～b3 各 1 個 A 的一般 commit（各 +1）。預期預設 A 多 3 個 commit；`--include-merges` 再多 1 個 commit，行數不變。

**FX-04 binary-rename**
- 1：Z，2025-12-01，新增 `big.txt` 100 行、`dir1/mv.txt` 50 行、`asset.bin`（binary）。
- 2：A，**2026-01-10**，只新增 `logo.png`（binary）與 `notes.txt` 10 行 → 預期當日 A 1／10／0。（AC-07）
- 3：A，**2026-01-20**，只 `git mv big.txt big-renamed.txt`，內容不變 → 預期當日 A 1／0／0。（AC-08）
- 4：A，2026-01-21，`git mv dir1/mv.txt dir2/mv.txt`（跨資料夾、內容不變）→ 1／0／0。
- 5：A，2026-01-22，修改 `asset.bin`（binary 修改）→ 1／0／0。
- 6：A，2026-01-23，刪除 `logo.png`（binary 刪除）→ 1／0／0。
- 7：A，2026-01-24，改名 `notes.txt`→`notes2.txt` 並修改其中 1 行（相似度 90%）→ 預期見待釐清 D-03（暫定 1／1／1）。

**FX-05 branches**
- main：1：A，2026-01-05，+1；2：A，2026-01-06，+1。
- 從 1 分出 `feature`：3、4：C / c@example.com，2026-01-07、2026-01-08，各 +5。不合併。
- 最後 `git checkout main`。
- 預期（HEAD=main）：只有 A 2／2／0，沒有 C。
- 變體：`git checkout feature` 後：A 1、C 2。`git checkout --detach <commit 2>`：A 2。

**FX-06 dates-tz**（測試時 `TZ=Asia/Taipei`）

| # | 作者 | author date | committer date | 用途 |
|---|------|-------------|----------------|------|
| 1 | D10 / d10@example.com | 2025-12-31T23:59:59+08:00 | 同左 | AC-10 起日前 1 秒 |
| 2 | D10 | 2026-01-01T00:00:00+08:00 | 同左 | AC-10 |
| 3 | D10 | 2026-01-15T12:00:00+08:00 | 同左 | AC-10 同日查詢 |
| 4 | D10 | 2026-01-31T23:59:59+08:00 | 同左 | AC-10 |
| 5 | D10 | 2026-02-01T00:00:00+08:00 | 同左 | AC-10 迄日後 |
| 6 | TZ / tz@example.com | 2025-12-31T23:30:00-08:00（＝台北 2026-01-01 15:30） | 同左 | AC-11 |
| 7 | TZ2 / tz2@example.com | 2026-01-01T01:00:00+14:00（＝台北 2025-12-31 19:00） | 同左 | AC-11 反向 |
| 8 | R / r@example.com | 2026-01-10T10:00:00+08:00 | 2026-02-10T10:00:00+08:00 | AC-12 |
| 9 | R2 / r2@example.com | 2026-02-15T10:00:00+08:00 | 2026-01-15T10:00:00+08:00 | AC-12 反向 |

（AC-10 的 PRD 例子只有 3 個 commit；FX-06 多加 #1、#3 做邊界，所以 D10 在 1 月的預期 commits=3（#2、#3、#4）；只看 PRD 例子的 3 個時間點時，計入的是 #2、#4，與 PRD「前兩個計入」一致。）

**FX-07 sorting**：A / a@example.com 5 個、B / b@example.com 9 個、C / c@example.com 1 個、X / x@example.com 4 個、M / m@example.com 4 個、N / `N@Example.com` 4 個（大寫 email），全部在 2026-01，每個 commit +1。預期順序：B(9)、A(5)、m@(4)、n@(4)、x@(4)、C(1)。（n@ 驗證「不分大小寫」：若誤用區分大小寫的排序，`N` 會排到 `m` 前面。）

**FX-08 unicode-csv**（每人 1 個 commit，+1）
`王小明 / wang@example.com`、`山田太郎 / yamada@example.jp`、`Zoë 🚀 / zoe@example.com`、`李四 / 李四@例子.台灣`、`Wang, Alice / walice@example.com`、`Bob "The Builder" / bob@example.com`、`Al / al@example.com`（純 ASCII 對照）。

**FX-09 empty**：只執行 `git init`（`main` 分支、0 commit）。

**FX-10 misc-lines**（範圍 2026-01）
- E / e@example.com：`--allow-empty` 空 commit 1 個 → 1／0／0。
- F / f@example.com：新增 20 行 → 刪除整個檔案 → 2／20／20。
- G / g@example.com：新增 CRLF 換行的 5 行檔案 → 1／5／0；再新增最後一行無換行的 3 行檔案 → 累計 2／8／0。
- H / h@example.com（author）＋ K / k@example.com（committer）：1 個 commit +2 → 列在 h@，不出現 k@（A-03）。
- L / l@example.com：一個 commit 新增 12,345 行 → added=12345（不得輸出 `12,345`）。
- 歷史順序與日期順序打亂（後 commit 的 author date 比前一個早）→ 統計不受影響。

**FX-11 bad-paths**（由測試腳本在暫存區建立）
- `<tmp>/no/such/dir`（不存在）；`<tmp>/plain-dir`（存在、不在任何 repo 內；須確認其所有上層目錄都不是 repo）；`<tmp>/plain-file.txt`；`<tmp>/no-perm`（`chmod 000` 的 repo，僅 POSIX，且以非 root 身分執行）；`<tmp>/corrupt`（複製 FX-01 後刪除一個 object 檔）；`<tmp>/含 空白 與中文/repo`（FX-01 的複本，路徑含空白與中文）。

**FX-12 perf**：見第 6.2 節。

---

## 5. 測試案例

格式：每個 TC 有 ID、等級（Must／Should，沿用 PRD）、層級（I＝Integration、E＝E2E、P＝效能、M＝人工）、前置條件、步驟、預期結果。所有 TC 另外套用第 3 節通用檢查 G-1～G-6。

### F1 作者統計（Must）

#### AC-01 基本統計（Must）
- **TC-AC01-01**（I）前置：FX-01。步驟：`git-stats <FX-01> --since 2026-01-01 --until 2026-01-31`。預期：恰好 2 列資料；A：commits=3、added=30、deleted=5；B：commits=1、added=2、deleted=0；exit 0。
- **TC-AC01-02**（I）同上加 `--format csv`。預期：CSV 解析後恰好 2 筆，數值同上。（用 CSV 交叉驗證表格解析，避免表格版面問題掩蓋數字錯誤。）
- **TC-AC01-03**（I，邊界：範圍外不計）前置：FX-01。步驟：`--since 2025-01-01 --until 2026-12-31`。預期：A 4／37／5、B 2／6／0。
- **TC-AC01-04**（I，邊界：單一 commit）前置：FX-01。步驟：`--since 2026-01-06 --until 2026-01-06`。預期：只有 B 1 列，1／2／0。
- **TC-AC01-05**（I，邊界：作者＝author 不是 committer，A-03）前置：FX-10。預期：h@example.com 1 列 commits=1、added=2；輸出中沒有 k@example.com。
- **TC-AC01-06**（I，邊界：行數語意，A-04）前置：FX-10。預期：E 1／0／0（空 commit 仍計 commit）；F 2／20／20（刪檔計刪除行）；G 2／8／0（CRLF 與無結尾換行不影響行數）。
- **TC-AC01-07**（I，邊界：大數值，G-6）前置：FX-10。CSV 與表格中 L 的 added 均為 `12345`（表格見 D-05）。
- **TC-AC01-08**（I，邊界：root commit）前置：FX-01 範圍 2025-12-01～2025-12-31。預期：A 1／7／0（第一個 commit 的新增行數要計入）。

#### AC-02 同 email 不同名字合併（Must）
- **TC-AC02-01**（I）前置：FX-02。步驟：範圍 2026-01-01～2026-01-31，`--format csv`。預期：email 為 `alice@example.com` 的資料列恰好 1 筆，commits=3、added=12、deleted=3。
- **TC-AC02-02**（I）同上用表格格式。預期：只有 1 列含 `alice@example.com`。
- **TC-AC02-03**（I，反向：不同 email 同名字不合併）前置：在 FX-02 複本加 1 個 `Alice / alice2@example.com` 的 commit。預期：`alice@` 與 `alice2@` 是 2 列。
- **TC-AC02-04**（I，範圍外功能不得生效）前置：FX-02 複本在根目錄加 `.mailmap` 把 `Alice` 對到別的 email 並 commit（此 commit 本身也會被計入，預期值相應調整）。預期：統計結果不受 `.mailmap` 影響（PRD 4.2：MVP 不支援 `.mailmap`）。見待釐清 D-09。

#### AC-03 合併後顯示的名字（Must）
- **TC-AC03-01**（I）前置：FX-02。預期：`alice@example.com` 列的名字為 `Alice Wang`。
- **TC-AC03-02**（I，邊界：author date 與歷史順序相反）前置：FX-02。預期：`carol@example.com` 列名字為 `Carol Lee`（author date 最晚），不是歷史上最後一個 commit 的 `Carol`。
- **TC-AC03-03**（I，邊界：名字依「期間內」決定）前置：FX-02 範圍改為 2026-01-03～2026-01-04。預期：alice 列名字為 `Alice`（期間內只有署名 Alice 的 commit）。
- **TC-AC03-04**（I，邊界：同一 author date 兩個不同名字）見待釐清 D-01；先只驗證「輸出穩定」：連續執行 10 次，名字相同。

#### AC-04 email 大小寫視為同一人（Must）
- **TC-AC04-01**（I）前置：FX-02。預期：bob 只有 1 列，email 顯示 `bob@example.com`，commits=2、added=3。
- **TC-AC04-02**（I，邊界：只有大寫形式）範圍 2026-01-06～2026-01-06（只含 `Bob@Example.com`）。預期：email 仍顯示為 `bob@example.com`。
- **TC-AC04-03**（I，CSV）`--format csv` 時 email 欄同樣是全小寫。

#### AC-05 預設排除 merge commit（Must）
- **TC-AC05-01**（I）前置：FX-03。步驟：範圍 2026-01，不加 merge 選項。預期：A commits=2、added=13、deleted=2；Z 1／1／1。
- **TC-AC05-02**（I，邊界：octopus merge）前置：FX-03b。預期：octopus merge commit 不計入。
- **TC-AC05-03**（I，邊界：只有 merge commit 的作者）前置：FX-03 複本，由另一位作者 M2 做一次 `--no-ff` merge（無其他 commit）。預期：預設輸出中沒有 M2 這一列。

#### AC-06 選項納入 merge commit（Must）
- **TC-AC06-01**（I）前置：FX-03。步驟：加 `--include-merges`。預期：A commits=3、added=13、deleted=2（與 TC-AC05-01 的行數相同）。
- **TC-AC06-02**（I，邊界：evil merge）同上，確認 merge 解衝突時多加的 2 行**不**計入。
- **TC-AC06-03**（I，邊界：octopus）FX-03b 加 `--include-merges`：commit 數多 1，行數不變。
- **TC-AC06-04**（I，邊界：只有 merge 的作者）TC-AC05-03 的 repo 加 `--include-merges`：出現 M2 一列，1／0／0。
- **TC-AC06-05**（I，CSV）以 `--format csv --include-merges` 執行，數值同 TC-AC06-01。

#### AC-07 binary 檔的增刪行數（Must）
- **TC-AC07-01**（I）前置：FX-04。步驟：`--since 2026-01-10 --until 2026-01-10`。預期：A 1／10／0。
- **TC-AC07-02**（I，邊界：binary 修改）範圍 2026-01-22～2026-01-22。預期：A 1／0／0。
- **TC-AC07-03**（I，邊界：binary 刪除）範圍 2026-01-23～2026-01-23。預期：A 1／0／0。

#### AC-08 檔案改名不重複計算（Must）
- **TC-AC08-01**（I）前置：FX-04。範圍 2026-01-20～2026-01-20。預期：A 1／0／0。
- **TC-AC08-02**（I，邊界：跨資料夾改名）範圍 2026-01-21～2026-01-21。預期：A 1／0／0。
- **TC-AC08-03**（I，邊界：使用者 Git 設定 `diff.renames=false`）在執行環境的 `GIT_CONFIG_GLOBAL` 指向含 `diff.renames=false` 的設定檔後，重跑 TC-AC08-01。預期：仍為 1／0／0（結果不應因使用者個人 Git 設定而改變）。見待釐清 D-04。
- **TC-AC08-04**（I，邊界：改名＋小改）範圍 2026-01-24～2026-01-24。預期：見待釐清 D-03，暫定 1／1／1；判定前標為「無法驗證」，不作退件依據。

#### AC-09 統計範圍為目前 HEAD（Must）
- **TC-AC09-01**（I）前置：FX-05（HEAD=main）。預期：輸出不含 c@example.com；A 2／2／0。
- **TC-AC09-02**（I，反向）FX-05 複本 `git checkout feature`。預期：出現 C 2／10／0、A 1／1／0。
- **TC-AC09-03**（I，邊界：detached HEAD）FX-05 複本 `git checkout --detach <commit 2>`。預期：A 2／2／0，exit 0。
- **TC-AC09-04**（I，邊界：遠端追蹤分支與 tag 不影響）FX-05 複本加一個指向 feature 的 tag 與 `refs/remotes/origin/feature`。預期：同 TC-AC09-01。

### F2 日期範圍（Must）

#### AC-10 起訖日期都包含在內（Must）
- **TC-AC10-01**（I）前置：FX-06，`TZ=Asia/Taipei`。步驟：`--since 2026-01-01 --until 2026-01-31`。預期：D10 commits=3（#2 00:00:00、#3、#4 23:59:59 計入；#1、#5 不計入）。
- **TC-AC10-02**（I，同日查詢）`--since 2026-01-15 --until 2026-01-15`。預期：D10 commits=1（#3）。
- **TC-AC10-03**（I，邊界：只含起日當天 00:00:00）`--since 2026-01-01 --until 2026-01-01`。預期：D10 commits=1（#2），並含 TZ（見 AC-11）。
- **TC-AC10-04**（I，邊界：只含迄日當天 23:59:59）`--since 2026-01-31 --until 2026-01-31`。預期：D10 commits=1（#4）。
- **TC-AC10-05**（I，邊界：閏日）在 FX-06 複本加一個 2028-02-29T12:00+08:00 的 commit，`--since 2028-02-29 --until 2028-02-29`。預期：commits=1。

#### AC-11 時區以本地時區解讀（Must）
- **TC-AC11-01**（I）前置：FX-06，`TZ=Asia/Taipei`。`--since 2026-01-01 --until 2026-01-01` 時 TZ 列（#6）計入；`--since 2025-12-31 --until 2025-12-31` 時 TZ 列不出現。
- **TC-AC11-02**（I，反向）同一 fixture：`--since 2025-12-31 --until 2025-12-31` 時 TZ2 列（#7，字面日期 2026-01-01，台北時間 2025-12-31 19:00）計入；`--since 2026-01-01 --until 2026-01-01` 時不出現。
- **TC-AC11-03**（I，跨平台設定時區）在 Linux／macOS 用環境變數 `TZ=Asia/Taipei`；Windows 用 `tzutil /s "Taipei Standard Time"`（測完還原）。三平台結果須與 TC-AC11-01、02 相同。
- **TC-AC11-04**（I，換時區結果會變）同一 fixture 改用 `TZ=America/Los_Angeles` 執行 `--since 2025-12-31 --until 2025-12-31`。預期：TZ 列（洛杉磯時間 2025-12-31 23:30）計入。證明確實是「執行環境本地時區」，不是寫死 UTC+8。
- **TC-AC11-05**（I，邊界：夏令時間切換日）`TZ=America/New_York`，fixture 加 2026-03-08T00:30-05:00 與 2026-03-08T23:30-04:00 兩個 commit，`--since 2026-03-08 --until 2026-03-08`。預期：兩個都計入（該日只有 23 小時，仍是整天）。

#### AC-12 以 author date 判斷（Must）
- **TC-AC12-01**（I）前置：FX-06。`--since 2026-01-01 --until 2026-01-31`：R 列計入；`--since 2026-02-01 --until 2026-02-28`：R 列不出現。
- **TC-AC12-02**（I，反向）R2（author 2/15、committer 1/15）：1 月不出現、2 月出現。

#### AC-13 日期格式錯誤（Must）
- **TC-AC13-01**（I）前置：FX-01。依序以下列值作為 `--since`（`--until` 給合法值）：`2026/01/01`、`01-31-2026`、`2026-1-1`、`abc`。預期每次：stderr 含該錯誤值原文與字串 `YYYY-MM-DD`；stdout 0 byte；exit 2。
- **TC-AC13-02**（I）同上，但錯誤值放在 `--until`。
- **TC-AC13-03**（I，邊界）錯誤值：空字串 `""`、`20260101`、`2026-01-01T00:00:00`、`2026-01-01 13:00`（含時間，PRD 4.2／A-05 不支援）、` 2026-01-01`（前後空白）、全形數字 `２０２６-０１-０１`、`2026-01-1`、`02026-01-01`。預期：皆同 TC-AC13-01（exit 2）。空白是否自動去除見待釐清 D-06。
- **TC-AC13-04**（I，邊界：相對日期不支援）`yesterday`、`last month`、`7 days ago`。預期：exit 2（PRD 4.2）。

#### AC-14 不存在的日期（Must）
- **TC-AC14-01**（I）`--since 2026-02-30` 與 `--since 2026-13-01`（分別執行）。預期：同 AC-13 報錯方式，exit 2。
- **TC-AC14-02**（I，邊界）`2026-02-29`（非閏年）、`2026-04-31`、`2026-00-10`、`2026-01-00`、`2026-01-32`。預期：exit 2。
- **TC-AC14-03**（I，反向：合法閏日）`2028-02-29`。預期：被接受（exit 0）。
- **TC-AC14-04**（I，不得自動換算）對 TC-AC14-01 確認 stdout 為 0 byte（若工具把 `2026-02-30` 換算成 3/2 並輸出結果，即失敗）。

#### AC-15 起日晚於迄日（Must）
- **TC-AC15-01**（I）`--since 2026-03-01 --until 2026-02-01`。預期：stderr 有「起日晚於迄日」類型的英文訊息（建議包含兩個日期）；stdout 0 byte；exit 2。
- **TC-AC15-02**（I，邊界：只差一天）`--since 2026-01-02 --until 2026-01-01`。預期：exit 2。
- **TC-AC15-03**（I，反向：相同日期合法）見 TC-AC10-02。

#### AC-16 缺少日期參數（Must）
- **TC-AC16-01**（I）只給 `--until`。預期：stderr 含錯誤訊息＋用法說明（usage），stdout 0 byte，exit 2。
- **TC-AC16-02**（I）只給 `--since`。預期同上。
- **TC-AC16-03**（I）兩者都不給。預期同上。
- **TC-AC16-04**（I，邊界：選項有名稱沒有值）`git-stats <FX-01> --since --until 2026-01-31` 與結尾 `--until`（無值）。預期：exit 2，stdout 0 byte。

### F3 表格輸出與排序（Must；AC-21 為 Should）

#### AC-17 預設表格輸出與欄位（Must）
- **TC-AC17-01**（I）前置：FX-01。不加 `--format`。預期：第 1 個非框線行為表頭；欄位順序為 名字、email、commit 數、新增行數、刪除行數；之後每位作者一列。
- **TC-AC17-02**（I，對齊）以表頭各欄起訖位置為基準：文字欄（名字、email）每列左緣在同一字元位置；數字欄每列右緣在同一字元位置（以 FX-01、FX-07 這類純 ASCII 名字驗證）。
- **TC-AC17-03**（I，邊界：數字位數不同）FX-10（含 12345 與 0）。預期：數字欄仍右對齊。
- **TC-AC17-04**（I，`--format table` 明確指定）結果與不指定時逐位元組相同。

#### AC-18 依 commit 數由多到少排序（Must）
- **TC-AC18-01**（I）前置：FX-07，表格。預期：資料列順序 B、A、（m、n、x）、C。
- **TC-AC18-02**（I）同上 `--format csv`，順序相同。

#### AC-19 commit 數相同時的次排序（Must）
- **TC-AC19-01**（I）前置：FX-07。預期：4 個 commit 的三人順序為 `m@example.com`、`n@example.com`、`x@example.com`（不分大小寫比較）。表格與 CSV 都驗。
- **TC-AC19-02**（I，穩定性）表格與 CSV 各連續執行 10 次，stdout 逐位元組相同。
- **TC-AC19-03**（I，邊界：email 共同前綴）加 `a@example.com` 與 `a.b@example.com`、`ab@example.com` 各 1 個 commit 的變體。預期：依字元逐一比較的字典序排列（待釐清 D-07：「字母順序」對非英文字母、標點符號的定義）。

#### AC-20 查無資料時的表格輸出（Must）
- **TC-AC20-01**（I）前置：FX-01，範圍 `--since 2030-01-01 --until 2030-01-31`。預期：stdout 恰好 1 行（非空）提示訊息（例如 "No commits in the specified period"），**不含表頭與框線**；exit 0；stderr 空。
- **TC-AC20-02**（I，邊界：只有 merge commit 落在範圍內）TC-AC05-03 的 repo，範圍只涵蓋 M2 的 merge 那天，預設模式。預期：同 TC-AC20-01（沒有「符合條件」的 commit）。

#### AC-21 中日韓全形字元對齊（Should）
- **TC-AC21-01**（I）前置：FX-08。表格輸出。預期：以「全形（East Asian Wide/Fullwidth）字元＝2 欄寬、其他＝1 欄寬」計算，每列各欄起始顯示欄位相同。驗證方式：測試程式依 Unicode East Asian Width 計算顯示寬度後比對；另在 Windows Terminal、macOS Terminal、GNOME Terminal 截圖人工確認。
- **TC-AC21-02**（I，邊界：emoji、帶附加符號字母）`Zoë 🚀` 的寬度判定見待釐清 D-08；判定前只要求「不當機、不亂碼」（NFR-07），對齊結果記錄但不判失敗。

### F4 CSV 輸出（Must；AC-26 為 Should）

#### AC-22 CSV 格式與欄位（Must）
- **TC-AC22-01**（I）前置：FX-01，`--format csv`。預期：第 1 行逐字為 `name,email,commits,added,deleted`；之後 2 行；分隔符為半形逗號；順序 A、B。
- **TC-AC22-02**（I，邊界：數字）FX-10：`added` 為 `12345`，不得出現 `12,345`、`12345.0`、空白。
- **TC-AC22-03**（I，排序同 AC-18／19）FX-07 CSV 順序同 TC-AC19-01。
- **TC-AC22-04**（I，換行字元）記錄行尾是 LF 或 CRLF，兩者都可被標準 CSV 解析器讀；三平台輸出應一致（見待釐清 D-10）。

#### AC-23 CSV 特殊字元跳脫（Must）
- **TC-AC23-01**（I）前置：FX-08，`--format csv`。預期：`Wang, Alice` 輸出為 `"Wang, Alice"`；`Bob "The Builder"` 輸出為 `"Bob ""The Builder"""`。
- **TC-AC23-02**（I，回讀）用標準 RFC 4180 解析器讀回全部列，每列恰好 5 欄，名字、email 與 fixture 原始值逐字相同。
- **TC-AC23-03**（I，反向：不需要時不加引號）`Al` 的那行為 `Al,al@example.com,1,1,0`（PRD 只要求「含特殊字元才包」；若一律加引號，判為待釐清而非失敗，見 D-11）。
- **TC-AC23-04**（Unit 建議，非 QA 驗收）名字含換行：Git 不允許 author 名字含換行，無法用真實 repo 造出，請 TechLead 在 Unit 層覆蓋。QA 在報告中標「無法以黑箱驗證」。

#### AC-24 CSV 編碼（Must）
- **TC-AC24-01**（I）前置：FX-08，`--format csv`，stdout 以 byte 讀取。預期：開頭 3 byte 不是 `EF BB BF`（無 BOM）；整體可用嚴格 UTF-8 解碼成功；`王小明` 一列的 name 欄解碼後為 `王小明`。
- **TC-AC24-02**（I，跨平台）Windows 上用 `--output` 寫檔（見 AC-26）與 Git Bash 導出兩種方式，檔案都是 UTF-8 無 BOM。另記錄 PowerShell 5.1 用 `>` 導出的結果（PRD 已知會變 UTF-16，僅記錄、不判失敗）。
- **TC-AC24-03**（I，非中文非 ASCII）`山田太郎`、`Zoë 🚀`、`李四@例子.台灣` 回讀一致。

#### AC-25 查無資料時的 CSV 輸出（Must）
- **TC-AC25-01**（I）前置：FX-01，範圍 2030-01。`--format csv`。預期：stdout 恰好一行 `name,email,commits,added,deleted`（後接一個換行或無換行均可，見 D-10），沒有其他內容；exit 0。

#### AC-26 直接寫入檔案（Should）
- **TC-AC26-01**（I）前置：FX-08，空暫存資料夾。步驟：`git-stats <FX-08> --since ... --until ... --format csv --output report.csv`。預期：產生 `report.csv`；內容與不加 `--output` 時的 stdout **逐位元組相同**；stdout 0 byte；stderr 0 byte；exit 0。
- **TC-AC26-02**（I，覆寫）先建一個內容較長的 `report.csv`，再執行。預期：檔案被完整覆寫（不殘留舊內容、長度等於新內容）。
- **TC-AC26-03**（I，查無資料）範圍 2030-01。預期：檔案只有表頭一行。
- **TC-AC26-04**（I，異常：無法寫入）`--output` 指向不存在的資料夾或唯讀資料夾。預期：exit 非 0（建議 1）、stdout 0 byte、stderr 有說明。見待釐清 D-12。
- **TC-AC26-05**（I，邊界：與表格格式併用）`--output report.txt` 不加 `--format csv`。預期：見待釐清 D-13；判定前只記錄行為。
- **TC-AC26-06**（I，NFR-08）只有 `--output` 指定的檔案被建立，其他位置無新檔。

### F5 輸入檢查、錯誤處理與基本指令行為（Must）

#### AC-27 路徑不存在（Must）
- **TC-AC27-01**（I）`git-stats /no/such/dir --since 2026-01-01 --until 2026-01-31`（Windows 用 `C:\no\such\dir`）。預期：stderr 含「路徑不存在」類型英文訊息與該路徑原文；stdout 0 byte；exit 1。
- **TC-AC27-02**（I，邊界：路徑含空白與中文）`<tmp>/不存在 的/資料夾`。預期：同上，訊息中路徑完整不亂碼。
- **TC-AC27-03**（I，邊界：路徑合法但日期也錯）見待釐清 D-14（錯誤優先順序）；判定前只檢查 exit code 為 1 或 2 且 stdout 0 byte。

#### AC-28 路徑不是 Git repo（Must）
- **TC-AC28-01**（I）FX-11 `plain-dir`。預期：stderr 含「不是 Git repo」類型訊息與該路徑；stdout 0 byte；exit 1。
- **TC-AC28-02**（I）FX-11 `plain-file.txt`（一般檔案）。預期同上（exit 1）。
- **TC-AC28-03**（I，異常：無讀取權限，僅 POSIX、非 root）FX-11 `no-perm`。預期：exit 1，stdout 0 byte，stderr 有說明（依約定「讀取失敗」屬 1）。
- **TC-AC28-04**（I，異常：損壞的 repo）FX-11 `corrupt`，範圍涵蓋缺失的 object。預期：exit 1，stdout 0 byte（不得輸出部分結果）。
- **TC-AC28-05**（I，邊界：bare repo、`.git` 資料夾本身當路徑）見待釐清 D-15；只記錄行為。

#### AC-29 路徑是 repo 子資料夾（Must）
- **TC-AC29-01**（I）FX-01，以 `<FX-01>/src` 執行。預期：stdout 與以 `<FX-01>` 執行逐位元組相同（B 只改了 `docs/`、A 的 `c.txt` 在根目錄，仍要計入）。
- **TC-AC29-02**（I，CSV）同上 `--format csv`。
- **TC-AC29-03**（I，邊界：相對路徑與 `.`）在 `<FX-01>/src` 目錄下以 `.` 為路徑執行；在 `<FX-01>` 的上層以相對路徑 `FX-01` 執行。結果皆與 TC-AC29-01 相同。

#### AC-30 沒有任何 commit 的新 repo（Must）
- **TC-AC30-01**（I）FX-09，表格。預期：同 AC-20（1 行提示、exit 0、stderr 空）。
- **TC-AC30-02**（I）FX-09，`--format csv`。預期：同 AC-25（只有表頭）。
- **TC-AC30-03**（I，邊界：HEAD 指向尚未建立的分支，但其他分支有 commit）FX-05 複本 `git checkout --orphan newbranch` 並清空 index。預期：同 TC-AC30-01（依 AC-09，HEAD 可追溯的 commit 為 0）。

#### AC-31 缺少 repo 路徑參數（Must）
- **TC-AC31-01**（I）在 FX-01 目錄內執行 `git-stats --since 2026-01-01 --until 2026-01-31`。預期：stderr 有錯誤訊息＋用法說明；stdout 0 byte；exit 2（不得預設用目前資料夾，PRD Q-09）。
- **TC-AC31-02**（I，邊界：多給一個位置參數）`git-stats <FX-01> <FX-02> --since ... --until ...`。預期：exit 2（見待釐清 D-16）。
- **TC-AC31-03**（I，異常：未知選項）`--foo`、`--include-merge`（拼錯）。預期：exit 2（第 5 節約定「未知選項」屬 2）。

#### AC-32 未知的輸出格式（Must）
- **TC-AC32-01**（I）`--format json`。預期：stderr 含錯誤訊息並列出 `table`、`csv`；stdout 0 byte；exit 2。
- **TC-AC32-02**（I，邊界）`--format xml`、`--format ""`、`--format`（無值）、`--format CSV`（大寫）。前三者 exit 2；大寫是否接受見待釐清 D-17。

#### AC-33 說明（Must）
- **TC-AC33-01**（I）`git-stats --help`。預期：stdout 含 repo 路徑、起日、迄日、輸出格式、納入 merge commit 選項、日期格式 `YYYY-MM-DD` 的說明（逐項以關鍵字比對：`--since`、`--until`、`--format`、`table`、`csv`、`--include-merges`、`YYYY-MM-DD`，以及路徑參數名稱）；exit 0；stderr 空。
- **TC-AC33-02**（I，邊界）`--help` 在非 Git 資料夾執行、或與其他參數並用（`git-stats --help --format json`）。預期：仍顯示說明、exit 0（見待釐清 D-18）。
- **TC-AC33-03**（E）在 E2E 安裝後的乾淨機器上執行 TC-AC33-01。

#### AC-34 版本（Must）
- **TC-AC34-01**（I，G3 階段）`git-stats --version`。預期：stdout 含符合 `X.Y.Z`（語意化版本）的版本號；exit 0；stderr 空。
- **TC-AC34-02**（E，G4 階段）從 GitHub Release 下載 tag `vX.Y.Z` 的檔案安裝後執行。預期：版本號＝tag 去掉 `v`。G3 時尚無 Release，本 TC 標「無法驗證（階段未到）」，不影響 G3，但 G4 前必須通過。

#### AC-35 唯讀，不修改 repo（Must）
- **TC-AC35-01**（I）FX-01，執行前記錄 `git status --porcelain`、`git for-each-ref`、`.git/config`；執行有資料的查詢（表格、CSV、`--include-merges` 各一次）。預期：三者執行後完全相同。
- **TC-AC35-02**（I）同上，但查詢範圍無資料（2030-01）。預期同上。
- **TC-AC35-03**（I，加強檢查）在 FX-01 的工作目錄放一個未追蹤檔與一個已修改但未 commit 的檔案，再執行。預期：`git status --porcelain` 前後相同、工作目錄檔案內容不變。
- **TC-AC35-04**（I，加強檢查，記錄用）執行前後對整個 `.git/` 內所有檔案計算雜湊值（hash）比對，並排除 `.git/index` 以外不做排除。若有差異：refs、config、objects、HEAD 有變 → 判為 AC-35／PRD 4.3 違反（高）；只有 `.git/index` 的時間戳記資訊被更新 → 記錄並列為待釐清 D-19，不直接退件。
- 另：G-4 讓整個 Integration 套件的每次執行都做 TC-AC35-01 的三項比對。

#### AC-36 錯誤與結果分流（Must）
- **TC-AC36-01**（I）對 TC-AC13-01、TC-AC14-01、TC-AC15-01、TC-AC16-01～03、TC-AC27-01、TC-AC28-01、TC-AC31-01、TC-AC32-01 的每個指令，執行 `git-stats ... > out.txt 2> err.txt`。預期：`out.txt` 為 0 byte；`err.txt` 非空。
- **TC-AC36-02**（I）對 TC-AC01-01、TC-AC20-01、TC-AC22-01、TC-AC25-01 的指令同樣導向。預期：`err.txt` 為 0 byte。
- **TC-AC36-03**（E，Windows）在 PowerShell 與 cmd 各做一次 TC-AC36-01 的代表案例（AC-13、AC-27），確認 Windows 上同樣分流。

### F6 安裝與發佈（Must）

#### AC-37 跨平台單一指令安裝（Must）
- **TC-AC37-01**（E）乾淨 Windows 11 x86_64（只裝 Git for Windows）：依 README 安裝（1 行指令，或下載 Release 檔案後 1 個步驟）→ 關閉並重開終端機 → `git-stats --version`。預期：符合 AC-34；全程計時並記錄（M2 ≤5 分鐘）。
- **TC-AC37-02**（E）乾淨 macOS（當時最新兩個大版本各一台；至少一台 Apple Silicon，NFR-03 另需 x86_64 一台）。預期同上。
- **TC-AC37-03**（E）乾淨 Ubuntu 24.04 x86_64。預期同上。
- **TC-AC37-04**（E，NFR-03 延伸）Ubuntu 22.04 x86_64。預期同上。
- **TC-AC37-05**（E，安裝後跑核心案例）在 TC-AC37-01～04 的每台機器上，以安裝好的指令跑 TC-AC01-01、TC-AC22-01、TC-AC24-01、TC-AC27-01。
- 「1 個步驟」的判定：計算使用者需要輸入的指令數或手動動作數（解壓縮＋放到 PATH 算 2 步）。見待釐清 D-20。

#### AC-38 GitHub Release 發佈（Must）
- **TC-AC38-01**（E，G4 階段）查看 `https://github.com/hansheng08/git-stats-cli/releases`（或 `gh release view vX.Y.Z`）。預期：tag 格式為 `vX.Y.Z`；有對應 Release；說明內有變更摘要與 Windows／macOS／Linux 三平台安裝方式。
- **TC-AC38-02**（E）tag 指向的 commit 在 main 上。
- **TC-AC38-03**（E）PyPI 與 npm 搜尋產品名稱（`git-stats`、`git-stats-cli`）不存在本專案發佈的套件（若有同名他人套件，確認發佈者不是本專案）。
- G3 時本 AC 標「無法驗證（階段未到）」，G4 前必須通過。

### 非功能需求（NFR）

- **TC-NFR01-01**（P）見第 6 節：perf-10k、整段歷史、表格與 CSV 各一組量測，全部 ≤ 10 秒。
- **TC-NFR02-01**（P）見第 6 節：perf-1k、整段歷史，全部 ≤ 2 秒。
- **TC-NFR03-01**（I）全部 Integration 案例在 NFR-03 OS 矩陣（Windows 11 x86_64、macOS 兩個大版本含 Apple Silicon 與 x86_64、Ubuntu 22.04、24.04 x86_64）上通過。見第 8 節。
- **TC-NFR04-01**（I）在無網路環境跑完整 Integration 套件：Linux 用 `unshare -rn` 或 `docker run --network none`。預期：結果與有網路時完全相同。
- **TC-NFR04-02**（I）Linux 上以 `strace -f -e trace=network` 執行 TC-AC01-01 與 TC-AC20-01。預期：沒有 `connect` 到任何網路位址（AF_INET／AF_INET6）的系統呼叫。
- **TC-NFR05-01**（M）行為面同 TC-NFR04-01／02；Code review 部分由 TechLead 在 PR 中確認「無遙測、無上傳」並記錄，QA 只檢查 PR 上有此紀錄，不讀程式碼（見 1.2）。
- **TC-NFR06-01**（M）對所有錯誤類 TC（AC-13～16、27、28、31、32）的 stderr 以 G-3 清單人工判定，結果表附在 QA_REPORT。
- **TC-NFR07-01**（I）FX-08 表格與 CSV 都不當機、不出現 `?`、`\uXXXX`、亂碼；名字與 email 回讀一致。三平台都要驗（Windows 在 Windows Terminal＋UTF-8 字碼頁與預設字碼頁各一次，見待釐清 D-21）。
- **TC-NFR08-01**（I）設定空的 `HOME`（Windows 另設 `USERPROFILE`、`APPDATA`、`LOCALAPPDATA`）與 `TMPDIR`（Windows `TEMP`、`TMP`），執行代表性案例前後列出這些目錄所有檔案。預期：除 `--output` 指定的檔案外無新增或修改。
- **TC-M2-01**（E）同 TC-AC37-01～03 的計時，三平台各 ≤ 5 分鐘（從開啟 README 到看到第一次統計結果）。

### PRD 情境（E2E 串接）
- **TC-S1-01**（E）老闆情境：對 FX-01 查上個月整月（2026-01-01～2026-01-31），表格輸出，最多的人在最前。
- **TC-S2-01**（E）技術主管情境：FX-08 輸出 CSV 檔，用 Google Sheets 匯入與 Excel「從文字匯入」開啟（人工截圖），中文不亂碼、逗號名字不錯欄。
- **TC-S3-01**（E）維護者情境：FX-02＋FX-03 組合，同 email 合併、預設不含 merge。
- **TC-S4-01**（E）腳本情境：用 shell 腳本連跑成功與失敗各一次，依 exit code 分支處理；失敗時輸出檔為空。

---

## 6. 效能測試（NFR-01、NFR-02）

### 6.1 參考機器（Reference Machine）
| 項目 | 規格 |
|------|------|
| 平台 | GitHub Actions hosted runner `ubuntu-24.04`（標準規格、x86_64） |
| 最低硬體 | 2 vCPU、7 GB RAM、SSD（GitHub 標準 runner 的公開規格；**不確定**：GitHub 依 repo 公開／私有與方案不同可能給 2 或 4 vCPU，因此每次量測必須記錄實際值） |
| 作業系統 | Ubuntu 24.04 LTS |
| Git 版本 | Ubuntu 24.04 套件庫內的 git（記錄 `git --version`；若 README 規定最低 Git 版本，另以該最低版本跑一次） |
| 檔案系統 | runner 預設工作磁碟（ext4） |
| 每次記錄 | `nproc`、`free -m`、`lscpu` 的 CPU 型號、`git --version`、`uname -a`、工具版本 |

說明：
- 選 2 vCPU／7 GB 是「較弱」的規格，在它上面達標，通常在使用者的筆電上也能達標；比用 QA 自己的雲端機器更容易重現。
- 若實際分到的 runner 規格高於最低規格，結果照常判定，但在報告中註明；若量測結果在 9～10 秒之間（接近門檻），QA 另在最低規格（2 vCPU、7 GB）的 VM 上複測一次再判定。
- 其他平台（Windows、macOS）的效能僅記錄、不作為 NFR-01 判定依據（PRD 只要求「參考測試機器」）。

### 6.2 測試資料產生器（perf-10k／perf-1k）
- 用 `git fast-import`（Git 內建的批次匯入工具，可在幾秒內造出上萬個 commit）讀取一個由腳本產生的固定串流，不依賴實作語言；產生腳本用固定的亂數種子（seed），保證每次 HEAD SHA 相同（記入 `EXPECTED_SHA`）。
- perf-10k 的形狀（模擬「中型真實專案」）：
  - HEAD 可追溯的 commit 總數恰好 **10,000**（含 merge commit），全在 main 線上。
  - 作者 50 位（其中 5 位有 2 種名字、3 位 email 大小寫不同、5 位中文或 emoji 名字）。
  - 約 2%（200 個）為 merge commit（兩個父 commit），其餘為一般 commit。
  - 檔案 500 個文字檔、分在 20 個資料夾，外加 20 個 binary 檔；每個一般 commit 改 1～5 個檔，每檔新增 1～50 行、刪除 0～30 行；約 1% commit 含改名。
  - author date 平均分布在 2021-01-01～2025-12-31。
- perf-1k：同一產生器、commit 數 1,000，其他比例相同。
- 產生器同時輸出「預期總數表」（每位作者 commit／增刪行數），效能測試順便驗證結果正確（M1）。產生器本身的正確性：QA 用 `git log --numstat` 對產生的 repo 抽 3 位作者手算核對一次。

### 6.3 量測方式
- 指令（整段歷史）：`git-stats <perf-10k> --since 1970-01-01 --until 2099-12-31`（表格）與加 `--format csv`（stdout 導向 `/dev/null` 以排除終端機顯示時間）。「整段歷史」需要給一個涵蓋全部的日期範圍，因 MVP 起訖日必填（Q-09），見待釐清 D-02。
- 工具：`hyperfine --warmup 3 --runs 10 --export-json perf.json '<指令>'`；若無法安裝 hyperfine，改用 `/usr/bin/time -v` 手動跑 3 次暖機＋10 次量測。
- 主要判定（warm cache）：10 次量測的**最大值**（也就是每一次都）≤ 10 秒（NFR-02 為 ≤ 2 秒）。另在報告中列出 median、p95、平均、標準差。
  - 選用「最大值」而不是中位數：PRD 寫「≤ 10 秒完成」，使用者每一次執行都應達到，用中位數會放過偶發超時。
- 輔助量測（cold cache，只記錄）：每次量測前執行 `sync; echo 3 | sudo tee /proc/sys/vm/drop_caches`（hyperfine `--prepare`），跑 5 次記錄 median 與最大值。若 cold cache 超過 10 秒，記為「中」嚴重度風險並請老闆決定是否納入判定（見待釐清 D-02）。
- 同時記錄最大記憶體用量（`/usr/bin/time -v` 的 Maximum resident set size），僅記錄。
- 雜訊控制：量測時 runner 上不跑其他 job；每個 job 只跑效能測試；結果超標時自動重跑整組一次，兩組都超標才判失敗（避免共用機器的偶發干擾），兩組結果都寫入報告。

---

## 7. 回歸測試（Regression）清單

### 7.1 每個 PR（CI 必跑，約數分鐘）
| 編號 | 內容 |
|------|------|
| RG-01 | 全部 Integration 案例（第 5 節標 I 者）在 Ubuntu 24.04 上執行 |
| RG-02 | 黃金檔（golden file，事先存好、確認正確的預期輸出檔）比對：FX-01、FX-02、FX-03、FX-07、FX-08 的表格與 CSV 輸出逐位元組比對 |
| RG-03 | G-4 唯讀檢查（每次執行都做） |
| RG-04 | 所有錯誤類案例的 exit code 與 stdout 0 byte（AC-13～16、27、28、31、32、36） |
| RG-05 | fixture `EXPECTED_SHA` 比對（fixture 本身沒被改壞） |

### 7.2 每次合併前（G3 前，QA 執行）
| 編號 | 內容 |
|------|------|
| RG-10 | RG-01～05 在 NFR-03 全部 OS 矩陣執行 |
| RG-11 | 效能測試 TC-NFR01-01、TC-NFR02-01 |
| RG-12 | 無網路與 NFR-08 檢查（TC-NFR04-01、TC-NFR08-01） |
| RG-13 | 先前所有已修復缺陷的「鎖住測試」（fix/Q-xxx 分支加入的失敗測試，修好後永久保留） |

### 7.3 發佈前（G4 前）
| 編號 | 內容 |
|------|------|
| RG-20 | TC-AC37-01～05、TC-AC38-01～03、TC-AC34-02、TC-M2-01 |
| RG-21 | TC-S1-01～TC-S4-01 |

### 7.4 M3 核對範本（給老闆，非 QA 驗收）
對真實 repo 隨機抽 3 位作者：`git log --author=<email> --since=<起日> --until=<迄日 23:59:59> --no-merges --numstat --format=` 手動加總（注意 `--author` 是部分比對，email 大小寫與名字合併需人工處理），與工具輸出比對。

---

## 8. 環境需求

| 項目 | 需求 |
|------|------|
| CI | GitHub Actions；矩陣：`ubuntu-22.04`、`ubuntu-24.04`、`windows-2025`（或當時對應 Windows 11 等級的 runner；**不確定** GitHub 是否提供「Windows 11」runner，若只有 Windows Server，Windows 11 改用 VM 人工驗證）、`macos-latest` 與前一版（Apple Silicon）、一個 x86_64 macOS（若 GitHub 已無 x86_64 macOS runner，改用人工或標「無法驗證」並請老闆決定） |
| Git | 每個平台的 Git 記錄版本；Windows 用 Git for Windows（內附 Git Bash，用來跑 fixture 腳本） |
| 工具 | POSIX shell、`git`、標準 CSV 解析器（例如 Python 3 `csv` 模組）、Unicode East Asian Width 判定（例如 Python `unicodedata`）、`hyperfine`、`strace`（Linux）、`unshare` 或 Docker（無網路測試）、`sha256sum` |
| 時區 | Linux／macOS 用 `TZ` 環境變數；Windows 用 `tzutil`（需要管理員權限，GitHub Windows runner 有） |
| 權限 | 權限測試（TC-AC28-03）需以非 root 身分執行 |
| 雲端副本 | QA 在 `/workspace/projects/git-stats-cli` 自己 checkout 分支執行；缺工具依 `docs/RUNBOOK.md` 安裝，RUNBOOK 沒寫則回報幕僚長請 SRE 補上 |
| E2E 機器 | 乾淨的 Windows 11、macOS（最新兩個大版本）、Ubuntu 22.04／24.04 VM 或 runner，只預裝 Git |
| 無網路 | Linux 上以 network namespace 或 Docker `--network none` |

---

## 9. 通過／退件標準

### 9.1 每條 AC 的判定
- **通過**：該 AC 所有對應 TC（不含標「待釐清、只記錄」者）在要求的平台全部通過。
- **失敗**：任一對應 TC 的預期結果不符。
- **無法驗證**：環境或階段不允許（例如 G3 時尚無 Release）、fixture `EXPECTED_SHA` 不符、或 PRD 有待釐清事項導致無法判定。必須寫明原因與何時能驗證。

### 9.2 缺陷嚴重度
| 等級 | 定義 | 例子 |
|------|------|------|
| 致命（Critical） | 修改或損壞使用者 repo、對外傳送資料、無法安裝或無法執行 | AC-35 違反、NFR-04／05 違反 |
| 高（High） | 統計數字錯誤、Must AC 不符、exit code 錯誤、錯誤時 stdout 有輸出（會污染腳本產出的 CSV）、效能不達標、在某支援平台無法使用 | AC-01 數字錯、CSV 跳脫錯誤導致錯欄 |
| 中（Medium） | Should AC 不符、錯誤訊息不夠清楚（NFR-06）、只在罕見邊界才出錯且不影響數字 | AC-21 對齊偏差 |
| 低（Low） | 文字、排版等不影響使用的小問題 | help 文字錯字 |

### 9.3 G3 前的判定規則
1. **退件**：任一 Must AC 判「失敗」，或有任一「高」以上缺陷。不得因為趕時間而放行。
2. **Must AC 判「無法驗證」**：
   - 因階段未到（AC-34 的 Release 比對、AC-37、AC-38、M2）→ 不影響 G3，但必須在 G4 前驗證，並在回報中列出。
   - 因其他原因（環境、待釐清）→ 狀態寫「卡住」，回報幕僚長說明需要誰提供什麼；QA 不自行假設通過。
3. Should AC（AC-21、AC-26）失敗：記為「中」缺陷，不單獨構成退件；是否放行由老闆決定，QA 在回報中附建議。
4. NFR：PRD 未標 Must／Should（見待釐清 D-22）。QA 暫定 NFR-01～NFR-08 的違反依 9.2 判嚴重度（安全類＝致命，效能與平台＝高），因此實務上會導致退件。
5. 全部 Must 通過、無高以上缺陷 → 回報幕僚長，建議送 G3。
6. 只對「本次 PR 影響的 AC」做完整驗證；其餘 AC 由回歸測試（第 7 節）覆蓋，回歸任一失敗＝該 AC 失敗。

### 9.4 判定紀錄
每次驗證在 `docs/QA_REPORT.md` 新增一節（不覆蓋舊紀錄），包含：被驗證的分支與 commit SHA、執行環境、CI 結果連結、AC 逐條判定表、缺陷清單（重現步驟、預期、實際、嚴重度）、效能數據。

---

## 10. AC 對照表（AC → TC → 層級）

| AC | 等級 | TC | 層級 | Fixture |
|----|------|----|------|---------|
| AC-01 | Must | TC-AC01-01～08 | I | FX-01、FX-10 |
| AC-02 | Must | TC-AC02-01～04 | I | FX-02 |
| AC-03 | Must | TC-AC03-01～04 | I | FX-02 |
| AC-04 | Must | TC-AC04-01～03 | I | FX-02 |
| AC-05 | Must | TC-AC05-01～03 | I | FX-03、FX-03b |
| AC-06 | Must | TC-AC06-01～05 | I | FX-03、FX-03b |
| AC-07 | Must | TC-AC07-01～03 | I | FX-04 |
| AC-08 | Must | TC-AC08-01～04 | I | FX-04 |
| AC-09 | Must | TC-AC09-01～04 | I | FX-05 |
| AC-10 | Must | TC-AC10-01～05 | I | FX-06 |
| AC-11 | Must | TC-AC11-01～05 | I | FX-06 |
| AC-12 | Must | TC-AC12-01～02 | I | FX-06 |
| AC-13 | Must | TC-AC13-01～04 | I | FX-01 |
| AC-14 | Must | TC-AC14-01～04 | I | FX-01 |
| AC-15 | Must | TC-AC15-01～03 | I | FX-01 |
| AC-16 | Must | TC-AC16-01～04 | I | FX-01 |
| AC-17 | Must | TC-AC17-01～04 | I | FX-01、FX-07、FX-10 |
| AC-18 | Must | TC-AC18-01～02 | I | FX-07 |
| AC-19 | Must | TC-AC19-01～03 | I | FX-07 |
| AC-20 | Must | TC-AC20-01～02 | I | FX-01、FX-03 變體 |
| AC-21 | Should | TC-AC21-01～02 | I＋人工截圖 | FX-08 |
| AC-22 | Must | TC-AC22-01～04 | I | FX-01、FX-07、FX-10 |
| AC-23 | Must | TC-AC23-01～04 | I（04 為 Unit 建議） | FX-08 |
| AC-24 | Must | TC-AC24-01～03 | I | FX-08 |
| AC-25 | Must | TC-AC25-01 | I | FX-01 |
| AC-26 | Should | TC-AC26-01～06 | I | FX-08、FX-01 |
| AC-27 | Must | TC-AC27-01～03 | I | FX-11 |
| AC-28 | Must | TC-AC28-01～05 | I | FX-11 |
| AC-29 | Must | TC-AC29-01～03 | I | FX-01 |
| AC-30 | Must | TC-AC30-01～03 | I | FX-09、FX-05 變體 |
| AC-31 | Must | TC-AC31-01～03 | I | FX-01 |
| AC-32 | Must | TC-AC32-01～02 | I | FX-01 |
| AC-33 | Must | TC-AC33-01～03 | I、E | — |
| AC-34 | Must | TC-AC34-01～02 | I（G3）、E（G4） | — |
| AC-35 | Must | TC-AC35-01～04（＋G-4 全面套用） | I | FX-01 |
| AC-36 | Must | TC-AC36-01～03 | I、E | 錯誤類 TC 共用 |
| AC-37 | Must | TC-AC37-01～05 | E | 乾淨機器 |
| AC-38 | Must | TC-AC38-01～03 | E（G4） | GitHub Releases |

覆蓋檢查：AC-01～AC-38 共 38 條，每條至少 1 個 TC，**全數覆蓋**。

NFR 對照：NFR-01→TC-NFR01-01；NFR-02→TC-NFR02-01；NFR-03→TC-NFR03-01、TC-AC37-01～04；NFR-04→TC-NFR04-01～02、AC-35；NFR-05→TC-NFR05-01；NFR-06→TC-NFR06-01；NFR-07→TC-NFR07-01、TC-AC24-03；NFR-08→TC-NFR08-01、TC-AC26-06。M2→TC-M2-01。

案例數量：AC 類 TC 135 個、NFR／M2 類 10 個、情境類 4 個，合計 149 個。

---

## 11. 待釐清事項（不猜，等 PM／Architect 回覆）

| # | 相關 AC/NFR | 問題 | 為什麼不能自己決定 | QA 暫定做法 |
|---|-------------|------|--------------------|-------------|
| D-01 | AC-03 | 同一 email、author date 完全相同但名字不同時，顯示哪個名字？ | PRD 只說「最新」，平手未定義 | 只驗證輸出穩定（10 次相同） |
| D-02 | NFR-01／02 | 「統計整段歷史」但起訖日必填；用 `--since 1970-01-01 --until 2099-12-31` 是否等同整段歷史？效能是否只看 warm cache？ | PRD 未定義 | 用此範圍；warm cache 判定、cold cache 只記錄 |
| D-03 | AC-08 | 改名同時修改內容（例如相似度 90%）怎麼算？只算實際改的行，還是整檔刪除＋新增？相似度門檻多少？ | AC-08 只定義「內容完全不變」 | 暫定只算實際改的行（1／1）；判定前標無法驗證 |
| D-04 | AC-07、08 | 結果是否要不受使用者個人 Git 設定（`diff.renames`、`core.quotepath`、`log.showSignature` 等）影響？ | PRD 未提 | QA 認為應不受影響（否則同 repo 不同人結果不同，違反 M1），暫列為高風險測試 |
| D-05 | AC-17 | 表格的表頭文字、框線或分隔符、數字是否可有千分位？ | AC-17 只規定欄位順序與對齊；Q-15 只說表頭用英文 | 只驗欄位順序、對齊、數字可解析；請 Architect 在 ARCHITECTURE.md 定義版面 |
| D-06 | AC-13 | 日期值前後有空白時，報錯還是自動去除？ | PRD 未提 | 依 A-05「只接受 `YYYY-MM-DD`」暫定報錯（exit 2） |
| D-07 | AC-19 | 「email 字母順序（不分大小寫）」對數字、`.`、`-`、`+`、非 ASCII 字元的順序？ | 「字母順序」可解讀為字元碼順序或語系排序 | 暫定為「轉小寫後依 Unicode 字元碼逐字比較」 |
| D-08 | AC-21 | emoji、組合字元（如 `ë` 拆成 e＋附加符號）的顯示寬度如何計算？ | AC-21 只定義全形＝2 | 只記錄、不判失敗 |
| D-09 | AC-02 | repo 有 `.mailmap` 時，是否必須「不套用」？ | 4.2 只說 MVP 不「支援」，沒說遇到時要忽略 | 暫定必須忽略（結果只依 commit 上的 email） |
| D-10 | AC-22、25 | CSV 行尾用 LF 還是 CRLF（RFC 4180 寫 CRLF）？最後一行要不要換行？三平台是否一致？ | PRD 未規定 | 兩者都接受，但要求三平台一致 |
| D-11 | AC-23 | 不含特殊字元的欄位能否也加引號？ | RFC 4180 允許，PRD 例子未明說 | 接受，記錄 |
| D-12 | AC-26 | `--output` 寫檔失敗時 exit code 是 1 嗎？ | 約定的 1 只列「讀取失敗」 | 暫定 1 |
| D-13 | AC-26 | `--output` 搭配表格格式時的行為？ | AC-26 只描述 CSV | 只記錄 |
| D-14 | AC-27、13 | 同時有路徑錯誤與參數錯誤時，exit code 用 1 還是 2？ | PRD 未定義優先順序 | 接受 1 或 2，stdout 必須 0 byte |
| D-15 | AC-28 | bare repo、或直接把 `.git` 資料夾當路徑時怎麼處理？ | PRD 未提 | 只記錄 |
| D-16 | AC-31 | 多給一個位置參數時是否屬於「使用方式錯誤」？ | PRD 只列「缺參數」與「未知選項」 | 暫定 exit 2 |
| D-17 | AC-32 | `--format CSV`（大寫）是否接受？ | PRD 未提 | 只記錄 |
| D-18 | AC-33 | `--help` 與其他參數或無效參數並用時，是否仍 exit 0？ | PRD 未提 | 暫定 exit 0 |
| D-19 | AC-35、4.3 | 工具執行時若觸發 Git 更新 `.git/index` 的檔案時間戳記快取（內容不變），算不算違反「永遠唯讀」？ | AC-35 的三項比對不會發現，但 4.3 寫「任何寫入」 | 記錄，若發生回報老闆決定 |
| D-20 | AC-37 | 「下載 Release 檔案後 1 個步驟」：解壓縮＋加入 PATH 算幾步？Windows 是否需要處理 SmartScreen、macOS 是否需要處理 Gatekeeper（未簽章程式的安全警告）？ | 會直接影響 AC-37 判定 | 每個手動動作算 1 步，安全警告的處理也算 1 步 |
| D-21 | NFR-07 | Windows 預設主控台字碼頁（非 UTF-8）下，表格中文顯示是否也要正確？ | PRD 未定義 Windows 終端機設定 | 兩種都測，預設字碼頁的結果只記錄 |
| D-22 | NFR 全部 | NFR 是 Must 還是 Should？ | PRD 第 6 節未標 | 依 9.3 第 4 點處理 |

---

## 12. 後續工作
1. 本計畫推上後，等各開發任務（T-xxx）的 PR 開好，QA 依第 4 節產出 Q-xxx 任務單，請 Claude Code 在該 task/T-xxx 分支上撰寫 fixture 腳本、Integration 測試與 CI job。
2. `ARCHITECTURE.md` 推上後，QA 只讀它的對外介面（指令與選項名稱、表格版面）更新第 1.3 節的對照與 D-05，預期結果不因實作而改變。
3. 待釐清 D-01～D-22 由幕僚長轉 PM／Architect；回覆後更新本文件版本。

## 變更紀錄
| 版本 | 日期 | 作者 | 變更內容 |
|------|------|------|----------|
| v0.1 | 2026-11-22 | QA | 初版（交接單 TP-001），依 PRD v1.0 設計 |
