
Bugs

* Line wrap on InputBar text issues
* Fix command echo to only be user entered commands, no resolved commands
* Fix ctrl+c and quitting -- it is messy and requires double quit sometimes
* Fix bugs with color propagation
* Resize should automatically re-render


Features


* /buffer list & switch command implementation (extend buffer.go)
* Ctrl+R FZF history mode
* Tab-autocomplete on input bar
* Config settings for command echo, command seperator, etc.
* Configurable keybindings
* Scroll mode enhancements (quickly escape scroll mode to live mode, scroll by half page, etc.)
* Search mode implementation
* implement and verify timers, gags, highlights, triggers
* Save servers / search mud database, connect to saved servers
