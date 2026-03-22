// Package rediskey 定义 tk-business 服务使用的全部 Redis key 模板。
package rediskey

import (
	"fmt"
	"time"
)

// ─────────────────────────────────────────────
// 首页聚合缓存
// ─────────────────────────────────────────────

const (
	// HomeOverviewTTL 首页概览数据缓存有效期（5 分钟）
	// 后台修改 Banner/广播/外链等配置后需主动调用 InvalidateHomeOverview() 清除
	HomeOverviewTTL = 5 * time.Minute
)

// KeyHomeOverview 首页聚合数据缓存（banners/broadcasts/tabs/links等）。
// 格式: tk:home:overview:v1
// TTL:  HomeOverviewTTL
func KeyHomeOverview() string {
	return "tk:home:overview:v1"
}

// ─────────────────────────────────────────────
// 开奖看板缓存
// ─────────────────────────────────────────────

const (
	// DashboardTTL 开奖看板数据缓存有效期（30 秒）
	// 短 TTL 保证开奖数据更新后快速生效；开奖写入后可主动清除以立即生效
	DashboardTTL = 30 * time.Second

	// DashboardLockTTL 缓存击穿互斥锁最大持有时间（10 秒）
	// 超过此时间未重建则自动释放锁，允许其他请求重新触发重建
	DashboardLockTTL = 10 * time.Second
)

// KeyDashboard 开奖看板缓存（按彩种 ID 隔离）。
// 格式: tk:business:dashboard:{specialLotteryID}
// TTL:  DashboardTTL
func KeyDashboard(specialLotteryID uint) string {
	return fmt.Sprintf("tk:business:dashboard:%d", specialLotteryID)
}

// KeyDashboardLock 开奖看板缓存重建互斥锁（防缓存击穿）。
// 格式: lock:business:dashboard:{specialLotteryID}
// TTL:  DashboardLockTTL
func KeyDashboardLock(specialLotteryID uint) string {
	return fmt.Sprintf("lock:business:dashboard:%d", specialLotteryID)
}

// ─────────────────────────────────────────────
// 开奖历史缓存
// ─────────────────────────────────────────────

const (
	// DrawHistoryTTL 开奖历史列表缓存有效期（1 小时）
	DrawHistoryTTL = 1 * time.Hour
)

// KeyDrawHistory 开奖历史列表缓存。
// 格式: tk:business:draw:history:{specialLotteryID}:{year}
// TTL:  DrawHistoryTTL
func KeyDrawHistory(specialLotteryID uint, year int) string {
	return fmt.Sprintf("tk:business:draw:history:%d:%d", specialLotteryID, year)
}

// ─────────────────────────────────────────────
// 图纸卡片列表缓存
// ─────────────────────────────────────────────

const (
	// LotteryCardsTTL 图纸卡片列表缓存有效期（10 分钟）
	LotteryCardsTTL = 10 * time.Minute
)

// KeyLotteryCards 图纸卡片列表缓存（按分类 key 隔离，空字符串代表"全部"）。
// 格式: tk:business:lottery:cards:{category}
// TTL:  LotteryCardsTTL
func KeyLotteryCards(category string) string {
	if category == "" {
		category = "all"
	}
	return fmt.Sprintf("tk:business:lottery:cards:%s", category)
}

// ─────────────────────────────────────────────
// 开奖现场页缓存（直播聚合）
// ─────────────────────────────────────────────

const (
	// LiveSceneTTL 开奖现场聚合数据缓存有效期（默认 15 秒，可通过配置覆盖）
	LiveSceneTTL = 15 * time.Second
)

// KeyLiveScene 开奖现场页聚合数据缓存。
// 格式: tk:live_scene:page:v1:{specialLotteryID}
// TTL:  LiveSceneTTL（通过 SceneTTLSeconds 配置项覆盖）
func KeyLiveScene(specialLotteryID uint) string {
	return fmt.Sprintf("tk:live_scene:page:v1:%d", specialLotteryID)
}

// ─────────────────────────────────────────────
// 分布式锁
// ─────────────────────────────────────────────

const (
	// DrawWriteLockTTL 开奖结果写入锁最大持有时间（5 秒）
	DrawWriteLockTTL = 5 * time.Second
)

// KeyDrawWriteLock 开奖结果写入分布式锁（防多实例重复写入）。
// 格式: lock:business:draw:{specialLotteryID}:{issue}
// TTL:  DrawWriteLockTTL
func KeyDrawWriteLock(specialLotteryID uint, issue string) string {
	return fmt.Sprintf("lock:business:draw:%d:%s", specialLotteryID, issue)
}
