package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// defaultDSN 默认连接当前开发库；生产环境请通过 TK_DB_DSN 覆盖。
const defaultDSN = "root:12345678@tcp(127.0.0.1:3306)/nb_sys_001?charset=utf8mb4&parseTime=True&loc=Local"

// main 执行 w_* -> tk_* 数据迁移。
func main() {
	// 1) 优先读取环境变量 DSN，便于在不同环境复用同一脚本。
	dsn := strings.TrimSpace(os.Getenv("TK_DB_DSN"))
	if dsn == "" {
		dsn = defaultDSN
	}

	// 2) 建立数据库连接并探活，提前暴露连接错误。
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql failed: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping mysql failed: %v", err)
	}

	log.Println("start migrate w_* -> tk_* ...")

	// 3) 先做结构对齐，避免后续导入时因表/字段不存在失败。
	if err := ensureSchema(db); err != nil {
		log.Fatalf("ensure schema failed: %v", err)
	}

	// 4) 再做数据迁移：外链、分类、图纸三块按需求导入。
	if err := migrateExternalLinks(db); err != nil {
		log.Fatalf("migrate external links failed: %v", err)
	}
	if err := migrateCategories(db); err != nil {
		log.Fatalf("migrate categories failed: %v", err)
	}
	if err := migrateLotteryInfo(db); err != nil {
		log.Fatalf("migrate lottery_info failed: %v", err)
	}

	// 5) 输出迁移后的统计信息，便于快速核对结果。
	printCount(db, "tk_external_link")
	printCount(db, "tk_lottery_category")
	printCount(db, "tk_lottery_info")
	log.Println("migrate done")
}

// ensureSchema 保证目标 tk_* 表结构满足迁移和运行需要。
func ensureSchema(db *sql.DB) error {
	// A) 外链表：若不存在则创建；若已存在但缺字段则补字段。
	if !tableExists(db, "tk_external_link") {
		if err := execSQL(db, "create tk_external_link", `
CREATE TABLE IF NOT EXISTS tk_external_link (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  name VARCHAR(80) NOT NULL COMMENT '外链名称',
  url VARCHAR(255) NOT NULL COMMENT '外链地址',
  position VARCHAR(32) NOT NULL COMMENT '展示位置',
  icon_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标地址（用于金刚导航）',
  group_key VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分组键（如：aocai/hkcai/default）',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用；0停用',
  sort BIGINT NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_tk_external_link_position (position),
  KEY idx_tk_external_link_group_key (group_key),
  KEY idx_tk_external_link_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='外链配置表';
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_external_link", "icon_url") {
		if err := execSQL(db, "alter tk_external_link add icon_url", `
ALTER TABLE tk_external_link
  ADD COLUMN icon_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标地址（用于金刚导航）' AFTER position;
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_external_link", "group_key") {
		if err := execSQL(db, "alter tk_external_link add group_key", `
ALTER TABLE tk_external_link
  ADD COLUMN group_key VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分组键（如：aocai/hkcai/default）' AFTER icon_url;
`); err != nil {
			return err
		}
	}

	// B) 分类表：当前库缺 tk_lottery_category 时直接建表。
	if !tableExists(db, "tk_lottery_category") {
		if err := execSQL(db, "create tk_lottery_category", `
CREATE TABLE IF NOT EXISTS tk_lottery_category (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  category_key VARCHAR(32) NOT NULL COMMENT '分类键（唯一）',
  name VARCHAR(32) NOT NULL COMMENT '分类名称',
  search_keywords VARCHAR(255) NOT NULL DEFAULT '' COMMENT '搜索关键字',
  show_on_home TINYINT NOT NULL DEFAULT 1 COMMENT '是否首页展示',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用；0停用',
  sort BIGINT NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_tk_lottery_category_key (category_key),
  KEY idx_tk_lottery_category_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='图库分类配置表';
`); err != nil {
			return err
		}
	}

	// C) 图纸表：补齐分类与开奖号码新字段，供“分类单选 + 6+1 开奖录入”使用。
	if !columnExists(db, "tk_lottery_info", "category_id") {
		if err := execSQL(db, "alter tk_lottery_info add category_id", `
ALTER TABLE tk_lottery_info
  ADD COLUMN category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图库分类ID（关联tk_lottery_category.id）' AFTER special_lottery_id;
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_lottery_info", "category_tag") {
		if err := execSQL(db, "alter tk_lottery_info add category_tag", `
ALTER TABLE tk_lottery_info
  ADD COLUMN category_tag VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类标识兼容字段（通常等于category_key）' AFTER category_id;
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_lottery_info", "normal_draw_result") {
		if err := execSQL(db, "alter tk_lottery_info add normal_draw_result", `
ALTER TABLE tk_lottery_info
  ADD COLUMN normal_draw_result VARCHAR(64) NOT NULL DEFAULT '' COMMENT '普通号码（6个，逗号分隔）' AFTER draw_code;
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_lottery_info", "special_draw_result") {
		if err := execSQL(db, "alter tk_lottery_info add special_draw_result", `
ALTER TABLE tk_lottery_info
  ADD COLUMN special_draw_result VARCHAR(16) NOT NULL DEFAULT '' COMMENT '特别号码（1个）' AFTER normal_draw_result;
`); err != nil {
			return err
		}
	}
	if !columnExists(db, "tk_lottery_info", "playback_url") {
		if err := execSQL(db, "alter tk_lottery_info add playback_url", `
ALTER TABLE tk_lottery_info
  ADD COLUMN playback_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '直播回放地址（直播结束后录入）' AFTER draw_at;
`); err != nil {
			return err
		}
	}
	if !indexExists(db, "tk_lottery_info", "idx_tk_lottery_info_category_id") {
		if err := execSQL(db, "add idx_tk_lottery_info_category_id", `
ALTER TABLE tk_lottery_info
  ADD KEY idx_tk_lottery_info_category_id (category_id);
`); err != nil {
			return err
		}
	}
	if !indexExists(db, "tk_lottery_info", "idx_tk_lottery_info_category_tag") {
		if err := execSQL(db, "add idx_tk_lottery_info_category_tag", `
ALTER TABLE tk_lottery_info
  ADD KEY idx_tk_lottery_info_category_tag (category_tag);
`); err != nil {
			return err
		}
	}
	return nil
}

// migrateExternalLinks 将 w_external_link 合并到 tk_external_link。
func migrateExternalLinks(db *sql.DB) error {
	// 旧表不存在时直接跳过，保证脚本可在新环境重复执行。
	if !tableExists(db, "w_external_link") {
		log.Println("skip external_link: w_external_link not exists")
		return nil
	}

	// 1) 先更新同名同链接同位置的记录，保持已有 tk_* ID 稳定。
	if err := execSQL(db, "update tk_external_link from w_external_link", `
UPDATE tk_external_link t
JOIN w_external_link w
  ON BINARY t.name = BINARY w.name
 AND BINARY t.url = BINARY w.url
 AND BINARY t.position = BINARY w.position
SET
  t.icon_url = IFNULL(w.icon_url, ''),
  t.group_key = IFNULL(w.group_key, ''),
  t.status = w.status,
  t.sort = w.sort,
  t.updated_at = IFNULL(w.updated_at, t.updated_at);
`); err != nil {
		return err
	}

	// 2) 再插入 tk_* 中不存在的记录，避免漏数。
	if err := execSQL(db, "insert tk_external_link missing rows", `
INSERT INTO tk_external_link (
  name, url, position, icon_url, group_key, status, sort, created_at, updated_at
)
SELECT
  w.name, w.url, w.position, IFNULL(w.icon_url, ''), IFNULL(w.group_key, ''), w.status, w.sort,
  IFNULL(w.created_at, NOW(3)), IFNULL(w.updated_at, NOW(3))
FROM w_external_link w
LEFT JOIN tk_external_link t
  ON BINARY t.name = BINARY w.name
 AND BINARY t.url = BINARY w.url
 AND BINARY t.position = BINARY w.position
WHERE t.id IS NULL;
`); err != nil {
		return err
	}
	return nil
}

// migrateCategories 将 w_lottery_category 合并到 tk_lottery_category。
func migrateCategories(db *sql.DB) error {
	// 旧表不存在时跳过。
	if !tableExists(db, "w_lottery_category") {
		log.Println("skip lottery_category: w_lottery_category not exists")
		return nil
	}

	// 1) 按 category_key 对齐更新，避免重复分类。
	if err := execSQL(db, "update tk_lottery_category from w_lottery_category", `
UPDATE tk_lottery_category t
JOIN w_lottery_category w
  ON BINARY t.category_key = BINARY w.category_key
SET
  t.name = w.name,
  t.search_keywords = IFNULL(w.search_keywords, ''),
  t.show_on_home = w.show_on_home,
  t.status = w.status,
  t.sort = w.sort,
  t.updated_at = IFNULL(w.updated_at, t.updated_at);
`); err != nil {
		return err
	}

	// 2) 插入缺失分类。
	if err := execSQL(db, "insert tk_lottery_category missing rows", `
INSERT INTO tk_lottery_category (
  category_key, name, search_keywords, show_on_home, status, sort, created_at, updated_at
)
SELECT
  w.category_key, w.name, IFNULL(w.search_keywords, ''), w.show_on_home, w.status, w.sort,
  IFNULL(w.created_at, NOW(3)), IFNULL(w.updated_at, NOW(3))
FROM w_lottery_category w
LEFT JOIN tk_lottery_category t
  ON BINARY t.category_key = BINARY w.category_key
WHERE t.id IS NULL;
`); err != nil {
		return err
	}
	return nil
}

// migrateLotteryInfo 将 w_lottery_info 合并到 tk_lottery_info。
func migrateLotteryInfo(db *sql.DB) error {
	// 旧表不存在时跳过。
	if !tableExists(db, "w_lottery_info") {
		log.Println("skip lottery_info: w_lottery_info not exists")
		return nil
	}

	// 1) 按 (special_lottery_id + issue + title) 更新已有 tk 记录。
	if err := execSQL(db, "update tk_lottery_info from w_lottery_info", `
UPDATE tk_lottery_info t
JOIN w_lottery_info w
  ON t.special_lottery_id = w.special_lottery_id
 AND BINARY t.issue = BINARY w.issue
 AND BINARY t.title = BINARY w.title
SET
  t.category_id = COALESCE((
    SELECT c.id FROM tk_lottery_category c
    WHERE c.category_key COLLATE utf8mb4_general_ci = IFNULL(w.category_tag, '') COLLATE utf8mb4_general_ci
       OR c.name COLLATE utf8mb4_general_ci = IFNULL(w.category_tag, '') COLLATE utf8mb4_general_ci
    ORDER BY c.id ASC LIMIT 1
  ), t.category_id),
  t.category_tag = IFNULL(w.category_tag, ''),
  t.year = w.year,
  t.cover_image_url = w.cover_image_url,
  t.detail_image_url = w.detail_image_url,
  t.draw_code = w.draw_code,
  t.normal_draw_result = TRIM(BOTH ',' FROM SUBSTRING_INDEX(REPLACE(IFNULL(w.draw_result, ''), ' ', ''), ',', 6)),
  t.special_draw_result = TRIM(BOTH ',' FROM SUBSTRING_INDEX(REPLACE(IFNULL(w.draw_result, ''), ' ', ''), ',', -1)),
  t.draw_result = w.draw_result,
  t.draw_at = w.draw_at,
  t.playback_url = IFNULL(t.playback_url, ''),
  t.is_current = w.is_current,
  t.status = w.status,
  t.sort = w.sort,
  t.likes_count = w.likes_count,
  t.comment_count = w.comment_count,
  t.favorite_count = w.favorite_count,
  t.read_count = w.read_count,
  t.poll_enabled = w.poll_enabled,
  t.poll_default_expand = w.poll_default_expand,
  t.recommend_info_ids = IFNULL(w.recommend_info_ids, ''),
  t.updated_at = IFNULL(w.updated_at, t.updated_at);
`); err != nil {
		return err
	}

	// 2) 插入缺失图纸记录。
	if err := execSQL(db, "insert tk_lottery_info missing rows", `
INSERT INTO tk_lottery_info (
  special_lottery_id, category_id, category_tag, issue, year, title,
  cover_image_url, detail_image_url, draw_code, normal_draw_result, special_draw_result, draw_result, draw_at, playback_url,
  is_current, status, sort, likes_count, comment_count, favorite_count, read_count,
  poll_enabled, poll_default_expand, recommend_info_ids, created_at, updated_at
)
SELECT
  w.special_lottery_id,
  COALESCE((
    SELECT c.id FROM tk_lottery_category c
    WHERE c.category_key COLLATE utf8mb4_general_ci = IFNULL(w.category_tag, '') COLLATE utf8mb4_general_ci
       OR c.name COLLATE utf8mb4_general_ci = IFNULL(w.category_tag, '') COLLATE utf8mb4_general_ci
    ORDER BY c.id ASC LIMIT 1
  ), 0),
  IFNULL(w.category_tag, ''), w.issue, w.year, w.title,
  w.cover_image_url, w.detail_image_url, w.draw_code,
  TRIM(BOTH ',' FROM SUBSTRING_INDEX(REPLACE(IFNULL(w.draw_result, ''), ' ', ''), ',', 6)),
  TRIM(BOTH ',' FROM SUBSTRING_INDEX(REPLACE(IFNULL(w.draw_result, ''), ' ', ''), ',', -1)),
  w.draw_result, w.draw_at, '',
  w.is_current, w.status, w.sort, w.likes_count, w.comment_count, w.favorite_count, w.read_count,
  w.poll_enabled, w.poll_default_expand, IFNULL(w.recommend_info_ids, ''),
  IFNULL(w.created_at, NOW(3)), IFNULL(w.updated_at, NOW(3))
FROM w_lottery_info w
LEFT JOIN tk_lottery_info t
  ON t.special_lottery_id = w.special_lottery_id
 AND BINARY t.issue = BINARY w.issue
 AND BINARY t.title = BINARY w.title
WHERE t.id IS NULL;
`); err != nil {
		return err
	}
	return nil
}

// tableExists 判断目标表是否存在。
func tableExists(db *sql.DB, table string) bool {
	var c int
	if err := db.QueryRow(`
SELECT COUNT(1)
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`, table).Scan(&c); err != nil {
		return false
	}
	return c > 0
}

// columnExists 判断目标字段是否存在。
func columnExists(db *sql.DB, table, column string) bool {
	var c int
	if err := db.QueryRow(`
SELECT COUNT(1)
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, column).Scan(&c); err != nil {
		return false
	}
	return c > 0
}

// indexExists 判断目标索引是否存在。
func indexExists(db *sql.DB, table, index string) bool {
	var c int
	if err := db.QueryRow(`
SELECT COUNT(1)
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, table, index).Scan(&c); err != nil {
		return false
	}
	return c > 0
}

// printCount 打印表记录数，便于迁移后核对结果。
func printCount(db *sql.DB, table string) {
	if !tableExists(db, table) {
		log.Printf("%s not exists", table)
		return
	}
	var c int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(1) FROM %s", table)).Scan(&c); err != nil {
		log.Printf("%s count err: %v", table, err)
		return
	}
	log.Printf("%s count=%d", table, c)
}

// execSQL 统一执行 SQL 并打印受影响行数，方便排错。
func execSQL(db *sql.DB, label, query string) error {
	res, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	aff, _ := res.RowsAffected()
	log.Printf("%s ok, affected=%d", label, aff)
	return nil
}
