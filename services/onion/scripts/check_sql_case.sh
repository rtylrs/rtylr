#!/bin/bash
set -e

# SQL Case Checking Script
# This script checks SQL keywords are uppercase in SQLC query files

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# SQL keywords to check (common ones)
SQL_KEYWORDS=(
    "SELECT" "INSERT" "UPDATE" "DELETE" "FROM" "WHERE" "AND" "OR" "NOT"
    "ORDER BY" "GROUP BY" "HAVING" "LIMIT" "OFFSET" "JOIN" "LEFT JOIN" "RIGHT JOIN" "INNER JOIN"
    "ON" "AS" "IN" "EXISTS" "BETWEEN" "LIKE" "IS NULL" "IS NOT NULL"
    "COUNT" "SUM" "AVG" "MAX" "MIN" "DISTINCT" "UNION" "UNION ALL"
    "WITH" "CASE" "WHEN" "THEN" "ELSE" "END" "RETURNING"
    "CREATE" "ALTER" "DROP" "INDEX" "TABLE" "VIEW" "FUNCTION" "PROCEDURE"
    "TRANSACTION" "COMMIT" "ROLLBACK" "SAVEPOINT"
    "EXCLUDED" "CONFLICT" "DO UPDATE SET" "NOW" "INTERVAL"
)

echo -e "${BLUE}🔍 Checking SQL keyword case in SQLC query files...${NC}"

# Find tracked SQL files only. Untracked agent worktrees and local scratch
# directories can contain stale copies that should not affect repo builds.
SQL_FILES=$(git ls-files '*.sql')

found_issues=false

for file in $SQL_FILES; do
    echo -e "${YELLOW}Checking: $file${NC}"
    file_has_issues=false
    
    # Check each SQL keyword
    for keyword in "${SQL_KEYWORDS[@]}"; do
        # Convert keyword to lowercase for search
        lowercase_keyword=$(echo "$keyword" | tr '[:upper:]' '[:lower:]')
        
        # Search for non-uppercase versions of the keyword, but exclude comments,
        # string literals, and quoted identifiers. Seed data and enum values often
        # contain ordinary words like "and", "in", or "drop" that are not SQL.
        processed_content=$(perl -0777 -pe 's/--[^\n]*//g; s{/\*.*?\*/}{}gs; s/'\''(?:'\'''\''|[^'\''])*'\''//g; s/"(?:""|[^"])*"//g; s/`(?:``|[^`])*`//g' "$file")
        
        # Check for lowercase version
        if echo "$processed_content" | grep -q "\b${lowercase_keyword}\b"; then
            echo -e "  ${RED}Found lowercase: $lowercase_keyword (should be: $keyword)${NC}"
            file_has_issues=true
            found_issues=true
        fi
        
        # Check for mixed case versions (like wHERE, WhErE, etc.)
        # Use case-insensitive search to find the keyword, then check if it's not exactly uppercase
        found_mixed=$(echo "$processed_content" | grep -oi "\b${keyword}\b" | grep -v "^${keyword}$" | head -1)
        if [ -n "$found_mixed" ]; then
            echo -e "  ${RED}Found mixed case: $found_mixed (should be: $keyword)${NC}"
            file_has_issues=true
            found_issues=true
        fi
    done
    
    echo
done

if [ "$found_issues" = true ]; then
    echo -e "${RED}❌ Found SQL keywords that should be uppercase${NC}"
    echo -e "${YELLOW}To fix, update the SQL keywords to uppercase in the files above${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All SQL keywords are properly cased!${NC}"
fi
