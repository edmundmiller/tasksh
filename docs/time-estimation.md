# Time Estimation System

Tasksh now includes a robust, multi-source time estimation system that combines multiple data sources to provide accurate time estimates for tasks.

## Features

### 1. **Multi-Source Estimation**
The system combines estimates from multiple sources in priority order:

- **Timewarrior Data** (highest priority if available)
  - Actual tracked time for completed tasks
  - Partial time for in-progress tasks with completion percentage
  
- **Historical Database** (medium priority)
  - Enhanced similarity matching using keywords, project, and priority
  - Time-weighted to prefer recent completions
  
- **AI Estimation** (optional, medium priority)
  - Uses OpenAI API for intelligent analysis
  - Cached to reduce API calls
  - Enable with `TASKSH_AI_ESTIMATION=true`
  
- **Keyword-Based Fallback** (lowest priority)
  - Quick tasks: "email", "call" → 0.5 hours
  - Small tasks: "review", "test" → 1.0 hours
  - Large tasks: "implement", "create" → 3.0 hours
  - Research tasks: "analyze", "investigate" → 2.5 hours

### 2. **Confidence Scoring**
Each estimate includes a confidence score (0-1) based on:
- Data source reliability
- Number of similar tasks found
- Match quality (exact, project, keywords)
- Recency of data

### 3. **Learning System**
- Tracks estimation accuracy over time
- Applies calibration factors based on historical accuracy
- Identifies patterns in over/under-estimation

### 4. **Timewarrior Integration**
**Automatic Background Sync**: Timewarrior data is automatically synced in the background every 4 hours when you use the planning features. No manual sync needed!

Manual sync commands (optional):
```bash
# Sync last 30 days (default)
tasksh sync

# Sync all time data
tasksh sync all

# Sync specific time period
tasksh sync week
tasksh sync month
tasksh sync year
tasksh sync 2024-01-01
```

Check sync status:
```bash
tasksh diagnostics  # Shows last sync time and next sync schedule
```

## Usage

### Daily Planning
When you use `tasksh plan daily`, tasks are automatically estimated using all available data sources:

```
📋 Select tasks for today

AVAILABLE TASKS:
[ ]  ▶  1. Fix parser bug in config module (3.5h, Critical)
         High priority • Based on 5 similar tasks in project 'tasksh' (85% confidence)
[✓]     2. Review pull request #123 (1.2h, Important)  
         Medium priority • Already tracked 0.8 hours (estimated 70% complete) (90% confidence)
[ ]     3. Update documentation (1.0h, Flexible)
         Low priority • Small task based on keywords (45% confidence)
```

### Configuration

Enable AI estimation (optional):
```bash
export TASKSH_AI_ESTIMATION=true
export OPENAI_API_KEY=your-key-here
# or
export OPENAI_API_KEY_CMD='op read "op://Private/api.openai.com/apikey"'
```

Configure estimation preferences in `~/.config/tasksh/settings.json`:
```json
{
  "estimation": {
    "preferTimewarrior": true,
    "minConfidence": 0.3,
    "aiEnabled": false,
    "cacheAIEstimates": true,
    "fallbackHours": 2.0,
    "autoSyncEnabled": true,
    "autoSyncInterval": "4h"
  }
}
```

## How It Works

1. **Task Analysis**: When estimating a task, the system:
   - Checks if timewarrior has tracked time for this task UUID
   - Searches for similar completed tasks in the database
   - Optionally queries AI for analysis
   - Falls back to keyword-based estimation

2. **Similarity Matching**: Enhanced algorithm considers:
   - Exact description matches (100% weight)
   - Same project (50% weight)
   - Shared keywords using Jaccard similarity (40% weight)
   - Priority matching (10% weight)
   - Time decay (recent tasks weighted higher)

3. **Estimate Selection**: The system:
   - Collects all available estimates
   - Sorts by confidence and configured preferences
   - Selects the highest confidence estimate above threshold
   - Applies learning-based calibration
   - Adds 15% buffer for planning safety

## Best Practices

1. **Track Time Consistently**: Use timewarrior to track actual time:
   ```bash
   timew start task_<uuid> project_<name> "Task description"
   timew stop
   ```

2. **Sync Regularly**: Keep estimates updated:
   ```bash
   tasksh sync week  # Weekly sync is usually sufficient
   ```

3. **Use Descriptive Task Names**: Better descriptions lead to better matches:
   - Good: "Implement user authentication with OAuth2"
   - Poor: "Fix stuff"

4. **Review Estimation Accuracy**: Check how well estimates match reality:
   ```bash
   # Future feature: tasksh estimate stats
   ```

## Technical Details

The estimation system is implemented across several packages:

- `internal/estimation/`: Core estimation logic and multi-source combiner
- `internal/timewarrior/`: Timewarrior client for reading tracked time
- `internal/timedb/`: Historical database with similarity matching
- `internal/ai/`: Optional AI-powered analysis

Data flows through a confidence-weighted pipeline that ensures the best available estimate is always used while maintaining transparency about the source and confidence level.