#!/bin/bash

# Script to create all 68 GitHub issues from backlog.md

REPO="AINative-Studio/ainative-code"
BACKLOG_FILE="/Users/aideveloper/AINative-Code/backlog.md"

# Read the backlog file and parse tasks
# We'll process the file line by line and extract task information

task_number=""
task_title=""
priority=""
effort=""
dependencies=""
labels=""
description=""
acceptance_criteria=""
in_task=false
in_description=false
in_acceptance=false

while IFS= read -r line; do
    # Detect task header (### TASK-XXX: Title)
    if [[ $line =~ ^###[[:space:]]TASK-([0-9]+):[[:space:]](.+)$ ]]; then
        # If we have a previous task, create the issue
        if [ -n "$task_number" ]; then
            # Construct issue body
            issue_body="**Priority**: $priority
**Effort**: $effort
**Dependencies**: $dependencies

## Description

$description

## Acceptance Criteria

$acceptance_criteria"

            # Create the issue
            echo "Creating issue for TASK-$task_number..."
            gh issue create --repo "$REPO" \
                --title "TASK-$task_number: $task_title" \
                --body "$issue_body" \
                --label "$labels" 2>&1
        fi

        # Reset for new task
        task_number="${BASH_REMATCH[1]}"
        task_title="${BASH_REMATCH[2]}"
        priority=""
        effort=""
        dependencies=""
        labels=""
        description=""
        acceptance_criteria=""
        in_task=true
        in_description=false
        in_acceptance=false
        continue
    fi

    # Only process lines if we're in a task
    if [ "$in_task" = true ]; then
        # Extract Priority
        if [[ $line =~ ^\*\*Priority\*\*:[[:space:]](.+)$ ]]; then
            priority="${BASH_REMATCH[1]}"
        fi

        # Extract Effort
        if [[ $line =~ ^\*\*Effort\*\*:[[:space:]](.+)$ ]]; then
            effort="${BASH_REMATCH[1]}"
        fi

        # Extract Dependencies
        if [[ $line =~ ^\*\*Dependencies\*\*:[[:space:]](.+)$ ]]; then
            dependencies="${BASH_REMATCH[1]}"
        fi

        # Extract Labels
        if [[ $line =~ ^\*\*Labels\*\*:[[:space:]](.+)$ ]]; then
            # Remove backticks and convert to comma-separated
            labels=$(echo "${BASH_REMATCH[1]}" | sed 's/`//g' | tr ',' '\n' | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | paste -sd ',' -)
        fi

        # Detect Description section
        if [[ $line =~ ^\*\*Description\*\*: ]]; then
            in_description=true
            in_acceptance=false
            continue
        fi

        # Detect Acceptance Criteria section
        if [[ $line =~ ^\*\*Acceptance[[:space:]]Criteria\*\*: ]]; then
            in_acceptance=true
            in_description=false
            continue
        fi

        # Collect description text
        if [ "$in_description" = true ] && [ -n "$line" ]; then
            description="$description$line"$'\n'
        fi

        # Collect acceptance criteria text
        if [ "$in_acceptance" = true ] && [ -n "$line" ]; then
            acceptance_criteria="$acceptance_criteria$line"$'\n'
        fi
    fi
done < "$BACKLOG_FILE"

# Create the last task
if [ -n "$task_number" ]; then
    issue_body="**Priority**: $priority
**Effort**: $effort
**Dependencies**: $dependencies

## Description

$description

## Acceptance Criteria

$acceptance_criteria"

    echo "Creating issue for TASK-$task_number..."
    gh issue create --repo "$REPO" \
        --title "TASK-$task_number: $task_title" \
        --body "$issue_body" \
        --label "$labels" 2>&1
fi

echo "All issues created successfully!"
