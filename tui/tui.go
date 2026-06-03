package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yakushev/OpenMacUpdate/db"
	"github.com/yakushev/OpenMacUpdate/scan"
	"github.com/yakushev/OpenMacUpdate/update"
)

type state int

const (
	stateLoading state = iota
	stateReady
)

type filterMode int

const (
	filterUpdates filterMode = iota
	filterAll
	filterErrors
)

type logEntry struct {
	time    time.Time
	message string
	level   string
}

type Model struct {
	state    state
	database *db.Database
	apps     []appRow
	filtered []int
	table    table.Model
	spinner  spinner.Model
	progress progress.Model
	search   textinput.Model
	searching bool
	filter   filterMode

	checkTotal     int
	checkDone      int
	checking       bool
	checkComplete  bool

	updatingErr string
	updatingOut string

	logs    []logEntry
	maxLogs int

	err      error
	width    int
	height   int
	ready    bool
	quitting bool
}

type appRow struct {
	App     scan.AppInfo
	Entry   db.AppEntry
	InDB    bool
	Checked bool
	Status  string
	Update  *update.UpdateInfo
	Source  string
}

type (
	scanDoneMsg   struct {
		database *db.Database
		apps     []appRow
	}
	autoCheckStartMsg struct{}
	checkDoneMsg      struct {
		idx  int
		info *update.UpdateInfo
		err  error
	}
	updateDoneMsg struct {
		err error
		out string
	}
	brewDoneMsg struct {
		outdated []update.BrewCaskInfo
		err      error
	}
	masDoneMsg struct {
		outdated []update.MasAppInfo
		err      error
	}
	errMsg struct{ err error }
)

var (
	cPrimary   = lipgloss.Color("39")
	cGreen     = lipgloss.Color("46")
	cOrange    = lipgloss.Color("208")
	cRed       = lipgloss.Color("196")
	cGray      = lipgloss.Color("240")
	cDarkGray  = lipgloss.Color("238")
	cWhite     = lipgloss.Color("255")
	cHighlight = lipgloss.Color("69")
)

func New() Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(cHighlight)
	s.Spinner = spinner.MiniDot

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(30),
		progress.WithoutPercentage(),
	)

	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 50
	ti.Width = 25

	return Model{
		state:    stateLoading,
		spinner:  s,
		progress: p,
		search:   ti,
		filter:   filterUpdates,
		maxLogs:  4,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.progress.Init(), loadDB)
}

func loadDB() tea.Msg {
	d, err := db.Load()
	if err != nil {
		return errMsg{err}
	}
	apps, err := scan.ScanDefault()
	if err != nil {
		return errMsg{err}
	}

	var rows []appRow
	for _, app := range apps {
		r := appRow{App: app, Status: "unknown"}
		if entry, ok := d.Lookup(app.BundleID); ok {
			r.InDB = true
			r.Entry = entry
			if entry.IsIgnored != 0 {
				r.Status = "ignored"
			}
		}
		rows = append(rows, r)
	}
	return scanDoneMsg{database: d, apps: rows}
}

func (m *Model) addLog(msg, level string) {
	m.logs = append(m.logs, logEntry{time: time.Now(), message: msg, level: level})
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[len(m.logs)-m.maxLogs:]
	}
}

func (m Model) columns() []table.Column {
	w := m.width - 4
	if w < 70 {
		w = 70
	}
	appW := w * 25 / 100
	verW := w * 20 / 100
	latestW := w * 20 / 100
	srcW := w * 7 / 100
	statusW := w - 3 - appW - verW - latestW - srcW
	return []table.Column{
		{Title: " ", Width: 3},
		{Title: "App", Width: appW},
		{Title: "Installed", Width: verW},
		{Title: "Latest", Width: latestW},
		{Title: "", Width: srcW},
		{Title: "Status", Width: statusW},
	}
}

func (m Model) visibleRows() []int {
	var indices []int
	query := strings.ToLower(m.search.Value())
	for i, r := range m.apps {
		switch m.filter {
		case filterUpdates:
			if r.Status != "newer" {
				continue
			}
		case filterErrors:
			if r.Status != "error" {
				continue
			}
		}
		if query != "" {
			name := strings.ToLower(r.App.Name)
			bid := strings.ToLower(r.App.BundleID)
			if !strings.Contains(name, query) && !strings.Contains(bid, query) {
				continue
			}
		}
		indices = append(indices, i)
	}
	return indices
}

func (m Model) toRows(indices []int) []table.Row {
	var tr []table.Row
	for _, i := range indices {
		r := m.apps[i]
		check := " "
		if r.Checked {
			check = lipgloss.NewStyle().Foreground(cGreen).Bold(true).Render("*")
		}

		ver := r.App.Version
		if ver == "" {
			ver = r.App.Build
		}
		if ver == "" {
			ver = "—"
		}

		latest := ""
		source := ""
		status := ""

		switch r.Status {
		case "current":
			latest = lipgloss.NewStyle().Foreground(cGray).Render(ver)
			status = lipgloss.NewStyle().Foreground(cGreen).Render("OK")
		case "newer":
			if r.Update != nil {
				latest = lipgloss.NewStyle().Foreground(cOrange).Bold(true).Render(r.Update.LatestVersion)
			}
			status = lipgloss.NewStyle().Foreground(cOrange).Bold(true).Render("UPDATE")
		case "ignored":
			latest = lipgloss.NewStyle().Foreground(cDarkGray).Render("—")
			status = lipgloss.NewStyle().Foreground(cDarkGray).Render("skip")
		case "checking":
			status = lipgloss.NewStyle().Foreground(cHighlight).Render("...")
		case "error":
			status = lipgloss.NewStyle().Foreground(cRed).Render("ERR")
		default:
			if r.InDB && r.Entry.FeedURL != "" {
				status = lipgloss.NewStyle().Foreground(cDarkGray).Render("pending")
			} else {
				status = lipgloss.NewStyle().Foreground(cDarkGray).Render("—")
			}
		}

		switch r.Source {
		case "brew":
			source = lipgloss.NewStyle().Foreground(cOrange).Render("brew")
		case "mas":
			source = lipgloss.NewStyle().Foreground(cPrimary).Render("mas")
		case "":
			if r.InDB && r.Entry.FeedURL != "" {
				source = lipgloss.NewStyle().Foreground(cDarkGray).Render("xml")
			}
		default:
			source = lipgloss.NewStyle().Foreground(cHighlight).Render(r.Source)
		}

		tr = append(tr, table.Row{check, r.App.Name, ver, latest, source, status})
	}
	return tr
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		if m.state == stateReady {
			m.ensureTable()
		}
		return m, nil

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "/":
			m.searching = true
			m.search.Focus()
			return m, nil
		case "esc":
			if m.searching {
				m.searching = false
				m.search.Blur()
				m.search.SetValue("")
				m.refreshTable()
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.searching {
				m.searching = false
				m.search.Blur()
				m.refreshTable()
				return m, nil
			}
			sel := m.selectedAppIndex()
			if sel >= 0 {
				m.addLog("Launching "+m.apps[sel].App.Name+"...", "info")
				go exec.Command("open", m.apps[sel].App.Path).Run()
			}
		case "f":
			if !m.searching {
				m.filter = (m.filter + 1) % 3
				m.refreshTable()
			}
		case " ":
			if !m.searching {
				sel := m.selectedAppIndex()
				if sel >= 0 {
					m.apps[sel].Checked = !m.apps[sel].Checked
					m.refreshTable()
				}
			}
		case "a":
			if !m.searching {
				allChecked := true
				for _, i := range m.filtered {
					if !m.apps[i].Checked {
						allChecked = false
						break
					}
				}
				for _, i := range m.filtered {
					m.apps[i].Checked = !allChecked
				}
				m.refreshTable()
			}
		case "u":
			if !m.searching && !m.checking {
				return m, m.startAutoCheck()
			}
		case "U":
			if !m.searching {
				return m, m.startUpdate()
			}
		case "r":
			if !m.searching {
				sel := m.selectedAppIndex()
				if sel >= 0 && m.apps[sel].Update != nil && m.apps[sel].Update.ReleaseNotes != "" {
					go exec.Command("open", m.apps[sel].Update.ReleaseNotes).Run()
				}
			}
		}

	case scanDoneMsg:
		m.database = msg.database
		m.apps = msg.apps
		m.state = stateReady
		m.refreshTable()
		m.addLog(fmt.Sprintf("Scanned %d apps [db: %s]", len(m.apps), msg.database.Source), "info")
		return m, func() tea.Msg { return autoCheckStartMsg{} }

	case autoCheckStartMsg:
		return m, m.startAutoCheck()

	case checkDoneMsg:
		appName := m.apps[msg.idx].App.Name
		if msg.err != nil {
			m.apps[msg.idx].Status = "error"
		} else if msg.info != nil {
			m.apps[msg.idx].Update = msg.info
			cmp := update.CompareVersions(m.apps[msg.idx].App.Version, msg.info.LatestVersion)
			if cmp < 0 {
				m.apps[msg.idx].Status = "newer"
				m.addLog(fmt.Sprintf("%s  %s -> %s", appName, m.apps[msg.idx].App.Version, msg.info.LatestVersion), "warn")
			} else {
				m.apps[msg.idx].Status = "current"
			}
		} else {
			m.apps[msg.idx].Status = "current"
		}
		m.checkDone++
		if m.checkDone >= m.checkTotal && !m.checkComplete {
			m.checking = false
			m.checkComplete = true
			updates := 0
			for _, a := range m.apps {
				if a.Status == "newer" {
					updates++
				}
			}
			if updates > 0 {
				m.addLog(fmt.Sprintf("Done — %d updates available", updates), "warn")
			} else {
				m.addLog("Done — all apps up to date", "success")
			}
		}
		m.refreshTable()
		return m, nil

	case updateDoneMsg:
		if msg.err != nil {
			m.updatingErr = msg.err.Error()
			m.addLog("Update failed: "+msg.err.Error(), "error")
		} else {
			m.updatingOut = msg.out
			m.addLog(msg.out, "success")
		}
		m.refreshTable()
		return m, nil

	case brewDoneMsg:
		if msg.err == nil {
			for _, cask := range msg.outdated {
				for i := range m.apps {
					if m.apps[i].Entry.CaskName == cask.Name || strings.EqualFold(m.apps[i].App.Name, cask.Name) {
						m.apps[i].Status = "newer"
						m.apps[i].Source = "brew"
						m.apps[i].Update = &update.UpdateInfo{LatestVersion: cask.Latest}
					}
				}
			}
		}
		m.refreshTable()
		return m, nil

	case masDoneMsg:
		if msg.err == nil {
			for _, app := range msg.outdated {
				for i := range m.apps {
					if strings.Contains(m.apps[i].App.BundleID, "com.apple") && strings.EqualFold(m.apps[i].App.Name, app.Name) {
						m.apps[i].Status = "newer"
						m.apps[i].Source = "mas"
						m.apps[i].Update = &update.UpdateInfo{LatestVersion: app.Version}
					}
				}
			}
		}
		m.refreshTable()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	}

	if m.searching {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}
	if m.state == stateReady {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) ensureTable() {
	rows := m.toRows(m.filtered)
	if m.table.Columns() == nil {
		t := table.New(
			table.WithColumns(m.columns()),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.calcTableHeight()),
		)
		s := table.DefaultStyles()
		s.Header = lipgloss.NewStyle().
			Foreground(cGray).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(cDarkGray)
		s.Selected = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(lipgloss.Color("237")).
			Bold(false)
		t.SetStyles(s)
		m.table = t
	} else {
		m.table.SetRows(rows)
		if m.width > 0 {
			m.table.SetColumns(m.columns())
			m.table.SetHeight(m.calcTableHeight())
		}
	}
}

func (m *Model) refreshTable() {
	m.filtered = m.visibleRows()
	m.ensureTable()
}

func (m *Model) selectedAppIndex() int {
	sel := m.table.Cursor()
	if sel >= 0 && sel < len(m.filtered) {
		return m.filtered[sel]
	}
	return -1
}

func (m Model) calcTableHeight() int {
	h := m.height - 12
	if h < 3 {
		h = 3
	}
	return h
}

func (m Model) startAutoCheck() tea.Cmd {
	m.checking = true
	m.checkComplete = false
	m.checkTotal = 0
	m.checkDone = 0

	var cmds []tea.Cmd

	cmds = append(cmds, func() tea.Msg {
		outdated, err := update.BrewListOutdated()
		return brewDoneMsg{outdated: outdated, err: err}
	})
	cmds = append(cmds, func() tea.Msg {
		outdated, err := update.MasListOutdated()
		return masDoneMsg{outdated: outdated, err: err}
	})

	for i, app := range m.apps {
		if app.Status == "ignored" {
			continue
		}
		feedURL := app.Entry.GetFeedURL(runtime.GOARCH)
		if feedURL == "" {
			feedURL = app.App.FeedURL
		}
		if feedURL != "" {
			m.checkTotal++
			idx := i
			cmds = append(cmds, func() tea.Msg {
				info, err := update.FetchFeed(feedURL)
				return checkDoneMsg{idx: idx, info: info, err: err}
			})
		}
	}

	if m.checkTotal == 0 {
		m.checking = false
		m.checkComplete = true
	}
	return tea.Batch(cmds...)
}

func (m Model) startUpdate() tea.Cmd {
	var cmds []tea.Cmd
	for i, app := range m.apps {
		if !app.Checked || app.Status != "newer" || app.Update == nil {
			continue
		}
		idx := i
		if app.Source == "brew" && app.Entry.CaskName != "" {
			cmds = append(cmds, func() tea.Msg {
				out, err := update.BrewUpgrade(m.apps[idx].Entry.CaskName)
				return updateDoneMsg{err: err, out: out}
			})
		} else if app.Source == "mas" {
			cmds = append(cmds, func() tea.Msg {
				out, err := update.MasUpgrade(m.apps[idx].App.BundleID)
				return updateDoneMsg{err: err, out: out}
			})
		} else if app.Update.DownloadURL != "" {
			cmds = append(cmds, func() tea.Msg {
				return updateApp(m.apps[idx])
			})
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func updateApp(r appRow) tea.Msg {
	dl := r.Update.DownloadURL
	if dl == "" {
		return updateDoneMsg{err: fmt.Errorf("no download URL for %s", r.App.Name)}
	}
	tmpPath := "/tmp/" + r.App.Name + ".download"
	cmd := exec.Command("curl", "-L", "-o", tmpPath, "--progress-bar", dl)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return updateDoneMsg{err: fmt.Errorf("download %s: %w\n%s", r.App.Name, err, string(out))}
	}
	if err := exec.Command("open", tmpPath).Run(); err != nil {
		return updateDoneMsg{err: fmt.Errorf("open %s: %w", tmpPath, err)}
	}
	return updateDoneMsg{out: fmt.Sprintf("Downloaded %s %s", r.App.Name, r.Update.LatestVersion)}
}

func (m Model) statsView() string {
	total := len(m.apps)
	updates, current, unknown, errors := 0, 0, 0, 0
	for _, a := range m.apps {
		switch a.Status {
		case "newer":
			updates++
		case "current":
			current++
		case "unknown":
			unknown++
		case "error":
			errors++
		}
	}

	s := fmt.Sprintf(" %d apps ", total)
	out := lipgloss.NewStyle().Foreground(cPrimary).Background(lipgloss.Color("236")).Bold(true).Render(s)

	if updates > 0 {
		out += " " + lipgloss.NewStyle().Foreground(cOrange).Background(lipgloss.Color("236")).Bold(true).Render(fmt.Sprintf(" %d updates ", updates))
	}
	if current > 0 {
		out += " " + lipgloss.NewStyle().Foreground(cGreen).Background(lipgloss.Color("236")).Render(fmt.Sprintf(" %d ok ", current))
	}
	if unknown > 0 {
		out += " " + lipgloss.NewStyle().Foreground(cGray).Background(lipgloss.Color("236")).Render(fmt.Sprintf(" %d pending ", unknown))
	}
	if errors > 0 {
		out += " " + lipgloss.NewStyle().Foreground(cRed).Background(lipgloss.Color("236")).Render(fmt.Sprintf(" %d err ", errors))
	}

	return out
}

func (m Model) progressView() string {
	if !m.checking || m.checkTotal == 0 {
		return ""
	}
	pct := float64(m.checkDone) / float64(m.checkTotal)
	return fmt.Sprintf(" %s %d/%d", m.progress.ViewAs(pct), m.checkDone, m.checkTotal)
}

func (m Model) logView() string {
	if len(m.logs) == 0 {
		return ""
	}
	var lines []string
	for _, l := range m.logs {
		ts := lipgloss.NewStyle().Foreground(cDarkGray).Render(l.time.Format("15:04:05"))
		var msg string
		switch l.level {
		case "success":
			msg = lipgloss.NewStyle().Foreground(cGreen).Render(l.message)
		case "error":
			msg = lipgloss.NewStyle().Foreground(cRed).Render(l.message)
		case "warn":
			msg = lipgloss.NewStyle().Foreground(cOrange).Render(l.message)
		default:
			msg = lipgloss.NewStyle().Foreground(cGray).Render(l.message)
		}
		lines = append(lines, "  "+ts+" "+msg)
	}
	return strings.Join(lines, "\n")
}

func (m Model) helpView() string {
	fLabel := "updates"
	switch m.filter {
	case filterAll:
		fLabel = "all"
	case filterErrors:
		fLabel = "errors"
	}

	keys := []string{
		"q:quit", "/:search", "enter:open", "u:scan",
		"space:select", "a:all", "U:update",
		"f:" + fLabel, "r:notes",
	}
	return "  " + lipgloss.NewStyle().Foreground(cDarkGray).Render(strings.Join(keys, "  "))
}

func (m Model) View() string {
	if m.err != nil {
		return "\n  " + lipgloss.NewStyle().Foreground(cRed).Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}
	if m.quitting {
		return "\n  bye\n\n"
	}

	switch m.state {
	case stateLoading:
		return fmt.Sprintf("\n  %s Scanning...\n", m.spinner.View())
	}

	var b strings.Builder
	b.WriteString("\n")

	title := lipgloss.NewStyle().Foreground(cPrimary).Bold(true).Render(" OpenMacUpdate")
	b.WriteString(" " + title)
	if m.checking {
		b.WriteString("  " + m.spinner.View() + m.progressView())
	}
	b.WriteString("\n")
	b.WriteString(" " + m.statsView() + "\n\n")

	if m.searching {
		b.WriteString("  " + m.search.View() + "\n\n")
	}

	if len(m.filtered) == 0 && m.checkComplete {
		msg := "No updates available"
		if m.filter == filterErrors {
			msg = "No errors"
		}
		b.WriteString("  " + lipgloss.NewStyle().Foreground(cGray).Render(msg) + "\n\n")
	} else {
		b.WriteString(m.table.View() + "\n")
	}

	logs := m.logView()
	if logs != "" {
		b.WriteString("\n" + logs + "\n")
	}

	b.WriteString("\n" + m.helpView())
	return b.String()
}
