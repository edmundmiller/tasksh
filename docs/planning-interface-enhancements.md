# Daily Planning Interface Enhancements

This document describes the enhanced visual design and user experience improvements made to the daily planning interface in tasksh, specifically focusing on steps 3-5 of the planning workflow.

## Overview

The daily planning interface guides users through a 5-step process:
1. Reflection - Review yesterday's work
2. Task Selection - Choose tasks for today
3. **Workload Assessment** - Analyze capacity and sustainability
4. **Finalization** - Set daily focus
5. **Summary** - Review complete plan

Steps 3-5 have been significantly enhanced to provide better visual feedback, clearer information hierarchy, and actionable insights.

## Step 3: Workload Assessment

### Purpose
Helps users understand if their planned workload is sustainable and well-balanced, preventing overcommitment and burnout.

### Visual Components

#### 1. Context Introduction
- Clear explanation of the assessment's purpose
- Sets expectation for sustainable productivity

#### 2. Workload Overview Box
A bordered box displaying key metrics:
- 📋 **Selected Tasks**: Total task count
- ⏱ **Total Estimated Time**: Sum of all task estimates
- 🎯 **Available Focus Time**: Realistic work hours (65% of workday)
- 📊 **Capacity Utilization**: Percentage of focus time used

#### 3. Capacity Visualization Bar
```
Capacity: [████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░] 96%
```
- Visual progress bar showing workload capacity
- Dynamic coloring based on utilization:
  - **Green** (< 90%): Good capacity, sustainable pace
  - **Yellow** (90-100%): Near capacity, full day
  - **Orange** (100-120%): Overloaded, needs adjustment
  - **Red** (> 120%): Significantly overloaded

#### 4. Task Breakdown by Priority
Shows distribution of work across priority levels:
```
TASK BREAKDOWN BY PRIORITY:
  🔴 Critical: 2.0h (40%)
  🟡 Important: 1.5h (30%)
  🟢 Flexible: 1.5h (30%)
```

#### 5. Energy Requirements
Displays cognitive load distribution:
```
ENERGY REQUIREMENTS:
  ⚡ High energy:    3.0h
  🔋 Medium energy:  1.5h
  🔌 Low energy:     0.5h
```

#### 6. Smart Recommendations
Context-aware tips based on capacity:
- **Overloaded**: "Consider deferring some flexible tasks or breaking larger tasks into smaller chunks"
- **Underutilized**: "You have room for more tasks if needed, or enjoy a lighter day!"

### Design Decisions
- Uses color coding consistently (red/yellow/green) for quick visual scanning
- Provides both numerical and visual representations of capacity
- Includes actionable recommendations rather than just data
- Groups related information in bordered sections for clarity

## Step 4: Finalization

### Purpose
Helps users set a daily focus statement that guides decision-making throughout the day.

### Visual Components

#### 1. Purpose Explanation
- Explains why daily focus matters
- Emphasizes alignment and decision-making benefits

#### 2. Focus Display (After Setting)
When a focus is set, displays in an attractive bordered box:
```
┌──────────────────────────────────────────────────────────┐
│                    ✨ Today's Focus ✨                    │
│                                                          │
│        Ship the authentication feature with tests        │
└──────────────────────────────────────────────────────────┘
```

#### 3. Example Focus Statements
Provides inspiration with real examples:
- "Ship the authentication feature with comprehensive tests"
- "Deep work on algorithm optimization - no meetings"
- "Clear technical debt and improve code documentation"
- "Customer support and bug fixes - be responsive"

#### 4. Quick Summary
After focus is set, shows:
- Confirmation checkmark
- Task count and total hours
- Transition prompt to final summary

### Design Decisions
- Centers the focus statement for emphasis
- Uses warm yellow color (#11) for the focus box border
- Provides concrete examples to guide users
- Maintains context about the plan while setting focus

## Step 5: Summary

### Purpose
Provides a comprehensive overview of the daily plan with visual hierarchy and time projections.

### Visual Components

#### 1. Date Header
- Uses calendar emoji (📅) for visual interest
- Bold, prominent date display
- Consistent formatting across all planning sessions

#### 2. Daily Focus Box
If set, prominently displays the focus statement:
```
┌──────────────────────────────────────────────────────────┐
│    🎯 Ship the authentication feature with tests         │
└──────────────────────────────────────────────────────────┘
```

#### 3. Categorized Task Lists
Tasks grouped by priority with visual indicators:

```
🔴 CRITICAL TASKS
  1. Fix authentication bug                           2.0h
     Project: auth-service • Due: 2024-01-15
  2. Deploy hotfix to production                      0.5h

🟡 IMPORTANT TASKS
  1. Code review for team                             1.0h
  2. Update API documentation                         1.5h
     Project: api-docs

🟢 FLEXIBLE TASKS
  1. Refactor user service                           2.0h
  2. Research new testing framework                   1.0h
```

Features:
- Color-coded priority sections
- Right-aligned time estimates for easy scanning
- Metadata (project, due date) shown in gray italics
- Consistent numbering within each category

#### 4. Statistics Summary Box
Bordered box with key metrics:
```
┌──────────────────────────────────────────────────┐
│ 📋 Total Tasks: 6                                │
│ ⏱  Total Time: 8.0 hours                        │
│ 🕐 If starting at 9 AM: Done by 5:00 PM         │
└──────────────────────────────────────────────────┘
```

#### 5. Motivational Status Message
Dynamic message based on workload:
- **Overloaded**: "⚠️ Challenging day ahead - stay focused on priorities!"
- **Full day**: "💪 Full day planned - you've got this!"
- **Balanced**: "✨ Well-balanced day - room for the unexpected!"

#### 6. Completion Prompt
Encouraging message with rocket emoji to start the day

### Design Decisions
- Uses visual hierarchy with boxes, colors, and spacing
- Provides time projections for realistic expectations
- Groups related information together
- Maintains consistent color coding throughout
- Adds personality with contextual messages and emojis

## Technical Implementation

### Color Palette
The interface uses ANSI color codes for terminal compatibility:
- **Color 1**: Red - Critical/errors
- **Color 2**: Green - Flexible/success
- **Color 3**: Yellow - Important/warning
- **Color 6**: Cyan - Headers/emphasis
- **Color 7**: White - Normal text
- **Color 8**: Bright black - Subtle elements
- **Color 10**: Bright green - Success messages
- **Color 11**: Bright yellow - Focus/highlights
- **Color 15**: Bright white - Bold text

### Styling Approach
- Uses `lipgloss` library for consistent styling
- Applies styles programmatically for maintainability
- Ensures compatibility across different terminal themes
- Maintains readability with appropriate contrast

### Layout Principles
1. **Visual Hierarchy**: Important information is larger/bolder
2. **Consistent Spacing**: Predictable margins and padding
3. **Grouped Information**: Related data in boxes or sections
4. **Progressive Disclosure**: Details shown where relevant
5. **Color Meaning**: Consistent color usage across steps

## User Experience Benefits

1. **Reduced Cognitive Load**: Visual elements help quick comprehension
2. **Better Decision Making**: Clear capacity indicators prevent overcommitment
3. **Increased Motivation**: Positive messaging and progress visualization
4. **Improved Planning**: Examples and recommendations guide better choices
5. **Enhanced Clarity**: Organized display makes plan easy to review

## Future Enhancements

Potential improvements to consider:
- Customizable work hours and break times
- Historical capacity accuracy tracking
- Integration with calendar for meeting conflicts
- Export options for daily plans
- Themed color schemes for accessibility