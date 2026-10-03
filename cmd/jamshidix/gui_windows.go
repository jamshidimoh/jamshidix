//go:build windows

package main

import (
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	guiClassName   = "JamshidixWindow"
	btnRefresh     = 1001
	btnConnect     = 1002
	btnDisconnect  = 1003
	cmbSort        = 1202
	cmbRegion      = 1203
	lstServers     = 1101
	lblStatus      = 1201
	msgRefreshDone = 0x8001
	msgActionDone  = 0x8002

	wsOverlapped   = 0x00CF0000
	wsVisible      = 0x10000000
	wsChild        = 0x40000000
	wsVScroll      = 0x00200000
	wsExClientEdge = 0x00000200
	lbNotify       = 0x00000001
	lbReset        = 0x0184
	lbAddString    = 0x0180
	lbGetCurSel    = 0x0188

	cbAddString    = 0x0143
	cbResetContent = 0x014B
	cbGetCurSel    = 0x0147
	cbGetLBText    = 0x0148
	cbSetCurSel    = 0x014E
	cbnSelChange   = 1
	cbDropDownList = 0x0003

	wmDestroy      = 0x0002
	wmCommand      = 0x0111
	wmSetFont      = 0x0030
	wmSize         = 0x0005
	bnClicked      = 0
	lbnDblClk      = 2
	swShow         = 5
	defaultGuiFont = 17
)

var (
	guiUser32      = syscall.NewLazyDLL("user32.dll")
	guiKernel32    = syscall.NewLazyDLL("kernel32.dll")
	registerClass  = guiUser32.NewProc("RegisterClassExW")
	createWindow   = guiUser32.NewProc("CreateWindowExW")
	defWindowProc  = guiUser32.NewProc("DefWindowProcW")
	showWindow     = guiUser32.NewProc("ShowWindow")
	updateWindow   = guiUser32.NewProc("UpdateWindow")
	getMessage     = guiUser32.NewProc("GetMessageW")
	translateMsg   = guiUser32.NewProc("TranslateMessage")
	dispatchMsg    = guiUser32.NewProc("DispatchMessageW")
	postQuit       = guiUser32.NewProc("PostQuitMessage")
	postMessage    = guiUser32.NewProc("PostMessageW")
	setWindowText = guiUser32.NewProc("SetWindowTextW")
	sendMessage   = guiUser32.NewProc("SendMessageW")
	setWindowPos  = guiUser32.NewProc("SetWindowPos")
	getStockObject = guiUser32.NewProc("GetStockObject")
	loadCursor    = guiUser32.NewProc("LoadCursorW")
	getModule     = guiKernel32.NewProc("GetModuleHandleW")
)

type guiState struct {
	hwnd        uintptr
	list        uintptr
	status      uintptr
	sortCombo   uintptr
	regionCombo uintptr
	refresh     uintptr
	connect     uintptr
	disconnect  uintptr
	allNodes    []Node
	nodes       []Node
	regions     []string
	sortMode    string
	regionMode  string
	busy        bool
	pending     Directory
	pendingMsg  string
}

var appGUI *guiState

type guiWndClass struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type guiMsg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

func guiUTF16(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func createControl(class, title string, style, exStyle uint32, x, y, w, h int32, id uintptr) uintptr {
	instance, _, _ := getModule.Call(0)
	r, _, _ := createWindow.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(guiUTF16(class))),
		uintptr(unsafe.Pointer(guiUTF16(title))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		appGUI.hwnd, id, instance, 0,
	)
	return r
}

func setGUIText(hwnd uintptr, text string) {
	setWindowText.Call(hwnd, uintptr(unsafe.Pointer(guiUTF16(text))))
}

func setGUIFont(hwnd uintptr) {
	font, _, _ := getStockObject.Call(defaultGuiFont)
	sendMessage.Call(hwnd, wmSetFont, font, 1)
}

func runGUI() error {
	if !isAdmin() {
		if err := relaunchElevatedArgs([]string{"--gui-elevated"}); err != nil {
			msg("اجرای Jamshidix نیازمند دسترسی Administrator است.\n\n"+err.Error(), mbError)
			return err
		}
		return nil
	}
	return runGUIElevated()
}

func runGUIElevated() error {
	d, err := loadDirectory()
	if err != nil {
		d = Directory{Version: 1}
	}
	appGUI = &guiState{
		allNodes:   append([]Node(nil), d.Nodes...),
		sortMode:   "اولویت",
		regionMode: "همه مناطق",
	}

	instance, _, _ := getModule.Call(0)
	cursor, _, _ := loadCursor.Call(0, 32512)
	className := guiUTF16(guiClassName)
	wc := guiWndClass{
		cbSize:        uint32(unsafe.Sizeof(guiWndClass{})),
		style:         0x0003,
		lpfnWndProc:   syscall.NewCallback(guiWndProc),
		hInstance:     instance,
		hCursor:       cursor,
		hbrBackground: 6,
		lpszClassName: className,
	}
	registerClass.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := createWindow.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(guiUTF16("Jamshidix — Free Tunnel"))),
		wsOverlapped,
		0x80000000, 0x80000000, 800, 600,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed")
	}
	appGUI.hwnd = hwnd
	createGUIControls()
	appGUI.applyView()
	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)

	go initialRefresh()
	go periodicRefresh()

	var m guiMsg
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) == -1 {
			return fmt.Errorf("GetMessageW failed")
		}
		if r == 0 {
			break
		}
		translateMsg.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMsg.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func createGUIControls() {
	appGUI.status = createControl("STATIC", "وضعیت: در حال راه‌اندازی…", wsChild|wsVisible, 0, 20, 18, 740, 26, lblStatus)
	appGUI.sortCombo = createControl("COMBOBOX", "", wsChild|wsVisible|wsVScroll|cbDropDownList, wsExClientEdge, 20, 50, 190, 28, cmbSort)
	appGUI.regionCombo = createControl("COMBOBOX", "", wsChild|wsVisible|wsVScroll|cbDropDownList, wsExClientEdge, 225, 50, 220, 28, cmbRegion)
	appGUI.list = createControl("LISTBOX", "", wsChild|wsVisible|wsVScroll|lbNotify, wsExClientEdge, 20, 92, 740, 410, lstServers)
	appGUI.refresh = createControl("BUTTON", "بروزرسانی سرورها", wsChild|wsVisible, 0, 20, 520, 180, 38, btnRefresh)
	appGUI.connect = createControl("BUTTON", "اتصال", wsChild|wsVisible, 0, 215, 520, 120, 38, btnConnect)
	appGUI.disconnect = createControl("BUTTON", "قطع اتصال", wsChild|wsVisible, 0, 350, 520, 120, 38, btnDisconnect)

	for _, h := range []uintptr{
		appGUI.status, appGUI.sortCombo, appGUI.regionCombo,
		appGUI.list, appGUI.refresh, appGUI.connect, appGUI.disconnect,
	} {
		setGUIFont(h)
	}
	populateSortCombo()
	populateRegionCombo()
	populateGUIList()
}

func (g *guiState) applyView() {
	if g == nil {
		return
	}
	d := Directory{Version: 1, Nodes: append([]Node(nil), g.allNodes...)}
	g.nodes = applyNodeView(d, g.sortMode, g.regionMode).Nodes
}

func populateSortCombo() {
	if appGUI == nil {
		return
	}
	sendMessage.Call(appGUI.sortCombo, cbResetContent, 0, 0)
	for _, mode := range []string{"اولویت", "سرعت", "منطقه"} {
		sendMessage.Call(appGUI.sortCombo, cbAddString, 0, uintptr(unsafe.Pointer(guiUTF16(mode))))
	}
	sendMessage.Call(appGUI.sortCombo, cbSetCurSel, 0, 0)
}

func populateRegionCombo() {
	if appGUI == nil {
		return
	}
	seen := map[string]bool{"همه مناطق": true}
	regions := []string{"همه مناطق"}
	for _, n := range appGUI.allNodes {
		r := nodeRegion(n)
		if !seen[r] {
			seen[r] = true
			regions = append(regions, r)
		}
	}
	appGUI.regions = regions
	sendMessage.Call(appGUI.regionCombo, cbResetContent, 0, 0)
	for _, r := range regions {
		sendMessage.Call(appGUI.regionCombo, cbAddString, 0, uintptr(unsafe.Pointer(guiUTF16(r))))
	}
	sendMessage.Call(appGUI.regionCombo, cbSetCurSel, 0, 0)
}

func comboText(hwnd uintptr) string {
	idx, _, _ := sendMessage.Call(hwnd, cbGetCurSel, 0, 0)
	if int(idx) < 0 {
		return ""
	}
	var buf [128]uint16
	sendMessage.Call(hwnd, cbGetLBText, idx, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf[:])
}

func applyGUIFilters() {
	if appGUI == nil || appGUI.busy {
		return
	}
	switch comboText(appGUI.sortCombo) {
	case "سرعت":
		appGUI.sortMode = "سرعت"
	case "منطقه":
		appGUI.sortMode = "منطقه"
		default:
		appGUI.sortMode = "اولویت"
	}
	r := comboText(appGUI.regionCombo)
	if r == "" {
		r = "همه مناطق"
	}
	appGUI.regionMode = r
	appGUI.applyView()
	populateGUIList()
}

func populateGUIList() {
	if appGUI == nil || appGUI.list == 0 {
		return
	}
	sendMessage.Call(appGUI.list, lbReset, 0, 0)
	for _, n := range appGUI.nodes {
		latency := "-"
		if n.LocalOK {
			latency = fmt.Sprintf("%dms", n.LocalLatencyMs)
		} else if n.LocalLatencyMs < 0 {
			latency = "محلی: مسدود/نامشخص"
		} else if n.RemoteLatencyMs > 0 {
			latency = fmt.Sprintf("remote %dms", n.RemoteLatencyMs)
		}
		label := fmt.Sprintf("%s | منطقه %s | %s | اولویت %d", n.Name, nodeRegion(n), latency, nodePriority(n))
		sendMessage.Call(appGUI.list, lbAddString, 0, uintptr(unsafe.Pointer(guiUTF16(label))))
	}
	if len(appGUI.nodes) == 0 {
		setGUIText(appGUI.status, "وضعیت: سروری مطابق فیلتر فعلی پیدا نشد.")
		return
	}
	setGUIText(appGUI.status, fmt.Sprintf(
		"وضعیت: قطع | %d سرور | مرتب‌سازی: %s | منطقه: %s",
		len(appGUI.nodes), comboText(appGUI.sortCombo), comboText(appGUI.regionCombo),
	))
}

func initialRefresh() {
	setGUIBusy(true)
	d, mirror, err := refreshDirectory()
	if err == nil {
		d = decorateLocalReachability(d)
		appGUI.pending = d
		appGUI.pendingMsg = fmt.Sprintf("فهرست بروزرسانی شد؛ mirror: %s", mirror)
	} else if cached, cacheErr := loadDirectory(); cacheErr == nil {
		cached = decorateLocalReachability(cached)
		appGUI.pending = cached
		appGUI.pendingMsg = "دریافت جدید ناموفق بود؛ آخرین فهرست سالم از cache استفاده شد."
	} else {
		appGUI.pending = Directory{Version: 1}
		appGUI.pendingMsg = "هیچ فهرست سروری در دسترس نیست."
	}
	postMessage.Call(appGUI.hwnd, msgRefreshDone, 0, 0)
}

func periodicRefresh() {
	for {
		time.Sleep(20 * time.Minute)
		if appGUI != nil && !appGUI.busy {
			go initialRefresh()
		}
	}
}

func setGUIBusy(b bool) {
	if appGUI != nil {
		appGUI.busy = b
	}
}

func selectedNode() (Node, error) {
	idx, _, _ := sendMessage.Call(appGUI.list, lbGetCurSel, 0, 0)
	if int(idx) < 0 || int(idx) >= len(appGUI.nodes) {
		return Node{}, fmt.Errorf("لطفاً یک سرور را انتخاب کنید.")
	}
	return appGUI.nodes[int(idx)], nil
}

func nodeToProfile(n Node) Profile {
	return Profile{
		Server: n.Server, Port: n.Port, UUID: n.UUID, PublicKey: n.PublicKey,
		ShortID: n.ShortID, SNI: n.SNI, Flow: n.Flow, Fingerprint: n.Fingerprint,
	}
}

func connectSelected() {
	if appGUI.busy {
		return
	}
	n, err := selectedNode()
	if err != nil {
		setGUIText(appGUI.status, "وضعیت: "+err.Error())
		return
	}
	setGUIBusy(true)
	setGUIText(appGUI.status, "وضعیت: در حال اتصال به "+n.Name+"…")
	go func() {
		err := connectNode(n)
		if err != nil {
			appGUI.pendingMsg = "اتصال ناموفق: " + err.Error()
		} else if ip := probeEgressIP(); ip != "" {
			appGUI.pendingMsg = "متصل شد | IP خروجی: " + ip
		} else {
			appGUI.pendingMsg = "متصل شد؛ IP خروجی قابل تأیید نبود."
		}
		postMessage.Call(appGUI.hwnd, msgActionDone, 0, 0)
	}()
}

func connectNode(n Node) error {
	if !existingSingBoxIsExpected() {
		if err := installSingBox(); err != nil {
			return err
		}
	}
	if ok, _ := probeNode(n); !ok {
		return fmt.Errorf("سرور از شبکه محلی قابل دسترسی نیست")
	}
	if err := saveProfile(nodeToProfile(n)); err != nil {
		return err
	}
	if err := checkConfigQuiet(); err != nil {
		return err
	}
	if err := stopSingBox(); err != nil {
		return err
	}
	if err := startDetached(); err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if isSingBoxRunning() && probeInternet() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("تونل برقرار نشد؛ لاگ: %s", logPath())
}

func disconnectSelected() {
	if appGUI.busy {
		return
	}
	setGUIBusy(true)
	go func() {
		err := stopSingBox()
		if err != nil {
			appGUI.pendingMsg = "قطع اتصال ناموفق: " + err.Error()
		} else {
			appGUI.pendingMsg = "اتصال قطع شد."
		}
		postMessage.Call(appGUI.hwnd, msgActionDone, 0, 0)
	}()
}

func probeEgressIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var b [128]byte
	n, _ := resp.Body.Read(b[:])
	return strings.TrimSpace(string(b[:n]))
}

func guiWndProc(hwnd uintptr, m uint32, wParam, lParam uintptr) uintptr {
	switch m {
	case wmSize:
		width := int32(uint16(lParam))
		height := int32(uint16(lParam >> 16))
		if appGUI != nil {
			setWindowPos.Call(appGUI.status, 0, 20, 18, uintptr(width-40), 26, 0)
			setWindowPos.Call(appGUI.sortCombo, 0, 20, 50, 190, 28, 0)
			setWindowPos.Call(appGUI.regionCombo, 0, 225, 50, 220, 28, 0)
			setWindowPos.Call(appGUI.list, 0, 20, 92, uintptr(width-40), uintptr(height-165), 0)
			setWindowPos.Call(appGUI.refresh, 0, 20, uintptr(height-72), 180, 38, 0)
			setWindowPos.Call(appGUI.connect, 0, 215, uintptr(height-72), 120, 38, 0)
			setWindowPos.Call(appGUI.disconnect, 0, 350, uintptr(height-72), 120, 38, 0)
		}
	case wmCommand:
		id := uint16(wParam)
		notify := uint16(wParam >> 16)
		switch uintptr(id) {
		case cmbSort, cmbRegion:
			if notify == cbnSelChange {
				applyGUIFilters()
			}
		case btnRefresh:
			if notify == bnClicked && !appGUI.busy {
				go initialRefresh()
			}
		case btnConnect:
			if notify == bnClicked {
				connectSelected()
			}
		case btnDisconnect:
			if notify == bnClicked {
				disconnectSelected()
			}
		case lstServers:
			if notify == lbnDblClk && !appGUI.busy {
				connectSelected()
			}
		}
	case msgRefreshDone:
		appGUI.allNodes = append([]Node(nil), appGUI.pending.Nodes...)
		populateRegionCombo()
		appGUI.applyView()
		appGUI.busy = false
		populateGUIList()
		setGUIText(appGUI.status, "وضعیت: "+appGUI.pendingMsg)
	case msgActionDone:
		appGUI.busy = false
		setGUIText(appGUI.status, "وضعیت: "+appGUI.pendingMsg)
	case wmDestroy:
		postQuit.Call(0)
		return 0
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(m), wParam, lParam)
	return result
}
