# git-stats-cli 測試計畫（TEST_PLAN）

- 文件擁有者：QA
- 版本：v0.4（交接單 TP-004：對齊 ARCHITECTURE v1.1、PRD v1.3）
- 最後更新：2026-11-22
- 預期結果的依據（唯一）：`docs/PRD.md` v1.3（commit `0c55219`；AC-01～AC-45、NFR-01～NFR-08，含第 6 節 NFR 優先度、嚴重度與判定條件）。v1.1→v1.3 只改第 6 節與 Q-17（v1.3 定案 NFR-06～08＝Should），**AC 內容未變更**
- 只用於「對外介面」與「測試基礎設施」的參考：`docs/ARCHITECTURE.md` v1.1（commit `5033d3c`；G2 已於 2026-11-22 00:44 核准：Go＋呼叫 git CLI；PoC A-001 已核准）的 §3 CLI 規格、§7 輸出格式、§8 錯誤訊息格式、§12／§12.1 測試做法（含 §12.1.9 對本計畫待釐清事項的回覆）。若 ARCHITECTURE 與 PRD 衝突，以 PRD 為準並列入第 11 節。

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
1. PRD 第 5 節 F1～F6 全部 AC（AC-01～AC-45，Must 43 條、Should 2 條：AC-21、AC-26；AC-39～AC-45 為 v1.1 新增）。
2. PRD 第 6 節 NFR-01～NFR-08（優先度、違反時嚴重度）與「NFR 判定條件」。
3. PRD 第 3 節成功指標中可由 QA 驗證的部分：M1（統計正確率，由本計畫的 fixture 承擔）、M2（三平台安裝時間，不含安裝 Git，G4 前實測）、M4（Must AC 通過率，記在 QA_REPORT）。
4. PRD 7.1 假設 A-03（作者＝author，不是 committer）、A-04（行數語意同 `git log --numstat`）、A-05（日期只接受 `YYYY-MM-DD`）衍生的邊界條件。
5. 第 5 節開頭的 exit code 約定（0／1／2）與「錯誤時 stdout 無輸出」規則，套用到所有案例。

### 1.2 不測範圍（PRD 4.2／4.3 範圍外，或非 QA 黑箱可驗證）
- JSON 輸出、分支篩選、檔案／路徑篩選、自訂排序、相對日期、含時間的日期輸入、時區選項、`.mailmap`、`Co-authored-by`、總計列／百分比／趨勢、多 repo、PyPI／npm 上架（4.2）。
  - 例外：「不支援」本身會被測（例如 `--format json` 必須報錯＝AC-32；含時間的日期必須報錯＝AC-13 邊界；`.mailmap` 存在時不得被套用＝TC-AC02-04）。
- GUI、網頁儀表板、績效評分（4.3）。
- M3（發佈後老闆在真實 repo 抽查）：由老闆回報，QA 只提供核對步驟範本（見 7.4）。
- 實作內部的單元測試內容：由 TechLead／Claude Code 在開發任務中撰寫；QA 只檢查它們存在並在 CI 通過，**不作為驗收依據**。
- NFR-05 的 Code review 部分：依獨立性原則 QA 不讀實作程式碼；QA 只做行為面驗證（無網路、無對外連線），Code review 由 TechLead 負責並在 PR 記錄（見 TC-NFR05-01）。

### 1.3 名稱約定（依 ARCHITECTURE §3）
- 指令 `git-stats`；選項 `--since`、`--until`、`--format table|csv`、`--include-merges`、`--output FILE`、`--help`／`-h`、`--version`；repo 路徑為恰好 1 個位置參數（positional argument，直接寫在指令後面、沒有選項名稱的參數）。名稱與 PRD 建議名稱相同（恆等對照）。
- ARCHITECTURE §3 另外允許：旗標放在路徑前或後、`--since=2026-01-01` 寫法、`--` 之後一律視為位置參數、同一旗標重複時以最後一次為準。這些是 PRD 沒有規定的對外介面，本計畫以「記錄並驗證與 ARCHITECTURE 一致」處理（TC-CLI-01～03），不作為 PRD AC 的判定依據。
- 所有案例的指令以 `git-stats <repo> --since <起日> --until <迄日> [其他選項]` 表示；`<FX-xx>` 代表該 fixture repo 的路徑。

---

## 2. 測試層級分工（對齊 ARCHITECTURE §12）

G2 已決定以 Go 實作，ARCHITECTURE §12／§12.1 是**測試基礎設施的唯一依據**。v0.1 規劃的「POSIX shell fixture 腳本＋`tests/fixtures/`＋另一套 Integration 套件」全部取消，全專案只有一套 fixture 機制與一套黑箱套件（ARCHITECTURE §12.1 開頭；回應 TASKS.md 問題 #1）。v0.2 起層級名稱改用 ARCHITECTURE 的叫法：v0.1 的「Integration（I）」＝現在的 **E2E**；v0.1 的「E2E（E）」＝現在的 **REL**（安裝／發佈驗證）。

| 層級代號 | ARCHITECTURE §12 對應 | 位置 | 驗什麼 | 誰寫 | 是否為 QA 驗收依據 |
|----------|----------------------|------|--------|------|--------------------|
| Unit | 單元測試、Fuzz | 各 package `*_test.go` | 純邏輯 | T-xxx（Claude Code） | 否（QA 只確認 CI 全綠） |
| —（不另編 TC） | 整合測試 | `internal/gitlog` | 真 git 輸出被正確解析 | T-010 | 否（輔助） |
| **E2E** | 端對端（E2E） | `test/e2e/`（§12.1.2、§12.1.5） | `go build` 出的執行檔以子程序執行，只看 stdout／stderr／exit code／檔案（黑箱） | T-013～T-015 與 QA 的 Q-xxx（Claude Code 撰寫） | **是（主要）** |
| — | 架構測試 | `internal/archtest` | 含 `ac_list` 覆蓋檢查（第 13 節） | T-001、T-016 | 否（但 `ac_list` 內容由 QA 決定） |
| **P** | 效能（§12.1.8，build tag `perf`） | `test/e2e`，build tag `perf` | NFR-01、NFR-02 | 負責人待幕僚長指定（TASKS 問題 #6）；判定方式依本計畫第 6 節 | **是** |
| **REL** | 安裝／發佈（CI＋G4 人工） | 乾淨機器、GitHub Releases | AC-34（Release 比對）、AC-37、AC-38、AC-45、M2、S1～S4 | QA 執行 | **是**（G4） |
| **M** | — | 人工 | 錯誤訊息品質、截圖、Code review 紀錄檢查 | QA | 是 |

原則：
- **黑箱不變**：E2E 層只能透過執行檔的對外介面觀察結果；測試程式可以 import `internal/testutil/fixture`（建 repo 用），**不得 import 產品 package**（`cli`、`stats`、`render`、`gitlog`、`daterange` 等）來算預期值。成功案例的預期 stdout 是 QA 手寫的黃金檔（§12.1.7，不提供 `-update`），錯誤案例的預期值寫死在測試中，來源都是本計畫第 4.3 節的手算表。
- **測試命名與位置（依 ARCHITECTURE §12.1.2）**：fixture 定義在 `test/e2e/fixtures_test.go`（每個 FX 一個 `buildFXnn(t)`，內容逐字依第 4.3 節）；測試函式名 `TestE2E_ACnn_<描述>`，子測試名＝本計畫的 TC ID（例：`t.Run("TC-AC22-01", …)`）；黃金檔 `test/e2e/testdata/<FX-ID>/<TC-ID>.stdout`。T 任務與 Q-xxx 用同一套命名，以 TC ID 對應本計畫。QA 不相信開發方寫的預期值：驗證時逐一比對黃金檔、`fixtures_test.go` 與第 4.3 節是否一致，不一致即為缺陷。
- **不在 `go test` 內的 TC**：REL、M 層，以及標「只記錄」的案例，不需要（也不應該）寫成會失敗的斷言。
- **OS 矩陣**：E2E 在 NFR-03 全部平台的 CI 上執行（第 8 節）。

---

## 3. 通用檢查（每個案例都要套用）

為避免每條案例重複，以下規則套用到所有 TC，任一條不符即判該 TC 失敗：

| 編號 | 適用 | 檢查內容 | 依據 |
|------|------|----------|------|
| G-1 | 所有預期成功的執行 | exit code = 0，且 **stderr 為空（0 byte）** | AC-36、Q-07 |
| G-2 | 所有預期失敗的執行 | stdout **為 0 byte**；stderr 非空；exit code 依約定（1＝執行期錯誤、2＝使用方式錯誤） | 第 5 節約定、AC-36 |
| G-3 | 所有錯誤訊息 | 英文；同時說明「出了什麼錯」與「怎麼修正」（例如附正確格式或用法），以人工檢查清單判定：①指出錯誤的參數或路徑 ②給出修正方式 | NFR-06、Q-15 |
| G-4 | 所有對 fixture 的執行（`--output` 指向 repo 內者除外，見 P-07） | 執行前後 fixture 的 `git status --porcelain`、`git for-each-ref`、`.git/config` 完全相同 | AC-35（在 E2E 套件中全面套用，不只 AC-35 本身） |
| G-5 | 所有執行 | 測試環境無網路（見 NFR-04）時結果不變 | NFR-04 |
| G-6 | 數字欄位 | 十進位整數、不含千分位、不含小數點 | AC-22（CSV 明訂；表格依 ARCHITECTURE §7.1，D-05） |

「資料列」定義：表格輸出中除表頭（與可能存在的框線／分隔線）以外、每位作者一行的內容；CSV 中除第 1 行表頭以外的每一筆紀錄。

---

## 4. 測試資料（fixture repo）

### 4.1 建置原則（依 ARCHITECTURE §12.1）
1. **測試當下建立**：每個 fixture 由 `internal/testutil/fixture` 的 API（§12.1.3：`New`、`Commit(Change{Author, Committer, Write, WriteBytes, Rename, Delete, AllowEmpty})`、`Branch`、`Checkout`、`Detach`、`Merge`（含解衝突與 octopus）、`Git`（逃生口：tag、`update-ref`、orphan、bare clone 等）、`Snapshot`、`Generate`）在 `t.TempDir()` 中建立；不把 `.git` 資料夾 commit 進 repo。v0.1 的 shell 腳本與 `tests/fixtures/` 取消。
2. **隔離**：照 §12.1.4（建 fixture 與執行產品都套用）：`HOME`／`USERPROFILE`／`XDG_CONFIG_HOME` 指向空暫存目錄、`GIT_CONFIG_NOSYSTEM=1`、`GIT_CONFIG_GLOBAL=<空檔>`、`GIT_TERMINAL_PROMPT=0`；repo 內 `core.autocrlf=false`、`core.safecrlf=false`、`core.filemode=false`、`commit.gpgsign=false`、`gc.auto=0`、`core.quotepath=off`；每個 commit 以 `GIT_AUTHOR_*`／`GIT_COMMITTER_*` 環境變數固定名字、email、日期。
3. **日期**：一律用含時區偏移的 ISO 8601（例如 `2026-01-10T12:00:00+08:00`）；除非案例特別要求（AC-12、A-03），committer 與 author 相同。
4. **內容**：文字檔以固定字串（`line 1`…`line N`、LF、最後一行有換行，除非案例另有指定）；binary 檔用固定 bytes（PNG 檔頭 `\x89PNG\r\n\x1a\n` 加固定 64 byte，含 NUL）。
5. **可重現性檢查（`EXPECTED_SHA`，§12.1.4）**：每個 FX 的 HEAD SHA 記在 `test/e2e/testdata/EXPECTED_SHA`（首次建置產生、QA 確認後鎖定）；不符時測試以「fixture untrusted」失敗，該輪判為「無法驗證」而不是通過。三平台／不同 git 版本 SHA 是否相同由 PoC A-001 驗證。
6. **預期值**：每個 fixture 的預期結果表在第 4.3 節，由 QA 依 PRD 手算，不是跑工具得到的。
7. **時區（§12.1.6）**：Linux／macOS 以環境變數 `TZ=Asia/Taipei` 啟動產品。Go 在 Windows **不讀** `TZ`（ARCHITECTURE v1.0 已查證並更正），Windows 做法（W1：`tzutil` 改系統時區；W2：產品主動讀 `TZ`）等 PoC A-001 與 ADR-005 決定；在那之前 Windows 的 AC-11 案例標「待 A-001」。
8. **使用者 Git 設定注入**：TC-AC08-03、TC-CFG-01 在「執行產品時」用 `RunOpts.Env` 把 `GIT_CONFIG_GLOBAL` 改指向含搗亂設定的檔案（§12.1.4、§12.1.5）；AC-39 用 `RunOpts.Env` 改 `PATH`。
9. **執行產品（§12.1.5）**：`TestMain` 編譯一次（版本注入 `9.8.7`）；`run(t, RunOpts{Args, Dir, Env, Timeout})` 不經 shell 啟動，分開擷取 stdout、stderr 原始 bytes 與 exit code；工作目錄預設在 repo 外。每個跑在 fixture 上的案例前後都做 `Snapshot()` 比對（G-4）。

### 4.2 Fixture 清單

| ID | 名稱 | 內容摘要 | 使用的 TC |
|----|------|----------|-----------|
| FX-01 | basic | A、B 兩位作者；期間內外各有 commit；含 `src/` 子資料夾 | AC-01、17、22、26、29、33～36、Regression |
| FX-02 | identity | 同 email 不同名字、email 大小寫不同、author date 與 commit 順序相反；`.mailmap` 變體 | AC-02～04、AC-40 |
| FX-03 | merges | `--no-ff` merge、含衝突解決內容的 merge、octopus merge（3 個父 commit） | AC-05、06 |
| FX-04 | binary-rename | binary 新增／修改／刪除、純改名、跨資料夾改名、改名＋小改 | AC-07、08 |
| FX-05 | branches | main＋未合併的 feature 分支、detached HEAD | AC-09、AC-41 |
| FX-06 | dates-tz | 日期邊界、跨時區 author date、author≠committer date | AC-10～12 |
| FX-07 | sorting | A=5、B=9、C=1、x@=4、m@=4、大小寫混合 email 平手 | AC-18、19 |
| FX-08 | unicode-csv | 中文、日文、emoji 名字與 email；含逗號、雙引號、公式字元的名字 | AC-21、23、24、43、44、NFR-07 |
| FX-09 | empty | 剛 `git init`，0 個 commit | AC-20、25、30 |
| FX-10 | misc-lines | 空 commit、刪檔、CRLF 檔、無結尾換行、author≠committer、大量行數 | AC-01 邊界、A-03、A-04、G-6 |
| FX-11 | bad-paths | 不存在路徑、一般資料夾、一般檔案、無讀取權限的資料夾、損壞的 repo、bare repo、含空白與中文的路徑 | AC-27、28、36、42 |
| FX-12 | perf-10k / perf-1k | `fixture.Generate`（`git fast-import`）產生 10,000／1,000 個 commit（§12.1.8，build tag `perf`） | NFR-01、02 |

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
- 7：A，2026-01-24，改名 `notes.txt`→`notes2.txt` 並修改其中 1 行（相似度 90%）→ 預期 1／1／1（D-03 已關閉）。

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
- `<tmp>/no/such/dir`（不存在）；`<tmp>/plain-dir`（存在、不在任何 repo 內；須確認其所有上層目錄都不是 repo）；`<tmp>/plain-file.txt`；`<tmp>/no-perm`（`chmod 000` 的 repo，僅 POSIX，且以非 root 身分執行）；`<tmp>/corrupt`（複製 FX-01 後刪除一個 object 檔）；`<tmp>/含 空白 與中文/repo`（FX-01 的複本，路徑含空白與中文）；`<tmp>/bare.git`（`git clone --bare` FX-01）。
（FX-11 中非 repo 的路徑由測試直接以檔案系統操作建立；bare repo 以 `r.Git("clone", "--bare", …)` 或 `git init --bare` 建立。）

**FX-12 perf**：見第 6.2 節。

---

## 5. 測試案例

格式：每個 TC 有 ID、等級（Must／Should，沿用 PRD）、層級（E2E＝`test/e2e` 黑箱、P＝效能、REL＝安裝／發佈、M＝人工；定義見第 2 節）、前置條件、步驟、預期結果。所有 TC 另外套用第 3 節通用檢查 G-1～G-6。標「介面檢查」的句子是依 ARCHITECTURE 的對外介面（訊息字串、版面）所加的檢查：不符時記為「低」缺陷並回報 Architect／TechLead 對齊，**不影響該 AC 依 PRD 的判定**。

### F1 作者統計（Must）

#### AC-01 基本統計（Must）
- **TC-AC01-01**（E2E）前置：FX-01。步驟：`git-stats <FX-01> --since 2026-01-01 --until 2026-01-31`。預期：恰好 2 列資料；A：commits=3、added=30、deleted=5；B：commits=1、added=2、deleted=0；exit 0。
- **TC-AC01-02**（E2E）同上加 `--format csv`。預期：CSV 解析後恰好 2 筆，數值同上。（用 CSV 交叉驗證表格解析，避免表格版面問題掩蓋數字錯誤。）
- **TC-AC01-03**（E2E，邊界：範圍外不計）前置：FX-01。步驟：`--since 2025-01-01 --until 2026-12-31`。預期：A 4／37／5、B 2／6／0。
- **TC-AC01-04**（E2E，邊界：單一 commit）前置：FX-01。步驟：`--since 2026-01-06 --until 2026-01-06`。預期：只有 B 1 列，1／2／0。
- **TC-AC01-05**（E2E，邊界：作者＝author 不是 committer，A-03）前置：FX-10（以 `Change.Committer` 設定與 author 不同的 committer）。預期：h@example.com 1 列 commits=1、added=2；輸出中沒有 k@example.com。
- **TC-AC01-06**（E2E，邊界：行數語意，A-04）前置：FX-10。預期：E 1／0／0（空 commit 仍計 commit）；F 2／20／20（刪檔計刪除行）；G 2／8／0（CRLF 與無結尾換行不影響行數）。
- **TC-AC01-07**（E2E，邊界：大數值，G-6）前置：FX-10。CSV 與表格中 L 的 added 均為 `12345`（表格見 D-05）。
- **TC-AC01-08**（E2E，邊界：root commit）前置：FX-01 範圍 2025-12-01～2025-12-31。預期：A 1／7／0（第一個 commit 的新增行數要計入）。

#### AC-02 同 email 不同名字合併（Must）
- **TC-AC02-01**（E2E）前置：FX-02。步驟：範圍 2026-01-01～2026-01-31，`--format csv`。預期：email 為 `alice@example.com` 的資料列恰好 1 筆，commits=3、added=12、deleted=3。
- **TC-AC02-02**（E2E）同上用表格格式。預期：只有 1 列含 `alice@example.com`。
- **TC-AC02-03**（E2E，反向：不同 email 同名字不合併）前置：在 FX-02 複本加 1 個 `Alice / alice2@example.com` 的 commit。預期：`alice@` 與 `alice2@` 是 2 列。
- **TC-AC02-04**（E2E，範圍外功能不得生效）前置：FX-02 複本在根目錄加 `.mailmap` 把 `alice@example.com` 對到 `other@example.com`、名字改成 `Mapped`，以 author `Z / z@example.com`、日期 2026-02-01（範圍外）commit。預期：1 月的結果與 TC-AC02-01、TC-AC03-01 完全相同，輸出中沒有 `other@example.com`、`Mapped`（PRD v1.1 AC-40 已定案「明確忽略 `.mailmap`」；完整案例見 TC-AC40-01～02）。

#### AC-03 合併後顯示的名字（Must）
- **TC-AC03-01**（E2E）前置：FX-02。預期：`alice@example.com` 列的名字為 `Alice Wang`。
- **TC-AC03-02**（E2E，邊界：author date 與歷史順序相反）前置：FX-02。預期：`carol@example.com` 列名字為 `Carol Lee`（author date 最晚），不是歷史上最後一個 commit 的 `Carol`。
- **TC-AC03-03**（E2E，邊界：名字依「期間內」決定）前置：FX-02 範圍改為 2026-01-03～2026-01-04。預期：alice 列名字為 `Alice`（期間內只有署名 Alice 的 commit）。
- **TC-AC03-04**（E2E，邊界：同一 author date 兩個不同名字，D-01 已關閉）前置：FX-02 複本加 2 個 email 為 `tie@example.com` 的 commit，author date 都是 2026-01-25T09:00:00+08:00；先 commit 的署名 `Tie Old`（committer date 2026-01-25T09:00），後 commit（它的子 commit）署名 `Tie New`（committer date 2026-01-25T10:00）。預期：名字為 `Tie New`（ARCHITECTURE §6 規則 4：時間相同時取 git log 輸出中較先出現者，即較新的那個）；連續執行 10 次結果相同。

#### AC-04 email 大小寫視為同一人（Must）
- **TC-AC04-01**（E2E）前置：FX-02。預期：bob 只有 1 列，email 顯示 `bob@example.com`，commits=2、added=3。
- **TC-AC04-02**（E2E，邊界：只有大寫形式）範圍 2026-01-06～2026-01-06（只含 `Bob@Example.com`）。預期：email 仍顯示為 `bob@example.com`。
- **TC-AC04-03**（E2E，CSV）`--format csv` 時 email 欄同樣是全小寫。

#### AC-05 預設排除 merge commit（Must）
- **TC-AC05-01**（E2E）前置：FX-03。步驟：範圍 2026-01，不加 merge 選項。預期：A commits=2、added=13、deleted=2；Z 1／1／1。
- **TC-AC05-02**（E2E，邊界：octopus merge）前置：FX-03b。預期：octopus merge commit 不計入。
- **TC-AC05-03**（E2E，邊界：只有 merge commit 的作者）前置：FX-03 複本，由另一位作者 M2 做一次 `--no-ff` merge（無其他 commit）。預期：預設輸出中沒有 M2 這一列。

#### AC-06 選項納入 merge commit（Must）
- **TC-AC06-01**（E2E）前置：FX-03。步驟：加 `--include-merges`。預期：A commits=3、added=13、deleted=2（與 TC-AC05-01 的行數相同）。
- **TC-AC06-02**（E2E，邊界：evil merge）同上，確認 merge 解衝突時多加的 2 行**不**計入。
- **TC-AC06-03**（E2E，邊界：octopus）FX-03b 加 `--include-merges`：commit 數多 1，行數不變。
- **TC-AC06-04**（E2E，邊界：只有 merge 的作者）TC-AC05-03 的 repo 加 `--include-merges`：出現 M2 一列，1／0／0。
- **TC-AC06-05**（E2E，CSV）以 `--format csv --include-merges` 執行，數值同 TC-AC06-01。

#### AC-07 binary 檔的增刪行數（Must）
- **TC-AC07-01**（E2E）前置：FX-04。步驟：`--since 2026-01-10 --until 2026-01-10`。預期：A 1／10／0。
- **TC-AC07-02**（E2E，邊界：binary 修改）範圍 2026-01-22～2026-01-22。預期：A 1／0／0。
- **TC-AC07-03**（E2E，邊界：binary 刪除）範圍 2026-01-23～2026-01-23。預期：A 1／0／0。

#### AC-08 檔案改名不重複計算（Must）
- **TC-AC08-01**（E2E）前置：FX-04。範圍 2026-01-20～2026-01-20。預期：A 1／0／0。
- **TC-AC08-02**（E2E，邊界：跨資料夾改名）範圍 2026-01-21～2026-01-21。預期：A 1／0／0。
- **TC-AC08-03**（E2E，邊界：使用者 Git 設定 `diff.renames=false`）執行產品時把 `GIT_CONFIG_GLOBAL` 指向含 `diff.renames=false` 的設定檔，重跑 TC-AC08-01。預期：仍為 1／0／0（D-04 已關閉：結果不得因使用者個人 Git 設定而改變，否則同一 repo 不同人跑出不同數字，違反 PRD M1）。其他設定見 TC-CFG-01。
- **TC-AC08-04**（E2E，邊界：改名＋小改）範圍 2026-01-24～2026-01-24。預期：A 1／1／1（D-03 已關閉：PRD A-04「行數與 `git log --numstat` 概念一致」；改名判定門檻依 git `-M` 預設 50%，ARCHITECTURE §12.1.9）。另加一個相似度低於 50% 的變體（改名且改掉 10 行中的 8 行）：預期當作「刪舊檔＋新增新檔」，記錄結果；此變體的行數只記錄、不作退件依據，因為 PRD 沒有規定門檻。

#### AC-09 統計範圍為目前 HEAD（Must）
- **TC-AC09-01**（E2E）前置：FX-05（HEAD=main）。預期：輸出不含 c@example.com；A 2／2／0。
- **TC-AC09-02**（E2E，反向）FX-05 複本 `git checkout feature`。預期：出現 C 2／10／0、A 1／1／0。
- **TC-AC09-04**（E2E，邊界：遠端追蹤分支與 tag 不影響）FX-05 複本加一個指向 feature 的 tag 與 `refs/remotes/origin/feature`。預期：同 TC-AC09-01。

### F2 日期範圍（Must）

#### AC-10 起訖日期都包含在內（Must）
- **TC-AC10-01**（E2E）前置：FX-06，`TZ=Asia/Taipei`。步驟：`--since 2026-01-01 --until 2026-01-31`。預期：D10 commits=3（#2 00:00:00、#3、#4 23:59:59 計入；#1、#5 不計入）。
- **TC-AC10-02**（E2E，同日查詢）`--since 2026-01-15 --until 2026-01-15`。預期：D10 commits=1（#3）。
- **TC-AC10-03**（E2E，邊界：只含起日當天 00:00:00）`--since 2026-01-01 --until 2026-01-01`。預期：D10 commits=1（#2），並含 TZ（見 AC-11）。
- **TC-AC10-04**（E2E，邊界：只含迄日當天 23:59:59）`--since 2026-01-31 --until 2026-01-31`。預期：D10 commits=1（#4）。
- **TC-AC10-05**（E2E，邊界：閏日）在 FX-06 複本加一個 2028-02-29T12:00+08:00 的 commit，`--since 2028-02-29 --until 2028-02-29`。預期：commits=1。

#### AC-11 時區以本地時區解讀（Must）
- **TC-AC11-01**（E2E）前置：FX-06，`TZ=Asia/Taipei`。`--since 2026-01-01 --until 2026-01-01` 時 TZ 列（#6）計入；`--since 2025-12-31 --until 2025-12-31` 時 TZ 列不出現。
- **TC-AC11-02**（E2E，反向）同一 fixture：`--since 2025-12-31 --until 2025-12-31` 時 TZ2 列（#7，字面日期 2026-01-01，台北時間 2025-12-31 19:00）計入；`--since 2026-01-01 --until 2026-01-01` 時不出現。
- **TC-AC11-03**（E2E，Windows 時區）Linux／macOS 用 `TZ=Asia/Taipei`。Windows：Go 不讀 `TZ`（ARCHITECTURE §12.1.6），做法依 PoC A-001／ADR-005（W1：以 `tzutil /s "Taipei Standard Time"` 改系統時區、測完還原；W2：產品讀 `TZ`，另保留 1 個 `tzutil` 案例驗證「未設 `TZ` 時跟隨系統時區」）。不論採哪一個，Windows 結果都必須與 TC-AC11-01、02 相同。A-001 完成前此 TC 在 Windows 標「待 A-001」（無法驗證，不阻擋其他任務，但 G3 前必須有結果）。
- **TC-AC11-04**（E2E，換時區結果會變）同一 fixture 改用 `TZ=America/Los_Angeles` 執行 `--since 2025-12-31 --until 2025-12-31`。預期：TZ 列（洛杉磯時間 2025-12-31 23:30）計入。證明確實是「執行環境本地時區」，不是寫死 UTC+8。
- **TC-AC11-05**（E2E，邊界：夏令時間切換日）`TZ=America/New_York`，fixture 加 2026-03-08T00:30-05:00 與 2026-03-08T23:30-04:00 兩個 commit，`--since 2026-03-08 --until 2026-03-08`。預期：兩個都計入（該日只有 23 小時，仍是整天）。

#### AC-12 以 author date 判斷（Must）
- **TC-AC12-01**（E2E）前置：FX-06。`--since 2026-01-01 --until 2026-01-31`：R 列計入；`--since 2026-02-01 --until 2026-02-28`：R 列不出現。
- **TC-AC12-02**（E2E，反向）R2（author 2/15、committer 1/15）：1 月不出現、2 月出現。

#### AC-13 日期格式錯誤（Must）
- **TC-AC13-01**（E2E）前置：FX-01。依序以下列值作為 `--since`（`--until` 給合法值）：`2026/01/01`、`01-31-2026`、`2026-1-1`、`abc`。預期每次：stderr 含該錯誤值原文與字串 `YYYY-MM-DD`；stdout 0 byte；exit 2。（訊息格式參考 ARCHITECTURE §8：第一行以 `git-stats: error:` 開頭、另有一行修正方式；測試以子字串斷言，不比對整句。）
- **TC-AC13-02**（E2E）同上，但錯誤值放在 `--until`。
- **TC-AC13-03**（E2E，邊界）錯誤值：空字串 `""`、`20260101`、`2026-01-01T00:00:00`、`2026-01-01 13:00`（含時間，PRD 4.2／A-05 不支援）、` 2026-01-01`（前後空白）、全形數字 `２０２６-０１-０１`、`2026-01-1`、`02026-01-01`。預期：皆同 TC-AC13-01（exit 2）。前後空白不自動去除（D-06 已關閉，依 PRD A-05 只接受 `YYYY-MM-DD`）。
- **TC-AC13-04**（E2E，邊界：相對日期不支援）`yesterday`、`last month`、`7 days ago`。預期：exit 2（PRD 4.2）。

#### AC-14 不存在的日期（Must）
- **TC-AC14-01**（E2E）`--since 2026-02-30` 與 `--since 2026-13-01`（分別執行）。預期：同 AC-13 報錯方式，exit 2。
- **TC-AC14-02**（E2E，邊界）`2026-02-29`（非閏年）、`2026-04-31`、`2026-00-10`、`2026-01-00`、`2026-01-32`。預期：exit 2。
- **TC-AC14-03**（E2E，反向：合法閏日）`2028-02-29`。預期：被接受（exit 0）。
- **TC-AC14-04**（E2E，不得自動換算）對 TC-AC14-01 確認 stdout 為 0 byte（若工具把 `2026-02-30` 換算成 3/2 並輸出結果，即失敗）。

#### AC-15 起日晚於迄日（Must）
- **TC-AC15-01**（E2E）`--since 2026-03-01 --until 2026-02-01`。預期：stderr 有「起日晚於迄日」類型的英文訊息（建議包含兩個日期）；stdout 0 byte；exit 2。
- **TC-AC15-02**（E2E，邊界：只差一天）`--since 2026-01-02 --until 2026-01-01`。預期：exit 2。
- **TC-AC15-03**（E2E，反向：相同日期合法）見 TC-AC10-02。

#### AC-16 缺少日期參數（Must）
- **TC-AC16-01**（E2E）只給 `--until`。預期：stderr 含錯誤訊息＋用法說明（usage；ARCHITECTURE §8 為 `Run 'git-stats --help' for usage.` 加精簡 usage 行），stdout 0 byte，exit 2。
- **TC-AC16-02**（E2E）只給 `--since`。預期同上。
- **TC-AC16-03**（E2E）兩者都不給。預期同上。
- **TC-AC16-04**（E2E，邊界：選項有名稱沒有值）`git-stats <FX-01> --since --until 2026-01-31` 與結尾 `--until`（無值）。預期：exit 2，stdout 0 byte。

### F3 表格輸出與排序（Must；AC-21 為 Should）

#### AC-17 預設表格輸出與欄位（Must）
- **TC-AC17-01**（E2E）前置：FX-01。不加 `--format`。預期：第 1 行為表頭，欄位順序為 名字、email、commit 數、新增行數、刪除行數；之後每位作者一列。表頭文字依 ARCHITECTURE §7.1 為 `Name`、`Email`、`Commits`、`Added`、`Deleted`（英文，Q-15），沒有框線或分隔線（D-05 已關閉）。
- **TC-AC17-02**（E2E，對齊）以表頭各欄位置為基準：文字欄（名字、email）每列左緣在同一字元位置；數字欄每列右緣在同一字元位置（以 FX-01、FX-07 這類純 ASCII 名字驗證）。介面檢查（ARCHITECTURE §7.1）：欄間至少 2 個空白、行尾沒有空白。
- **TC-AC17-03**（E2E，邊界：數字位數不同）FX-10（含 12345 與 0）。預期：數字欄仍右對齊；數字以十進位整數顯示，不含千分位（D-05 已關閉，ARCHITECTURE §12.1.9）。
- **TC-AC17-04**（E2E，`--format table` 明確指定）結果與不指定時逐位元組相同。

#### AC-18 依 commit 數由多到少排序（Must）
- **TC-AC18-01**（E2E）前置：FX-07，表格。預期：資料列順序 B、A、（m、n、x）、C。
- **TC-AC18-02**（E2E）同上 `--format csv`，順序相同。

#### AC-19 commit 數相同時的次排序（Must）
- **TC-AC19-01**（E2E）前置：FX-07。預期：4 個 commit 的三人順序為 `m@example.com`、`n@example.com`、`x@example.com`（不分大小寫比較）。表格與 CSV 都驗。
- **TC-AC19-02**（E2E，穩定性）表格與 CSV 各連續執行 10 次，stdout 逐位元組相同。
- **TC-AC19-03**（E2E，邊界：email 共同前綴）加 `a@example.com` 與 `a.b@example.com`、`ab@example.com` 各 1 個 commit 的變體。預期：email 轉小寫後依 Unicode 碼位逐字比較的字典序（D-07 已關閉，ARCHITECTURE §6 規則 5）：`a.b@example.com`、`a@example.com`、`ab@example.com`（`.`＝0x2E 小於 `@`＝0x40 小於 `b`＝0x62）。

#### AC-20 查無資料時的表格輸出（Must）
- **TC-AC20-01**（E2E）前置：FX-01，範圍 `--since 2030-01-01 --until 2030-01-31`。預期：stdout 恰好 1 行（非空）提示訊息，**不含表頭**；exit 0；stderr 空。介面檢查：內容為 ARCHITECTURE §7.1 的 `No commits found between 2030-01-01 and 2030-01-31.`。
- **TC-AC20-02**（E2E，邊界：只有 merge commit 落在範圍內）TC-AC05-03 的 repo，範圍只涵蓋 M2 的 merge 那天，預設模式。預期：同 TC-AC20-01（沒有「符合條件」的 commit）。

#### AC-21 中日韓全形字元對齊（Should）
- **TC-AC21-01**（E2E）前置：FX-08。表格輸出。預期：以「全形（East Asian Wide/Fullwidth）字元＝2 欄寬、其他＝1 欄寬」計算，每列各欄起始顯示欄位相同。驗證方式：測試程式依 Unicode East Asian Width 計算顯示寬度後比對；另在 Windows Terminal、macOS Terminal、GNOME Terminal 截圖人工確認。
- **TC-AC21-02**（E2E，邊界：emoji、帶附加符號字母）`Zoë 🚀` 的寬度判定 PRD 沒有規定（D-08 已關閉為「只記錄」）：只要求「不當機、不亂碼」（NFR-07），對齊結果記錄但不判失敗。
- **TC-AC21-03**（E2E，邊界：控制字元，只記錄）名字含 tab（fixture 名字 `Tab\tName`；若 git 不接受，本案例取消）。介面檢查（ARCHITECTURE §7.1）：表格中以字面 `\t` 顯示，各欄仍對齊；CSV 中保持原始 tab（TC-AC23-02 回讀一致）。

### F4 CSV 輸出（Must；AC-26 為 Should）

#### AC-22 CSV 格式與欄位（Must）
- **TC-AC22-01**（E2E）前置：FX-01，`--format csv`。預期：第 1 行逐字為 `name,email,commits,added,deleted`；之後 2 行；分隔符為半形逗號；順序 A、B。
- **TC-AC22-02**（E2E，邊界：數字）FX-10：`added` 為 `12345`，不得出現 `12,345`、`12345.0`、空白。
- **TC-AC22-03**（E2E，排序同 AC-18／19）FX-07 CSV 順序同 TC-AC19-01。

#### AC-23 CSV 特殊字元跳脫（Must）
- **TC-AC23-01**（E2E）前置：FX-08，`--format csv`。預期：`Wang, Alice` 輸出為 `"Wang, Alice"`；`Bob "The Builder"` 輸出為 `"Bob ""The Builder"""`。
- **TC-AC23-02**（E2E，回讀）用標準 RFC 4180 解析器讀回全部列，每列恰好 5 欄，名字、email 與 fixture 原始值逐字相同。
- **TC-AC23-03**（E2E，反向：不需要時不加引號）`Al` 的那行為 `Al,al@example.com,1,1,0`。介面檢查（D-11 已關閉，ARCHITECTURE §12.1.9：只在需要時加引號，欄位以空白開頭時也會加）；PRD 允許加引號，因此不符只記為「低」。
- **TC-AC23-04**（Unit 建議，非 QA 驗收）名字含換行：Git 不允許 author 名字含換行，無法用真實 repo 造出，請 TechLead 在 Unit 層覆蓋。QA 在報告中標「無法以黑箱驗證」。

#### AC-24 CSV 編碼（Must）
- **TC-AC24-01**（E2E）前置：FX-08，`--format csv`，stdout 以 byte 讀取。預期：開頭 3 byte 不是 `EF BB BF`（無 BOM）；整體可用嚴格 UTF-8 解碼成功；`王小明` 一列的 name 欄解碼後為 `王小明`。
- **TC-AC24-02**（E2E，跨平台）Windows 上用 `--output` 寫檔（見 AC-26）與 Git Bash 導出兩種方式，檔案都是 UTF-8 無 BOM。另記錄 PowerShell 5.1 用 `>` 導出的結果（PRD 已知會變 UTF-16，僅記錄、不判失敗）。
- **TC-AC24-03**（E2E，非中文非 ASCII）`山田太郎`、`Zoë 🚀`、`李四@例子.台灣` 回讀一致。

#### AC-25 查無資料時的 CSV 輸出（Must）
- **TC-AC25-01**（E2E）前置：FX-01，範圍 2030-01。`--format csv`。預期：stdout 逐位元組為 `name,email,commits,added,deleted\n`，沒有其他內容；exit 0（最後一行也以 LF 結尾，PRD AC-43）。

#### AC-26 直接寫入檔案（Should）
- **TC-AC26-01**（E2E）前置：FX-08，空暫存資料夾。步驟：`git-stats <FX-08> --since ... --until ... --format csv --output report.csv`。預期：產生 `report.csv`；內容與不加 `--output` 時的 stdout **逐位元組相同**；stdout 0 byte；stderr 0 byte；exit 0。
- **TC-AC26-02**（E2E，覆寫）先建一個內容較長的 `report.csv`，再執行。預期：檔案被完整覆寫（不殘留舊內容、長度等於新內容）。
- **TC-AC26-03**（E2E，查無資料）範圍 2030-01。預期：檔案只有表頭一行。
- **TC-AC26-04**（E2E，異常：無法寫入）`--output` 指向不存在的資料夾、唯讀資料夾（POSIX、非 root）、或一個已存在的資料夾名稱。預期：exit 1、stdout 0 byte、stderr 有說明（D-12 已關閉，ARCHITECTURE §7.3、§8）。
- **TC-AC26-05**（E2E，與表格格式併用，D-13 已關閉）`--output report.txt` 不加 `--format`。預期：產生 `report.txt`，內容與不加 `--output` 時的表格 stdout 逐位元組相同；stdout、stderr 皆 0 byte；exit 0（ARCHITECTURE §3、§7.3）。
- **TC-AC26-07**（E2E，相對路徑）在 FX-01 以外的工作目錄執行 `--output out/report.csv`（`out/` 已存在）。預期：檔案寫在「目前工作目錄」下的 `out/`，不是 repo 下（ARCHITECTURE §3）。
- **TC-AC26-06**（E2E，NFR-08）只有 `--output` 指定的檔案被建立，其他位置無新檔。

### F5 輸入檢查、錯誤處理與基本指令行為（Must）

#### AC-27 路徑不存在（Must）
- **TC-AC27-01**（E2E）`git-stats /no/such/dir --since 2026-01-01 --until 2026-01-31`（Windows 用 `C:\no\such\dir`）。預期：stderr 含「路徑不存在」類型英文訊息與該路徑原文；stdout 0 byte；exit 1。
- **TC-AC27-02**（E2E，邊界：路徑含空白與中文）`<tmp>/不存在 的/資料夾`。預期：同上，訊息中路徑完整不亂碼。
- **TC-AC27-03**（E2E，邊界：路徑不存在且日期也錯，D-14 已關閉）`git-stats /no/such/dir --since 2026/01/01 --until 2026-01-31`。預期：exit 2，stderr 報日期錯誤，stdout 0 byte（ARCHITECTURE §3 驗證順序：使用方式錯誤一律先於執行期錯誤）。

#### AC-28 路徑不是 Git repo（Must）
- **TC-AC28-01**（E2E）FX-11 `plain-dir`。預期：stderr 含「不是 Git repo」類型訊息與該路徑；stdout 0 byte；exit 1。
- **TC-AC28-02**（E2E）FX-11 `plain-file.txt`（一般檔案）。預期同上（exit 1）。
- **TC-AC28-03**（E2E，異常：無讀取權限，僅 POSIX、非 root）FX-11 `no-perm`。預期：exit 1，stdout 0 byte，stderr 有說明（依約定「讀取失敗」屬 1）。
- **TC-AC28-04**（E2E，異常：損壞的 repo）FX-11 `corrupt`，範圍涵蓋缺失的 object。預期：exit 1，stdout 0 byte（不得輸出部分結果）。
- **TC-AC28-06**（E2E，`.git` 資料夾本身當路徑，D-15 已關閉）`git-stats <FX-01>/.git …`。預期：exit 1、stderr 為「不是 Git repo」類型訊息並含該路徑、stdout 0 byte（ARCHITECTURE §12.1.9；與 PRD AC-42 同理，沒有工作目錄就不是可統計的 repo）。

#### AC-29 路徑是 repo 子資料夾（Must）
- **TC-AC29-01**（E2E）FX-01，以 `<FX-01>/src` 執行。預期：stdout 與以 `<FX-01>` 執行逐位元組相同（B 只改了 `docs/`、A 的 `c.txt` 在根目錄，仍要計入）。
- **TC-AC29-02**（E2E，CSV）同上 `--format csv`。
- **TC-AC29-03**（E2E，邊界：相對路徑與 `.`）在 `<FX-01>/src` 目錄下以 `.` 為路徑執行；在 `<FX-01>` 的上層以相對路徑 `FX-01` 執行。結果皆與 TC-AC29-01 相同。

#### AC-30 沒有任何 commit 的新 repo（Must）
- **TC-AC30-01**（E2E）FX-09，表格。預期：同 AC-20（1 行提示、exit 0、stderr 空）。
- **TC-AC30-02**（E2E）FX-09，`--format csv`。預期：同 AC-25（只有表頭）。
- **TC-AC30-03**（E2E，邊界：HEAD 指向尚未建立的分支，但其他分支有 commit）FX-05 複本 `git checkout --orphan newbranch` 並清空 index。預期：同 TC-AC30-01（依 AC-09，HEAD 可追溯的 commit 為 0）。

#### AC-31 缺少 repo 路徑參數（Must）
- **TC-AC31-01**（E2E）在 FX-01 目錄內執行 `git-stats --since 2026-01-01 --until 2026-01-31`。預期：stderr 有錯誤訊息＋用法說明；stdout 0 byte；exit 2（不得預設用目前資料夾，PRD Q-09）。
- **TC-AC31-02**（E2E，邊界：多給一個位置參數，D-16 已關閉）`git-stats <FX-01> <FX-02> --since ... --until ...`。預期：exit 2、stdout 0 byte（ARCHITECTURE §3「超過 1 個 → 使用方式錯誤」）。
- **TC-AC31-03**（E2E，異常：未知選項）`--foo`、`--include-merge`（拼錯）。預期：exit 2（第 5 節約定「未知選項」屬 2）。

#### AC-32 未知的輸出格式（Must）
- **TC-AC32-01**（E2E）`--format json`。預期：stderr 含錯誤訊息並列出 `table`、`csv`；stdout 0 byte；exit 2。
- **TC-AC32-02**（E2E，邊界）`--format xml`、`--format ""`、`--format`（無值）、`--format CSV`、`--format Table`（大小寫不同）。預期：全部 exit 2、stderr 列出 `table`、`csv`、stdout 0 byte（D-17 已關閉：區分大小寫，ARCHITECTURE §3）。

#### AC-33 說明（Must）
- **TC-AC33-01**（E2E）`git-stats --help`。預期：stdout 含 repo 路徑、起日、迄日、輸出格式、納入 merge commit 選項、日期格式 `YYYY-MM-DD` 的說明（逐項以關鍵字比對：`--since`、`--until`、`--format`、`table`、`csv`、`--include-merges`、`YYYY-MM-DD`，以及路徑參數名稱）；exit 0；stderr 空。
- **TC-AC33-02**（E2E，邊界，D-18 已關閉，ARCHITECTURE §3）以下都要 stdout 顯示說明、stderr 空、exit 0：在非 Git 資料夾執行 `git-stats --help`；`git-stats -h`；`git-stats --help --format json`；`git-stats /no/such/dir --help`；`git-stats --help --since 2026/01/01`。
- **TC-AC33-04**（E2E，邊界，D-18 已關閉）`git-stats --help --bogus` 與 `git-stats --bogus --help`（未知旗標，兩種順序）。預期：兩者都 exit 2、stdout 0 byte、stderr 指出未知旗標；結果不得因順序而不同（PRD 第 5 節約定「未知選項」屬使用方式錯誤；ARCHITECTURE v1.1 §3）。
- **TC-AC33-05**（E2E，D-18b 已關閉）`git-stats --help --version` 與 `git-stats --version --help`。預期：兩者都 stdout 只有用法說明（同 TC-AC33-01 的內容檢查）、**不含版本號那一行**、stderr 空、exit 0（`--help` 優先，ARCHITECTURE v1.1 §3；PRD AC-33／AC-34 只各自定義單獨使用的情況，沒有規定並用，屬介面細節）。另 `git-stats <FX-01> --since --help --until 2026-01-31`：`--help` 被當成 `--since` 的值，預期 exit 2、stderr 為日期格式錯誤（含值 `--help` 與 `YYYY-MM-DD`）、stdout 0 byte（PRD AC-13）。結尾 `--until`（缺值）→ exit 2（PRD AC-16，同 TC-AC16-04）。
- **TC-AC33-03**（REL）在 E2E 安裝後的乾淨機器上執行 TC-AC33-01。

#### AC-34 版本（Must）
- **TC-AC34-01**（E2E，G3 階段）`git-stats --version`。預期：stdout 含符合 `X.Y.Z`（語意化版本）的版本號；exit 0；stderr 空。介面檢查（ARCHITECTURE §3）：只有版本號一行、在非 Git 資料夾也可執行。E2E 以 `-ldflags` 注入 `9.8.7`，斷言 stdout 逐位元組為 `9.8.7\n`（ARCHITECTURE §12.1.5）。
- **TC-AC34-02**（REL，G4 階段）從 GitHub Release 下載 tag `vX.Y.Z` 的檔案安裝後執行。預期：版本號＝tag 去掉 `v`。G3 時尚無 Release，本 TC 標「無法驗證（階段未到）」，不影響 G3，但 G4 前必須通過。

#### AC-35 唯讀，不修改 repo（Must）
- **TC-AC35-01**（E2E）FX-01，執行前記錄 `git status --porcelain`、`git for-each-ref`、`.git/config`；執行有資料的查詢（表格、CSV、`--include-merges` 各一次）。預期：三者執行後完全相同。
- **TC-AC35-02**（E2E）同上，但查詢範圍無資料（2030-01）。預期同上。
- **TC-AC35-03**（E2E，加強檢查）在 FX-01 的工作目錄放一個未追蹤檔與一個已修改但未 commit 的檔案，再執行。預期：`git status --porcelain` 前後相同、工作目錄檔案內容不變。
- **TC-AC35-04**（E2E，加強檢查）執行前以 `r.SnapshotGitDir()` 記錄整個 `.git/` 的逐項清單（路徑、類型、大小、檔案內容 SHA-256、**檔案與資料夾的 mtime**），執行後再記一次，以 `fixture.DiffGitDir(before, after)` 比對（ARCHITECTURE v1.1 §12.1.3.1）；不排除任何檔案或資料夾，不比對 atime 與權限。`DiffGitDir` 回傳任何一項都判為缺陷（D-19：PRD 4.3「任何寫入」）：refs、config、objects、HEAD 有變＝致命；只有 `.git/index` 或其他快取／鎖檔痕跡（例如資料夾 mtime 因建立又刪除 `index.lock` 而改變）＝高。表格、CSV、`--include-merges`、查無資料各跑一次。已知限制（只記錄）：mtime 解析度內的同內容寫入無法偵測。
- **TC-AC35-05**（E2E）PRD v1.1 的 AC-35 已限定範圍：`--output` 指向 repo **外**時，三項比對完全相同（必測）。`--output <FX-01>/report.csv`（指向 repo 內）不在 AC-35 範圍：只記錄，預期 refs 與 `.git/config` 不變、`git status --porcelain` 只多一行 `?? report.csv`。
- 另：G-4 讓整個 E2E 套件的每次執行都做 TC-AC35-01 的三項比對（`--output` 指向 repo 內的案例除外）。

#### AC-36 錯誤與結果分流（Must）
- **TC-AC36-01**（E2E）對 TC-AC13-01、TC-AC14-01、TC-AC15-01、TC-AC16-01～03、TC-AC27-01、TC-AC28-01、TC-AC31-01、TC-AC32-01 的每個指令，執行 `git-stats ... > out.txt 2> err.txt`。預期：`out.txt` 為 0 byte；`err.txt` 非空。
- **TC-AC36-02**（E2E）對 TC-AC01-01、TC-AC20-01、TC-AC22-01、TC-AC25-01 的指令同樣導向。預期：`err.txt` 為 0 byte。
- **TC-AC36-03**（REL，Windows）在 PowerShell 與 cmd 各做一次 TC-AC36-01 的代表案例（AC-13、AC-27），確認 Windows 上同樣分流。

### F6 安裝與發佈（Must）

#### AC-37 跨平台單一指令安裝（Must）
- **TC-AC37-01**（REL）乾淨 Windows 11 x86_64（只裝 Git for Windows）：依 README 安裝（1 行指令，或下載 Release 檔案後 1 個步驟）→ 關閉並重開終端機 → `git-stats --version`。預期：符合 AC-34；全程計時並記錄（M2 ≤5 分鐘）。
- **TC-AC37-02**（REL）乾淨 macOS（當時最新兩個大版本各一台；至少一台 Apple Silicon，NFR-03 另需 x86_64 一台）。預期同上。
- **TC-AC37-03**（REL）乾淨 Ubuntu 24.04 x86_64。預期同上。
- **TC-AC37-04**（REL，NFR-03 延伸）Ubuntu 22.04 x86_64。預期同上。
- **TC-AC37-05**（REL，安裝後跑核心案例）在 TC-AC37-01～04 的每台機器上，以安裝好的指令跑 TC-AC01-01、TC-AC22-01、TC-AC24-01、TC-AC27-01。
- 「1 個步驟」的判定：計算使用者需要輸入的指令數或手動動作數（解壓縮＋放到 PATH 算 2 步）。見待釐清 D-20。

#### AC-38 GitHub Release 發佈（Must）
- **TC-AC38-01**（REL，G4 階段）查看 `https://github.com/hansheng08/git-stats-cli/releases`（或 `gh release view vX.Y.Z`）。預期：tag 格式為 `vX.Y.Z`；有對應 Release；說明內有變更摘要與 Windows／macOS／Linux 三平台安裝方式。
- **TC-AC38-02**（REL）tag 指向的 commit 在 main 上。
- **TC-AC38-03**（REL）PyPI 與 npm 搜尋產品名稱（`git-stats`、`git-stats-cli`）不存在本專案發佈的套件（若有同名他人套件，確認發佈者不是本專案）。
- G3 時本 AC 標「無法驗證（階段未到）」，G4 前必須通過。

### 非功能需求（NFR）

- **TC-NFR01-01**（P）見第 6 節：perf-10k、整段歷史、表格與 CSV 各一組量測，全部 ≤ 10 秒。
- **TC-NFR02-01**（P）見第 6 節：perf-1k、整段歷史，全部 ≤ 2 秒。
- **TC-NFR03-01**（E2E）全部 E2E 案例在 NFR-03 OS 矩陣（Windows 11 x86_64、macOS 兩個大版本含 Apple Silicon 與 x86_64、Ubuntu 22.04、24.04 x86_64）上通過。見第 8 節。
- **TC-NFR04-01**（E2E）在無網路環境跑完整 E2E 套件（`go test ./test/e2e/...`，測試二進位先在有網路時建好）：Linux 用 `unshare -rn` 或 `docker run --network none`。預期：結果與有網路時完全相同。
- **TC-NFR04-02**（E2E）Linux 上以 `strace -f -e trace=network` 執行 TC-AC01-01 與 TC-AC20-01。預期：沒有 `connect` 到任何網路位址（AF_INET／AF_INET6）的系統呼叫。
- **TC-NFR05-01**（M）行為面同 TC-NFR04-01／02；Code review 部分由 TechLead 在 PR 中確認「無遙測、無上傳」並記錄，QA 只檢查 PR 上有此紀錄，不讀程式碼（見 1.2）。
- **TC-NFR06-01**（M）對所有錯誤類 TC（AC-13～16、27、28、31、32）的 stderr 以 G-3 清單人工判定，結果表附在 QA_REPORT。
- **TC-NFR07-01**（E2E）FX-08 表格與 CSV 都不當機、不出現 `?`、`\uXXXX`、亂碼；名字與 email 回讀一致。三平台都要驗（Windows 在 Windows Terminal＋UTF-8 字碼頁與預設字碼頁各一次，見待釐清 D-21）。
- **TC-NFR08-01**（E2E）使用 §12.1.4 預設的隔離環境：`HOME`／`USERPROFILE`／`XDG_CONFIG_HOME`，以及 `TMPDIR`／`TEMP`／`TMP`／`APPDATA`／`LOCALAPPDATA`（ARCHITECTURE v1.1 起**預設**就指向每個測試自己的空資料夾，不需另外覆寫）。執行代表性案例（成功、查無資料、錯誤、`--output`）前後列出這些目錄的所有檔案。預期：除 `--output` 指定的檔案外，沒有新增或修改任何檔案（PRD NFR-08）。另確認測試啟動時這些目錄確實是空的且不是系統原本的位置，否則本 TC 判「無法驗證」。
- **TC-M2-01**（REL）同 TC-AC37-01～03 的計時，三平台各 ≤ 5 分鐘（從開啟 README 到看到第一次統計結果）。**不含安裝 Git 的時間**（PRD v1.1 M2）：計時從「已裝好 Git 的乾淨機器開啟 README」開始。

### PRD v1.1 新增 AC（AC-39～AC-45，全部 Must）

#### AC-39 找不到 Git 或 Git 版本過舊（Must）
- **TC-AC39-01**（E2E）以 `RunOpts.Env` 把 `PATH` 設為不含 `git` 的暫存資料夾，執行 `git-stats <FX-01> --since 2026-01-01 --until 2026-01-31`。預期：stderr 說明找不到 Git 並提示如何安裝；stdout 0 byte；exit 1。
- **TC-AC39-02**（E2E）`PATH` 最前面放一個假的 `git`（測試在 `TestMain` 另外編譯的小程式，對 `--version` 回 `git version 2.30.0`，Windows 為 `git.exe`）。預期：stderr 含目前版本 `2.30.0` 與最低需求 `2.34`，並提示升級；stdout 0 byte；exit 1。
- **TC-AC39-03**（E2E，邊界）假 git 回 `git version 2.34.0`（剛好等於最低版本；ARCHITECTURE v1.1 §12.1.5 的 fakegit，對 `--version` 以外的參數一律失敗）。預期：stderr **不含**「版本過舊／需要 2.34」的訊息（PRD AC-39 只對「低於 2.34」報版本錯誤）；因假 git 無法提供資料，後續 git 呼叫失敗屬執行期錯誤，預期 **exit 1**、stdout 0 byte（PRD 第 5 節約定：讀取失敗＝1）。另在 CI 的 Git 2.34 job（§12.1.5）以真正的 Git 2.34 跑完整 E2E 套件，全部通過（NFR-03、AC-39）。
- **TC-AC39-04**（E2E，邊界：使用方式錯誤優先）沒有 git 時執行 `git-stats <FX-01> --since 2026/01/01 --until 2026-01-31`。預期：exit 2（日期錯誤），不是 1（PRD 約定：參數錯誤不受環境影響；ARCHITECTURE §3 驗證順序）。
- **TC-AC39-05**（E2E）沒有 git 時 `git-stats --help`、`git-stats --version` 仍 exit 0。

#### AC-40 明確忽略 `.mailmap`（Must）
- **TC-AC40-01**（E2E）前置：FX-02 變體，範圍內有 `Old / old@example.com` 1 個、`New / new@example.com` 1 個 commit，根目錄 `.mailmap` 內容為 `Mapped Name <new@example.com> <old@example.com>`（該檔由範圍外的 commit 加入）。預期：`old@example.com` 與 `new@example.com` 各 1 列，名字分別為 `Old`、`New`，沒有 `Mapped Name`。
- **TC-AC40-02**（E2E）同一 repo 以 `r.Git("rm", ".mailmap")` 再以範圍外日期 commit 後重跑（範圍不變）。預期：stdout 與 TC-AC40-01 逐位元組相同（表格與 CSV 各驗一次）。
- **TC-AC40-03**（E2E，邊界）使用者 Git 設定 `log.mailmap=true`、`mailmap.file=<另一個 mailmap>` 時，結果仍與 TC-AC40-01 相同（見 TC-CFG-01）。

#### AC-41 detached HEAD 照常統計（Must）
- **TC-AC41-01**（E2E）FX-05 複本 `r.Detach(<commit 2 的 SHA>)`。預期：exit 0；stdout 與「在指向 commit 2 的分支（main）上執行」逐位元組相同（A 2／2／0）。
- **TC-AC41-02**（E2E，邊界）detached 到 feature 上的 commit 4。預期：與在 `feature` 分支上執行相同（A 1、C 2）。

#### AC-42 bare repo 視為不是 Git repo（Must）
- **TC-AC42-01**（E2E）`git init --bare <tmp>/empty-bare.git`。預期：stderr 為「不是 Git repo」類型訊息並含該路徑；stdout 0 byte；exit 1。
- **TC-AC42-02**（E2E）`git clone --bare <FX-01> <tmp>/bare.git`（有 commit 的 bare repo）。預期：同 TC-AC42-01（不得輸出統計結果）。

#### AC-43 CSV 換行固定為 LF（Must）
- **TC-AC43-01**（E2E）FX-01、FX-08 以 `--format csv` 執行，讀 stdout 原始 bytes。預期：每一行（含表頭與最後一行）都以 `0x0A` 結尾；整段輸出不含 `0x0D`；最後一個 byte 是 `0x0A`。
- **TC-AC43-02**（E2E）同上改用 `--output report.csv`。預期：檔案 bytes 與 TC-AC43-01 的 stdout 相同。
- **TC-AC43-03**（E2E，跨平台）同一份黃金檔在 Windows、macOS、Linux 三平台的 CI 比對都通過（＝三平台逐位元組相同；黃金檔需 `-text` 屬性，避免 Windows checkout 自動轉 CRLF，ARCHITECTURE §12.1.7）。
- **TC-AC43-04**（E2E，邊界，只記錄）名字本身含 CR 的作者（若 git 接受）：PRD 允許欄位值內的 CR，只記錄。

#### AC-44 CSV 公式字元原樣輸出（Must）
- **TC-AC44-01**（E2E）FX-08 加 4 位作者，名字分別為 `=1+1`、`+cmd`、`-x`、`@SUM(A1)`，各 1 個 commit。`--format csv`。預期：這 4 行的 name 欄原樣出現（不加 `'`、空白或其他前綴；`@SUM(A1)` 不含逗號所以不加引號）；以 `encoding/csv` 讀回與原值逐字相同。
- **TC-AC44-02**（E2E，邊界）名字 `=HYPERLINK("http://x","y")`（含逗號與雙引號）。預期：依 AC-23 加引號並把 `"` 寫成 `""`，此外不做任何處理；讀回一致。

#### AC-45 Release 產物平台（Must，G4）
- **TC-AC45-01**（REL，G4）`gh release view vX.Y.Z --json assets` 列出下載檔。預期：含 Windows x86_64、macOS（Apple Silicon 與 x86_64）、Linux x86_64 的產物，以及 **Linux arm64** 的壓縮檔；**沒有** Windows arm64 的產物。
- **TC-AC45-02**（REL，G4）在 Linux arm64 機器（或 runner）下載該產物執行 `git-stats --version`。預期：版本號符合 AC-34。

### 對外介面與使用者設定（依 ARCHITECTURE，不屬於 PRD AC）
- **TC-CFG-01**（E2E，D-04）執行產品時的使用者 Git 設定含：`log.mailmap=true`、`diff.renames=false`、`log.showSignature=true`、`core.quotepath=true`、`color.ui=always`、`log.diffMerges=on`、`core.pager=false`、`diff.external=<一個會寫檔的程式>`、`i18n.logOutputEncoding=ISO-8859-1`。對 FX-01、FX-02、FX-03、FX-04、FX-08 重跑 TC-AC01-01、TC-AC02-01、TC-AC05-01、TC-AC06-01、TC-AC07-01、TC-AC08-01、TC-AC24-01、TC-AC40-01。預期：結果與乾淨設定時逐位元組相同；外部 diff 程式沒有被執行（它要寫的檔案不存在）。
- **TC-CLI-01**（E2E，介面）旗標放在路徑前：`git-stats --since 2026-01-01 --until 2026-01-31 <FX-01>`；`--since=2026-01-01` 寫法。預期：結果與 TC-AC01-01 相同（ARCHITECTURE §3）。
- **TC-CLI-02**（E2E，介面）同一旗標重複：`--format csv --format table`。預期：以最後一次為準，輸出表格（ARCHITECTURE §3）。
- **TC-CLI-03**（E2E，介面）路徑以 `-` 開頭：建立名為 `-repo` 的 FX-01 複本，執行 `git-stats -- -repo --since … --until …`（`--` 之後視為位置參數）。預期：正常統計、exit 0（ARCHITECTURE §3、§9 指令注入對策）。

### PRD 情境（E2E 串接）
- **TC-S1-01**（REL）老闆情境：對 FX-01 查上個月整月（2026-01-01～2026-01-31），表格輸出，最多的人在最前。
- **TC-S2-01**（REL）技術主管情境：FX-08 輸出 CSV 檔，用 Google Sheets 匯入與 Excel「從文字匯入」開啟（人工截圖），中文不亂碼、逗號名字不錯欄。
- **TC-S3-01**（REL）維護者情境：FX-02＋FX-03 組合，同 email 合併、預設不含 merge。
- **TC-S4-01**（REL）腳本情境：用 shell 腳本連跑成功與失敗各一次，依 exit code 分支處理；失敗時輸出檔為空。

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

### 6.2 測試資料產生器（perf-10k／perf-1k，依 ARCHITECTURE §12.1）
- 以 `fixture.Generate(t, dir, GenSpec{Commits: n, Seed: seed, …})`（回傳 `(*Repo, Totals)`；`GenSpec` 的其餘參數逐項對應本節的形狀；內部用 `git fast-import`，Git 內建的批次匯入工具，可在幾秒內造出上萬個 commit）建立，放在 build tag `perf` 的測試中，不在每次 `go test` 執行（ARCHITECTURE §12.1.8）。固定 seed，同一版本每次造出的 HEAD SHA 必須相同（記入 `EXPECTED_SHA`）。
- **形狀以本節為準**（ARCHITECTURE v1.1 §12.1.8 已採納）：A-001 的效能數據交給 QA，由 QA 判斷是否修改本節；產生器不得自行採用其他參數。保留在磁碟上：`go test -tags perf -run TestPerfGenerate -count=1 ./test/e2e/... -args -perf.out=<目錄>` 產生 `<目錄>/perf-1k`、`<目錄>/perf-10k` 與 `perf-*.totals.json`，供 `hyperfine` 使用。
- perf-10k 的形狀（模擬「中型真實專案」）：
  - HEAD 可追溯的 commit 總數恰好 **10,000**（含 merge commit），全在 main 線上。
  - 作者 50 位（其中 5 位有 2 種名字、3 位 email 大小寫不同、5 位中文或 emoji 名字）。
  - 約 2%（200 個）為 merge commit（兩個父 commit），其餘為一般 commit。
  - 檔案 500 個文字檔、分在 20 個資料夾，外加 20 個 binary 檔；每個一般 commit 改 1～5 個檔，每檔新增 1～50 行、刪除 0～30 行；約 1% commit 含改名，且**一律是純改名**（被改名的檔案內容完全不變，numstat 0／0；同一 commit 可另外修改其他檔案），每一行文字在整個歷史中唯一，使產生器算出的預期總數與 git numstat 必然一致（ARCHITECTURE v1.1 §12.1.8 的產生器限制）。「改名＋修改」不放進 perf fixture，由 FX-04 的 TC-AC08-04 覆蓋（見第 11.2 節說明）。
  - author date 平均分布在 2021-01-01～2025-12-31。
- perf-1k：同一產生器、commit 數 1,000，其他比例相同。
- 產生器同時回傳「預期總數表」`Totals`（鍵＝小寫 email，每位作者 commit／增刪行數；由產生器依自己寫入的內容計算，**不得呼叫產品程式**），效能測試順便驗證結果正確（M1）。產生器本身的正確性：QA 用 `git log --numstat` 對產生的 repo 抽 3 位作者手算核對一次。

### 6.3 量測方式
- 指令（整段歷史）：`git-stats <perf-10k> --since 1970-01-01 --until 2099-12-31`（表格）與加 `--format csv`（stdout 導向 `/dev/null` 以排除終端機顯示時間）。「整段歷史」需要給一個涵蓋全部的日期範圍，因 MVP 起訖日必填（Q-09）；perf fixture 的日期全落在 2021～2025，此範圍即等於整段歷史（D-02 已關閉；PRD v1.2 第 6 節已明定此範圍，且要求 fixture 日期全部落在範圍內）。
- **判定以本節的量測為準**：ARCHITECTURE §12.1.8 的 Go perf 測試（暖機 1 次、取 5 次中位數）是開發中的快速檢查；NFR 通過與否以本節的方法在參考機器上量測（ARCHITECTURE v1.1 §12.1.8 已同意以本節為準）。
- 工具：`hyperfine --warmup 3 --runs 10 --export-json perf.json '<指令>'`（執行檔為 `go build` 的正式版本、fixture 由 `fixture.Generate` 產生後保留在磁碟上）；若無法安裝 hyperfine，改用 `/usr/bin/time -v` 手動跑 3 次暖機＋10 次量測。
- **判定只看 warm cache**（PRD v1.3 第 6 節「NFR 判定條件」，D-02）：10 次量測的**最大值**（也就是每一次都）≤ 10 秒（NFR-02 為 ≤ 2 秒）。另在報告中列出 median、p95、平均、標準差。
  - 選用「最大值」而不是中位數：PRD 寫「≤ 10 秒完成」，使用者每一次執行都應達到，用中位數會放過偶發超時。
- 輔助量測（cold cache，只記錄）：每次量測前執行 `sync; echo 3 | sudo tee /proc/sys/vm/drop_caches`（hyperfine `--prepare`），跑 5 次記錄 median 與最大值。**cold cache 結果只記錄、不判定**（PRD v1.3 第 6 節）；若超過 10 秒，在 QA_REPORT 列為資訊。
- 同時記錄最大記憶體用量（`/usr/bin/time -v` 的 Maximum resident set size），僅記錄。
- 雜訊控制：量測時 runner 上不跑其他 job；每個 job 只跑效能測試；結果超標時自動重跑整組一次，兩組都超標才判失敗（避免共用機器的偶發干擾），兩組結果都寫入報告。

---

## 7. 回歸測試（Regression）清單

### 7.1 每個 PR（CI 必跑，約數分鐘）
| 編號 | 內容 |
|------|------|
| RG-01 | `go test ./...` 全部（含 `test/e2e` 的 `TestE2E_*`、`internal/archtest` 的 `ac_list` 覆蓋檢查）在 Ubuntu 24.04 上執行 |
| RG-02 | 黃金檔（golden file，事先存好、確認正確的預期輸出檔）比對：`test/e2e/testdata/<FX-ID>/<TC-ID>.stdout` 逐位元組比對。黃金檔由 QA 依第 4.3 節手算後手寫（ARCHITECTURE §12.1.7，不提供 `-update`）；**不得以修改黃金檔讓測試通過**，任何修改或放寬都需經老闆核准 |
| RG-03 | G-4 唯讀檢查（每次執行都做） |
| RG-04 | 所有錯誤類案例的 exit code 與 stdout 0 byte（AC-13～16、27、28、31、32、36） |
| RG-05 | fixture 可重現性：`test/e2e/testdata/EXPECTED_SHA` 比對（ARCHITECTURE §12.1.4） |

### 7.2 每次合併前（G3 前，QA 執行）
| 編號 | 內容 |
|------|------|
| RG-10 | RG-01～05 在 NFR-03 全部 OS 矩陣執行，另加 Git 2.34 一組（TC-AC39-03） |
| RG-11 | 效能測試 TC-NFR01-01、TC-NFR02-01 |
| RG-12 | 無網路、NFR-08、使用者 Git 設定干擾檢查（TC-NFR04-01、TC-NFR08-01、TC-CFG-01） |
| RG-13 | 先前所有已修復缺陷的「鎖住測試」（fix/Q-xxx 分支加入的失敗測試，修好後永久保留） |

### 7.3 發佈前（G4 前）
| 編號 | 內容 |
|------|------|
| RG-20 | TC-AC37-01～05、TC-AC38-01～03、TC-AC34-02、TC-AC45-01～02、TC-M2-01 |
| RG-21 | TC-S1-01～TC-S4-01 |

### 7.4 M3 核對範本（給老闆，非 QA 驗收）
對真實 repo 隨機抽 3 位作者：`git log --author=<email> --since=<起日> --until=<迄日 23:59:59> --no-merges --numstat --format=` 手動加總（注意 `--author` 是部分比對，email 大小寫與名字合併需人工處理），與工具輸出比對。

---

## 8. 環境需求

| 項目 | 需求 |
|------|------|
| CI | GitHub Actions；矩陣：`ubuntu-22.04`、`ubuntu-24.04`、`windows-2025`（或當時對應 Windows 11 等級的 runner；**不確定** GitHub 是否提供「Windows 11」runner，若只有 Windows Server，Windows 11 改用 VM 人工驗證）、`macos-latest` 與前一版（Apple Silicon）、一個 x86_64 macOS（若 GitHub 已無 x86_64 macOS runner，改用人工或標「無法驗證」並請老闆決定） |
| Git | 每個平台記錄版本；最低版本 2.34（PRD AC-39），CI 另跑一組 2.34；Windows 用 Git for Windows |
| 語言與測試工具 | Go（版本依 `go.mod`）、`go test`；CSV 回讀用 Go 標準庫 `encoding/csv`（只在測試中用來讀輸出，不呼叫產品 package）；顯示寬度判定依 Unicode East Asian Width（測試中自行判定，不得 import 產品的 `render`） |
| 量測與檢查工具 | `hyperfine`、`/usr/bin/time`、`strace`（Linux）、`unshare` 或 Docker（無網路測試）、`sha256sum`；雲端缺工具依 `docs/RUNBOOK.md`（O-001 撰寫中）安裝 |
| 時區 | Linux／macOS 用 `TZ`；Windows 依 PoC A-001／ADR-005（W1 `tzutil` 需要管理員權限，GitHub Windows runner 有；ARCHITECTURE §12.1.6） |
| 權限 | 權限測試（TC-AC28-03）需以非 root 身分執行 |
| 雲端副本 | QA 在自己的 worktree 或 clone（`/workspace/qa/` 底下，不在共用的 `/workspace/projects/git-stats-cli` 工作區暫存或 commit）checkout 分支執行；缺工具依 `docs/RUNBOOK.md` 安裝，RUNBOOK 沒寫則回報幕僚長請 SRE 補上 |
| E2E 機器 | 乾淨的 Windows 11、macOS（最新兩個大版本）、Ubuntu 22.04／24.04 VM 或 runner，只預裝 Git |
| 無網路 | Linux 上以 network namespace 或 Docker `--network none` |

---

## 9. 通過／退件標準

### 9.1 每條 AC 的判定
- **通過**：該 AC 所有對應 TC（不含標「待釐清、只記錄」者）在要求的平台全部通過。
- **失敗**：任一對應 TC 的預期結果不符。
- **無法驗證**：環境或階段不允許（例如 G3 時尚無 Release）、fixture 不可重現（HEAD SHA 跨平台不同）、或 PRD 有待釐清事項導致無法判定。必須寫明原因與何時能驗證。

### 9.2 缺陷嚴重度
| 等級 | 定義 | 例子 |
|------|------|------|
| 致命（Critical） | 修改或損壞使用者 repo、對外傳送資料、無法安裝或無法執行 | AC-35 違反、NFR-04／05 違反 |
| 高（High） | 統計數字錯誤、Must AC 不符、exit code 錯誤、錯誤時 stdout 有輸出（會污染腳本產出的 CSV）、效能不達標、在某支援平台無法使用 | AC-01 數字錯、CSV 跳脫錯誤導致錯欄 |
| 中（Medium） | Should AC 不符、錯誤訊息不夠清楚（NFR-06）、只在罕見邊界才出錯且不影響數字 | AC-21 對齊偏差 |
| 低（Low） | 文字、排版等不影響使用的小問題 | help 文字錯字 |

### 9.3 G3 前的判定規則
1. **退件**：任一 Must AC 或 Must NFR 判「失敗」，或有任一來自 Must 項目、或不屬於任何 Should 項目的「高」以上缺陷。不得因為趕時間而放行。（Should 項目的缺陷見第 3、4 點。）
2. **Must AC 判「無法驗證」**：
   - 因階段未到（AC-34 的 Release 比對、AC-37、AC-38、AC-45、M2）→ 不影響 G3，但必須在 G4 前驗證，並在回報中列出。
   - Windows 的 AC-11（TC-AC11-03）在 PoC A-001 完成前標「待 A-001」；G3 前必須有結果。
   - 因其他原因（環境、待釐清）→ 狀態寫「卡住」，回報幕僚長說明需要誰提供什麼；QA 不自行假設通過。
3. Should AC（AC-21、AC-26）失敗：記為「中」缺陷，不單獨構成退件；是否放行由老闆決定，QA 在回報中附建議。
4. **NFR 的優先度與嚴重度（PRD v1.3 第 6 節）**：

   | NFR | 優先度 | 違反時嚴重度 | G3 判定 |
   |-----|--------|--------------|---------|
   | NFR-04、NFR-05（安全） | Must | 致命 | 未通過即**退件** |
   | NFR-01、NFR-02（效能）、NFR-03（平台） | Should | 高 | 記錄缺陷、**不阻擋 G3**（見 4b） |
   | NFR-06（可用性） | Should | 中 | 記錄缺陷、**不阻擋 G3**（見 4b） |
   | NFR-07（相容性）、NFR-08（資料保存） | Should | 高 | 記錄缺陷、**不阻擋 G3**（見 4b） |

   4b. **Should 類 NFR 未通過的處理**：照樣執行、照樣判定；未通過時在 QA_REPORT 記為缺陷，嚴重度依上表（NFR-01～03＝高），附重現步驟與量測數據。它本身**不構成退件、不阻擋 G3**。但 QA 在回報單「需要老闆決定的事」逐條列出這些缺陷，建議老闆在 G3 明確表示接受或要求修正；QA 不會因為它是 Should 就略過不報。Must AC、Must NFR 的失敗，以及非 Should 項目的高以上缺陷，仍一律退件（第 1 點）。
   4c. **NFR-06～08 的優先度依據**：PRD v1.3 第 6 節與 Q-17（已定案，2026-11-22 01:10 老闆）明定 NFR-06、NFR-07、NFR-08＝Should。v0.3.1 註記的「PRD 尚未反映」差異已消除。
4d. 「介面檢查」不符（ARCHITECTURE 有寫、PRD 沒要求的訊息字串或版面）＝低，不構成退件。
5. 全部 Must AC 與 Must NFR 通過，且沒有第 1 點所列的高以上缺陷 → 回報幕僚長，建議送 G3（若有 4b 的 Should 缺陷，一併列出請老闆決定）。
6. 只對「本次 PR 影響的 AC」做完整驗證；其餘 AC 由回歸測試（第 7 節）覆蓋，回歸任一失敗＝該 AC 失敗。

### 9.4 判定紀錄
每次驗證在 `docs/QA_REPORT.md` 新增一節（不覆蓋舊紀錄），包含：被驗證的分支與 commit SHA、執行環境、CI 結果連結、AC 逐條判定表、缺陷清單（重現步驟、預期、實際、嚴重度）、效能數據。

---

## 10. AC 對照表（AC → TC → 層級）

| AC | 等級 | TC | 層級 | Fixture | archtest `ac_list` |
|----|------|----|------|---------|-----|
| AC-01 | Must | TC-AC01-01～08、TC-CFG-01、TC-CLI-01 | E2E | FX-01、FX-10 | 列入 |
| AC-02 | Must | TC-AC02-01～04、TC-CFG-01 | E2E | FX-02 | 列入 |
| AC-03 | Must | TC-AC03-01～04 | E2E | FX-02 | 列入 |
| AC-04 | Must | TC-AC04-01～03 | E2E | FX-02 | 列入 |
| AC-05 | Must | TC-AC05-01～03、TC-CFG-01 | E2E | FX-03、FX-03b | 列入 |
| AC-06 | Must | TC-AC06-01～05、TC-CFG-01 | E2E | FX-03、FX-03b | 列入 |
| AC-07 | Must | TC-AC07-01～03、TC-CFG-01 | E2E | FX-04 | 列入 |
| AC-08 | Must | TC-AC08-01～04、TC-CFG-01 | E2E | FX-04 | 列入 |
| AC-09 | Must | TC-AC09-01、02、04 | E2E | FX-05 | 列入 |
| AC-10 | Must | TC-AC10-01～05 | E2E | FX-06 | 列入 |
| AC-11 | Must | TC-AC11-01～05（03 的 Windows 部分待 A-001） | E2E | FX-06 | 列入 |
| AC-12 | Must | TC-AC12-01～02 | E2E | FX-06 | 列入 |
| AC-13 | Must | TC-AC13-01～04 | E2E | FX-01 | 列入 |
| AC-14 | Must | TC-AC14-01～04 | E2E | FX-01 | 列入 |
| AC-15 | Must | TC-AC15-01～03 | E2E | FX-01 | 列入 |
| AC-16 | Must | TC-AC16-01～04 | E2E | FX-01 | 列入 |
| AC-17 | Must | TC-AC17-01～04 | E2E | FX-01、FX-07、FX-10 | 列入 |
| AC-18 | Must | TC-AC18-01～02 | E2E | FX-07 | 列入 |
| AC-19 | Must | TC-AC19-01～03 | E2E | FX-07 | 列入 |
| AC-20 | Must | TC-AC20-01～02 | E2E | FX-01、FX-03 變體 | 列入 |
| AC-21 | Should | TC-AC21-01～03 | E2E＋人工截圖 | FX-08 | 列入 |
| AC-22 | Must | TC-AC22-01～03 | E2E | FX-01、FX-07、FX-10 | 列入 |
| AC-23 | Must | TC-AC23-01～04（04 為 Unit 建議） | E2E | FX-08 | 列入 |
| AC-24 | Must | TC-AC24-01～03、TC-CFG-01 | E2E | FX-08 | 列入 |
| AC-25 | Must | TC-AC25-01 | E2E | FX-01 | 列入 |
| AC-26 | Should | TC-AC26-01～07 | E2E | FX-08、FX-01 | 列入 |
| AC-27 | Must | TC-AC27-01～03 | E2E | FX-11 | 列入 |
| AC-28 | Must | TC-AC28-01～04、06 | E2E | FX-11 | 列入 |
| AC-29 | Must | TC-AC29-01～03 | E2E | FX-01 | 列入 |
| AC-30 | Must | TC-AC30-01～03 | E2E | FX-09、FX-05 變體 | 列入 |
| AC-31 | Must | TC-AC31-01～03 | E2E | FX-01 | 列入 |
| AC-32 | Must | TC-AC32-01～02 | E2E | FX-01 | 列入 |
| AC-33 | Must | TC-AC33-01～05 | E2E、REL | — | 列入 |
| AC-34 | Must | TC-AC34-01（G3）、TC-AC34-02（G4） | E2E、REL | — | 列入 |
| AC-35 | Must | TC-AC35-01～05（＋G-4 全面套用） | E2E | FX-01 | 列入 |
| AC-36 | Must | TC-AC36-01～03 | E2E、REL | 錯誤類 TC 共用 | 列入 |
| AC-37 | Must | TC-AC37-01～05 | REL（G4） | 乾淨機器 | **排除** |
| AC-38 | Must | TC-AC38-01～03 | REL（G4） | GitHub Releases | **排除** |
| AC-39 | Must | TC-AC39-01～05 | E2E | FX-01 | 列入 |
| AC-40 | Must | TC-AC40-01～03、TC-AC02-04、TC-CFG-01 | E2E | FX-02 變體 | 列入 |
| AC-41 | Must | TC-AC41-01～02 | E2E | FX-05 | 列入 |
| AC-42 | Must | TC-AC42-01～02 | E2E | FX-11 | 列入 |
| AC-43 | Must | TC-AC43-01～04 | E2E | FX-01、FX-08 | 列入 |
| AC-44 | Must | TC-AC44-01～02 | E2E | FX-08 | 列入 |
| AC-45 | Must | TC-AC45-01～02 | REL（G4） | GitHub Releases | **排除** |

覆蓋檢查：AC-01～AC-45 共 45 條（v1.0 的 38 條＋v1.1 新增 7 條），每條至少 1 個 TC，**全數覆蓋**。

NFR 對照：NFR-01→TC-NFR01-01；NFR-02→TC-NFR02-01；NFR-03→TC-NFR03-01、TC-AC37-01～04、TC-AC39-03；NFR-04→TC-NFR04-01～02、AC-35；NFR-05→TC-NFR05-01；NFR-06→TC-NFR06-01；NFR-07→TC-NFR07-01、TC-AC24-03；NFR-08→TC-NFR08-01、TC-AC26-06。M2→TC-M2-01。ARCHITECTURE 對外介面→TC-CFG-01、TC-CLI-01～03。

案例數量：見第 11 節末。

## 11. 待釐清事項

### 11.1 已關閉（v0.2）
| # | 問題 | 結論 | 依據 | 影響的 TC |
|---|------|------|------|-----------|
| D-01 | 同 email、author date 相同但名字不同時顯示哪個 | 取 git log 輸出中較先出現者（較新的 commit） | ARCHITECTURE §6-4、§12.1.9 | TC-AC03-04 |
| D-02 | 效能的日期範圍與 cache | 範圍 `1970-01-01～2099-12-31`（perf fixture 全在 2021～2025）；只以 warm cache 判定，cold cache 只記錄 | PRD v1.2 第 6 節（範圍於 v1.2 明定）；PRD v1.1 第 6 節（老闆 G2 決定）；ARCHITECTURE §12.1.8 | 第 6.3 節 |
| D-03 | 改名同時修改內容 | 只計實際改動的行（與 `git log --numstat` 一致）；改名判定門檻為 git `-M` 預設 50%。低於門檻的變體只記錄 | PRD A-04；ARCHITECTURE §4.2、§12.1.9。**PRD v1.1 沒有為這一項新增 AC**，因此 QA 只把 A-04 的語意當判定依據，門檻本身不判 | TC-AC08-04 |
| D-04 | 結果是否不受使用者 Git 設定影響 | 不受影響 | PRD M1（同一 repo 結果必須一致）；ARCHITECTURE §4.2、§12.1.9 | TC-AC08-03、TC-CFG-01 |
| D-05 | 表格版面 | 表頭 `Name  Email  Commits  Added  Deleted`、欄間 2 空白、文字左對齊／數字右對齊、行尾無空白、無框線、**無千分位**；查無資料訊息 `No commits found between … and ….` | ARCHITECTURE §7.1、§12.1.9（介面檢查） | TC-AC17-01～03、TC-AC20-01 |
| D-06 | 日期值前後有空白 | 不去除，exit 2 | PRD A-05；ARCHITECTURE §12.1.9 | TC-AC13-03 |
| D-07 | email 次排序的字元順序 | 轉小寫後依 Unicode 碼位逐字比較 | ARCHITECTURE §6-5、§12.1.9 | TC-AC19-03 |
| D-08 | emoji、組合字元寬度 | PRD 未規定，只記錄（ARCHITECTURE 依 go-runewidth 預設） | ARCHITECTURE §12.1.9 | TC-AC21-02 |
| D-09 | `.mailmap` 是否忽略 | 必須忽略 | **PRD v1.1 AC-40**（P-03） | TC-AC02-04、TC-AC40-01～03 |
| D-10 | CSV 換行 | 每一行（含最後一行）以 LF 結尾、不含 CR、三平台逐位元組相同 | **PRD v1.1 AC-43**（P-05） | TC-AC25-01、TC-AC43-01～04 |
| D-11 | 不含特殊字元時是否加引號 | 只在需要時加（欄位以空白開頭時也會加）；PRD 允許加引號，故只做介面檢查 | ARCHITECTURE §12.1.9 | TC-AC23-03 |
| D-12 | `--output` 寫檔失敗 | exit 1 | ARCHITECTURE §7.3、§8 | TC-AC26-04 |
| D-13 | `--output` 搭配表格 | 適用；內容與表格 stdout 逐位元組相同 | ARCHITECTURE §3、§7.3 | TC-AC26-05 |
| D-14 | 路徑錯誤與參數錯誤同時發生 | 使用方式錯誤（2）優先 | ARCHITECTURE §3 驗證順序 | TC-AC27-03、TC-AC39-04 |
| D-15 | bare repo、`.git` 資料夾當路徑 | 都視為「不是 Git repo」，exit 1 | bare：**PRD v1.1 AC-42**（P-04）；`.git`：ARCHITECTURE §12.1.9 | TC-AC28-06、TC-AC42-01～02 |
| D-16 | 多給位置參數 | exit 2 | ARCHITECTURE §3 | TC-AC31-02 |
| D-17 | `--format` 大小寫 | 區分大小寫；`CSV`、`Table` → exit 2 並列出 `table`、`csv` | ARCHITECTURE §3（v1.0 新增）、§12.1.9 | TC-AC32-02 |
| D-18 | `--help` 與其他參數並用 | 未知旗標（`--help --bogus`）→ exit 2；`--help` 搭配其他合法旗標或錯誤的旗標「值」（例如 `--format json`、錯誤日期、不存在路徑）→ exit 0。`--help` 與 `--version` 同時出現誰優先拆出 D-18b | ARCHITECTURE §3（v1.0 新增）、§12.1.9 | TC-AC33-02、04、05 |
| D-18b | `--help` 與 `--version` 並用；旗標缺值與 `--help` 並用 | `--help` 優先（exit 0、只印用法）；`--since --help` → `--help` 被當成起日值，日期錯誤 exit 2；結尾缺值 → exit 2 | ARCHITECTURE v1.1 §3、§12.1.9（PRD 未規定並用，屬介面細節；後兩者由 PRD AC-13、AC-16 推得） | TC-AC33-04、05 |
| D-19 | Git 更新 `.git/index` 快取算不算寫入 | 算，視為缺陷；以 `SnapshotGitDir`／`DiffGitDir` 檢查（含資料夾 mtime） | PRD 4.3「任何寫入」；ARCHITECTURE v1.1 §12.1.3.1 | TC-AC35-04 |
| D-22 | NFR 等級 | 安全類（NFR-04、05）＝致命；效能（NFR-01、02）、平台（NFR-03）＝高 | PRD v1.1／v1.2 第 6 節 | 第 9.3 節 |
| D-22b | NFR-06～08 嚴重度與優先度 | 嚴重度 NFR-06＝中、NFR-07／08＝高；優先度三項皆 Should | **PRD v1.2／v1.3 第 6 節**、Q-17（老闆 2026-11-22 01:06、01:10） | 第 9.3 節 |

**D-17、D-18 的說明**：TP-002 開工時（ARCHITECTURE v0.1）這兩項沒有答案，QA 原本的建議是：D-17「區分大小寫、exit 2」；D-18「`--help` 優先於所有參數內容檢查，但未知旗標仍 exit 2」。ARCHITECTURE v1.0 在 §3 加入的規則和 QA 建議一致，所以兩項直接關閉。和 v0.1 本計畫的衝突只有一處：v0.1 暫定「`--help` 與任何參數並用都 exit 0」，v0.2 改為 `--help --bogus` 判 exit 2（理由：PRD 第 5 節約定「未知選項」屬使用方式錯誤，AC-33 只定義 `--help` 單獨使用的情況）。仍未寫清楚的部分是 D-18b（`--help` 和 `--version` 同時出現），以及「旗標缺值」與 `--help` 並用時的情況（例如 `--since --help`：`--help` 可能被當成 `--since` 的值，結果變成日期錯誤 exit 2）。這兩點都只記錄（TC-AC33-05），請 Architect 在 §3 補一句。

### 11.2 仍開放
| # | 相關 AC/NFR | 問題 | 為什麼不能自己決定 | QA 暫定做法 |
|---|-------------|------|--------------------|-------------|
| D-20 | AC-37 | 「下載 Release 檔案後 1 個步驟」怎麼算、SmartScreen／Gatekeeper 警告處理算不算步驟 | ARCHITECTURE §12.1.9 沒有回覆；§14 主要路徑是 1 行安裝指令，手動下載為備案；R-04 未驗證 | 以 1 行安裝指令為主要判定；手動路徑每個動作算 1 步，記錄 |
| D-21 | NFR-07 | Windows 非 UTF-8 字碼頁下的表格中文顯示 | ARCHITECTURE §12.1.9 預期與字碼頁無關都正確，但**未驗證**（PoC A-001 附帶檢查） | 兩種字碼頁都要求顯示正確；A-001 結果出來前標「待 A-001」 |
| 覆蓋說明 | AC-08、PRD A-04 | perf fixture 只做純改名，「改名＋修改」不在 perf 資料中 | ARCHITECTURE v1.1 §12.1.8 為了讓產生器的預期值與 numstat 必然一致而限制；這不影響 AC 覆蓋：AC-08 本身只要求純改名，「改名＋修改」（D-03，PRD A-04）仍由 FX-04 #7 的 TC-AC08-04 以手算預期值覆蓋 | 不是覆蓋缺口；若之後有人要求 FX-04 也改成只有純改名，QA 會視為移除 D-03 的覆蓋而拒絕，需經老闆核准 |

### 11.3 PRD v1.1（ARCHITECTURE §18 P-01～P-08 定案）對本計畫的影響
| P | PRD v1.1 | 對本計畫的影響 |
|---|----------|----------------|
| P-01 | AC-39 找不到 Git／版本過舊 → exit 1 | 新增 TC-AC39-01～05 |
| P-02 | M2 不含安裝 Git 的時間 | TC-M2-01 計時起點改為「已裝 Git 的乾淨機器」 |
| P-03 | AC-40 明確忽略 `.mailmap` | 關閉 D-09；新增 TC-AC40-01～03 |
| P-04 | AC-41 detached HEAD 照常統計；AC-42 bare repo 視為不是 Git repo | 關閉 D-15；新增 TC-AC41-01～02、TC-AC42-01～02（原 TC-AC09-03、TC-AC28-05 併入） |
| P-05 | AC-43 CSV 換行固定 LF（含最後一行） | 關閉 D-10；新增 TC-AC43-01～04（原 TC-AC22-04 併入）；TC-AC25-01 改為逐位元組 `…deleted\n` |
| P-06 | AC-44 公式字元原樣輸出；7.3 R-01 已接受風險 | 新增 TC-AC44-01～02 |
| P-07 | AC-35 限定為不使用 `--output` 或 `--output` 指向 repo 外 | TC-AC35-05 改為「repo 外必測、repo 內只記錄」；G-4 排除 repo 內 `--output` |
| P-08 | AC-45 Release 含 linux/arm64、不含 windows/arm64 | 新增 TC-AC45-01～02（REL，G4）；`ac_list` 排除 |

D-03 不在 P-01～P-08 範圍內；本計畫依 PRD A-04＋ARCHITECTURE §12.1.9 關閉（見 11.1）。

**案例統計（v0.4，與 v0.2 相同；PRD v1.2、v1.3 沒有新增或修改 AC；TP-004 只修改既有 TC）**：AC 類 TC 158 個、NFR／M2 類 10 個、對外介面類 4 個（TC-CFG-01、TC-CLI-01～03）、情境類 4 個，合計 **176 個**（v0.1 為 149 個）。

## 12. 後續工作
1. 各 T-xxx 的 PR 開好、TechLead Review 通過後，QA 產出 Q-xxx 任務單，請 Claude Code 在該 task/T-xxx 分支依 ARCHITECTURE §12.1.2 的位置與命名追加本計畫的 TC（子測試名＝TC ID）與 QA 手寫的黃金檔。
2. T-013～T-015 的 `fixtures_test.go`、黃金檔與預期值，QA 驗證時逐一比對第 4.3 節；不一致即為缺陷。
3. `EXPECTED_SHA` 首次產生後由 QA 確認並鎖定。
4. PoC A-001 結果出來後：更新 TC-AC11-03（Windows 時區，ADR-005）、D-21、第 6.2 節形狀（若 A-001 證明不合理）、TC-AC39-03（Git 2.34）。

## 13. archtest `ac_list` 內容（決定）

**`ac_list` 是什麼**：ARCHITECTURE §11 規定「每條 Must AC 至少一個測試」；`internal/archtest` 讀 `test/e2e/ac_list.txt` 裡的 AC 編號，掃描測試函式名稱中的 `ACnn`，清單中只要有任何一個 AC 找不到對應測試，`go test` 就失敗（T-016）。所以清單中**只能放能在 `go test` 裡自動驗證的 AC**；放了無法自動化的 AC，CI 就永遠是紅的。

**決定**：`ac_list` 包含 **AC-01～AC-36、AC-39～AC-44（共 42 條）**，排除 **AC-37、AC-38、AC-45（共 3 條）**。

| AC | 決定 | 理由（一行） |
|----|------|--------------|
| AC-37 | 排除 | 需要乾淨的三平台機器、從真實 GitHub Release 安裝並重開終端機，`go test` 無法重現；改由 TC-AC37-01～05 在 G4 驗證 |
| AC-38 | 排除 | 驗的是 GitHub Releases 頁面與 tag 這類外部狀態，release 前不存在；改由 TC-AC38-01～03 在 G4 驗證 |
| AC-45 | 排除 | 驗的是已發佈 Release 的下載檔清單，同樣是 G4 才存在的外部狀態；改由 TC-AC45-01～02 在 G4 驗證 |
| AC-34 | **列入** | `--version` 輸出可在 `go test` 驗（TC-AC34-01，注入 `9.8.7`）；只有「與 Release tag 一致」的部分在 G4 由 TC-AC34-02 驗，不靠 archtest |
| AC-21、AC-26 | **列入** | 雖是 Should，但能完全自動化，列入可避免漏寫；archtest 只檢查「有測試」，不改變 Should 的退件規則 |
| AC-11 | 列入 | Linux／macOS 可在 `go test` 中以 `TZ` 測；Windows 做法待 A-001，不影響「有測試」 |
| AC-39～AC-44 | 列入 | 都能用 fixture＋`RunOpts`（改 `PATH`、bare repo、detached HEAD、bytes 比對）自動驗證 |

規則：
- 清單格式（每行 `AC-01` 或 `AC01`）由 T-016 決定，**內容以本節為準**；清單有任何增刪都要先改本節。
- 之後 PRD 新增的 AC：能在 `go test` 驗證的一律列入，在新增 AC 的同一份 TEST_PLAN 更新中決定。
- 排除不代表不測：AC-37、AC-38、AC-45 仍是 Must，G4 前必須通過（第 9.3 節第 2 點）。
- 與 ARCHITECTURE 一致：v1.1 的 §11、§12.1.2 已改為引用本節（列入 AC-34、AC-21、AC-26，排除 AC-37、AC-38、AC-45），v0.2 註記的差異已消除。

## 14. ARCHITECTURE §12.1 補充事項（v0.2 提出；**ARCHITECTURE v1.1 已全部採納**，保留作為紀錄）

狀態（v0.4）：S-01～S-09 均已由 ARCHITECTURE v1.1（`5033d3c`）採納：S-01→§12.1.4、S-02→§12.1.8、S-03→§5.1／§11、S-04→§12.1.5、S-05→§11／§12.1.2、S-06→§12.1.5 假 git、S-07→§12.1.3.1、S-08→§12.1.1、S-09→§3。以下為 v0.2 時的原文。

ARCHITECTURE v1.0 的 §12.1 已涵蓋 QA 在 v0.2 草稿中提出的大部分需求：committer 與 author 不同、空 commit、解衝突 merge、octopus、detached HEAD、`Git()` 逃生口、`EXPECTED_SHA`、`RunOpts.Env` 注入設定、原始 bytes 擷取、QA 手寫黃金檔，以及無網路 job（交給 SRE）。剩下的補充如下：

| # | 補充需求 | 為什麼需要 | 用到的 TC |
|---|----------|------------|-----------|
| S-01 | §12.1.4 的 Windows 例外清單會繼承 `TEMP`、`TMP`；請允許（或在 helper 中預設）把 `TMPDIR`／`TEMP`／`TMP`、`APPDATA`、`LOCALAPPDATA` 指向每個測試自己的暫存目錄 | NFR-08 要比對「暫存目錄與家目錄沒有新增本工具檔案」，繼承系統 TEMP 就無法判定 | TC-NFR08-01 |
| S-02 | §12.1.8 效能：① 形狀改為「依 TEST_PLAN §6.2」，A-001 的數據交給 QA 改 TEST_PLAN，而不是由產生器自行採用（§12.1.2 規定 FX 內容由 QA 擁有）；② `Generate` 要另外輸出「預期總數表」，算法不得呼叫產品程式；③ 寫明 Go perf 測試（暖機 1 次、5 次中位數）只是快速檢查，NFR 判定依 TEST_PLAN §6.3（暖機 3 次、10 次最大值 ≤ 門檻、參考機器 `ubuntu-24.04` 標準 runner）；④ perf fixture 要能保留在磁碟上給 `hyperfine` 用 | 效能判定方法與參考機器由 QA 在 TEST_PLAN 定義（PRD NFR-01）；目前兩邊的方法不同 | TC-NFR01-01、TC-NFR02-01 |
| S-03 | 由 archtest 強制 `test/e2e` 只能 import `internal/testutil/...` 與標準庫，不得 import 產品 package | §12.1.1 寫「不呼叫內部函式」，但沒有自動檢查；這是 QA 黑箱獨立性的基礎 | 全部 E2E |
| S-04 | CI 矩陣固定保留一組 Git 2.34 的 E2E job，不只在 PoC A-001 跑一次 | PRD AC-39 把 2.34 定為最低版本，之後每次改動都要確認沒有用到更新版 git 才有的旗標 | TC-AC39-03 |
| S-05 | §12.1.2「AC 清單」那一列改為引用 TEST_PLAN §13（AC-34 列入；排除 AC-37、AC-38、AC-45） | 避免兩份文件對 `ac_list` 的描述不一致 | 第 13 節 |
| S-06 | AC-39 的假 git：§12.1.3 規定 fixture「不建立可執行檔」，請在 §12.1.5 明訂假 git 的做法（例如 `TestMain` 另外 `go build` 一個小程式到暫存目錄，Windows 產生 `git.exe`），讓它不受 fixture 規則限制 | TC-AC39-02、03 需要一個回報指定版本號的 git | TC-AC39-02、03 |
| S-07 | `Snapshot()` 另提供選用的「整個 `.git/` 內容＋修改時間雜湊」 | D-19 已定為缺陷，但 `Snapshot` 的三項比對抓不到 index 快取被更新 | TC-AC35-04 |
| S-08 | §12.1.1 名詞對照表更新：TEST_PLAN v0.2 已改用 ARCHITECTURE 的名稱（E2E＝`test/e2e`，REL＝安裝／發佈驗證） | 舊對照（Integration＝E2E）會造成混淆 | — |
| S-09 | §3 補一句：`--help` 與 `--version` 同時出現時誰優先；旗標缺值（`--since --help`）的行為 | D-18b | TC-AC33-05 |

## 變更紀錄
| 版本 | 日期 | 作者 | 變更內容 |
|------|------|------|----------|
| v0.1 | 2026-11-22 | QA | 初版（交接單 TP-001），依 PRD v1.0 設計 |
| v0.2 | 2026-11-22 | QA | 交接單 TP-002：依據改為 PRD v1.1（`f5dbf86`）與 ARCHITECTURE v1.0（`915474f`）；fixture 與 E2E 改依 ARCHITECTURE §12.1（取消 shell 腳本與 `tests/fixtures/`，改用 `fixtures_test.go`、`EXPECTED_SHA`、QA 手寫黃金檔、`TestE2E_ACnn`＋TC ID 子測試）；層級代號改為 E2E／P／REL／M；新增 AC-39～AC-45 的 TC 並更新對照表（45 條全覆蓋）；關閉 D-01～D-19、D-22（D-18b、D-20、D-21、D-22b 仍開放）；套用 PRD v1.1 的 NFR 判定條件；新增第 13 節 `ac_list`（排除 AC-37、AC-38、AC-45）、第 14 節 §12.1 補充需求 S-01～S-09；新增 TC-CFG、TC-CLI 等案例 |
| v0.3 | 2026-11-22 | QA | 交接單 TP-003：依據改為 PRD v1.2（`8b5d669`，AC 未變更）；第 9.3 節加入 NFR 優先度／嚴重度表與 Should 類 NFR 未通過的處理（記缺陷、不阻擋 G3、列給老闆決定）；NFR-06～08 優先度先標「待定」（PRD Q-17）；關閉 D-22b；D-02 改引 PRD v1.2 的量測範圍 |
| v0.3.1 | 2026-11-22 | QA | 交接單 TP-003 補充：依幕僚長通知，Q-17 決定 NFR-06、07、08 全部為 Should（嚴重度不變），改用 Should 類 NFR 規則（記錄缺陷、不阻擋 G3）；註記 PRD v1.2 文字尚未更新 |
| v0.4 | 2026-11-22 | QA | 交接單 TP-004：依據改為 PRD v1.3（`0c55219`）、ARCHITECTURE v1.1（`5033d3c`）。TC-AC33-05 改為確定檢查（`--help` 優先於 `--version`，關閉 D-18b）；TC-AC33-04 加 `--bogus --help`；TC-NFR08-01 改用預設隔離的 TEMP 等目錄；§6.2 改用新 `Generate` 簽名、perf 改名一律純改名（改名＋修改仍由 TC-AC08-04 覆蓋）；TC-AC35-04 改用 `SnapshotGitDir`／`DiffGitDir`（含資料夾 mtime）；TC-AC39-03 預期改為 exit 1；Q-17 差異已由 PRD v1.3 消除；§14 標示 S-01～S-09 已採納 |
