package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

// AppConfig holds paths and settings loaded from config.local.json.
type AppConfig struct {
	AccountBookExe string `json:"account_book_exe"`
	DBPath         string `json:"db_path"`
	BackupRepo     string `json:"backup_repo"`
	DefaultMonth   string `json:"default_month"`
}

// ConfigStatus reports health of each configured path.
type ConfigStatus struct {
	AccountBookExe       string   `json:"account_book_exe"`
	AccountBookExeExists bool     `json:"account_book_exe_exists"`
	DBPath               string   `json:"db_path"`
	DBPathExists         bool     `json:"db_path_exists"`
	BackupRepo           string   `json:"backup_repo"`
	BackupRepoExists     bool     `json:"backup_repo_exists"`
	DefaultMonth         string   `json:"default_month"`
	OverallStatus        string   `json:"overall_status"`
	Errors               []string `json:"errors"`
}

// App struct
type App struct {
	ctx    context.Context
	config *AppConfig
}

type DashboardStats struct {
	Month           string  `json:"month"`
	Income          float64 `json:"income"`
	Expense         float64 `json:"expense"`
	Balance         float64 `json:"balance"`
	Refund          float64 `json:"refund"`
	Reimbursement   float64 `json:"reimbursement"`
	NeedReviewCount int     `json:"need_review_count"`
	TotalBills      int     `json:"total_bills"`
	ActiveBills     int     `json:"active_bills"`
	DeletedBills    int     `json:"deleted_bills"`
	LastBackupTime  string  `json:"last_backup_time"`
}

type BillItem struct {
	ID              int64    `json:"id"`
	BillTime        string   `json:"bill_time"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	Category        string   `json:"category"`
	SubCategory     string   `json:"sub_category"`
	Merchant        string   `json:"merchant"`
	Tags            []string `json:"tags"`
	Note            string   `json:"note"`
	DisplayTitle    string   `json:"display_title"`
	DisplaySubtitle string   `json:"display_subtitle"`
}

type BillDetail struct {
	ID              int64    `json:"id"`
	BillTime        string   `json:"bill_time"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	Category        string   `json:"category"`
	SubCategory     string   `json:"sub_category"`
	Merchant        string   `json:"merchant"`
	Tags            []string `json:"tags"`
	Note            string   `json:"note"`
	DisplayTitle    string   `json:"display_title"`
	DisplaySubtitle string   `json:"display_subtitle"`
	RawCategory     string   `json:"raw_category"`
	RawSubCategory  string   `json:"raw_sub_category"`
	RawTags         []string `json:"raw_tags"`
	CreatedAt       string   `json:"created_at"`
}

type CreateBillInput struct {
	BillTime    string   `json:"bill_time"`
	Type        string   `json:"type"`
	Amount      float64  `json:"amount"`
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category"`
	Merchant    string   `json:"merchant"`
	Tags        []string `json:"tags"`
	Note        string   `json:"note"`
}

type UpdateBillBasicInput struct {
	ID       int64   `json:"id"`
	BillTime string  `json:"bill_time"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Merchant string  `json:"merchant"`
}

type CategoryNode struct {
	ID              int64          `json:"id"`
	ParentID        int64          `json:"parent_id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	TaxonomyVersion string         `json:"taxonomy_version"`
	SortOrder       int            `json:"sort_order"`
	IsActive        bool           `json:"is_active"`
	IconKey         string         `json:"icon_key"`
	BillCount       int            `json:"bill_count"`
	LastUsedAt      string         `json:"last_used_at"`
	Children        []CategoryNode `json:"children"`
}

type TagItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	GroupName  string `json:"group_name"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
	UseCount   int    `json:"use_count"`
	LastUsedAt string `json:"last_used_at"`
}

type CategoryInput struct {
	ParentID  int64  `json:"parent_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	SortOrder int    `json:"sort_order"`
	IconKey   string `json:"icon_key"`
}

type UpdateCategoryInput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IconKey   string `json:"icon_key"`
}

type TagInput struct {
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
	SortOrder int    `json:"sort_order"`
}

type UpdateTagInput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
	SortOrder int    `json:"sort_order"`
}

type MergeTagsInput struct {
	SourceID int64 `json:"source_id"`
	TargetID int64 `json:"target_id"`
}

type MerchantItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	SortOrder   int     `json:"sort_order"`
	IsActive    bool    `json:"is_active"`
	UseCount    int     `json:"use_count"`
	TotalAmount float64 `json:"total_amount"`
	LastUsedAt  string  `json:"last_used_at"`
}

type MerchantInput struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type UpdateMerchantInput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type monthlyStatsJSON struct {
	Month           string `json:"month"`
	Income          int64  `json:"income_cents"`
	Expense         int64  `json:"expense_cents"`
	Balance         int64  `json:"balance_cents"`
	Refund          int64  `json:"refund_cents"`
	Reimbursement   int64  `json:"reimbursement_cents"`
	NeedReviewCount int    `json:"need_review_count"`
}

const configFile = "config.local.json"

func defaultConfig() *AppConfig {
	return &AppConfig{
		AccountBookExe: `D:\Project\self\mira\mira-ledger\account-book.exe`,
		DBPath:         `D:\Project\self\mira\mira-ledger\data\account-book.db`,
		BackupRepo:     `D:\Project\self\mira\mira-ledger-data`,
		DefaultMonth:   "current",
	}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// LoadConfig reads config.local.json. If missing it writes defaults, validates,
// and stores the result in a.config. Returns any validation error so the UI can
// display it.
func (a *App) loadConfig() error {
	cfg := defaultConfig()

	data, err := os.ReadFile(configFile)
	if err != nil {
		// File missing — write defaults and continue with validation
		raw, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(configFile, raw, 0644)
	} else {
		if parseErr := json.Unmarshal(data, cfg); parseErr != nil {
			return fmt.Errorf("parse %s: %w", configFile, parseErr)
		}
	}

	cfg.AccountBookExe = strings.TrimSpace(cfg.AccountBookExe)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	cfg.BackupRepo = strings.TrimSpace(cfg.BackupRepo)
	cfg.DefaultMonth = strings.TrimSpace(cfg.DefaultMonth)
	if cfg.DefaultMonth == "" {
		cfg.DefaultMonth = "current"
	}

	// Validate
	if cfg.AccountBookExe == "" {
		return fmt.Errorf("account_book_exe is empty")
	}
	if _, err := os.Stat(cfg.AccountBookExe); err != nil {
		return fmt.Errorf("account_book_exe not found: %s", cfg.AccountBookExe)
	}
	if cfg.DBPath == "" {
		return fmt.Errorf("db_path is empty")
	}
	if _, err := os.Stat(cfg.DBPath); err != nil {
		return fmt.Errorf("db_path not found: %s", cfg.DBPath)
	}
	if cfg.BackupRepo == "" {
		return fmt.Errorf("backup_repo is empty")
	}
	if info, err := os.Stat(cfg.BackupRepo); err != nil || !info.IsDir() {
		return fmt.Errorf("backup_repo not found or not a directory: %s", cfg.BackupRepo)
	}

	a.config = cfg
	return nil
}

// GetConfig returns the current config (Wails-exported for frontend).
func (a *App) GetConfig() (*AppConfig, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	return a.config, nil
}

// GetConfigStatus reads config.local.json and reports health of each path.
func (a *App) GetConfigStatus() *ConfigStatus {
	cfg := defaultConfig()
	data, err := os.ReadFile(configFile)
	if err == nil {
		json.Unmarshal(data, cfg)
	}
	cfg.AccountBookExe = strings.TrimSpace(cfg.AccountBookExe)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	cfg.BackupRepo = strings.TrimSpace(cfg.BackupRepo)
	cfg.DefaultMonth = strings.TrimSpace(cfg.DefaultMonth)
	if cfg.DefaultMonth == "" {
		cfg.DefaultMonth = "current"
	}

	status := &ConfigStatus{
		AccountBookExe: cfg.AccountBookExe,
		DBPath:         cfg.DBPath,
		BackupRepo:     cfg.BackupRepo,
		DefaultMonth:   cfg.DefaultMonth,
		OverallStatus:  "ok",
		Errors:         []string{},
	}

	// Check account_book_exe
	if cfg.AccountBookExe == "" {
		status.Errors = append(status.Errors, "账本程序路径未配置")
	} else if _, err := os.Stat(cfg.AccountBookExe); err != nil {
		status.Errors = append(status.Errors, "账本程序不存在: "+cfg.AccountBookExe)
	} else {
		status.AccountBookExeExists = true
	}

	// Check db_path
	if cfg.DBPath == "" {
		status.Errors = append(status.Errors, "数据库路径未配置")
	} else if _, err := os.Stat(cfg.DBPath); err != nil {
		status.Errors = append(status.Errors, "数据库文件不存在: "+cfg.DBPath)
	} else {
		status.DBPathExists = true
	}

	// Check backup_repo
	if cfg.BackupRepo == "" {
		status.Errors = append(status.Errors, "备份仓库路径未配置")
	} else if info, err := os.Stat(cfg.BackupRepo); err != nil || !info.IsDir() {
		status.Errors = append(status.Errors, "备份仓库不存在或不是目录: "+cfg.BackupRepo)
	} else {
		status.BackupRepoExists = true
	}

	if len(status.Errors) > 0 {
		status.OverallStatus = "error"
	}

	return status
}

// UpdateConfig validates, persists, and applies a new config.
func (a *App) UpdateConfig(input AppConfig) error {
	input.AccountBookExe = strings.TrimSpace(input.AccountBookExe)
	input.DBPath = strings.TrimSpace(input.DBPath)
	input.BackupRepo = strings.TrimSpace(input.BackupRepo)
	input.DefaultMonth = strings.TrimSpace(input.DefaultMonth)
	if input.DefaultMonth == "" {
		input.DefaultMonth = "current"
	}

	if input.AccountBookExe == "" {
		return fmt.Errorf("account_book_exe is empty")
	}
	if _, err := os.Stat(input.AccountBookExe); err != nil {
		return fmt.Errorf("account_book_exe not found: %s", input.AccountBookExe)
	}
	if input.DBPath == "" {
		return fmt.Errorf("db_path is empty")
	}
	if _, err := os.Stat(input.DBPath); err != nil {
		return fmt.Errorf("db_path not found: %s", input.DBPath)
	}
	if input.BackupRepo == "" {
		return fmt.Errorf("backup_repo is empty")
	}
	if info, err := os.Stat(input.BackupRepo); err != nil || !info.IsDir() {
		return fmt.Errorf("backup_repo not found or not a directory: %s", input.BackupRepo)
	}

	raw, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configFile, raw, 0644); err != nil {
		return fmt.Errorf("write %s: %w", configFile, err)
	}

	a.config = &input
	return nil
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.loadConfig(); err != nil {
		// Config error will be surfaced to UI via GetConfig
		_ = err
	}
	if a.config != nil {
		ensureMigration(a.config.DBPath)
	}
}

func ensureMigration(dbPath string) {
	conn, err := openLedgerDB(dbPath)
	if err != nil {
		return
	}
	defer conn.Close()

	if !columnExists(conn, "bills", "is_deleted") {
		conn.Exec(`ALTER TABLE bills ADD COLUMN is_deleted INTEGER NOT NULL DEFAULT 0`)
	}
	if !columnExists(conn, "bills", "deleted_at") {
		conn.Exec(`ALTER TABLE bills ADD COLUMN deleted_at TEXT`)
	}
	if !columnExists(conn, "tags", "sort_order") {
		conn.Exec(`ALTER TABLE tags ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`)
		seedTagSortOrder(conn)
	}
	if !columnExists(conn, "merchants", "sort_order") {
		conn.Exec(`ALTER TABLE merchants ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`)
	}
	if !columnExists(conn, "categories", "icon_key") {
		conn.Exec(`ALTER TABLE categories ADD COLUMN icon_key TEXT NOT NULL DEFAULT ''`)
	}
}

func (a *App) GetVerify() string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("verify", "--db", a.config.DBPath)
}

func (a *App) GetMonthlyStats(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("stats", "--db", a.config.DBPath, "--month", month)
}

func (a *App) GetDashboardStats(month string) (*DashboardStats, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	statsOutput, err := a.runAccountBookOutput("stats", "--db", a.config.DBPath, "--month", month, "--format", "json")
	if err != nil {
		return nil, err
	}
	var monthly monthlyStatsJSON
	if err := json.Unmarshal([]byte(statsOutput), &monthly); err != nil {
		return nil, fmt.Errorf("parse stats json: %w", err)
	}

	verifyOutput, err := a.runAccountBookOutput("verify", "--db", a.config.DBPath)
	if err != nil {
		return nil, err
	}
	totalBills, err := parseIntField(verifyOutput, "total_bills")
	if err != nil {
		return nil, err
	}
	needReviewCount, err := parseIntField(verifyOutput, "need_review_count")
	if err != nil {
		return nil, err
	}

	lastBackupTime, _ := a.lastGitCommitTime()

	activeBills, deletedBills := 0, 0
	if conn, err2 := a.dbConn(); err2 == nil {
		defer conn.Close()
		conn.QueryRow(`SELECT count(*) FROM bills WHERE is_deleted = 0`).Scan(&activeBills)
		conn.QueryRow(`SELECT count(*) FROM bills WHERE is_deleted = 1`).Scan(&deletedBills)
	}

	return &DashboardStats{
		Month:           monthly.Month,
		Income:          centsToAmount(monthly.Income),
		Expense:         centsToAmount(monthly.Expense),
		Balance:         centsToAmount(monthly.Balance),
		Refund:          centsToAmount(monthly.Refund),
		Reimbursement:   centsToAmount(monthly.Reimbursement),
		NeedReviewCount: needReviewCount,
		TotalBills:      totalBills,
		ActiveBills:     activeBills,
		DeletedBills:    deletedBills,
		LastBackupTime:  lastBackupTime,
	}, nil
}

func (a *App) GetMonthlySummary(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("summary", "--db", a.config.DBPath, "--month", month, "--format", "json")
}

func (a *App) dbConn() (*sql.DB, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	return openLedgerDB(a.config.DBPath)
}

func (a *App) GetCategoryTree() ([]CategoryNode, error) {
	return a.ListCategories(false)
}

func (a *App) GetAllTags() ([]TagItem, error) {
	return a.ListTags(false)
}

func (a *App) ListCategories(includeInactive bool) ([]CategoryNode, error) {
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	where := ""
	if !includeInactive {
		where = "WHERE c.is_active = 1"
	}
	rows, err := conn.Query(`
SELECT
	c.id,
	c.parent_id,
	c.name,
	c.type,
	c.taxonomy_version,
	c.sort_order,
	c.is_active,
	ifnull(c.icon_key, '') AS icon_key,
	(
		SELECT count(*)
		FROM bills b
		WHERE b.is_deleted = 0
		  AND ((c.parent_id = 0 AND b.category_id = c.id) OR (c.parent_id != 0 AND b.subcategory_id = c.id))
	) AS bill_count,
	ifnull((
		SELECT max(b.bill_time)
		FROM bills b
		WHERE b.is_deleted = 0
		  AND ((c.parent_id = 0 AND b.category_id = c.id) OR (c.parent_id != 0 AND b.subcategory_id = c.id))
	), '') AS last_used_at
FROM categories c
` + where + `
ORDER BY c.type, c.parent_id, c.sort_order, c.id`)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	roots := []CategoryNode{}
	rootIndex := map[int64]int{}
	childrenByParent := map[int64][]CategoryNode{}
	for rows.Next() {
		var node CategoryNode
		var active int
		if err := rows.Scan(
			&node.ID,
			&node.ParentID,
			&node.Name,
			&node.Type,
			&node.TaxonomyVersion,
			&node.SortOrder,
			&active,
			&node.IconKey,
			&node.BillCount,
			&node.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		node.IsActive = active == 1
		node.Children = []CategoryNode{}
		if node.ParentID == 0 {
			rootIndex[node.ID] = len(roots)
			roots = append(roots, node)
			continue
		}
		childrenByParent[node.ParentID] = append(childrenByParent[node.ParentID], node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	for parentID, children := range childrenByParent {
		if idx, ok := rootIndex[parentID]; ok {
			roots[idx].Children = children
		}
	}
	return roots, nil
}

func (a *App) CreateCategory(input CategoryInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.TrimSpace(input.Type)
	if input.Name == "" {
		return fmt.Errorf("category name is required")
	}
	if input.ParentID == 0 && input.Type == "" {
		return fmt.Errorf("category type is required")
	}

	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	taxonomy := currentTaxonomyVersion(conn)
	parentID := input.ParentID
	categoryType := input.Type
	if parentID != 0 {
		if err := conn.QueryRow(`SELECT type, taxonomy_version FROM categories WHERE id = ?`, parentID).Scan(&categoryType, &taxonomy); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("parent category %d not found", parentID)
			}
			return fmt.Errorf("query parent category: %w", err)
		}
	}
	if categoryType != "expense" && categoryType != "income" {
		return fmt.Errorf("unsupported category type %q", categoryType)
	}
	if input.SortOrder <= 0 {
		input.SortOrder = nextCategorySortOrder(conn, parentID, categoryType, taxonomy)
	}
	iconKey := strings.TrimSpace(input.IconKey)
	now := nowString()
	_, err = conn.Exec(
		`INSERT INTO categories(parent_id, name, type, taxonomy_version, sort_order, is_active, icon_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`,
		parentID,
		input.Name,
		categoryType,
		taxonomy,
		input.SortOrder,
		iconKey,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (a *App) UpdateCategory(input UpdateCategoryInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.ID <= 0 {
		return fmt.Errorf("category id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("category name is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	iconKey := strings.TrimSpace(input.IconKey)
	result, err := conn.Exec(`UPDATE categories SET name = ?, sort_order = ?, icon_key = ?, updated_at = ? WHERE id = ?`, input.Name, input.SortOrder, iconKey, nowString(), input.ID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return ensureCategoryUpdated(result, input.ID)
}

func (a *App) SetCategoryActive(id int64, isActive bool) error {
	if id <= 0 {
		return fmt.Errorf("category id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	result, err := conn.Exec(`UPDATE categories SET is_active = ?, updated_at = ? WHERE id = ?`, boolToInt(isActive), nowString(), id)
	if err != nil {
		return fmt.Errorf("set category active: %w", err)
	}
	return ensureCategoryUpdated(result, id)
}

func (a *App) DeleteCategory(id int64) error {
	if id <= 0 {
		return fmt.Errorf("category id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	var childCount int
	if err := conn.QueryRow(`SELECT count(*) FROM categories WHERE parent_id = ?`, id).Scan(&childCount); err != nil {
		return fmt.Errorf("count category children: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("category has child categories; deactivate it instead")
	}
	var billCount int
	if err := conn.QueryRow(`SELECT count(*) FROM bills WHERE category_id = ? OR subcategory_id = ?`, id, id).Scan(&billCount); err != nil {
		return fmt.Errorf("count category usage: %w", err)
	}
	if billCount > 0 {
		return fmt.Errorf("category is used by bills; deactivate it instead")
	}
	result, err := conn.Exec(`DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return ensureCategoryUpdated(result, id)
}

func (a *App) ListTags(includeInactive bool) ([]TagItem, error) {
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	where := ""
	if !includeInactive {
		where = "WHERE t.is_active = 1"
	}
	rows, err := conn.Query(`
SELECT
	t.id,
	t.name,
	t.group_name,
	t.sort_order,
	t.is_active,
	count(b.id) AS use_count,
	ifnull(max(b.bill_time), '') AS last_used_at
FROM tags t
LEFT JOIN bill_tags bt ON bt.tag_id = t.id
LEFT JOIN bills b ON b.id = bt.bill_id AND b.is_deleted = 0
` + where + `
GROUP BY t.id
ORDER BY t.group_name, t.sort_order, t.name, t.id`)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	tags := []TagItem{}
	for rows.Next() {
		var tag TagItem
		var active int
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.GroupName, &tag.SortOrder, &active, &tag.UseCount, &tag.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tag.IsActive = active == 1
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}
	return tags, nil
}

func (a *App) CreateTag(input TagInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.GroupName = strings.TrimSpace(input.GroupName)
	if input.Name == "" {
		return fmt.Errorf("tag name is required")
	}
	if input.GroupName == "" {
		input.GroupName = "content"
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	if input.SortOrder <= 0 {
		input.SortOrder = nextTagSortOrder(conn, input.GroupName)
	}
	now := nowString()
	_, err = conn.Exec(`INSERT INTO tags(name, group_name, sort_order, is_active, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`, input.Name, input.GroupName, input.SortOrder, now, now)
	if err != nil {
		return fmt.Errorf("create tag: %w", err)
	}
	return nil
}

func (a *App) UpdateTag(input UpdateTagInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.GroupName = strings.TrimSpace(input.GroupName)
	if input.ID <= 0 {
		return fmt.Errorf("tag id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("tag name is required")
	}
	if input.GroupName == "" {
		return fmt.Errorf("tag group is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	result, err := conn.Exec(`UPDATE tags SET name = ?, group_name = ?, sort_order = ?, updated_at = ? WHERE id = ?`, input.Name, input.GroupName, input.SortOrder, nowString(), input.ID)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}
	return ensureTagUpdated(result, input.ID)
}

func (a *App) SetTagActive(id int64, isActive bool) error {
	if id <= 0 {
		return fmt.Errorf("tag id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	result, err := conn.Exec(`UPDATE tags SET is_active = ?, updated_at = ? WHERE id = ?`, boolToInt(isActive), nowString(), id)
	if err != nil {
		return fmt.Errorf("set tag active: %w", err)
	}
	return ensureTagUpdated(result, id)
}

func (a *App) DeleteTag(id int64) error {
	if id <= 0 {
		return fmt.Errorf("tag id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	var billCount int
	if err := conn.QueryRow(`SELECT count(*) FROM bill_tags WHERE tag_id = ?`, id).Scan(&billCount); err != nil {
		return fmt.Errorf("count tag usage: %w", err)
	}
	if billCount > 0 {
		return fmt.Errorf("tag is used by bills; deactivate or merge it instead")
	}
	result, err := conn.Exec(`DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return ensureTagUpdated(result, id)
}

func (a *App) MergeTags(input MergeTagsInput) error {
	if input.SourceID <= 0 || input.TargetID <= 0 {
		return fmt.Errorf("source and target tag ids are required")
	}
	if input.SourceID == input.TargetID {
		return fmt.Errorf("source and target tags must be different")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := ensureTagExistsTx(tx, input.SourceID); err != nil {
		return fmt.Errorf("source tag: %w", err)
	}
	if err := ensureTagExistsTx(tx, input.TargetID); err != nil {
		return fmt.Errorf("target tag: %w", err)
	}
	if _, err := tx.Exec(`
INSERT OR IGNORE INTO bill_tags(bill_id, tag_id)
SELECT bill_id, ? FROM bill_tags WHERE tag_id = ?`, input.TargetID, input.SourceID); err != nil {
		return fmt.Errorf("move bill tags: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM bill_tags WHERE tag_id = ?`, input.SourceID); err != nil {
		return fmt.Errorf("remove source bill tags: %w", err)
	}
	result, err := tx.Exec(`UPDATE tags SET is_active = 0, updated_at = ? WHERE id = ?`, nowString(), input.SourceID)
	if err != nil {
		return fmt.Errorf("deactivate source tag: %w", err)
	}
	if err := ensureTagUpdated(result, input.SourceID); err != nil {
		return err
	}
	return tx.Commit()
}

func (a *App) ListMerchants(includeInactive bool) ([]MerchantItem, error) {
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `
SELECT
	m.id,
	m.name,
	m.sort_order,
	m.is_active,
	COUNT(b.id) AS use_count,
	COALESCE(SUM(CASE WHEN b.type = 'expense' THEN b.amount ELSE 0 END), 0) AS total_amount,
	COALESCE(MAX(b.bill_time), '') AS last_used_at
FROM merchants m
LEFT JOIN bills b ON b.merchant_id = m.id AND b.is_deleted = 0
`
	if !includeInactive {
		query += `WHERE m.is_active = 1
`
	}
	query += `GROUP BY m.id
ORDER BY m.sort_order ASC, m.name ASC`

	rows, err := conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list merchants: %w", err)
	}
	defer rows.Close()

	items := []MerchantItem{}
	for rows.Next() {
		var item MerchantItem
		var active int
		if err := rows.Scan(&item.ID, &item.Name, &item.SortOrder, &active, &item.UseCount, &item.TotalAmount, &item.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan merchant: %w", err)
		}
		item.IsActive = active != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *App) CreateMerchant(input MerchantInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("merchant name is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	now := nowString()
	_, err = conn.Exec(
		`INSERT INTO merchants (name, sort_order, is_active, created_at, updated_at) VALUES (?, ?, 1, ?, ?)`,
		name, input.SortOrder, now, now,
	)
	if err != nil {
		return fmt.Errorf("create merchant: %w", err)
	}
	return nil
}

func (a *App) UpdateMerchant(input UpdateMerchantInput) error {
	if input.ID == 0 {
		return fmt.Errorf("merchant id is required")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("merchant name is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	now := nowString()
	result, err := conn.Exec(
		`UPDATE merchants SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
		name, input.SortOrder, now, input.ID,
	)
	if err != nil {
		return fmt.Errorf("update merchant: %w", err)
	}
	return ensureUpdated(result, input.ID)
}

func (a *App) SetMerchantActive(id int64, isActive bool) error {
	if id == 0 {
		return fmt.Errorf("merchant id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	active := 0
	if isActive {
		active = 1
	}
	now := nowString()
	result, err := conn.Exec(`UPDATE merchants SET is_active = ?, updated_at = ? WHERE id = ?`, active, now, id)
	if err != nil {
		return fmt.Errorf("set merchant active: %w", err)
	}
	return ensureUpdated(result, id)
}

func (a *App) DeleteMerchant(id int64) error {
	if id == 0 {
		return fmt.Errorf("merchant id is required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	var useCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM bills WHERE merchant_id = ? AND is_deleted = 0`, id).Scan(&useCount); err != nil {
		return fmt.Errorf("count merchant usage: %w", err)
	}
	if useCount > 0 {
		return fmt.Errorf("该商家已被 %d 笔账单使用，请停用而不是删除", useCount)
	}

	result, err := conn.Exec(`DELETE FROM merchants WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete merchant: %w", err)
	}
	return ensureUpdated(result, id)
}

func (a *App) GetBills(month string) ([]BillItem, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT
	b.id,
	b.bill_time,
	b.type,
	b.amount,
	ifnull(parent.name, ''),
	ifnull(child.name, ''),
	ifnull(m.name, ''),
	ifnull(tags.names, ''),
	ifnull(b.note, '')
FROM bills b
LEFT JOIN categories parent ON parent.id = b.category_id
LEFT JOIN categories child ON child.id = b.subcategory_id
LEFT JOIN merchants m ON m.id = b.merchant_id
LEFT JOIN (
	SELECT bill_id, group_concat(name, char(31)) AS names
	FROM (
		SELECT bt.bill_id, t.name
		FROM bill_tags bt
		JOIN tags t ON t.id = bt.tag_id
		ORDER BY t.name
	)
	GROUP BY bill_id
) tags ON tags.bill_id = b.id
WHERE b.bill_date >= ? AND b.bill_date < ? AND b.is_deleted = 0
ORDER BY b.bill_time DESC, b.id DESC
LIMIT 200`, start, end)
	if err != nil {
		return nil, fmt.Errorf("query bills: %w", err)
	}
	defer rows.Close()

	var bills []BillItem
	for rows.Next() {
		var bill BillItem
		var amountCents int64
		var tagsText string
		if err := rows.Scan(
			&bill.ID,
			&bill.BillTime,
			&bill.Type,
			&amountCents,
			&bill.Category,
			&bill.SubCategory,
			&bill.Merchant,
			&tagsText,
			&bill.Note,
		); err != nil {
			return nil, fmt.Errorf("scan bill: %w", err)
		}
		bill.Amount = centsToAmount(amountCents)
		bill.Tags = splitTags(tagsText)
		applyBillDisplay(&bill)
		bills = append(bills, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bills: %w", err)
	}
	return bills, nil
}

func (a *App) SearchBills(keyword string) ([]BillItem, error) {
	keyword = strings.TrimSpace(keyword)
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	like := "%" + keyword + "%"
	rows, err := conn.Query(`
SELECT
	b.id,
	b.bill_time,
	b.type,
	b.amount,
	ifnull(parent.name, ''),
	ifnull(child.name, ''),
	ifnull(m.name, ''),
	ifnull(tags.names, ''),
	ifnull(b.note, '')
FROM bills b
LEFT JOIN categories parent ON parent.id = b.category_id
LEFT JOIN categories child ON child.id = b.subcategory_id
LEFT JOIN merchants m ON m.id = b.merchant_id
LEFT JOIN (
	SELECT bill_id, group_concat(name, char(31)) AS names
	FROM (
		SELECT bt.bill_id, t.name
		FROM bill_tags bt
		JOIN tags t ON t.id = bt.tag_id
		ORDER BY t.name
	)
	GROUP BY bill_id
) tags ON tags.bill_id = b.id
WHERE b.is_deleted = 0
  AND (
	ifnull(parent.name, '') LIKE ? OR ifnull(parent.name, '') LIKE ? ESCAPE '\'
	OR ifnull(child.name, '') LIKE ? OR ifnull(child.name, '') LIKE ? ESCAPE '\'
	OR ifnull(m.name, '') LIKE ? OR ifnull(m.name, '') LIKE ? ESCAPE '\'
	OR ifnull(b.note, '') LIKE ? OR ifnull(b.note, '') LIKE ? ESCAPE '\'
	OR ifnull(tags.names, '') LIKE ? OR ifnull(tags.names, '') LIKE ? ESCAPE '\'
	OR b.bill_time LIKE ?
	OR CAST(b.amount AS TEXT) LIKE ?
	OR b.type LIKE ?
  )
ORDER BY b.bill_time DESC, b.id DESC
LIMIT 500`, like, like, like, like, like, like, like, like, like, like, like, like, like)
	if err != nil {
		return nil, fmt.Errorf("search bills: %w", err)
	}
	defer rows.Close()

	var bills []BillItem
	for rows.Next() {
		var bill BillItem
		var amountCents int64
		var tagsText string
		if err := rows.Scan(
			&bill.ID,
			&bill.BillTime,
			&bill.Type,
			&amountCents,
			&bill.Category,
			&bill.SubCategory,
			&bill.Merchant,
			&tagsText,
			&bill.Note,
		); err != nil {
			return nil, fmt.Errorf("scan bill: %w", err)
		}
		bill.Amount = centsToAmount(amountCents)
		bill.Tags = splitTags(tagsText)
		applyBillDisplay(&bill)
		bills = append(bills, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bills: %w", err)
	}
	return bills, nil
}

type DisplayFields struct {
	DisplayTitle    string `json:"display_title"`
	DisplaySubtitle string `json:"display_subtitle"`
}

func buildDisplayFields(note, merchant, category, subCategory, billType string) DisplayFields {
	note = strings.TrimSpace(note)
	merchant = strings.TrimSpace(merchant)
	category = strings.TrimSpace(category)
	subCategory = strings.TrimSpace(subCategory)

	var title string
	switch {
	case note != "":
		title = note
	case billType == "income" && (category != "" || subCategory != ""):
		title = strings.Join(nonEmptyStrings(category, subCategory), " / ")
	case merchant != "":
		title = merchant
	case category != "" || subCategory != "":
		title = strings.Join(nonEmptyStrings(category, subCategory), " / ")
	default:
		title = "未命名账单"
	}

	var subtitle string
	if title != merchant && merchant != "" {
		subtitle = merchant
	} else if billType != "" {
		subtitle = billTypeLabel(billType)
	}

	return DisplayFields{DisplayTitle: title, DisplaySubtitle: subtitle}
}

func applyBillDisplay(bill *BillItem) {
	fields := buildDisplayFields(bill.Note, bill.Merchant, bill.Category, bill.SubCategory, bill.Type)
	bill.DisplayTitle = fields.DisplayTitle
	bill.DisplaySubtitle = fields.DisplaySubtitle
}

func nonEmptyStrings(values ...string) []string {
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func billTypeLabel(billType string) string {
	labels := map[string]string{
		"expense":       "支出",
		"income":        "收入",
		"refund":        "退款",
		"reimbursement": "报销",
		"transfer":      "转账",
	}
	if label, ok := labels[billType]; ok {
		return label
	}
	return billType
}

func (a *App) GetBillDetail(id int64) (*BillDetail, error) {
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var detail BillDetail
	var amountCents int64
	var tagsText string
	var rawTagsText string
	err = conn.QueryRow(`
SELECT
	b.id,
	b.bill_time,
	b.type,
	b.amount,
	ifnull(parent.name, ''),
	ifnull(child.name, ''),
	ifnull(m.name, ''),
	ifnull(tags.names, ''),
	ifnull(b.note, ''),
	ifnull(b.raw_category, ''),
	ifnull(b.raw_subcategory, ''),
	ifnull(b.raw_tags, ''),
	b.created_at
FROM bills b
LEFT JOIN categories parent ON parent.id = b.category_id
LEFT JOIN categories child ON child.id = b.subcategory_id
LEFT JOIN merchants m ON m.id = b.merchant_id
LEFT JOIN (
	SELECT bill_id, group_concat(name, char(31)) AS names
	FROM (
		SELECT bt.bill_id, t.name
		FROM bill_tags bt
		JOIN tags t ON t.id = bt.tag_id
		ORDER BY t.name
	)
	GROUP BY bill_id
) tags ON tags.bill_id = b.id
WHERE b.id = ?`, id).Scan(
		&detail.ID,
		&detail.BillTime,
		&detail.Type,
		&amountCents,
		&detail.Category,
		&detail.SubCategory,
		&detail.Merchant,
		&tagsText,
		&detail.Note,
		&detail.RawCategory,
		&detail.RawSubCategory,
		&rawTagsText,
		&detail.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bill %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query bill detail: %w", err)
	}
	detail.Amount = centsToAmount(amountCents)
	detail.Tags = splitTags(tagsText)
	detail.RawTags = splitRawTags(rawTagsText)
	fields := buildDisplayFields(detail.Note, detail.Merchant, detail.Category, detail.SubCategory, detail.Type)
	detail.DisplayTitle = fields.DisplayTitle
	detail.DisplaySubtitle = fields.DisplaySubtitle
	return &detail, nil
}

func (a *App) CreateBill(input CreateBillInput) (*BillDetail, error) {
	input.Type = strings.TrimSpace(input.Type)
	input.Category = strings.TrimSpace(input.Category)
	input.SubCategory = strings.TrimSpace(input.SubCategory)
	input.Merchant = strings.TrimSpace(input.Merchant)
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if input.BillTime == "" {
		return nil, fmt.Errorf("bill_time is required")
	}
	if input.Category == "" || input.SubCategory == "" {
		return nil, fmt.Errorf("category and sub_category are required")
	}

	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	amountCents := amountToCents(input.Amount)
	billTime := strings.TrimSpace(input.BillTime)
	billDate, err := deriveBillDate(billTime)
	if err != nil {
		return nil, err
	}

	categoryID, subCategoryID, err := findCategoryPair(conn, input.Type, input.Category, input.SubCategory)
	if err != nil {
		return nil, fmt.Errorf("category lookup: %w", err)
	}

	var merchantID *int64
	if input.Merchant != "" {
		mid, err := findMerchantID(conn, input.Merchant)
		if err != nil {
			return nil, err
		}
		merchantID = mid
	}

	now := nowString()
	rawJSON := "{}"

	result, err := conn.Exec(`
INSERT INTO bills (
    bill_time, bill_date, type, amount, currency,
    category_id, subcategory_id, merchant_id, note,
    source, manual_override, need_review, review_status,
    raw_data_json, created_at, updated_at
) VALUES (?, ?, ?, ?, 'CNY', ?, ?, ?, ?, 'manual', 1, 0, 'reviewed', ?, ?, ?)`,
		billTime, billDate, input.Type, amountCents,
		categoryID, subCategoryID, merchantID,
		strings.TrimSpace(input.Note),
		rawJSON, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert bill: %w", err)
	}

	billID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	for _, tag := range input.Tags {
		tagID, err := findTagID(conn, tag)
		if err != nil {
			return nil, fmt.Errorf("tag %q: %w", tag, err)
		}
		if _, err := conn.Exec(
			`INSERT OR IGNORE INTO bill_tags(bill_id, tag_id) VALUES (?, ?)`,
			billID, tagID,
		); err != nil {
			return nil, fmt.Errorf("insert bill tag %q: %w", tag, err)
		}
	}

	return a.GetBillDetail(billID)
}

func (a *App) UpdateBillBasic(input UpdateBillBasicInput) (*BillDetail, error) {
	input.BillTime = strings.TrimSpace(input.BillTime)
	input.Type = strings.TrimSpace(input.Type)
	input.Merchant = strings.TrimSpace(input.Merchant)

	if input.ID <= 0 {
		return nil, fmt.Errorf("bill id is required")
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if input.BillTime == "" {
		return nil, fmt.Errorf("bill_time is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	amountCents := amountToCents(input.Amount)
	billDate, err := deriveBillDate(input.BillTime)
	if err != nil {
		return nil, err
	}

	var merchantID *int64
	if input.Merchant != "" {
		mid, err := findMerchantID(conn, input.Merchant)
		if err != nil {
			return nil, err
		}
		merchantID = mid
	}

	now := nowString()

	result, err := conn.Exec(`
UPDATE bills SET
    bill_time = ?, bill_date = ?, type = ?, amount = ?,
    merchant_id = ?, manual_override = 1, updated_at = ?
WHERE id = ?`,
		input.BillTime, billDate, input.Type, amountCents,
		merchantID, now, input.ID)
	if err != nil {
		return nil, fmt.Errorf("update bill %d: %w", input.ID, err)
	}
	if err := ensureUpdated(result, input.ID); err != nil {
		return nil, err
	}

	return a.GetBillDetail(input.ID)
}

func (a *App) UpdateBillCategory(id int64, category, subCategory string) error {
	category = strings.TrimSpace(category)
	subCategory = strings.TrimSpace(subCategory)
	if category == "" || subCategory == "" {
		return fmt.Errorf("category and subCategory are required")
	}
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	var billType string
	if err := conn.QueryRow(`SELECT type FROM bills WHERE id = ?`, id).Scan(&billType); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("bill %d not found", id)
		}
		return fmt.Errorf("query bill type: %w", err)
	}
	categoryID, subCategoryID, err := findCategoryPair(conn, billType, category, subCategory)
	if err != nil {
		return err
	}
	result, err := conn.Exec(
		`UPDATE bills SET category_id = ?, subcategory_id = ?, manual_override = 1, need_review = 0, review_status = 'reviewed', updated_at = ? WHERE id = ?`,
		categoryID,
		subCategoryID,
		nowString(),
		id,
	)
	if err != nil {
		return fmt.Errorf("update bill category: %w", err)
	}
	return ensureUpdated(result, id)
}

func (a *App) UpdateBillTags(id int64, tags []string) error {
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	cleanTags := normalizeTagList(tags)
	tagIDs := make([]int64, 0, len(cleanTags))
	for _, tag := range cleanTags {
		tagID, err := findTagID(conn, tag)
		if err != nil {
			return err
		}
		tagIDs = append(tagIDs, tagID)
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`UPDATE bills SET manual_override = 1, updated_at = ? WHERE id = ?`, nowString(), id)
	if err != nil {
		return fmt.Errorf("mark bill manual override: %w", err)
	}
	if err := ensureUpdated(result, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM bill_tags WHERE bill_id = ?`, id); err != nil {
		return fmt.Errorf("clear bill tags: %w", err)
	}
	for _, tagID := range tagIDs {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO bill_tags(bill_id, tag_id) VALUES (?, ?)`, id, tagID); err != nil {
			return fmt.Errorf("insert bill tag %d: %w", tagID, err)
		}
	}
	return tx.Commit()
}

func (a *App) UpdateBillNote(id int64, note string) error {
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := conn.Exec(
		`UPDATE bills SET note = ?, manual_override = 1, updated_at = ? WHERE id = ?`,
		strings.TrimSpace(note),
		nowString(),
		id,
	)
	if err != nil {
		return fmt.Errorf("update bill note: %w", err)
	}
	return ensureUpdated(result, id)
}

func (a *App) RunGitHubBackup(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("backup-github", "--db", a.config.DBPath, "--month", month, "--repo", a.config.BackupRepo)
}

func (a *App) SoftDeleteBill(id int64) error {
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	now := nowString()
	result, err := conn.Exec(
		`UPDATE bills SET is_deleted = 1, deleted_at = ?, manual_override = 1, updated_at = ? WHERE id = ? AND is_deleted = 0`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete bill %d: %w", id, err)
	}
	return ensureUpdated(result, id)
}

func (a *App) RestoreBill(id int64) error {
	conn, err := a.dbConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := conn.Exec(
		`UPDATE bills SET is_deleted = 0, deleted_at = NULL, manual_override = 1, updated_at = ? WHERE id = ? AND is_deleted = 1`,
		nowString(), id,
	)
	if err != nil {
		return fmt.Errorf("restore bill %d: %w", id, err)
	}
	return ensureUpdated(result, id)
}

func (a *App) GetDeletedBills(month string) ([]BillItem, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
	SELECT
		b.id,
		b.bill_time,
		b.type,
		b.amount,
		ifnull(parent.name, ''),
		ifnull(child.name, ''),
		ifnull(m.name, ''),
		ifnull(tags.names, ''),
		ifnull(b.note, '')
	FROM bills b
	LEFT JOIN categories parent ON parent.id = b.category_id
	LEFT JOIN categories child ON child.id = b.subcategory_id
	LEFT JOIN merchants m ON m.id = b.merchant_id
	LEFT JOIN (
		SELECT bill_id, group_concat(name, char(31)) AS names
		FROM (
			SELECT bt.bill_id, t.name
			FROM bill_tags bt
			JOIN tags t ON t.id = bt.tag_id
			ORDER BY t.name
		)
		GROUP BY bill_id
	) tags ON tags.bill_id = b.id
	WHERE b.bill_date >= ? AND b.bill_date < ? AND b.is_deleted = 1
	ORDER BY b.bill_time DESC, b.id DESC
	LIMIT 200`, start, end)
	if err != nil {
		return nil, fmt.Errorf("query deleted bills: %w", err)
	}
	defer rows.Close()

	var bills []BillItem
	for rows.Next() {
		var bill BillItem
		var amountCents int64
		var tagsText string
		if err := rows.Scan(
			&bill.ID,
			&bill.BillTime,
			&bill.Type,
			&amountCents,
			&bill.Category,
			&bill.SubCategory,
			&bill.Merchant,
			&tagsText,
			&bill.Note,
		); err != nil {
			return nil, fmt.Errorf("scan deleted bill: %w", err)
		}
		bill.Amount = centsToAmount(amountCents)
		bill.Tags = splitTags(tagsText)
		applyBillDisplay(&bill)
		bills = append(bills, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deleted bills: %w", err)
	}
	return bills, nil
}

func (a *App) RunLocalBackup(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("backup", "--db", a.config.DBPath, "--month", month,
		"--backup-dir", filepath.Join(a.config.BackupRepo, "backups"),
		"--export-dir", filepath.Join(a.config.BackupRepo, "exports"),
	)
}

func (a *App) RunReport(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("report", "--db", a.config.DBPath, "--month", month,
		"--out", filepath.Join(a.config.BackupRepo, "reports"),
	)
}

func (a *App) ExportData(month string) string {
	if a.config == nil {
		return "配置未加载"
	}
	return a.runAccountBook("summary", "--db", a.config.DBPath, "--month", month,
		"--format", "json", "--out", filepath.Join(a.config.BackupRepo, "exports"),
	)
}

func (a *App) OpenPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("路径不存在: %s", path)
	}
	arg := path
	if !info.IsDir() {
		arg = "/select," + path
	}
	return exec.Command("explorer", arg).Start()
}

func (a *App) runAccountBook(args ...string) string {
	out, err := a.runAccountBookOutput(args...)
	if err != nil && out == "" {
		return err.Error()
	}
	return out
}

func (a *App) runAccountBookOutput(args ...string) (string, error) {
	if a.config == nil {
		return "", fmt.Errorf("config not loaded")
	}
	cmd := exec.Command(a.config.AccountBookExe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Dir = filepath.Dir(a.config.AccountBookExe)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return string(out), fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		}
		return "", err
	}
	return string(out), nil
}

func (a *App) lastGitCommitTime() (string, error) {
	if a.config == nil {
		return "", fmt.Errorf("config not loaded")
	}
	gitDir := filepath.Join(a.config.BackupRepo, ".git")
	cmd := exec.Command("git", "--git-dir", gitDir, "log", "-1", "--format=%cI")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func parseIntField(output, name string) (int, error) {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `:\s*(\d+)`)
	match := re.FindStringSubmatch(output)
	if len(match) != 2 {
		return 0, fmt.Errorf("parse verify output: missing %s", name)
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("parse verify output %s: %w", name, err)
	}
	return value, nil
}

func centsToAmount(cents int64) float64 {
	return float64(cents) / 100
}

func amountToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func deriveBillDate(billTime string) (string, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
	for _, format := range formats {
		t, err := time.Parse(format, billTime)
		if err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("parse bill_time %q: unsupported format", billTime)
}

func monthRange(month string) (string, string, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", "", fmt.Errorf("invalid month %q, want YYYY-MM", month)
	}
	return t.Format("2006-01-02"), t.AddDate(0, 1, 0).Format("2006-01-02"), nil
}

func splitTags(tagsText string) []string {
	if strings.TrimSpace(tagsText) == "" {
		return []string{}
	}
	parts := strings.Split(tagsText, "\x1f")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func splitRawTags(tagsText string) []string {
	tagsText = strings.TrimSpace(tagsText)
	if tagsText == "" {
		return []string{}
	}
	parts := strings.FieldsFunc(tagsText, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '|' || r == '/'
	})
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func openLedgerDB(dbPath string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open ledger db: %w", err)
	}
	if _, err := conn.Exec(`PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("configure ledger db: %w", err)
	}
	return conn, nil
}

func columnExists(conn *sql.DB, table, column string) bool {
	rows, err := conn.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var dflt interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

func seedTagSortOrder(conn *sql.DB) {
	rows, err := conn.Query(`SELECT id, group_name FROM tags ORDER BY group_name, id`)
	if err != nil {
		return
	}
	defer rows.Close()
	orders := map[string]int{}
	type tagOrder struct {
		id    int64
		order int
	}
	var updates []tagOrder
	for rows.Next() {
		var id int64
		var group string
		if err := rows.Scan(&id, &group); err != nil {
			return
		}
		orders[group]++
		updates = append(updates, tagOrder{id: id, order: orders[group]})
	}
	for _, update := range updates {
		conn.Exec(`UPDATE tags SET sort_order = ? WHERE id = ?`, update.order, update.id)
	}
}

func currentTaxonomyVersion(conn *sql.DB) string {
	var taxonomy string
	if err := conn.QueryRow(`SELECT taxonomy_version FROM categories WHERE taxonomy_version != '' ORDER BY id LIMIT 1`).Scan(&taxonomy); err == nil && taxonomy != "" {
		return taxonomy
	}
	return "taxonomy_2026_v1"
}

func nextCategorySortOrder(conn *sql.DB, parentID int64, typ, taxonomy string) int {
	var maxOrder sql.NullInt64
	conn.QueryRow(`SELECT max(sort_order) FROM categories WHERE parent_id = ? AND type = ? AND taxonomy_version = ?`, parentID, typ, taxonomy).Scan(&maxOrder)
	if maxOrder.Valid {
		return int(maxOrder.Int64) + 1
	}
	return 1
}

func nextTagSortOrder(conn *sql.DB, groupName string) int {
	var maxOrder sql.NullInt64
	conn.QueryRow(`SELECT max(sort_order) FROM tags WHERE group_name = ?`, groupName).Scan(&maxOrder)
	if maxOrder.Valid {
		return int(maxOrder.Int64) + 1
	}
	return 1
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ensureCategoryUpdated(result sql.Result, id int64) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("category %d not found", id)
	}
	return nil
}

func ensureTagUpdated(result sql.Result, id int64) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("tag %d not found", id)
	}
	return nil
}

func ensureTagExistsTx(tx *sql.Tx, id int64) error {
	var exists int
	if err := tx.QueryRow(`SELECT count(*) FROM tags WHERE id = ?`, id).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("tag %d not found", id)
	}
	return nil
}

func findCategoryPair(conn *sql.DB, billType, category, subCategory string) (int64, int64, error) {
	var categoryID int64
	err := conn.QueryRow(
		`SELECT id FROM categories WHERE parent_id = 0 AND name = ? AND type = ? AND is_active = 1 ORDER BY id LIMIT 1`,
		category,
		billType,
	).Scan(&categoryID)
	if err == sql.ErrNoRows && billType != "expense" {
		err = conn.QueryRow(
			`SELECT id FROM categories WHERE parent_id = 0 AND name = ? AND type = 'expense' AND is_active = 1 ORDER BY id LIMIT 1`,
			category,
		).Scan(&categoryID)
	}
	if err == sql.ErrNoRows {
		return 0, 0, fmt.Errorf("category %q not found for type %q", category, billType)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("find category %q: %w", category, err)
	}

	var subCategoryID int64
	err = conn.QueryRow(
		`SELECT id FROM categories WHERE parent_id = ? AND name = ? AND is_active = 1 ORDER BY id LIMIT 1`,
		categoryID,
		subCategory,
	).Scan(&subCategoryID)
	if err == sql.ErrNoRows {
		return 0, 0, fmt.Errorf("subCategory %q not found under %q", subCategory, category)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("find subCategory %q: %w", subCategory, err)
	}
	return categoryID, subCategoryID, nil
}

func findTagID(conn *sql.DB, tag string) (int64, error) {
	var tagID int64
	err := conn.QueryRow(`SELECT id FROM tags WHERE name = ? AND is_active = 1 ORDER BY group_name, id LIMIT 1`, tag).Scan(&tagID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("tag %q not found", tag)
	}
	if err != nil {
		return 0, fmt.Errorf("find tag %q: %w", tag, err)
	}
	return tagID, nil
}

func findMerchantID(conn *sql.DB, name string) (*int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	var id int64
	err := conn.QueryRow(`SELECT id FROM merchants WHERE name = ? AND is_active = 1 LIMIT 1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find merchant %q: %w", name, err)
	}
	return &id, nil
}

func normalizeTagList(tags []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	return out
}

func ensureUpdated(result sql.Result, id int64) error {
	n, err := result.RowsAffected()
	if err != nil {
		return nil
	}
	if n == 0 {
		return fmt.Errorf("bill %d not found", id)
	}
	return nil
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
