# AINative-Code TUI - Keyboard Shortcuts Quick Reference

## Navigation

| Shortcut | Description |
|----------|-------------|
| `↑` or `k` | Scroll up one line |
| `↓` or `j` | Scroll down one line |
| `PgUp` or `Ctrl+U` | Scroll up half page |
| `PgDn` or `Ctrl+D` | Scroll down half page |
| `Home` or `g` | Jump to top of conversation |
| `End` or `G` | Jump to bottom of conversation |
| Mouse Wheel | Scroll up/down (3 lines) |

## Editing & Input

| Shortcut | Description |
|----------|-------------|
| `Enter` | Send message to AI |
| `Ctrl+L` | Clear conversation history |
| `ESC` | Cancel current input or close help |

## View Options

| Shortcut | Description |
|----------|-------------|
| `t` | Toggle thinking blocks on/off |
| `e` | Expand all thinking blocks |
| `c` | Collapse all thinking blocks |
| `Ctrl+R` | Refresh display |

## Help & System

| Shortcut | Description |
|----------|-------------|
| `?` | Show/hide help overlay |
| `Ctrl+H` | Alternative help toggle |
| `Ctrl+C` | Quit application |

## Status Bar Information

The status bar shows:
- **Left**: Connection status (Ready, Streaming, Error)
- **Middle**: Provider, model, token usage (when available)
- **Right**: Session duration, scroll position, keyboard hints

### Status Indicators

- `● Ready` - Green: Ready for input
- `● Streaming...` - Green: Receiving AI response
- `✗ Error` - Red: Error occurred
- `⚠ Disconnected` - Red: Connection lost

### Scroll Indicator

- `↑ Top` - At the beginning of conversation
- `↓ Bottom` - At the end of conversation
- `↕ XX%` - Scroll position percentage

## Thinking Blocks

Extended thinking from Claude can be toggled on/off:
- Press `t` to toggle display
- Press `e` to expand all blocks
- Press `c` to collapse all blocks
- Status bar shows current state (ON/OFF)

## Responsive Design

The TUI adapts to your terminal size:
- **Fullscreen (>=80 cols)**: Full status bar with all information
- **Medium (40-79 cols)**: Compact status bar
- **Small (<40 cols)**: Minimal view with essential info

## Tips

1. **Vim Users**: Use `k`/`j` for scrolling, `g`/`G` for jump navigation
2. **Quick Help**: Press `?` anytime to see all shortcuts
3. **Auto-Scroll**: During streaming, viewport automatically scrolls to bottom
4. **Mouse Support**: Mouse wheel works in compatible terminals
5. **Keyboard Only**: Everything accessible via keyboard - no mouse required
6. **Context Hints**: Watch status bar for context-sensitive hints

## Accessibility

- **High Contrast**: Clear color coding for all states
- **Keyboard Navigation**: Complete control without mouse
- **Screen Reader Friendly**: Structured text output
- **Visual Indicators**: Multiple cues for all states

## Color Coding

- **Blue**: User messages and prompts
- **Green**: Assistant messages and success states
- **Yellow**: Warnings and system messages
- **Red**: Errors
- **Purple**: Thinking blocks and borders
- **Gray**: Muted text and separators

Press `?` in the TUI to see this help anytime!
