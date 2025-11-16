# Fine-Tuning Data Collection and Conversion

This document explains how to collect training data from tool calls and convert it to Qwen fine-tuning format.

## Overview

The `tool_calls.log` file now captures full conversation context including:
- **User Query**: The original user request
- **Model Response**: The complete model response with tool calls
- **Tool Execution Results**: Status, output, and error details
- **Rating**: User-provided quality rating (1-5 stars)

## Data Collection

### Automatic Logging

Every tool call execution is automatically logged to `tool_calls.log` with:
- Original user query
- Full model response (including tool calls)
- Tool execution results
- User rating (optional)

### Log Format

Each entry in `tool_calls.log` is a JSON object:

```json
{
  "timestamp": "2025-11-16T13:10:33.85769013-05:00",
  "model": "qwen2.5-coder:3b",
  "user_query": "Check current users",
  "model_response": "{\"role\":\"assistant\",\"content\":\"\",\"tool_calls\":[...]}",
  "tool_name": "run_commands",
  "arguments": "{\"command\":\"who\"}",
  "status": "success",
  "message": "Command executed successfully",
  "output": "cal      pts/0        2025-11-16 10:06 (192.168.2.25)\n",
  "tools_enabled": true,
  "rating": 5
}
```

## Converting to Fine-Tuning Format

### Prerequisites

- Go 1.20+ installed
- `tool_calls.log` file with training data

### Conversion Script

Use the `convert_logs_for_finetuning.go` script to convert logs to Qwen fine-tuning format:

```bash
# Basic usage
go run convert_logs_for_finetuning.go tool_calls.log

# Specify output file
go run convert_logs_for_finetuning.go tool_calls.log finetuning_data.jsonl

# Filter by minimum rating (only include examples rated 4+)
go run convert_logs_for_finetuning.go tool_calls.log finetuning_data.jsonl --min-rating 4
```

### Output Format

The conversion script produces a JSONL file where each line is a fine-tuning example:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Check current users"
    },
    {
      "role": "assistant",
      "content": "I'll help you with that. Let me use the run_commands tool.",
      "tool_calls": [
        {
          "id": "call_1",
          "type": "function",
          "function": {
            "name": "run_commands",
            "arguments": "{\"command\":\"who\"}"
          }
        }
      ]
    },
    {
      "role": "tool",
      "content": "Tool execution result:\nStatus: success\nOutput: cal      pts/0        2025-11-16 10:06 (192.168.2.25)\n"
    }
  ]
}
```

## Fine-Tuning with Qwen

### Using the Converted Data

The converted JSONL file can be used with Qwen fine-tuning tools:

```bash
# Example with Qwen fine-tuning (adjust based on your setup)
python -m qwen.finetune \
  --model_name qwen2.5-coder:0.5b \
  --train_data finetuning_data.jsonl \
  --output_dir ./finetuned_model
```

### Quality Filtering

Use the `--min-rating` flag to filter examples by quality:

- `--min-rating 3`: Include examples rated 3+ (default)
- `--min-rating 4`: Include only high-quality examples (4-5 stars)
- `--min-rating 5`: Include only perfect examples (5 stars)

### Best Practices

1. **Collect Diverse Examples**: Use various types of commands and queries
2. **Rate Consistently**: Provide ratings after each tool execution
3. **Filter Quality**: Use `--min-rating 4` for production fine-tuning
4. **Review Output**: Check the converted data before fine-tuning
5. **Backup Logs**: Keep original `tool_calls.log` for future conversions

## Example Workflow

```bash
# 1. Use the CLI tool and rate interactions
tinypenguin-cli run "Check current users"
# Rate: 5

tinypenguin-cli run "Show disk usage"
# Rate: 4

# 2. Convert logs to fine-tuning format
go run convert_logs_for_finetuning.go tool_calls.log finetuning_data.jsonl --min-rating 4

# 3. Review the converted data
head -n 1 finetuning_data.jsonl | jq .

# 4. Use for fine-tuning
# (Follow Qwen fine-tuning documentation)
```

## Troubleshooting

### Old Format Logs

If you have old logs without `user_query` and `model_response` fields, the conversion script will attempt to reconstruct them. However, for best results, use the updated logging system.

### Missing Ratings

Entries without ratings are included by default. Use `--min-rating 1` to exclude unrated examples.

### Large Log Files

The conversion script processes logs line-by-line and can handle large files. For very large files (>100MB), consider splitting them first.

## Next Steps

1. Collect more training data through normal usage
2. Convert logs periodically to update fine-tuning dataset
3. Fine-tune your model with the collected data
4. Evaluate the fine-tuned model and iterate
