package main

import (
	"context"
	"os/exec"
)

const (
	accountBookExe = `D:\Project\self\mira-ledger\account-book.exe`
	ledgerRoot     = `D:\Project\self\mira-ledger`
	ledgerDB       = `D:\Project\self\mira-ledger\data\account-book.db`
	backupRepo     = `D:\Project\self\mira-ledger-data`
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetVerify() string {
	return runAccountBook("verify", "--db", ledgerDB)
}

func (a *App) GetMonthlyStats(month string) string {
	return runAccountBook("stats", "--db", ledgerDB, "--month", month)
}

func (a *App) GetMonthlySummary(month string) string {
	return runAccountBook("summary", "--db", ledgerDB, "--month", month, "--format", "json")
}

func (a *App) RunGitHubBackup(month string) string {
	return runAccountBook("backup-github", "--db", ledgerDB, "--month", month, "--repo", backupRepo)
}

func runAccountBook(args ...string) string {
	cmd := exec.Command(accountBookExe, args...)
	cmd.Dir = ledgerRoot
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return err.Error()
	}
	return string(out)
}
