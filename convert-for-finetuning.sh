#!/bin/bash

# Helper script to convert tool_calls.log to Qwen fine-tuning format

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="${1:-tool_calls.log}"
OUTPUT_FILE="${2:-finetuning_data.jsonl}"
MIN_RATING="${3:-3}"

if [ ! -f "$LOG_FILE" ]; then
    echo "Error: Log file '$LOG_FILE' not found"
    echo "Usage: $0 [log_file] [output_file] [min_rating]"
    exit 1
fi

echo "🔄 Converting $LOG_FILE to fine-tuning format..."
echo "   Output: $OUTPUT_FILE"
echo "   Min rating: $MIN_RATING+"
echo ""

cd "$SCRIPT_DIR"

go run convert_logs_for_finetuning.go "$LOG_FILE" "$OUTPUT_FILE" --min-rating "$MIN_RATING"

if [ -f "$OUTPUT_FILE" ]; then
    LINE_COUNT=$(wc -l < "$OUTPUT_FILE")
    FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
    echo ""
    echo "✅ Conversion complete!"
    echo "   Output file: $OUTPUT_FILE"
    echo "   Examples: $LINE_COUNT"
    echo "   Size: $FILE_SIZE"
    echo ""
    echo "📋 Preview (first example):"
    head -n 1 "$OUTPUT_FILE" | jq . 2>/dev/null || head -n 1 "$OUTPUT_FILE"
fi
