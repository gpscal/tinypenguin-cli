# Fine-Tuning Data Collection - Implementation Summary

## Changes Made

### 1. Enhanced Logging Structure (`cli/pkg/cli/task.go`)

**Updated `ToolCallLog` struct** to include full conversation context:
- ✅ Added `UserQuery` field - stores the original user query
- ✅ Added `ModelResponse` field - stores the complete model response with tool calls

**Before:**
```go
type ToolCallLog struct {
    Timestamp     time.Time
    Model         string
    ToolName      string
    Arguments     string
    // ... other fields
}
```

**After:**
```go
type ToolCallLog struct {
    Timestamp     time.Time
    Model         string
    UserQuery     string    // NEW: Original user query
    ModelResponse string    // NEW: Full model response
    ToolName      string
    Arguments     string
    // ... other fields
}
```

### 2. Updated Logging Calls

All `logToolCall()` invocations now capture:
- ✅ Original user query (`query` parameter)
- ✅ Full model response (serialized JSON of the `message` object)
- ✅ Tool execution results (existing)
- ✅ User ratings (existing)

### 3. Conversion Script (`convert_logs_for_finetuning.go`)

Created a comprehensive conversion script that:
- ✅ Reads `tool_calls.log` entries
- ✅ Converts to Qwen fine-tuning format (JSONL with messages array)
- ✅ Handles both old format (without user_query) and new format
- ✅ Filters by minimum rating (`--min-rating` flag)
- ✅ Reconstructs missing data for old format entries
- ✅ Includes tool execution results in conversation flow

### 4. Helper Script (`convert-for-finetuning.sh`)

Created a bash wrapper script for easy conversion:
```bash
./convert-for-finetuning.sh [log_file] [output_file] [min_rating]
```

### 5. Documentation (`FINETUNING.md`)

Created comprehensive documentation covering:
- ✅ Data collection process
- ✅ Log format explanation
- ✅ Conversion instructions
- ✅ Fine-tuning workflow
- ✅ Best practices and troubleshooting

## Usage

### Collecting Training Data

Simply use the CLI tool normally - all tool calls are automatically logged:

```bash
tinypenguin-cli run "Check current users"
# Rate: 5 ⭐

tinypenguin-cli run "Show disk usage"
# Rate: 4 ⭐
```

### Converting to Fine-Tuning Format

```bash
# Using the helper script
./convert-for-finetuning.sh tool_calls.log finetuning_data.jsonl 4

# Or directly with Go
go run convert_logs_for_finetuning.go tool_calls.log finetuning_data.jsonl --min-rating 4
```

### Output Format

The converted data follows Qwen's fine-tuning format:

```json
{
  "messages": [
    {"role": "user", "content": "Check current users"},
    {
      "role": "assistant",
      "content": "...",
      "tool_calls": [...]
    },
    {"role": "tool", "content": "Tool execution result: ..."}
  ]
}
```

## Benefits

1. **Complete Context**: Full conversation history preserved
2. **Quality Filtering**: Filter by user ratings (1-5 stars)
3. **Backward Compatible**: Handles old log format gracefully
4. **Ready for Fine-Tuning**: Output format matches Qwen requirements
5. **Easy to Use**: Simple conversion process

## Migration Notes

- **Old logs**: Will be reconstructed (best effort) during conversion
- **New logs**: Include full context automatically
- **No breaking changes**: Existing functionality unchanged

## Next Steps

1. Start using the CLI tool to collect training data
2. Rate interactions consistently (1-5 stars)
3. Periodically convert logs to fine-tuning format
4. Use converted data to fine-tune `qwen2.5-coder:0.5b`
