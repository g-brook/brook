package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/g-brook/brook/common/version"
)

// ============ Package Documentation ============

// Package cli provides a terminal user interface for the Brook tunnel client.

// ============ Constants ============

const (
	MaxLogLines          = 1000
	MinLogHeight         = 5
	MaxLogHeight         = 30
	DefaultLogHeight     = 12
	MinTunnelsHeight     = 8
	MaxTunnelsHeight     = 20
	DefaultTunnelsHeight = 10 // ← 添加这个常量
	DefaultTotalHeight   = 40
	MinTotalHeight       = 24
	MaxTotalHeight       = 200
)

var mainView tea.View

// 固定元素高度常量
const (
	BannerFixedHeight        = 5 // 4行横幅 + 1空行
	StatusFixedHeight        = 8 // 状态面板实际高度
	DividerFixedHeight       = 2 // 分隔线 + 换行
	HelpFixedHeight          = 1 // 帮助栏
	SectionHeaderFixedHeight = 2 // 每个区域标题（标题 + 换行）
)

// ============ Data Type Definitions ============

// TunnelCon represents tunnel connection information
type TunnelCon struct {
	Addr      string
	Port      string
	LocalAddr string
	Protocol  string
	State     string
}

// SyncMap is a simple thread-safe map
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func (m *SyncMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return f(k.(K), v.(V))
	})
}

func (m *SyncMap[K, V]) Values() []V {
	var values []V
	m.m.Range(func(_, v interface{}) bool {
		values = append(values, v.(V))
		return true
	})
	return values
}

func (m *SyncMap[K, V]) Len() int {
	length := 0
	m.m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

func (m *SyncMap[K, V]) Clear() {
	m.m.Range(func(key, _ interface{}) bool {
		m.m.Delete(key)
		return true
	})
}

// ============ Style Definitions ============

var (
	// Cyber color palette
	colorCyan    = lipgloss.Color("#00F5FF")
	colorMagenta = lipgloss.Color("#FF2D78")
	colorPurple  = lipgloss.Color("#BD93F9")
	colorGreen   = lipgloss.Color("#50FA7B")
	colorRed     = lipgloss.Color("#FF5555")
	colorYellow  = lipgloss.Color("#FFB86C")
	colorDim     = lipgloss.Color("#44475A")
	colorFg      = lipgloss.Color("#F8F8F2")
	// Banner / title
	bannerStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)
	// Top border line
	dividerStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Status panel
	statusPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Border(lipgloss.RoundedBorder(), true, true, true, true).
				BorderForeground(colorCyan).
				Padding(0, 2)

	statusPanelStyleTopBorder = lipgloss.NewStyle().
					Align(lipgloss.Center)
	statusLabelStyle = lipgloss.NewStyle().
				Foreground(colorPurple).
				Align(lipgloss.Right)

	statusValueStyle = lipgloss.NewStyle().
				Foreground(colorFg)

	statusOnlineStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	statusOfflineStyle = lipgloss.NewStyle().
				Foreground(colorRed).
				Bold(true)

	latencyGoodStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	latencyMidStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	latencyBadStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	noConnectionStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Italic(true).
				Padding(0, 2)

	connOKStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	connNOStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	logTitleStyle = lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true).
			Padding(0, 1)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// ASCII banner
var bannerLines = `Welcome to Brook v` + version.GetBuildVersion()

// ============ Custom Message Types ============

type (
	LogMsg          string
	StatusUpdateMsg struct {
		Status string
	}
	AddressUpdateMsg struct {
		Address string
	}
	ConnUpdateMsg struct {
		Address   string
		Port      int
		LocalAddr string
		Protocol  string
		IsClose   bool
	}
	ConnStateUpdateMsg struct {
		IsClose bool
	}
	LatencyUpdateMsg struct {
		Latency int64
	}
	FocusChangeMsg struct {
		View string
	}
	ResizeMsg struct {
		View      string
		Direction string
	}
	SetTotalHeightMsg struct {
		Height int
	}
	tickMsg time.Time
)

// ============ Helper Functions ============

func createTable() table.Model {
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "PROTO", Width: 10},
		{Title: "REMOTE → LOCAL", Width: 50},
		{Title: "STATE", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(8),
		table.WithWidth(84),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorDim).
		BorderBottom(true).
		Bold(true).
		Foreground(colorCyan)
	s.Cell = s.Cell.
		Foreground(colorFg)

	t.SetStyles(s)
	return t
}

// ============ View Interface ============

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View(width int) string
	GetHeight() int
	SetHeight(height int)
	GetMinHeight() int
	GetMaxHeight() int
	GetTitle() string
	HandleKeyMsg(msg tea.KeyMsg) bool
}

// ============ Tunnels View ============

type TunnelsView struct {
	ConnectionTable table.Model
	Connections     *SyncMap[string, *TunnelCon]
	Width           int
	Height          int
	lastConnHash    uint64
	cachedTableView string
	focused         bool
}

func NewTunnelsView() *TunnelsView {
	return &TunnelsView{
		ConnectionTable: createTable(),
		Connections:     NewSyncMap[string, *TunnelCon](),
		Height:          DefaultTunnelsHeight, // ← 现在可以正常使用
		focused:         true,
	}
}

func (v *TunnelsView) Init() tea.Cmd {
	return nil
}

func (v *TunnelsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case ConnUpdateMsg:
		state := "LIVE"
		if msg.IsClose {
			state = "DEAD"
		}
		key := fmt.Sprintf("%s:%d", msg.Address, msg.Port)
		v.Connections.Store(key, &TunnelCon{
			Addr:      msg.Address,
			Port:      strconv.Itoa(msg.Port),
			Protocol:  msg.Protocol,
			LocalAddr: msg.LocalAddr,
			State:     state,
		})
		v.invalidateCache()

	case ConnStateUpdateMsg:
		v.Connections.Range(func(key string, value *TunnelCon) bool {
			if msg.IsClose {
				value.State = "DEAD"
			} else {
				value.State = "LIVE"
			}
			return true
		})
		v.invalidateCache()

	case tea.WindowSizeMsg:
		v.Width = msg.Width
		v.handleResize()
	}

	return v, nil
}

func (v *TunnelsView) View(width int) string {
	v.Width = width
	v.handleResize()

	connCount := v.Connections.Len()

	var sb strings.Builder
	sb.WriteString("\n")

	if connCount == 0 {
		sb.WriteString(noConnectionStyle.Render("no active tunnels"))
		sb.WriteString("\n")
		return sb.String()
	}

	currentHash := v.calculateConnectionsHash()
	if currentHash != v.lastConnHash || v.cachedTableView == "" {
		v.lastConnHash = currentHash
		v.cachedTableView = v.buildConnectionsTable()
	}

	sb.WriteString(v.cachedTableView)
	sb.WriteString("\n")
	//sb.WriteString(sectionHeaderStyle.Render(fmt.Sprintf("⟫ TUNNELS  [%d active]", connCount)))

	return sb.String()
}

func (v *TunnelsView) handleResize() {
	v.ConnectionTable.SetWidth(v.Width - 4)
}

func (v *TunnelsView) GetHeight() int {
	return v.Height
}

func (v *TunnelsView) SetHeight(height int) {
	if height < v.GetMinHeight() {
		height = v.GetMinHeight()
	}
	if height > v.GetMaxHeight() {
		height = v.GetMaxHeight()
	}
	v.Height = height
}

func (v *TunnelsView) GetMinHeight() int {
	return MinTunnelsHeight
}

func (v *TunnelsView) GetMaxHeight() int {
	return MaxTunnelsHeight
}

func (v *TunnelsView) GetTitle() string {
	return "Tunnels"
}

func (v *TunnelsView) HandleKeyMsg(msg tea.KeyMsg) bool {
	return false
}

func (v *TunnelsView) invalidateCache() {
	v.lastConnHash = 0
	v.cachedTableView = ""
}

func (v *TunnelsView) calculateConnectionsHash() uint64 {
	var hash uint64 = 14695981039346656037
	v.Connections.Range(func(key string, value *TunnelCon) bool {
		data := key + value.State + value.Protocol + value.LocalAddr + value.Port
		for _, b := range []byte(data) {
			hash ^= uint64(b)
			hash *= 1099511628211
		}
		return true
	})
	return hash
}

func (v *TunnelsView) buildConnectionsTable() string {
	values := v.Connections.Values()
	rows := make([]table.Row, 0, len(values))

	for i, c := range values {
		targetIP := fmt.Sprintf("%s:%s  →  %s", c.Addr, c.Port, c.LocalAddr)

		var stateCell string
		switch c.State {
		case "LIVE":
			stateCell = connOKStyle.Render("◉ LIVE")
		case "DEAD":
			stateCell = connNOStyle.Render("◎ DEAD")
		default:
			stateCell = c.State
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			c.Protocol,
			targetIP,
			stateCell,
		})
	}

	v.ConnectionTable.SetRows(rows)
	return v.ConnectionTable.View()
}

// ============ Logs View ============

type LogsView struct {
	Logs     []string
	Viewport viewport.Model
	Width    int
	Height   int
	Ready    bool
	MaxLines int
	focused  bool
}

func NewLogsView() *LogsView {
	return &LogsView{
		Logs:     make([]string, 0),
		Height:   DefaultLogHeight,
		MaxLines: MaxLogLines,
		focused:  false,
	}
}

func (v *LogsView) Init() tea.Cmd {
	return nil
}

func (v *LogsView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case LogMsg:
		ts := time.Now().Format("15:04:05")
		line := fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(colorDim).Render(ts),
			string(msg),
		)
		v.Logs = append(v.Logs, line)

		if len(v.Logs) > v.MaxLines {
			v.Logs = v.Logs[len(v.Logs)-v.MaxLines:]
		}

		if v.Ready {
			v.Viewport.SetContent(strings.Join(v.Logs, "\n"))
			v.Viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		v.Width = msg.Width
		v.handleResize()

	case tea.KeyMsg:
		if v.Ready && v.focused {
			v.Viewport, cmd = v.Viewport.Update(msg)
		}
	}

	return v, cmd
}

func (v *LogsView) View(width int) string {
	v.Width = width
	v.handleResize()

	var sb strings.Builder
	sb.WriteString(logTitleStyle.Render(
		fmt.Sprintf("⟫ LOGS  [%d/%d lines  ·  +/- resize]", len(v.Logs), v.Height),
	))
	sb.WriteString("\n")

	if v.Ready {
		sb.WriteString(v.Viewport.View())
	} else {
		sb.WriteString("  initializing log stream...")
	}
	sb.WriteString("\n")

	return sb.String()
}

func (v *LogsView) handleResize() {
	contentWidth := v.Width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	if !v.Ready {
		v.Viewport = viewport.New(
			viewport.WithWidth(contentWidth),
			viewport.WithHeight(v.Height),
		)
		v.Ready = true
	} else {
		v.Viewport.SetWidth(contentWidth)
		v.Viewport.SetHeight(v.Height)
	}

	if len(v.Logs) > 0 {
		v.Viewport.SetContent(strings.Join(v.Logs, "\n"))
	}
}

func (v *LogsView) GetHeight() int {
	return v.Height
}

func (v *LogsView) SetHeight(height int) {
	if height < v.GetMinHeight() {
		height = v.GetMinHeight()
	}
	if height > v.GetMaxHeight() {
		height = v.GetMaxHeight()
	}
	v.Height = height
	v.MaxLines = height * 3
	if v.MaxLines > MaxLogLines {
		v.MaxLines = MaxLogLines
	}
	v.handleResize()
}

func (v *LogsView) GetMinHeight() int {
	return MinLogHeight
}

func (v *LogsView) GetMaxHeight() int {
	return MaxLogHeight
}

func (v *LogsView) GetTitle() string {
	return "Logs"
}

func (v *LogsView) HandleKeyMsg(msg tea.KeyMsg) bool {
	if !v.focused {
		return false
	}

	switch msg.String() {
	case "+":
		if v.Height < v.GetMaxHeight() {
			v.SetHeight(v.Height + 2)
		}
		return true
	case "-":
		if v.Height > v.GetMinHeight() {
			v.SetHeight(v.Height - 2)
		}
		return true
	case "b":
		if v.Ready {
			v.Viewport.GotoBottom()
		}
		return true
	}
	return false
}

// ============ Main TUI Model ============

type TUIModel struct {
	// State data
	RemoteAddress string
	Status        string
	Latency       int64

	// Spinner
	spinnerIdx int

	// Views
	TunnelsView *TunnelsView
	LogsView    *LogsView
	ActiveView  string

	// Layout related
	Width         int
	Height        int
	ContentHeight int

	// Fixed heights
	BannerHeight        int
	StatusHeight        int
	DividerHeight       int
	HelpHeight          int
	SectionHeaderHeight int
	MainView            tea.View

	// Ready flag
	ready bool
}

// NewTUIModel creates a new TUI model with default height
func NewTUIModel(remoteAddress string) *TUIModel {
	return &TUIModel{
		Status:        "offline",
		Latency:       0,
		RemoteAddress: remoteAddress,
		TunnelsView:   NewTunnelsView(),
		LogsView:      NewLogsView(),
		ActiveView:    "tunnels",

		Width:  80,
		Height: DefaultTotalHeight,

		BannerHeight:        BannerFixedHeight,
		StatusHeight:        StatusFixedHeight,
		DividerHeight:       DividerFixedHeight,
		HelpHeight:          HelpFixedHeight,
		SectionHeaderHeight: SectionHeaderFixedHeight,
		MainView:            tea.NewView(""),

		ready: false,
	}
}

// SetTotalHeight sets the total height of the interface
func (m *TUIModel) SetTotalHeight(height int) {
	if height < MinTotalHeight {
		height = MinTotalHeight
	}
	if height > MaxTotalHeight {
		height = MaxTotalHeight
	}
	m.Height = height
	m.handleResize()
}

// Init initializes the model
func (m *TUIModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
		m.TunnelsView.Init(),
		m.LogsView.Init(),
	)
}

// calculateFixedHeight returns the total height of fixed elements
func (m *TUIModel) calculateFixedHeight() int {
	return m.BannerHeight +
		m.StatusHeight +
		m.DividerHeight +
		m.HelpHeight +
		(m.SectionHeaderHeight * 2)
}

// calculateAvailableHeight returns height available for variable-sized views
func (m *TUIModel) calculateAvailableHeight() int {
	if m.Height == 0 {
		return 0
	}
	available := m.Height - m.calculateFixedHeight()
	if available < 10 {
		available = 10
	}
	return available
}

// ============ Update Logic ============

func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 只使用宽度，不使用终端高度
		m.Width = msg.Width
		m.handleResize()
		m.ready = true

		// 将窗口大小消息传递给视图
		_, tunnelsCmd := m.TunnelsView.Update(msg)
		_, logsCmd := m.LogsView.Update(msg)
		cmds = append(cmds, tunnelsCmd, logsCmd)

		return m, tea.Batch(cmds...)

	case tickMsg:
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}))

	case StatusUpdateMsg:
		m.Status = msg.Status
	case AddressUpdateMsg:
		m.RemoteAddress = msg.Address
	case LatencyUpdateMsg:
		m.Latency = msg.Latency

	case SetTotalHeightMsg:
		m.SetTotalHeight(msg.Height)
		return m, nil

	case FocusChangeMsg:
		m.ActiveView = msg.View
		m.TunnelsView.focused = msg.View == "tunnels"
		m.LogsView.focused = msg.View == "logs"

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.ActiveView == "tunnels" {
				m.ActiveView = "logs"
				m.TunnelsView.focused = false
				m.LogsView.focused = true
			} else {
				m.ActiveView = "tunnels"
				m.TunnelsView.focused = true
				m.LogsView.focused = false
			}
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		var handled bool
		if m.ActiveView == "tunnels" {
			handled = m.TunnelsView.HandleKeyMsg(msg)
			if !handled {
				_, cmd = m.TunnelsView.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			handled = m.LogsView.HandleKeyMsg(msg)
			if !handled {
				_, cmd = m.LogsView.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	default:
		_, tunnelsCmd := m.TunnelsView.Update(msg)
		_, logsCmd := m.LogsView.Update(msg)
		cmds = append(cmds, tunnelsCmd, logsCmd)
	}

	return m, tea.Batch(cmds...)
}

// ============ Layout Handling ============

// panelWidth returns the unified inner content width for all panels.
func (m *TUIModel) panelWidth() int {
	w := m.Width - 4
	if w < 60 {
		w = 60
	}
	return w
}

func (m *TUIModel) handleResize() {
	// 计算面板宽度
	_ = m.panelWidth() // 如果需要使用，可以取消注释

	// 计算可用高度
	availableHeight := m.calculateAvailableHeight()

	// 日志视图的首选高度
	logsPreferred := m.LogsView.GetHeight()

	// 确保日志高度不超过可用高度减去最小隧道高度
	if logsPreferred > availableHeight-MinTunnelsHeight {
		logsPreferred = availableHeight - MinTunnelsHeight
	}
	if logsPreferred < m.LogsView.GetMinHeight() {
		logsPreferred = m.LogsView.GetMinHeight()
	}

	// 隧道视图获取剩余高度
	tunnelsHeight := availableHeight - logsPreferred

	// 确保隧道高度在范围内
	if tunnelsHeight < m.TunnelsView.GetMinHeight() {
		tunnelsHeight = m.TunnelsView.GetMinHeight()
		logsPreferred = availableHeight - tunnelsHeight
	}
	if tunnelsHeight > m.TunnelsView.GetMaxHeight() {
		tunnelsHeight = m.TunnelsView.GetMaxHeight()
		logsPreferred = availableHeight - tunnelsHeight
	}

	// 设置最终高度
	m.TunnelsView.SetHeight(tunnelsHeight)
	m.LogsView.SetHeight(logsPreferred)
	m.ContentHeight = availableHeight
}

// ============ View Rendering ============

// View returns the rendered string
func (m *TUIModel) View() tea.View {
	if !m.ready || m.Width == 0 {
		m.MainView.Content = "Initializing......."
		return m.MainView
	}

	var sb strings.Builder
	pw := m.panelWidth() // ← 这里使用 panelWidth
	// Banner
	m.renderBanner(&sb)

	// Status panel
	m.renderStatus(&sb) // ← 传递 pw 参数

	// Tunnels View
	tunnelsView := m.TunnelsView.View(pw)
	sb.WriteString(tunnelsView)

	// Divider
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", m.Width-4)))
	sb.WriteString("\n")
	// Logs View
	logsView := m.LogsView.View(pw)
	sb.WriteString(logsView)
	// Help bar
	m.renderHelp(&sb)
	m.MainView.Content = sb.String()
	return m.MainView
}

func (m *TUIModel) renderBanner(sb *strings.Builder) {
	sb.WriteString("\n")
	sb.WriteString(statusPanelStyleTopBorder.Width(82).Render(bannerStyle.Render(bannerLines)))
	sb.WriteString("\n")
}

func (m *TUIModel) renderStatus(sb *strings.Builder) {
	var statusRendered string
	switch m.Status {
	case "online":
		statusRendered = statusOnlineStyle.Render("● ONLINE")
	case "offline":
		statusRendered = statusOfflineStyle.Render("○ OFFLINE")
	default:
		statusRendered = statusValueStyle.Render(strings.ToUpper(m.Status))
	}
	// Latency with color grading
	var latencyRendered string
	switch {
	case m.Latency == 0:
		latencyRendered = lipgloss.NewStyle().Foreground(colorDim).Render("— ms")
	case m.Latency < 80:
		latencyRendered = latencyGoodStyle.Render(fmt.Sprintf("%d ms", m.Latency))
	case m.Latency < 200:
		latencyRendered = latencyMidStyle.Render(fmt.Sprintf("%d ms", m.Latency))
	default:
		latencyRendered = latencyBadStyle.Render(fmt.Sprintf("%d ms", m.Latency))
	}
	var inner strings.Builder
	inner.WriteString(fmt.Sprintf("%s  %s\n",
		statusLabelStyle.Render("STATUS"),
		statusRendered,
	))
	inner.WriteString(fmt.Sprintf("%s  %s\n",
		statusLabelStyle.Render("REMOTE"),
		statusValueStyle.Render(m.RemoteAddress),
	))
	inner.WriteString(fmt.Sprintf("%s  %s",
		statusLabelStyle.Render("LATENCY"),
		latencyRendered,
	))
	panel := statusPanelStyle.Width(82).
		Render(inner.String())
	sb.WriteString(panel)
	sb.WriteString("\n")
}

func (m *TUIModel) renderHelp(sb *strings.Builder) {
	keys := []struct{ key, desc string }{
		{"↑/↓", "scroll"},
		{"tab", "switch view"},
		{"b", "log bottom"},
		{"+/-", "resize log"},
		{"q", "quit"},
	}

	var activeIndicator string
	if m.ActiveView == "tunnels" {
		activeIndicator = helpKeyStyle.Render("[Tunnels]") + " " + helpDescStyle.Render("active")
	} else {
		activeIndicator = helpKeyStyle.Render("[Logs]") + " " + helpDescStyle.Render("active")
	}

	var parts []string
	parts = append(parts, activeIndicator)

	for _, k := range keys {
		parts = append(parts,
			helpKeyStyle.Render(k.key)+" "+helpDescStyle.Render(k.desc),
		)
	}

	bar := helpStyle.Render(strings.Join(parts, helpDescStyle.Render("  ·  ")))
	sb.WriteString(bar)
}

// ============ Global Functions ============

var (
	globalProgram *tea.Program
	programMu     sync.RWMutex
)

func SetGlobalProgram(p *tea.Program) {
	programMu.Lock()
	defer programMu.Unlock()
	globalProgram = p
}

func getGlobalProgram() *tea.Program {
	programMu.RLock()
	defer programMu.RUnlock()
	return globalProgram
}

func UpdateStatus(status string) {
	if prog := getGlobalProgram(); prog != nil {
		prog.Send(StatusUpdateMsg{Status: status})
	}
}
func UpdateRemoteAddress(addr string) {
	if prog := getGlobalProgram(); prog != nil {
		prog.Send(AddressUpdateMsg{Address: addr})
	}
}

func UpdateConnections(addr string, port int, localAddr string, protocol string, isClose bool) {
	if prog := getGlobalProgram(); prog != nil {
		prog.Send(ConnUpdateMsg{
			Address:   addr,
			Port:      port,
			LocalAddr: localAddr,
			Protocol:  protocol,
			IsClose:   isClose,
		})
	}
}

func UpdateConnState(isClose bool) {
	if prog := getGlobalProgram(); prog != nil {
		prog.Send(ConnStateUpdateMsg{IsClose: isClose})
	}
}

func UpdateLatency(ms int64) {
	if prog := getGlobalProgram(); prog != nil {
		prog.Send(LatencyUpdateMsg{Latency: ms})
	}
}

func AddLog(format string, args ...interface{}) {
	if prog := getGlobalProgram(); prog != nil {
		msg := fmt.Sprintf(format, args...)
		prog.Send(LogMsg(msg))
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: TUI not initialized, log lost: %s\n",
			fmt.Sprintf(format, args...))
	}
}
