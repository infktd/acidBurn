# acidBurn Styling Enhancement Ideas

## Current Stack
- **Bubble Tea** - TUI framework ✅
- **Bubbles** - Component library ✅ (already included!)
- **Lip Gloss** - Styling library ✅

## Available Bubbles Components

### 1. **List** Component
Interactive list with built-in filtering, pagination, and help.

**Current:** Custom sidebar with manual rendering
**Could Use For:**
- Project sidebar (with fuzzy search built-in!)
- Service list
- Alert history

**Benefits:**
- Built-in fuzzy filtering
- Pagination for long lists
- Keyboard navigation handled automatically
- Status messages and spinner

### 2. **Table** Component
Formatted tables with scrolling and customization.

**Current:** Custom service rendering with manual spacing
**Could Use For:**
- Services view (STATUS, PID, CPU, MEM columns)
- Better alignment and borders

**Benefits:**
- Automatic column sizing
- Header/footer support
- Built-in scrolling
- Selectable rows

### 3. **Spinner** Component
Loading indicators with many built-in styles.

**Current:** Static "Starting acidBurn..." text
**Could Use For:**
- Project startup indicator
- Service restart indicator
- Background operations

**Available Styles:**
- `Line` - Simple rotating line
- `Dot` - Bouncing dot
- `Jump` - Jumping bar
- `Pulse` - Pulsing circle
- `Points` - Moving points
- `Globe` - Rotating globe
- `Meter` - Progress meter
- `Hamburger` - Loading hamburger

### 4. **Progress** Component
Customizable progress bars with gradients.

**Current:** Custom progress bar in splash
**Could Use For:**
- Startup progress (already using custom)
- Long-running operation progress
- Download/upload indicators

**Benefits:**
- Gradient fills
- Custom characters
- Percentage display
- Animation support

### 5. **Viewport** Component
Scrollable content area.

**Current:** Custom log view with manual scrolling
**Could Use For:**
- Log viewer (replace custom implementation)
- Alert history
- Help text

**Benefits:**
- Smooth scrolling
- Mouse support
- Auto-scroll to bottom
- Percentage indicator

### 6. **Paginator** Component
Page navigation for multi-page content.

**Could Use For:**
- Long project lists
- Alert history pages
- Service list pagination

### 7. **TextInput** Component
Single-line text input with cursor, clipboard, suggestions.

**Current:** Custom search input handling
**Could Use For:**
- Log search bar
- Project search
- Filter inputs
- Rename project dialog

**Benefits:**
- Cursor management
- Selection support
- Placeholder text
- Character/width limits
- Password mode

### 8. **Help** Component
Automatic help text generation from keybindings.

**Current:** Custom help modal
**Could Use For:**
- Generate help text automatically from KeyMap
- Context-sensitive help

### 9. **Stopwatch/Timer** Component
Time tracking and countdowns.

**Could Use For:**
- Service uptime display
- Auto-refresh countdown
- Operation timeout display

## Lip Gloss Styling Enhancements

### Border Styles
```go
// Available borders
lipgloss.NormalBorder()
lipgloss.RoundedBorder()
lipgloss.BlockBorder()
lipgloss.OuterHalfBlockBorder()
lipgloss.InnerHalfBlockBorder()
lipgloss.ThickBorder()
lipgloss.DoubleBorder()
lipgloss.HiddenBorder()
```

**Currently Using:** NormalBorder
**Could Try:** RoundedBorder for modals, ThickBorder for focused elements

### Gradients
```go
style := lipgloss.NewStyle().
    Foreground(lipgloss.AdaptiveColor{Light: "#000", Dark: "#FFF"}).
    Background(lipgloss.Color("#FF00FF"))
```

**Could Use For:**
- Status indicators (idle → running gradient)
- Progress bars
- Theme transitions

### Padding & Margin
```go
style.Padding(1, 2, 1, 2)  // top, right, bottom, left
style.Margin(0, 1)          // vertical, horizontal
```

**Could Improve:**
- Modal spacing (currently basic)
- Service list alignment
- Header/footer padding

### Alignment
```go
style.Align(lipgloss.Left, lipgloss.Center)
style.AlignHorizontal(lipgloss.Right)
style.AlignVertical(lipgloss.Bottom)
```

**Could Use For:**
- Right-align stats in header
- Center modal titles
- Bottom-align footer help

## Quick Wins for acidBurn

### 1. **Add Spinners for Loading States**
```go
import "github.com/charmbracelet/bubbles/spinner"

type Model struct {
    spinner spinner.Model
    loading bool
}

// When starting a project:
m.loading = true
m.spinner.Tick()
```

### 2. **Use Table for Services View**
```go
import "github.com/charmbracelet/bubbles/table"

columns := []table.Column{
    {Title: "Service", Width: 20},
    {Title: "Status", Width: 10},
    {Title: "PID", Width: 8},
    {Title: "CPU", Width: 8},
    {Title: "Memory", Width: 10},
}
```

### 3. **Enhanced Progress Bar**
```go
import "github.com/charmbracelet/bubbles/progress"

prog := progress.New(
    progress.WithDefaultGradient(),
    progress.WithWidth(40),
)
```

### 4. **Rounded Borders for Modals**
```go
modalStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(m.styles.theme.Primary).
    Padding(1, 2)
```

### 5. **List Component for Projects**
```go
import "github.com/charmbracelet/bubbles/list"

// Built-in fuzzy search!
projectList := list.New(items, list.NewDefaultDelegate(), 30, 20)
projectList.Title = "Projects"
```

## Suggested Enhancements Priority

### High Impact, Low Effort:
1. ✅ **Fix bracket colors in help** - DONE
2. **Add spinner for project start/stop operations**
3. **Rounded borders for modals**
4. **Table component for services view**

### Medium Impact, Medium Effort:
5. **List component for project sidebar** (fuzzy search!)
6. **Progress component for better loading bars**
7. **Enhanced borders and padding**

### Nice to Have:
8. **Viewport for logs** (replace custom)
9. **TextInput for search** (better than manual handling)
10. **Stopwatch for service uptime**

## Example: Adding a Spinner

**Before:**
```
Starting myproject...
```

**After:**
```
⠋ Starting myproject...
```

**Implementation:**
```go
// In model.go
import "github.com/charmbracelet/bubbles/spinner"

func New(...) *Model {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(styles.theme.Primary)

    return &Model{
        spinner: s,
        // ...
    }
}

// In Update:
case projectStartedMsg:
    m.loading = false
    // ...

// In View:
if m.loading {
    return m.spinner.View() + " " + m.loadingMessage
}
```

## Resources
- [Bubbles GitHub](https://github.com/charmbracelet/bubbles)
- [Bubbles Docs](https://pkg.go.dev/github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [Bubble Tea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
