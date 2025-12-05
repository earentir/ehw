package main

import (
	"fmt"
	"os"
	"os/signal"
	"retrotui"
	"sort"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/cobra"
)

type Page int

const (
	PageSummary Page = iota
	PageCPU
)

type App struct {
	hwInfo      *HardwareInfo
	currentPage Page
	screen      tcell.Screen
	done        chan bool
	scrollY     int // Scroll offset for current page
}

var (
	styleNormal  = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	styleReverse = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	styleTitle   = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)
	styleSection = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	styleBorder  = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
)

func runTUI(cmd *cobra.Command, args []string) {
	// Collect hardware info
	hwInfo, err := CollectHardwareInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting hardware info: %v\n", err)
		os.Exit(1)
	}

	// Initialize screen
	screen, err := retrotui.InitScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing screen: %v\n", err)
		os.Exit(1)
	}
	defer retrotui.ExitProgram(screen)

	// Enable mouse support
	screen.EnableMouse()

	app := &App{
		hwInfo:      hwInfo,
		currentPage: PageSummary,
		screen:      screen,
		done:        make(chan bool),
		scrollY:     0,
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		app.done <- true
	}()

	// Main event loop
	go app.eventLoop()

	app.render()

	<-app.done
}

func (app *App) eventLoop() {
	for {
		ev := app.screen.PollEvent()
		if ev == nil {
			break
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				app.done <- true
				return
			case tcell.KeyLeft:
				app.prevPage()
				app.scrollY = 0 // Reset scroll when changing pages
				app.render()
			case tcell.KeyRight:
				app.nextPage()
				app.scrollY = 0 // Reset scroll when changing pages
				app.render()
			case tcell.KeyUp:
				if app.scrollY > 0 {
					app.scrollY--
					app.render()
				}
			case tcell.KeyDown:
				app.scrollY++
				app.render()
			case tcell.KeyRune:
				if ev.Rune() == 'q' || ev.Rune() == 'Q' {
					app.done <- true
					return
				}
			}
		case *tcell.EventMouse:
			app.handleMouse(ev)
		case *tcell.EventResize:
			app.render()
		}
	}
}

func (app *App) handleMouse(ev *tcell.EventMouse) {
	width, height := app.screen.Size()
	mx, my := ev.Position()
	buttons := ev.Buttons()

	// Handle mouse wheel scrolling
	if buttons&tcell.WheelUp != 0 {
		if app.scrollY > 0 {
			app.scrollY--
			app.render()
		}
		return
	}
	if buttons&tcell.WheelDown != 0 {
		app.scrollY++
		app.render()
		return
	}

	// Handle mouse clicks on menu items (menu is inside border)
	if buttons&tcell.Button1 != 0 && (my == height-2 || my == height-3) {
		// Clicked on menu bar
		menuItems := []string{"Summary", "CPU"}
		menuWidth := 0
		for _, item := range menuItems {
			menuWidth += len(item) + 3
		}
		menuWidth -= 1
		startX := (width - menuWidth) / 2
		if startX < 2 {
			startX = 2
		}

		x := startX
		for i, item := range menuItems {
			itemWidth := len(item) + 3
			if mx >= x && mx < x+itemWidth {
				app.currentPage = Page(i)
				app.scrollY = 0
				app.render()
				return
			}
			x += itemWidth
		}
	}
}

func (app *App) nextPage() {
	totalPages := 2 // Summary, CPU
	app.currentPage = (app.currentPage + 1) % Page(totalPages)
}

func (app *App) prevPage() {
	totalPages := 2
	app.currentPage = (app.currentPage - 1 + Page(totalPages)) % Page(totalPages)
}

func (app *App) render() {
	app.screen.Clear()

	// Get terminal dimensions
	width, height := app.screen.Size()

	// Fill background with black
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			app.screen.SetContent(x, y, ' ', nil, styleNormal)
		}
	}

	// Draw border around the app (includes title in top border)
	app.drawBorder(width, height)

	// Render content based on current page
	switch app.currentPage {
	case PageSummary:
		app.renderSummary(width, height)
	case PageCPU:
		app.renderCPU(width, height)
	}

	// Render menu at bottom (last line)
	app.renderMenu(width, height)

	app.screen.Show()
}

func (app *App) drawBorder(width, height int) {
	// Box drawing characters (single line)
	topLeft := '┌'
	topRight := '┐'
	bottomLeft := '└'
	bottomRight := '┘'
	horizontal := '─'
	vertical := '│'
	doubleHorizontal := '═'

	// Get title for top border
	titles := map[Page]string{
		PageSummary: "HARDWARE SUMMARY",
		PageCPU:     "CPU INFORMATION",
	}
	title := titles[app.currentPage]
	if title == "" {
		title = "HARDWARE INFORMATION"
	}
	titlePart := "[ " + title + " ]"

	// Top border with integrated title
	app.screen.SetContent(0, 0, topLeft, nil, styleBorder)

	// Calculate title position (centered)
	titleStart := (width - len(titlePart)) / 2
	titleEnd := titleStart + len(titlePart)

	for x := 1; x < width-1; x++ {
		if x >= titleStart && x < titleEnd {
			// Draw title character - brackets in white, title text in yellow
			idx := x - titleStart
			ch := rune(titlePart[idx])
			if ch == '[' || ch == ']' || ch == ' ' {
				app.screen.SetContent(x, 0, ch, nil, styleBorder)
			} else {
				app.screen.SetContent(x, 0, ch, nil, styleTitle)
			}
		} else {
			// Draw double horizontal line around title area (white)
			app.screen.SetContent(x, 0, doubleHorizontal, nil, styleBorder)
		}
	}
	app.screen.SetContent(width-1, 0, topRight, nil, styleBorder)

	// Bottom border
	app.screen.SetContent(0, height-1, bottomLeft, nil, styleBorder)
	for x := 1; x < width-1; x++ {
		app.screen.SetContent(x, height-1, horizontal, nil, styleBorder)
	}
	app.screen.SetContent(width-1, height-1, bottomRight, nil, styleBorder)

	// Left and right borders
	for y := 1; y < height-1; y++ {
		app.screen.SetContent(0, y, vertical, nil, styleBorder)
		app.screen.SetContent(width-1, y, vertical, nil, styleBorder)
	}
}

func (app *App) renderSectionTitle(x, y, width int, title string) {
	if y < 1 {
		return
	}

	// Draw: ───[ Title ]───
	pos := x

	// Draw leading dashes
	for i := 0; i < 3; i++ {
		app.screen.SetContent(pos, y, '─', nil, styleSection)
		pos++
	}

	// Draw [ title ]
	titlePart := "[ " + title + " ]"
	for _, ch := range titlePart {
		app.screen.SetContent(pos, y, ch, nil, styleSection)
		pos++
	}

	// Draw trailing dashes
	for i := 0; i < 3; i++ {
		app.screen.SetContent(pos, y, '─', nil, styleSection)
		pos++
	}
}

func (app *App) renderMenu(width, height int) {
	menuY := height - 2 // Inside border
	menuItems := []string{"Summary", "CPU"}

	// Calculate menu width
	menuWidth := 0
	for _, item := range menuItems {
		menuWidth += len(item) + 3
	}
	menuWidth -= 1

	startX := (width - menuWidth) / 2
	if startX < 2 {
		startX = 2
	}

	x := startX
	for i, item := range menuItems {
		selected := int(app.currentPage) == i
		style := styleNormal
		if selected {
			style = styleReverse
		}
		retrotui.PrintAt(app.screen, x, menuY, fmt.Sprintf("[%s]", item), style)
		x += len(item) + 3
	}

	// Instructions on line above menu
	instructions := "← → Navigate | ↑ ↓ Scroll | Mouse: Click/Wheel | Q Quit"
	instX := (width - len(instructions)) / 2
	if instX < 2 {
		instX = 2
	}
	retrotui.PrintAt(app.screen, instX, menuY-1, instructions, styleNormal)
}

func (app *App) renderSummary(width, height int) {
	y := 2 - app.scrollY // Start below title border
	x := 3
	contentHeight := height - 4 // Account for border and menu

	// CPU Summary
	if y >= 2 && y < contentHeight {
		app.renderSectionTitle(x, y, width, "CPU")
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Vendor:     %s", app.hwInfo.CPU.Vendor), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Brand:      %s", truncateString(app.hwInfo.CPU.Brand, width-20)), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Cores:      %d", app.hwInfo.CPU.Cores), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Threads:    %d", app.hwInfo.CPU.Threads), styleNormal)
	}
	y += 2

	// Feature count summary
	if y >= 2 && y < contentHeight {
		app.renderSectionTitle(x, y, width, "Features")
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Total Features: %d", len(app.hwInfo.CPU.Features)), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Categories:     %d", len(app.hwInfo.CPU.FeatureCategories)), styleNormal)
	}
	y += 2

	// Cache summary
	if len(app.hwInfo.CPU.CacheDetails) > 0 {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, "Cache")
		}
		y++
		for _, cache := range app.hwInfo.CPU.CacheDetails {
			if y >= contentHeight {
				break
			}
			if y >= 2 {
				retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("L%d %s: %d KB", cache.Level, cache.Type, cache.SizeKB), styleNormal)
			}
			y++
		}
	}
}

func (app *App) renderCPU(width, height int) {
	y := 2 - app.scrollY // Start below title border
	x := 3
	contentHeight := height - 4 // Account for border and menu

	// Basic Info
	if y >= 2 && y < contentHeight {
		app.renderSectionTitle(x, y, width, "Basic Information")
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Vendor:        %s", app.hwInfo.CPU.Vendor), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Brand:         %s", truncateString(app.hwInfo.CPU.Brand, width-25)), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Model:         %s", app.hwInfo.CPU.Model), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Family:        %d", app.hwInfo.CPU.Family), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Model Number:  %d", app.hwInfo.CPU.ModelNumber), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Stepping:      %d", app.hwInfo.CPU.Stepping), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Cores:         %d", app.hwInfo.CPU.Cores), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Threads:       %d", app.hwInfo.CPU.Threads), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Max Func:      %d", app.hwInfo.CPU.MaxFunc), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Max Ext Func:  %d", app.hwInfo.CPU.MaxExtFunc), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Phys Addr Bits: %d", app.hwInfo.CPU.PhysicalAddrBits), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Linear Addr Bits: %d", app.hwInfo.CPU.LinearAddrBits), styleNormal)
	}
	y += 2

	// Processor Info Details
	if y >= 2 && y < contentHeight {
		app.renderSectionTitle(x, y, width, "Processor Details")
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Max Logical Processors: %d", app.hwInfo.CPU.ProcessorInfo.MaxLogicalProcessors), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Initial APIC ID: %d", app.hwInfo.CPU.ProcessorInfo.InitialAPICID), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Threads Per Core: %d", app.hwInfo.CPU.ProcessorInfo.ThreadPerCore), styleNormal)
	}
	y += 2

	// Model Data Details
	if y >= 2 && y < contentHeight {
		app.renderSectionTitle(x, y, width, "Model Data")
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Stepping ID: %d | Model ID: %d | Family ID: %d",
			app.hwInfo.CPU.ModelData.SteppingID, app.hwInfo.CPU.ModelData.ModelID, app.hwInfo.CPU.ModelData.FamilyID), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Extended Model: %d | Extended Family: %d",
			app.hwInfo.CPU.ModelData.ExtendedModel, app.hwInfo.CPU.ModelData.ExtendedFamily), styleNormal)
	}
	y++
	if y >= 2 && y < contentHeight {
		retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Processor Type: %d", app.hwInfo.CPU.ModelData.ProcessorType), styleNormal)
	}
	y += 2

	// Hybrid Info
	if app.hwInfo.CPU.HybridInfo.IsHybrid {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, "Hybrid CPU Information")
		}
		y++
		if y >= 2 && y < contentHeight {
			retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("Core Type: %s", app.hwInfo.CPU.HybridInfo.CoreType), styleNormal)
		}
		y += 2
	}

	// Detailed Cache Info
	if len(app.hwInfo.CPU.CacheDetails) > 0 {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, "Detailed Cache Information")
		}
		y++
		for _, cache := range app.hwInfo.CPU.CacheDetails {
			if y >= contentHeight {
				break
			}
			if y < 2 {
				y++
				continue
			}
			retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("L%d %s: %d KB, %d-way, %d bytes/line, %d sets",
				cache.Level, cache.Type, cache.SizeKB, cache.Ways, cache.LineSizeBytes, cache.TotalSets), styleNormal)
			y++
			if y >= contentHeight {
				break
			}
			if y >= 2 {
				retrotui.PrintAt(app.screen, x+8, y, fmt.Sprintf("Max Cores Sharing: %d | Max Processor IDs: %d",
					cache.MaxCoresSharing, cache.MaxProcessorIDs), styleNormal)
			}
			y++
			if y >= 2 && y < contentHeight {
				retrotui.PrintAt(app.screen, x+8, y, fmt.Sprintf("Write Policy: %s | Self-Init: %v | Fully Assoc: %v",
					cache.WritePolicy, cache.SelfInitializing, cache.FullyAssociative), styleNormal)
			}
			y++
		}
		y += 2
	}

	// TLB Info
	if len(app.hwInfo.CPU.TLBInfo.L1Data) > 0 || len(app.hwInfo.CPU.TLBInfo.L1Inst) > 0 || len(app.hwInfo.CPU.TLBInfo.L2Unified) > 0 {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, "TLB (Translation Lookaside Buffer)")
		}
		y++
		if len(app.hwInfo.CPU.TLBInfo.L1Data) > 0 {
			if y >= 2 && y < contentHeight {
				retrotui.PrintAt(app.screen, x+4, y, "L1 Data TLB:", styleSection)
			}
			y++
			for _, tlb := range app.hwInfo.CPU.TLBInfo.L1Data {
				if y >= contentHeight {
					break
				}
				if y >= 2 {
					retrotui.PrintAt(app.screen, x+8, y, fmt.Sprintf("%s: %d entries, %s associativity",
						tlb.PageSize, tlb.Entries, tlb.Associativity), styleNormal)
				}
				y++
			}
		}
		if len(app.hwInfo.CPU.TLBInfo.L1Inst) > 0 {
			if y >= 2 && y < contentHeight {
				retrotui.PrintAt(app.screen, x+4, y, "L1 Instruction TLB:", styleSection)
			}
			y++
			for _, tlb := range app.hwInfo.CPU.TLBInfo.L1Inst {
				if y >= contentHeight {
					break
				}
				if y >= 2 {
					retrotui.PrintAt(app.screen, x+8, y, fmt.Sprintf("%s: %d entries, %s associativity",
						tlb.PageSize, tlb.Entries, tlb.Associativity), styleNormal)
				}
				y++
			}
		}
		if len(app.hwInfo.CPU.TLBInfo.L2Unified) > 0 {
			if y >= 2 && y < contentHeight {
				retrotui.PrintAt(app.screen, x+4, y, "L2 Unified TLB:", styleSection)
			}
			y++
			for _, tlb := range app.hwInfo.CPU.TLBInfo.L2Unified {
				if y >= contentHeight {
					break
				}
				if y >= 2 {
					retrotui.PrintAt(app.screen, x+8, y, fmt.Sprintf("%s: %d entries, %s associativity",
						tlb.PageSize, tlb.Entries, tlb.Associativity), styleNormal)
				}
				y++
			}
		}
		y += 2
	}

	// Detailed Feature Categories - displayed in columns
	if len(app.hwInfo.CPU.FeatureCategories) > 0 {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, "Supported Features by Category")
		}
		y++

		// Sort category names for consistent ordering
		categoryNames := make([]string, 0, len(app.hwInfo.CPU.FeatureCategories))
		for category := range app.hwInfo.CPU.FeatureCategories {
			categoryNames = append(categoryNames, category)
		}
		sort.Strings(categoryNames)

		for _, category := range categoryNames {
			features := app.hwInfo.CPU.FeatureCategories[category]
			if y >= contentHeight {
				break
			}
			if y >= 2 {
				retrotui.PrintAt(app.screen, x+4, y, fmt.Sprintf("▸ %s (%d features)", category, len(features)), styleSection)
			}
			y++

			// Calculate column layout - max 4 columns, 30 chars wide
			colWidth := 30
			numCols := (width - x - 8) / colWidth
			if numCols < 1 {
				numCols = 1
			}
			if numCols > 4 {
				numCols = 4
			}

			// Calculate how many rows we need
			numRows := (len(features) + numCols - 1) / numCols

			// Display features in columns row by row
			for row := 0; row < numRows; row++ {
				if y >= contentHeight {
					break
				}
				if y >= 2 {
					for col := 0; col < numCols; col++ {
						idx := row*numCols + col
						if idx < len(features) {
							colX := x + 8 + (col * colWidth)
							retrotui.PrintAt(app.screen, colX, y, truncateString(features[idx].Name, colWidth-2), styleNormal)
						}
					}
				}
				y++
			}
			y++
		}
		y += 2
	}

	// All Features (displayed in columns)
	if len(app.hwInfo.CPU.Features) > 0 {
		if y >= 2 && y < contentHeight {
			app.renderSectionTitle(x, y, width, fmt.Sprintf("All Supported Features (%d total)", len(app.hwInfo.CPU.Features)))
		}
		y++

		// Calculate column layout - max 4 columns, 30 chars wide
		colWidth := 30
		numCols := (width - x - 4) / colWidth
		if numCols < 1 {
			numCols = 1
		}
		if numCols > 4 {
			numCols = 4
		}

		// Calculate how many rows we need
		numRows := (len(app.hwInfo.CPU.Features) + numCols - 1) / numCols

		// Display features in columns row by row
		for row := 0; row < numRows; row++ {
			if y >= contentHeight {
				break
			}
			if y >= 2 {
				for col := 0; col < numCols; col++ {
					idx := row*numCols + col
					if idx < len(app.hwInfo.CPU.Features) {
						colX := x + 4 + (col * colWidth)
						retrotui.PrintAt(app.screen, colX, y, truncateString(app.hwInfo.CPU.Features[idx], colWidth-2), styleNormal)
					}
				}
			}
			y++
		}
	}
}
