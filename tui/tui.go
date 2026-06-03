package tui

import (
	"fmt"
	"os/exec"
	"strings"

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
	stateChecking
	stateUpdating
)

type Model struct {
	state     state
	database  *db.Database
	apps      []appRow
	table     table.Model
	spinner   spinner.Model
	search    textinput.Model
	searching bool

	updatingIdx int
	updatingErr string
	updatingOut string

	err     error
	width   int
	height  int
	ready   bool
	quitting bool
}

type appRow struct {
	App     scan.AppInfo
	Entry   db.AppEntry
	InDB    bool
	Checked bool
	Status  string // "newer", "current", "unknown", "ignored"
	Update  *update.UpdateInfo
}

type (
	scanDoneMsg     struct {
		database *db.Database
		apps     []appRow
	}
	checkDoneMsg    struct{ idx int }
	updateDoneMsg   struct{ err error; out string }
	errMsg          struct{ err error }
)

func New() Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	s.Spinner = spinner.Dot

	ti := textinput.New()
	ti.Placeholder = "Search apps..."
	ti.CharLimit = 50
	ti.Width = 30

	return Model{
		state:   stateLoading,
		spinner: s,
		search:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadDB,
	)
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

func (m Model) columns() []table.Column {
	return []table.Column{
		{Title: "", Width: 3},
		{Title: "App", Width: max(25, m.width/4-5)},
		{Title: "Version", Width: 18},
		{Title: "Latest", Width: 18},
		{Title: "Source", Width: 8},
		{Title: "Status", Width: 10},
	}
}

func (m Model) toRows(rows []appRow) []table.Row {
	var tr []table.Row
	for _, r := range rows {
		check := " "
		if r.Checked {
			check = "✓"
		}

		ver := r.App.Version
		if ver == "" {
			ver = r.App.Build
		}

		latest := ""
		source := ""
		status := ""

		switch r.Status {
		case "current":
			latest = ver
			source = "✓"
			status = "up-to-date"
		case "newer":
			if r.Update != nil {
				latest = r.Update.LatestVersion
			}
			source = "↑"
			status = "update"
		case "ignored":
			latest = "—"
			source = "⛔"
			status = "ignored"
		case "checking":
			source = "⋯"
			status = "checking"
		case "error":
			source = "⚠"
			status = "error"
		default:
			if r.InDB && r.Entry.FeedURL != "" {
				source = "?"
				status = "tap to check"
			} else if r.InDB {
				source = "☰"
				status = "in db"
			} else {
				source = "—"
				status = "unknown"
			}
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
				return m, nil
			}
			if m.state == stateReady || m.state == stateChecking || m.state == stateUpdating {
				m.quitting = true
				return m, tea.Quit
			}
		case "enter":
			if m.searching {
				m.searching = false
				m.search.Blur()
				return m, nil
			}
			if m.state == stateReady {
				sel := m.table.Cursor()
				if sel >= 0 && sel < len(m.apps) && m.apps[sel].InDB && m.apps[sel].Entry.FeedURL != "" && m.apps[sel].Status == "unknown" {
					m.apps[sel].Status = "checking"
					idx := sel
					m.state = stateChecking
					return m, func() tea.Msg {
						info, err := update.FetchFeed(m.apps[idx].Entry.FeedURL)
						if err != nil {
							m.apps[idx].Status = "error"
							return checkDoneMsg{idx}
						}
						if info != nil {
							m.apps[idx].Update = info
							cmp := update.CompareVersions(m.apps[idx].App.Version, info.LatestVersion)
							if cmp < 0 {
								m.apps[idx].Status = "newer"
							} else {
								m.apps[idx].Status = "current"
							}
						}
						return checkDoneMsg{idx}
					}
				}
			}
		case " ":
			if m.state == stateReady {
				sel := m.table.Cursor()
				if sel >= 0 && sel < len(m.apps) {
					m.apps[sel].Checked = !m.apps[sel].Checked
					m.table.SetRows(m.toRows(m.apps))
				}
			}
		case "u":
			if m.state == stateReady {
				return m, bulkCheck(m)
			}
		case "U":
			if m.state == stateReady {
				return m, startUpdate(m)
			}
		}

	case scanDoneMsg:
		m.database = msg.database
		m.apps = msg.apps

		t := table.New(
			table.WithColumns(m.columns()),
			table.WithRows(m.toRows(m.apps)),
			table.WithFocused(true),
			table.WithHeight(m.height-6),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m.table = t
		m.state = stateReady
		return m, nil

	case checkDoneMsg:
		m.state = stateReady
		m.table.SetRows(m.toRows(m.apps))
		return m, nil

	case updateDoneMsg:
		if msg.err != nil {
			m.updatingErr = msg.err.Error()
		} else {
			m.updatingOut = msg.out
		}
		m.state = stateReady
		m.table.SetRows(m.toRows(m.apps))
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
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

func bulkCheck(m Model) tea.Cmd {
	var cmds []tea.Cmd
	for i, app := range m.apps {
		if app.InDB && app.Entry.FeedURL != "" && app.Status == "unknown" {
			m.apps[i].Status = "checking"
			idx := i
			cmds = append(cmds, func() tea.Msg {
				info, err := update.FetchFeed(m.apps[idx].Entry.FeedURL)
				if err != nil {
					m.apps[idx].Status = "error"
					return checkDoneMsg{idx}
				}
				if info != nil {
					m.apps[idx].Update = info
					cmp := update.CompareVersions(m.apps[idx].App.Version, info.LatestVersion)
					if cmp < 0 {
						m.apps[idx].Status = "newer"
					} else {
						m.apps[idx].Status = "current"
					}
				}
				return checkDoneMsg{idx}
			})
		}
	}
	return tea.Sequence(cmds...)
}

func startUpdate(m Model) tea.Cmd {
	var cmds []tea.Cmd
	for i, app := range m.apps {
		if app.Checked && app.Status == "newer" && app.Update != nil {
			idx := i
			cmds = append(cmds, func() tea.Msg {
				return updateApp(m.apps[idx])
			})
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Sequence(cmds...)
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

	openCmd := exec.Command("open", tmpPath)
	if err := openCmd.Run(); err != nil {
		return updateDoneMsg{err: fmt.Errorf("open %s: %w", tmpPath, err)}
	}

	return updateDoneMsg{out: fmt.Sprintf("Downloaded %s (%s)\nSaved to: %s", r.App.Name, r.Update.LatestVersion, tmpPath)}
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.quitting {
		return "\n  Bye!\n\n"
	}

	switch m.state {
	case stateLoading:
		return fmt.Sprintf("\n  %s Scanning /Applications...\n", m.spinner.View())

	case stateReady, stateChecking:
		var b strings.Builder

		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1).
			Render("OpenMacUpdate")

		help := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("  ↑↓ select • / search • ⏎ check • u check all • space select • U update selected • q quit")

		b.WriteString("\n  ")
		b.WriteString(title)
		b.WriteString("\n\n")

		if m.searching {
			b.WriteString("  " + m.search.View() + "\n\n")
		}

		b.WriteString(m.table.View())
		b.WriteString("\n")
		b.WriteString(help)

		return b.String()

	case stateUpdating:
		return fmt.Sprintf("\n  %s Updating...\n", m.spinner.View())
	}

	return "loading..."
}
