# TODO - Tasksh Improvements

This file tracks planned improvements and feature requests for tasksh.

## Recently Completed ✅

### High Priority Fixes
- [x] **Fix modify functionality** - Text input not receiving keystrokes in modify mode
  - Fixed by properly updating textInput component in updateModificationInput()
  - Also fixed for wait date and wait reason inputs
  - Fixed 'm' key appearing in input field by adding modeJustChanged flag
  
- [x] **Fix edit functionality** - Suspend Bubble Tea for external editor  
  - Implemented using tea.ExecProcess() for proper terminal handling
  - Added CreateEditCommand() in taskwarrior package
  - Edit now works properly without terminal conflicts

- [x] **Advanced auto-completion for modify command**
  - Real-time completion suggestions with popup display
  - Arrow key navigation through suggestions
  - Tab completion for selected suggestion
  - Context-aware suggestions based on input

### New Features  
- [x] **Calendar component for wait date selection**
  - Interactive calendar with keyboard navigation
  - Month/year navigation with p/n or </> keys
  - Arrow key navigation for days
  - Today highlighting and quick jump with 't' key
  
- [x] **Integrated calendar into wait flow**
  - Press 'w' to open calendar picker
  - Tab to toggle between calendar and text input
  - Supports both visual date picking and text parsing
  - Graceful fallback for complex date expressions

## Planned Improvements 🚀

### High Priority
- [ ] **Better error handling and user feedback**
  - More informative error messages for taskwarrior failures
  - Better validation of user inputs
  - Graceful handling of missing dependencies

- [ ] **Performance optimizations**
  - Cache task information to reduce taskwarrior calls
  - Lazy loading of task details
  - Optimize rendering for large task lists

### Medium Priority
- [ ] **Enhanced calendar features**
  - Support for more date input formats ("next friday", "in 2 weeks")
  - Holiday highlighting/awareness
  - Multiple date selection for recurring tasks
  - Visual indicators for due dates

- [ ] **Improved AI integration**
  - Better error handling when mods is unavailable
  - More detailed task analysis prompts
  - Integration with different AI providers
  - Smarter time estimation based on historical data

- [ ] **Enhanced task display**
  - Color coding by priority/urgency
  - Tags and annotations display
  - Task dependencies visualization
  - Progress indicators for recurring tasks

### Low Priority
- [ ] **Configuration improvements**
  - User-configurable key bindings
  - Themeable UI colors and styles
  - Customizable task display format
  - Plugin system for extensions

- [ ] **Export and reporting features**
  - Export review sessions to various formats
  - Productivity reports and analytics
  - Time tracking visualization
  - Integration with external time tracking tools

- [ ] **Advanced features**
  - Bulk operations on multiple tasks
  - Custom review filters and queries
  - Integration with external calendars
  - Mobile/web interface companion

## Technical Debt 🔧

- [ ] **Test coverage improvements**
  - Add integration tests for calendar functionality
  - Mock taskwarrior for more reliable testing
  - Performance benchmarking tests
  - UI interaction testing

- [ ] **Code organization**
  - Split large files into smaller modules
  - Better separation of concerns
  - Standardize error handling patterns
  - Documentation improvements

- [ ] **Dependency management**
  - Regular dependency updates
  - Minimize external dependencies where possible
  - Better handling of optional dependencies (mods, etc.)

## Implementation Notes 📝

### Auto-Completion System
- **Real-time suggestions**: Updates as user types with intelligent context parsing
- **Multi-type completion**: Supports attributes, projects, tags, priorities, status values, and dates
- **Keyboard navigation**: Arrow keys to navigate, Tab to complete, ESC to cancel
- **Visual feedback**: Color-coded suggestions by type with descriptions
- **Integration**: Uses existing taskwarrior data (_projects, _tags commands)
- **Smart parsing**: Understands context like "project:", "+tag", "-tag", attribute:value patterns

### Calendar Component
- Built with Bubble Tea framework for consistency
- Uses standard time.Time for date handling
- Supports both keyboard and visual navigation
- Integrates seamlessly with existing text input fallback

### Edit Functionality  
- Uses tea.ExecProcess() to suspend the Bubble Tea program
- Allows external editor ($EDITOR) to take full terminal control
- Automatically marks tasks as reviewed after successful edit
- Handles editor failures gracefully

### Text Input Fixes
- Fixed by ensuring textInput.Update() is called in all input modes
- Properly returns updated model and commands
- Maintains focus state correctly across mode transitions
- Added modeJustChanged flag to prevent triggering keys from appearing in input

## Contributing 🤝

When adding new features:
1. Update this TODO.md file with progress
2. Follow existing code patterns and naming conventions  
3. Add appropriate tests for new functionality
4. Update documentation as needed
5. Consider backwards compatibility

## Priority Levels

- **High**: Critical bugs, usability issues, or frequently requested features
- **Medium**: Nice-to-have improvements that enhance user experience
- **Low**: Polish, advanced features, or edge case handling

---

_Last updated: 2025-01-27_