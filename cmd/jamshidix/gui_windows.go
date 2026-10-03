//go:build windows

package main

// GUI source for the desktop client.
import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	guiClassName = "JamshidixWindow"
	btnRefresh    = 1001
	btnConnect    = 1002
	btnDisconnect = 1003
	cboSort        = 1004
	cboRegion      = 1005
	lstServers     = 1101
	lblStatus      = 1201
	msgRefreshDone uint32 = 0x8001
	msgActionDone  uint32 = 0x8002
	cboSelChange   uint16 = 1
	cbAddString    uint32 = 0x0143
	cbResetContent uint32 = 0x014B
	cbGetCurSel   uint32 = 0x0147
	cbSetCurSel   uint32 = 0x014E

	wsOverlapped   = 0x00CF0000
	wsVisible      = 0x10000000
	wsChild        = 0x40000000
	wsVScroll      = 0x00200000
	wsHScroll      = 0x00100000
	wsExClientEdge = 0x00000200
	lbNotify       = 0x00000001
	lbReset        = 0x0184
	lbAddString    = 0x0180
	lbGetCurSel    = 0x0188
	cbAddString    = 0x0143
	cbResetContent = 0x014B
	cbGetCurSel    = 0x0147
	cbSetCurSel    = 0x014E
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
	guiUser32 = syscall.NewLazyDLL("user32.dll")
	guiKernel32 = syscall.NewLazyDLL("kernel32.dll")
	registerClassExW = guiUser32.NewProc("RegisterClassExW")
	createWindowExW = guiUser32.NewProc("CreateWindowExW")
	defWindowProcW = guiUser32.NewProc("DefWindowProcW")
	showWindow = guiUser32.NewProc("ShowWindow")
	updateWindow = guiUser32.NewProc("UpdateWindow")
	getMessageW = guiUser32.NewProc("GetMessageW")
	translateMessage = guiUser32.NewProc("TranslateMessage")
	dispatchMessageW = guiUser32.NewProc("DispatchMessageW")
	destroyWindow = guiUser32.NewProc("DestroyWindow")
	postQuitMessage = guiUser32.NewProc("PostQuitMessage")
	postMessageW = guiUser32.NewProc("PostMessageW")
	setWindowTextW = guiUser32.NewProc("SetWindowTextW")
	sendMessageW = guiUser32.NewProc("SendMessageW")
	setWindowPos = guiUser32.NewProc("SetWindowPos")
	getStockObject = guiUser32.NewProc("GetStockObject")
	loadCursorW = guiUser32.NewProc("LoadCursorW")
	getModuleHandleW = guiKernel32.NewProc("GetModuleHandleW")
)

type guiState struct {
	hwnd       uintptr
	list       uintptr
	status     uintptr
	refresh    uintptr
	connect    uintptr
	disconnect uintptr
	nodes      []Node
	allNodes   []Node
	regions    []string
	sortMode   string
	regionMode string
	busy       bool
	pending    Directory
	pendingMsg string
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

func utf16(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func createControl(class, title string, style uint32, exStyle uint32, x, y, w, h int32, id uintptr) uintptr {
	instance, _, _ := getModuleHandleW.Call(0)
	r, _, _ := createWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16(class))),
		uintptr(unsafe.Pointer(utf16(title))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		appGUI.hwnd, id, instance, 0,
	)
	return r
}

func setText(hwnd uintptr, text string) {
	setWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16(text))))
}

func setGuiFont(hwnd uintptr) {
	font, _, _ := getStockObject.Call(defaultGuiFont)
	sendMessageW.Call(hwnd, wmSetFont, font, 1)
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
	appGUI = &guiState{allNodes: d.Nodes, sortMode: "اولویت", regionMode: "همه مناطق"}
	appGUI.applyView()
	instance, _, _ := getModuleHandleW.Call(0)
	cursor, _, _ := loadCursorW.Call(0, 32512)
	className := utf16(guiClassName)
	wc := guiWndClass{
		cbSize:        uint32(unsafe.Sizeof(guiWndClass{})),
		style:         0x0003,
		lpfnWndProc:   syscall.NewCallback(guiWndProc),
		hInstance:     instance,
		hCursor:       cursor,
		hbrBackground: 6,
		lpszClassName: className,
	}
	if r, _, _ := registerClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		// Registration may already exist after a second launch.
	}
	hwnd, _, _ := createWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16("Jamshidix — Free Tunnel"))),
		wsOverlapped,
		0x80000000, 0x80000000, 760, 560,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed")
	}
	appGUI.hwnd = hwnd
	createGUIControls()
	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)

	go initialRefresh()
	go periodicRefresh()
	var m guiMsg
	for {
		r, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) == -1 {
			return fmt.Errorf("GetMessageW failed")
		}
		if r == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func createGUIControls() {
	appGUI.status = createControl("STATIC", "وضعیت: در حال راه‌اندازی…", wsChild|wsVisible, 0, 20, 18, 700, 26, lblStatus)
	appGUI.sort = createControl("COMBOBOX", "", wsChild|wsVisible|wsVScroll, wsExClientEdge, 20, 50, 190, 28, cboSort)
	appGUI.region = createControl("COMBOBOX", "", wsChild|wsVisible|wsVScroll, wsExClientEdge, 225, 50, 220, 28, cboRegion)
	appGUI.list = createControl("LISTBOX", "", wsChild|wsVisible|wsVScroll|lbNotify, wsExClientEdge, 20, 92, 700, 370, lstServers)
	appGUI.refresh = createControl("BUTTON", "بروزرسانی سرورها", wsChild|wsVisible, 0, 20, 475, 170, 38, btnRefresh)
	appGUI.connect = createControl("BUTTON", "اتصال", wsChild|wsVisible, 0, 205, 475, 120, 38, btnConnect)
	appGUI.disconnect = createControl("BUTTON", "قطع اتصال", wsChild|wsVisible, 340, 475, 120, 38, btnDisconnect)
	setGuiFont(appGUI.status)
	setGuiFont(appGUI.sort)
	setGuiFont(appGUI.region)
	setGuiFont(appGUI.list)
	setGuiFont(appGUI.refresh)
	setGuiFont(appGUI.connect)
	setGuiFont(appGUI.disconnect)
	populateRegionCombo()
	populateSortCombo()
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
	if appGUI == nil || appGUI.sort == 0 { return }
	sendMessageW.Call(appGUI.sort, cbResetContent, 0, 0)
	for _, mode := range []string{"اولویت", "سرعت", "منطقه", "تازگی"} {
		sendMessageW.Call(appGUI.sort, cbAddString, 0, uintptr(unsafe.Pointer(utf16(mode))))
	}
	sendMessageW.Call(appGUI.sort, cbSetCurSel, 0, 0)
}

func populateRegionCombo() {
	if appGUI == nil || appGUI.region == 0 { return }
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
	sendMessageW.Call(appGUI.region, cbResetContent, 0, 0)
	for _, r := range regions {
		sendMessageW.Call(appGUI.region, cbAddString, 0, uintptr(unsafe.Pointer(utf16(r))))
	}
	sendMessageW.Call(appGUI.region, cbSetCurSel, 0, 0)
}

func comboIndex(hwnd uintptr) int {
	idx, _, _ := sendMessageW.Call(hwnd, cbGetCurSel, 0, 0)
	if int(idx) < 0 { return 0 }
	return int(idx)
}

func applyComboView() {
	if appGUI == nil { return }
	si := comboIndex(appGUI.sort)
	if si >= 0 && si < 4 {
		appGUI.sortMode = []string{"اولویت", "سرعت", "منطقه", "تازگی"}[si]
	}
	ri := comboIndex(appGUI.region)
	if ri >= 0 && ri < len(appGUI.regions) {
		appGUI.regionMode = appGUI.regions[ri]
	}
	appGUI.applyView()
	populateGUIList()
}

func populateGUIList() {
	if appGUI == nil || appGUI.list == 0 {
		return
	}
	sendMessageW.Call(appGUI.list, lbReset, 0, 0)
	for _, n := range appGUI.nodes {
		status := "تست‌نشده"
		if n.LocalOK {
			status = fmt.Sprintf("%dms", n.LocalLatencyMs)
		} else if n.LocalLatencyMs < 0 {
			status = "دسترسی محلی؟"
		}
		region := nodeRegion(n)
		label := fmt.Sprintf("%s | منطقه %s | %s | اولویت %d", n.Name, region, status, n.Priority)
		sendMessageW.Call(appGUI.list, lbAddString, 0, uintptr(unsafe.Pointer(utf16(label))))
	}
	if len(appGUI.nodes) == 0 {
		setText(appGUI.status, "وضعیت: فهرست سرور در دسترس نیست؛ روی «بروزرسانی سرورها» بزنید.")
	} else {
		setText(appGUI.status, fmt.Sprintf("وضعیت: قطع | %d سرور | مرتب‌سازی: %s | منطقه: %s", len(appGUI.nodes), appGUI.sortMode, appGUI.regionMode))
	}
}

func countryText(n Node) string { return nodeRegion(n) }

func speedText(n Node) string {
	lat := n.LocalLatencyMs
	if lat <= 0 { lat = n.RemoteLatencyMs }
	if lat <= 0 { return "-" }
	return fmt.Sprintf("%dms", lat)
}

func initialRefresh() {
	setGUIBusy(true)
	d, mirror, err := refreshDirectory()
	if err == nil {
		d = decorateLocalReachability(d)
		appGUI.pending = d
		appGUI.allNodes = d.Nodes
		appGUI.pendingMsg = fmt.Sprintf("فهرست به‌روزرسانی شد؛ منبع: %s", mirror)
	} else {
		if cached, e := loadDirectory(); e == nil {
			cached = decorateLocalReachability(cached)
			appGUI.pending = cached
			appGUI.allNodes = cached.Nodes
			appGUI.pendingMsg = "به‌روزرسانی ناموفق بود؛ آخرین فهرست ذخیره‌شده استفاده شد."
		} else {
			appGUI.pendingMsg = "فهرست سرورها قابل دریافت نیست."
		}
	}
	postMessageW.Call(appGUI.hwnd, msgRefreshDone, 0, 0)
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
	if appGUI == nil {
		return
	}
	appGUI.busy = b
	// Button state is left enabled; the click handlers themselves are idempotent.
}

func selectedNode() (Node, error) {
	idx, _, _ := sendMessageW.Call(appGUI.list, lbGetCurSel, 0, 0)
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
		setText(appGUI.status, "وضعیت: "+err.Error())
		return
	}
	setGUIBusy(true)
	setText(appGUI.status, "وضعیت: در حال اتصال به "+n.Name+"…")
	go func() {
		err := connectNode(n)
		if err != nil {
			appGUI.pendingMsg = "اتصال ناموفق: " + err.Error()
		} else {
			ip := probeEgressIP()
			if ip != "" {
				appGUI.pendingMsg = "متصل شد | IP خروجی: " + ip
			} else {
				appGUI.pendingMsg = "متصل شد، اما IP خروجی قابل تأیید نبود."
			}
		}
		postMessageW.Call(appGUI.hwnd, msgActionDone, 0, 0)
	}()
}

func connectNode(n Node) error {
	if !existingSingBoxIsExpected() {
		if err := installSingBox(); err != nil {
			return err
		}
	}
	p := nodeToProfile(n)
	if err := saveProfile(p); err != nil {
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
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if isSingBoxRunning() {
			if probeInternet() {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("تونل بالا نیامد؛ لاگ: %s", logPath())
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
		postMessageW.Call(appGUI.hwnd, msgActionDone, 0, 0)
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
	b, _ := ioReadLimited(resp, 128)
	s := strings.TrimSpace(string(b))
	return s
}

func ioReadLimited(resp *http.Response, max int64) ([]byte, error) {
	b := make([]byte, 0, max)
	buf := make([]byte, 64)
	for int64(len(b)) < max {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			remaining := int(max) - len(b)
			if n > remaining {
				n = remaining
			}
			b = append(b, buf[:n]...)
		}
		if err != nil {
			if len(b) > 0 {
				return b, nil
			}
			return b, err
		}
	}
	return b, nil
}

func guiWndProc(hwnd uintptr, m uint32, wParam, lParam uintptr) uintptr {
	switch m {
	case wmSize:
		width := int32(uint16(lParam))
		height := int32(uint16(lParam >> 16))
		if appGUI != nil {
			setWindowPos.Call(appGUI.status, 0, 20, 18, uintptr(width-40), 26, 0)
			setWindowPos.Call(appGUI.sort, 0, 20, 50, 190, 28, 0)
			setWindowPos.Call(appGUI.region, 0, 225, 50, 220, 28, 0)
			setWindowPos.Call(appGUI.list, 0, 20, 92, uintptr(width-40), uintptr(height-165), 0)
			setWindowPos.Call(appGUI.refresh, 0, 20, uintptr(height-72), 170, 38, 0)
			setWindowPos.Call(appGUI.connect, 0, 205, uintptr(height-72), 120, 38, 0)
			setWindowPos.Call(appGUI.disconnect, 0, 340, uintptr(height-72), 120, 38, 0)
		}
	case wmCommand:
		id := uint16(wParam)
		notify := uint16(wParam >> 16)
		switch uintptr(id) {
		case btnRefresh:
			if notify == bnClicked && !appGUI.busy { go initialRefresh() }
		case cboSort, cboRegion:
			if notify == cboSelChange && !appGUI.busy { applyComboView() }
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
		appGUI.allNodes = appGUI.pending.Nodes
		populateRegionCombo()
		appGUI.regionMode = "همه مناطق"
		appGUI.applyView()
		appGUI.busy = false
		populateGUIList()
		setText(appGUI.status, "وضعیت: "+appGUI.pendingMsg)
	case msgActionDone:
		appGUI.busy = false
		setText(appGUI.status, "وضعیت: "+appGUI.pendingMsg)
	case wmDestroy:
		if appGUI != nil {
			_ = stopSingBox()
		}
		postQuitMessage.Call(0)
		return 0
	}
	returnValue, _, _ := defWindowProcW.Call(hwnd, uintptr(m), wParam, lParam)
	return returnValue
}

