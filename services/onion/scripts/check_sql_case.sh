#!/bin/bash
# ============================================================================
#  check_sql_case.sh - SQL keywords must be uppercase in sqlc query files.
#
#  WHAT WAS WRONG WITH THE PREVIOUS VERSION
#
#  It failed `make build` on main over correct SQL. Every flagged occurrence was
#  a column alias, not a keyword:
#
#      SELECT COUNT(*) AS count ...      ORDER BY count DESC
#
#  The COUNT() call is already uppercase. `count` is the alias. But COUNT is in
#  the keyword list and the test was a bare \bcount\b, so the alias matched.
#  Three separate defects:
#
#    1. Identifiers sharing a name with a keyword were reported as lowercase
#       keywords. The list holds AS, COUNT, END, ON, NOW, INTERVAL, TABLE, VIEW,
#       CASE, SUM, MIN, MAX, INDEX and UNION, all plausible column names.
#
#    2. The mixed-case detector reported the same thing again. It ran
#           grep -oi "\bCOUNT\b" | grep -v "^COUNT$"
#       which matches the lowercase alias too, because -i is case-insensitive
#       and only the exact uppercase form was excluded.
#
#    3. It re-ran perl once per keyword per file. 60 keywords over 143 files is
#       more than 8,000 perl processes; the script takes minutes and timed out
#       when run under a tool with a request deadline.
#
#  WHAT THIS VERSION DOES DIFFERENTLY
#
#  Normalises each file once, collects the aliases the file itself declares, and
#  removes only their lowercase occurrences. An uppercase COUNT( survives and is
#  still checked; a lowercase `count` the file declared via `AS count` does not.
#  Line numbers survive normalisation and are reported, because "found lowercase:
#  and" with no location is not actionable.
#
#  KNOWN AND ACCEPTED LIMIT
#  If a file declares `AS count` and also genuinely writes lowercase `count(*)`,
#  the second is missed. Telling them apart needs a real parser. The trade is
#  deliberate: failing every local build over correct SQL is worse than missing a
#  rare slip in a file that already aliases that word.
# ============================================================================
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

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

echo -e "${BLUE}Checking SQL keyword case in sqlc query files${NC}"

# Tracked files only. Untracked agent worktrees and local scratch directories can
# hold stale copies that must not affect a repo build.
SQL_FILES=()
while IFS= read -r f; do SQL_FILES+=("$f"); done < <(git ls-files '*.sql')
if (( ${#SQL_FILES[@]} == 0 )); then
    echo -e "${YELLOW}no tracked .sql files${NC}"; exit 0
fi

tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
found_issues=false
checked=0

for file in "${SQL_FILES[@]}"; do
    norm="$tmp/norm.sql"

    # One pass. Comments, string literals and quoted identifiers go, but their
    # newlines stay, so reported line numbers still match the real file.
    perl -0777 -pe '
        s{--[^\n]*}{}g;
        s{^[ \t]*#[^\n]*}{}gm;
        s{/\*(.*?)\*/}{"\n" x (($1 =~ tr/\n//))}gse;
        s{'\''((?:'\'''\''|[^'\''])*)'\''}{"\n" x (($1 =~ tr/\n//))}gse;
        s{"((?:""|[^"])*)"}{"\n" x (($1 =~ tr/\n//))}gse;
        s{\$\d+}{}g;
    ' "$file" > "$norm"

    # Aliases this file declares. In this file their lowercase form is an
    # identifier, whatever the keyword list says.
    while IFS= read -r alias; do
        [[ -n $alias ]] || continue
        perl -i -pe "s/\\b${alias}\\b//g" "$norm"
    done < <(perl -0777 -ne 'while(/\bAS\s+("?)(\w+)\1/gi){print lc($2),"\n"}' "$norm" | sort -u)

    file_issues=""
    for keyword in "${SQL_KEYWORDS[@]}"; do
        # One check covers lowercase and mixed case: any case-insensitive hit that
        # is not exactly the uppercase form is a violation. Nothing is reported
        # twice, unlike the two overlapping checks this replaces.
        while IFS=: read -r lineno hit; do
            [[ -n ${hit:-} ]] || continue
            [[ $hit == "$keyword" ]] && continue
            # Numeric prefix so the report sorts by line, stripped before
            # printing. `sort -u -t: -k1,1n` was wrong: restricting the sort key
            # makes -u dedupe on that key alone, so every violation after the
            # first on any given line was silently dropped.
            file_issues+="$(printf '%06d' "$lineno")|    line ${lineno}: '${hit}' should be '${keyword}'"$'\n'
        done < <(grep -noiE "\\b${keyword}\\b" "$norm" 2>/dev/null || true)
    done

    checked=$((checked + 1))
    if [[ -n $file_issues ]]; then
        found_issues=true
        echo -e "${RED}${file}${NC}"
        printf '%s' "$file_issues" | sort -u | cut -d'|' -f2-
    fi
done

echo
if [[ $found_issues == true ]]; then
    echo -e "${RED}SQL keyword case violations found${NC}"
    echo -e "${YELLOW}Uppercase the keywords listed above. Column aliases are exempt,${NC}"
    echo -e "${YELLOW}so anything reported here is a genuine keyword.${NC}"
    exit 1
fi
echo -e "${GREEN}All SQL keywords correctly cased across ${checked} file(s)${NC}"
