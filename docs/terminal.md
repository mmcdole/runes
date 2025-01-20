# Terminal UI Implementation Approaches

When implementing a MUD client's terminal interface, handling scrolling behavior and ANSI color state management presents interesting challenges. A key consideration is how color states propagate across line breaks, particularly when the initiating color sequence has scrolled out of view. Consider this common scenario in MUD output:

```
Welcome to \033[31mThe Crimson Realm
A grand castle stands before you.
Its mighty walls stretch towards the sky.
The portcullis looms menacingly.\033[0m
```

In this example, even though the color red (\033[31m) is initiated on the first line, the subsequent lines should maintain this red coloring until explicitly reset. This behavior must persist even if the first line scrolls out of the visible viewport. Different MUD clients handle this challenge in distinct ways.

## TinTin++ Approach: Managed Screen Buffer

TinTin++ implements a sophisticated screen and buffer management system that provides complete control over the display state. Key components include:

1. A dedicated scroll buffer (`struct scroll`) that maintains the complete history of lines and their states
2. A screen management system that handles both the main viewport and split regions
3. Line-based state tracking where color states are preserved at line boundaries
4. Custom scrolling implementation that redraws the visible portion of the screen from the internal buffer

The scrolling mechanism works by maintaining two key components:
- A history buffer (`scroll->buffer`) that stores all lines with their associated metadata
- A screen buffer that represents the current visible state

When scrolling occurs, TinTin++ recalculates which portion of the history buffer should be visible and redraws the screen accordingly. This approach gives TinTin++ precise control over the display, as each scroll operation involves:
1. Calculating the new visible region
2. Preserving color states between lines
3. Redrawing the visible portion with correct color information
4. Managing split regions and input areas separately from the main scroll region

The critical code for color state propagation can be found in `get_color_codes()`:

```c
// Mix old (previous line) and current line color codes
void get_color_codes(char *old, char *str, char *buf, int flags)
{
    // First process the old (previous) line's color state
    pti = old;
    while (*pti) {
        // Process VT102 codes from previous line
        // Building up the color state
    }
    
    // Then process the current line's color codes
    pti = str;
    while (*pti) {
        // Process VT102 codes from current line
        // Updating the color state
    }
    
    // Generate final color state combining both
    ptb = buf;
    *ptb++ = '\e';
    *ptb++ = '[';
    *ptb++ = '0';
    
    // Add all active attributes from combined state
    if (HAS_BIT(vtc, COL_BLD)) {
        *ptb++ = ';';
        *ptb++ = '1';
    }
    // Add foreground/background colors
    if (HAS_BIT(vtc, COL_XTF)) {
        ptb += sprintf(ptb, ";38;5;%d", fgc);
    }
}
```

This color state propagation happens in `process_one_line()` before each line is added to the buffer:

```c
void process_one_line(struct session *ses, char *linebuf, int prompt)
{
    // ... other processing ...
    
    if (HAS_BIT(ses->config_flags, CONFIG_FLAG_COLORPATCH))
    {
        // First prepend the previous color state
        sprintf(temp, "%s%s%s", ses->color_patch, linebuf, "\e[0m");
        
        // Then update color_patch with the final state of this line
        // This will be used for the next line
        get_color_codes(ses->color_patch, linebuf, ses->color_patch, GET_ALL);
        
        strcpy(linebuf, temp);
    }
    
    // Store the line with color state in buffer
    add_line_buffer(ses, linebuf, prompt);
    // ... display handling ...
}
```

This shows how TinTin++ actively manages color state propagation by:
1. Prepending the previous line's color state to each new line
2. Extracting and storing the final color state from the current line
3. Using that stored state as the starting point for the next line
4. This chain ensures colors persist correctly even when scrolling

## Blightmud Approach: Native Terminal State

Blightmud takes advantage of the terminal's built-in capabilities for both scrolling and color state management. The key implementation details show how this works:

```rust
// Terminal setup in create_screen_writer()
fn create_screen_writer(mouse_support: bool) -> Result<Box<dyn Write>> {
    // Put terminal in raw mode and use alternate screen buffer
    let screen = stdout().into_raw_mode()?.into_alternate_screen()?;
    if mouse_support {
        Ok(Box::new(MouseTerminal::from(screen)))
    } else {
        Ok(Box::new(screen))
    }
}

// Direct line printing in ReaderScreen
impl ReaderScreen {
    fn print_line(&mut self, line: &Line) {
        if let Some(print_line) = &line.print_line() {
            // Store in history buffer
            self.history.append(print_line);
            
            // If not scrolling, write directly to terminal
            if !self.scroll_data.active {
                writeln!(
                    self.screen,
                    "{}\n{}{}",
                    Goto(1, self.height - 1),
                    print_line,  // Line contains raw ANSI codes
                    Goto(1, self.height)
                ).unwrap();
            }
        }
    }
}
```

This implementation shows how Blightmud:
1. Sets up a raw terminal mode that preserves ANSI escape sequences
2. Writes lines directly to the terminal with their original color codes intact
3. Lets the terminal emulator handle color state propagation between lines
4. Uses the terminal's native scrollback buffer for history
5. Only manages its own history buffer for features like searching and scrollback

The key difference from TinTin++ is that Blightmud doesn't need to explicitly track or propagate color states between lines. Instead, it relies on the terminal emulator to maintain color state according to the VT100/ANSI standard, which specifies that color states persist until explicitly reset. This creates a more lightweight implementation that integrates seamlessly with the terminal's built-in functionality.
