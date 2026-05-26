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
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Children []CategoryNode `json:"children"`
}

type TagItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
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

	var hasColumn int
	err = conn.QueryRow(`SELECT count(*) FROM pragma_table_info('bills') WHERE name = 'is_deleted'`).Scan(&hasColumn)
	if err != nil || hasColumn > 0 {
		return
	}
	conn.Exec(`ALTER TABLE bills ADD COLUMN is_deleted INTEGER NOT NULL DEFAULT 0`)
	conn.Exec(`ALTER TABLE bills ADD COLUMN deleted_at TEXT`)
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
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT id, parent_id, name, type
FROM categories
WHERE is_active = 1
ORDER BY parent_id, sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query category tree: %w", err)
	}
	defer rows.Close()

	roots := []CategoryNode{}
	rootIndex := map[int64]int{}
	childrenByParent := map[int64][]CategoryNode{}
	for rows.Next() {
		var id, parentID int64
		var node CategoryNode
		if err := rows.Scan(&id, &parentID, &node.Name, &node.Type); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		node.ID = id
		node.Children = []CategoryNode{}
		if parentID == 0 {
			rootIndex[id] = len(roots)
			roots = append(roots, node)
			continue
		}
		childrenByParent[parentID] = append(childrenByParent[parentID], node)
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

func (a *App) GetAllTags() ([]TagItem, error) {
	conn, err := a.dbConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT id, name, group_name
FROM tags
WHERE is_active = 1
ORDER BY group_name, name`)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	tags := []TagItem{}
	for rows.Next() {
		var tag TagItem
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.GroupName); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}
	return tags, nil
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
