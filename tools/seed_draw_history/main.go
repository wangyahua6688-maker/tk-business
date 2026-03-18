package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// defaultDSN 默认连接本地开发数据库。
const defaultDSN = "root:12345678@tcp(127.0.0.1:3306)/nb_sys_001?charset=utf8mb4&parseTime=True&loc=Local"

// specialSeed 彩种基础种子配置。
type specialSeed struct {
	// 处理当前语句逻辑。
	Code string
	// 处理当前语句逻辑。
	Name string
	// 处理当前语句逻辑。
	Sort int
}

// main 启动程序入口。
func main() {
	// 1) 读取数据库连接配置。
	dsn := strings.TrimSpace(os.Getenv("TK_DB_DSN"))
	// 判断条件并进入对应分支逻辑。
	if dsn == "" {
		// 更新当前变量或字段值。
		dsn = defaultDSN
	}

	// 2) 建立数据库连接并探活。
	db, err := sql.Open("mysql", dsn)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 调用log.Fatalf完成当前处理。
		log.Fatalf("open mysql failed: %v", err)
	}
	// 注册延迟执行逻辑。
	defer db.Close()
	// 判断条件并进入对应分支逻辑。
	if err := db.Ping(); err != nil {
		// 调用log.Fatalf完成当前处理。
		log.Fatalf("ping mysql failed: %v", err)
	}

	// 3) 初始化彩种（存在则更新，不存在则插入）。
	seeds := []specialSeed{
		// 处理当前语句逻辑。
		{Code: "macau", Name: "澳彩", Sort: 1},
		// 处理当前语句逻辑。
		{Code: "hk", Name: "港彩", Sort: 2},
	}
	// 判断条件并进入对应分支逻辑。
	if err := ensureSpecialLotteries(db, seeds); err != nil {
		// 调用log.Fatalf完成当前处理。
		log.Fatalf("ensure special lotteries failed: %v", err)
	}

	// 4) 读取彩种 ID 映射。
	specialIDByCode, err := loadSpecialIDMap(db)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 调用log.Fatalf完成当前处理。
		log.Fatalf("load special id map failed: %v", err)
	}

	// 5) 按彩种写入 12 期开奖记录（支持重复执行幂等覆盖）。
	baseDate := time.Date(2026, 3, 1, 21, 32, 0, 0, time.Local)
	// 定义并初始化当前变量。
	startIssue := 57
	// 定义并初始化当前变量。
	periodCount := 12
	// 循环处理当前数据集合。
	for idx, seed := range seeds {
		// 定义并初始化当前变量。
		sid := specialIDByCode[seed.Code]
		// 判断条件并进入对应分支逻辑。
		if sid == 0 {
			// 更新当前变量或字段值。
			log.Fatalf("special id missing for code=%s", seed.Code)
		}
		// 判断条件并进入对应分支逻辑。
		if err := seedDrawRecordsForSpecial(db, sid, seed, baseDate, startIssue, periodCount, idx*7); err != nil {
			// 调用log.Fatalf完成当前处理。
			log.Fatalf("seed draw records for %s failed: %v", seed.Code, err)
		}
	}

	// 6) 输出统计结果，便于联调确认。
	printCount(db, "tk_special_lottery")
	// 调用printCount完成当前处理。
	printCount(db, "tk_draw_record")
	// 调用fmt.Println完成当前处理。
	fmt.Println("seed draw history done")
}

// ensureSpecialLotteries 保证 tk_special_lottery 至少有澳彩/港彩两条基础数据。
func ensureSpecialLotteries(db *sql.DB, seeds []specialSeed) error {
	// 定义并初始化当前变量。
	now := time.Now()
	// 循环处理当前数据集合。
	for _, item := range seeds {
		// 判断条件并进入对应分支逻辑。
		if _, err := db.Exec(`
INSERT INTO tk_special_lottery (
  name, code, current_issue, next_draw_at, live_enabled, live_status, live_stream_url, status, sort, created_at, updated_at
) VALUES (?, ?, '', ?, 0, 'pending', '', 1, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  sort = VALUES(sort),
  status = 1,
  updated_at = NOW(3)
`, item.Name, item.Code, now, item.Sort); err != nil {
			// 返回当前处理结果。
			return err
		}
	}
	// 返回当前处理结果。
	return nil
}

// loadSpecialIDMap 读取 code -> id 的彩种映射。
func loadSpecialIDMap(db *sql.DB) (map[string]uint64, error) {
	// 定义并初始化当前变量。
	rows, err := db.Query(`SELECT id, code FROM tk_special_lottery WHERE status = 1`)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 注册延迟执行逻辑。
	defer rows.Close()

	// 定义并初始化当前变量。
	result := make(map[string]uint64, 8)
	// 循环处理当前数据集合。
	for rows.Next() {
		// 声明当前变量。
		var id uint64
		// 声明当前变量。
		var code string
		// 判断条件并进入对应分支逻辑。
		if err := rows.Scan(&id, &code); err != nil {
			// 返回当前处理结果。
			return nil, err
		}
		// 更新当前变量或字段值。
		result[strings.TrimSpace(code)] = id
	}
	// 返回当前处理结果。
	return result, nil
}

// seedDrawRecordsForSpecial 给单个彩种批量造开奖历史数据。
func seedDrawRecordsForSpecial(db *sql.DB, specialID uint64, seed specialSeed, baseDate time.Time, startIssue, count, seedOffset int) error {
	// 1) 先清空当前期标记，后续最后一期再标记为当前期。
	if _, err := db.Exec(`UPDATE tk_draw_record SET is_current = 0 WHERE special_lottery_id = ?`, specialID); err != nil {
		// 返回当前处理结果。
		return err
	}

	// 定义并初始化当前变量。
	latestIssue := ""
	// 定义并初始化当前变量。
	latestDrawAt := baseDate

	// 2) 逐期 upsert 开奖记录。
	for i := 0; i < count; i++ {
		// 定义并初始化当前变量。
		issueNo := startIssue + i
		// 定义并初始化当前变量。
		issue := fmt.Sprintf("2026-%03d", issueNo)
		// 定义并初始化当前变量。
		drawAt := baseDate.Add(time.Duration(i*24+seedOffset) * time.Hour)
		// 定义并初始化当前变量。
		numbers := buildSevenUniqueNumbers(issueNo + seedOffset + int(specialID))
		// 定义并初始化当前变量。
		normal := joinIntSlice(numbers[:6], ",")
		// 定义并初始化当前变量。
		special := fmt.Sprintf("%02d", numbers[6])
		// 定义并初始化当前变量。
		drawResult := normal + "," + special
		// 定义并初始化当前变量。
		drawLabels, zodiacLabels, wuxingLabels := buildDrawLabels(numbers)
		playbackURL := fmt.Sprintf("https://cdn.example.com/replay/%s/%s.m3u8", seed.Code, issue)
		// 定义并初始化当前变量。
		isCurrent := 0
		// 判断条件并进入对应分支逻辑。
		if i == count-1 {
			// 更新当前变量或字段值。
			isCurrent = 1
		}
		// 定义并初始化当前变量。
		sort := issueNo

		// 判断条件并进入对应分支逻辑。
		if _, err := db.Exec(`
INSERT INTO tk_draw_record (
  special_lottery_id, issue, year, draw_at,
  normal_draw_result, special_draw_result, draw_result, draw_labels, zodiac_labels, wuxing_labels, playback_url,
  status, is_current, sort, created_at, updated_at
) VALUES (?, ?, 2026, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  draw_at = VALUES(draw_at),
  normal_draw_result = VALUES(normal_draw_result),
  special_draw_result = VALUES(special_draw_result),
  draw_result = VALUES(draw_result),
  draw_labels = VALUES(draw_labels),
  zodiac_labels = VALUES(zodiac_labels),
  wuxing_labels = VALUES(wuxing_labels),
  playback_url = VALUES(playback_url),
  status = VALUES(status),
  is_current = VALUES(is_current),
  sort = VALUES(sort),
  updated_at = NOW(3)
`, specialID, issue, drawAt, normal, special, drawResult, drawLabels, zodiacLabels, wuxingLabels, playbackURL, isCurrent, sort); err != nil {
			// 返回当前处理结果。
			return err
		}

		// 更新当前变量或字段值。
		latestIssue = issue
		// 更新当前变量或字段值。
		latestDrawAt = drawAt
	}

	// 3) 同步彩种当前期与下期开奖时间。
	nextDrawAt := latestDrawAt.Add(24 * time.Hour)
	// 判断条件并进入对应分支逻辑。
	if _, err := db.Exec(`
UPDATE tk_special_lottery
SET current_issue = ?, next_draw_at = ?, status = 1, updated_at = NOW(3)
WHERE id = ?
`, latestIssue, nextDrawAt, specialID); err != nil {
		// 返回当前处理结果。
		return err
	}
	// 返回当前处理结果。
	return nil
}

// buildSevenUniqueNumbers 生成 1..49 范围内的 7 个不重复号码。
func buildSevenUniqueNumbers(seed int) []int {
	// 定义并初始化当前变量。
	out := make([]int, 0, 7)
	// 定义并初始化当前变量。
	used := make(map[int]struct{}, 7)
	// 定义并初始化当前变量。
	step := seed%11 + 5
	// 定义并初始化当前变量。
	cur := seed%49 + 1
	// 循环处理当前数据集合。
	for len(out) < 7 {
		// 判断条件并进入对应分支逻辑。
		if _, ok := used[cur]; !ok {
			// 更新当前变量或字段值。
			used[cur] = struct{}{}
			// 更新当前变量或字段值。
			out = append(out, cur)
		}
		// 更新当前变量或字段值。
		cur = ((cur + step - 1) % 49) + 1
		// 更新当前变量或字段值。
		step = (step % 13) + 3
	}
	// 返回当前处理结果。
	return out
}

// buildDrawLabels 按号码生成“属相/五行”标签串。
func buildDrawLabels(numbers []int) (string, string, string) {
	// 定义并初始化当前变量。
	zodiacs := []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	// 定义并初始化当前变量。
	wuxing := []string{"金", "木", "水", "火", "土"}
	// 定义并初始化当前变量。
	labels := make([]string, 0, len(numbers))
	// 定义并初始化当前变量。
	zodiacLabels := make([]string, 0, len(numbers))
	// 定义并初始化当前变量。
	wuxingLabels := make([]string, 0, len(numbers))
	// 循环处理当前数据集合。
	for _, n := range numbers {
		// 定义并初始化当前变量。
		zodiac := zodiacs[(n-1)%len(zodiacs)]
		// 定义并初始化当前变量。
		element := wuxing[(n-1)%len(wuxing)]
		// 更新当前变量或字段值。
		labels = append(labels, fmt.Sprintf("%s/%s", zodiac, element))
		// 更新当前变量或字段值。
		zodiacLabels = append(zodiacLabels, zodiac)
		// 更新当前变量或字段值。
		wuxingLabels = append(wuxingLabels, element)
	}
	// 返回当前处理结果。
	return strings.Join(labels, ","), strings.Join(zodiacLabels, ","), strings.Join(wuxingLabels, ",")
}

// joinIntSlice 将 int 数组按分隔符拼接，并对号码补齐两位。
func joinIntSlice(nums []int, sep string) string {
	// 定义并初始化当前变量。
	parts := make([]string, 0, len(nums))
	// 循环处理当前数据集合。
	for _, n := range nums {
		// 更新当前变量或字段值。
		parts = append(parts, fmt.Sprintf("%02d", n))
	}
	// 返回当前处理结果。
	return strings.Join(parts, sep)
}

// printCount 输出目标表行数。
func printCount(db *sql.DB, table string) {
	// 声明当前变量。
	var c int64
	// 判断条件并进入对应分支逻辑。
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&c); err != nil {
		// 调用log.Printf完成当前处理。
		log.Printf("%s count error: %v", table, err)
		// 返回当前处理结果。
		return
	}
	// 调用log.Printf完成当前处理。
	log.Printf("%s rows: %d", table, c)
}
