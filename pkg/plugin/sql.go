package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
)

var blockedReadOnlyTokens = map[string]struct{}{
	"ALTER":    {},
	"CALL":     {},
	"CREATE":   {},
	"DELETE":   {},
	"DROP":     {},
	"EXEC":     {},
	"EXECUTE":  {},
	"GRANT":    {},
	"INSERT":   {},
	"MERGE":    {},
	"REPLACE":  {},
	"REVOKE":   {},
	"TRUNCATE": {},
	"UPDATE":   {},
}

func interpolateSQL(rawSQL string, query backend.DataQuery) (string, error) {
	sqlQuery := &sqlutil.Query{
		RawSQL:        rawSQL,
		RefID:         query.RefID,
		Interval:      query.Interval,
		TimeRange:     query.TimeRange,
		MaxDataPoints: query.MaxDataPoints,
	}

	return sqlutil.Interpolate(sqlQuery, sqlutil.Macros{
		"timeFilter": irisMacroTimeFilter,
		"timeFrom":   irisMacroTimeFrom,
		"timeTo":     irisMacroTimeTo,
		"timeGroup":  irisMacroTimeGroup,
	})
}

func irisMacroTimeFilter(query *sqlutil.Query, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%w: $__timeFilter expects 1 argument", sqlutil.ErrorBadArgumentCount)
	}

	from := irisTimestampLiteral(query.TimeRange.From)
	to := irisTimestampLiteral(query.TimeRange.To)
	return fmt.Sprintf("%s >= %s AND %s <= %s", args[0], from, args[0], to), nil
}

func irisMacroTimeFrom(query *sqlutil.Query, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%w: $__timeFrom expects 1 argument", sqlutil.ErrorBadArgumentCount)
	}

	return fmt.Sprintf("%s >= %s", args[0], irisTimestampLiteral(query.TimeRange.From)), nil
}

func irisMacroTimeTo(query *sqlutil.Query, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%w: $__timeTo expects 1 argument", sqlutil.ErrorBadArgumentCount)
	}

	return fmt.Sprintf("%s <= %s", args[0], irisTimestampLiteral(query.TimeRange.To)), nil
}

func irisMacroTimeGroup(query *sqlutil.Query, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("%w: $__timeGroup expects 2 arguments", sqlutil.ErrorBadArgumentCount)
	}

	seconds, err := intervalSeconds(args[1], query.Interval)
	if err != nil {
		return "", err
	}

	origin := "{ts '1970-01-01 00:00:00'}"
	return fmt.Sprintf(
		"DATEADD(second, FLOOR(DATEDIFF(second, %s, %s) / %d) * %d, %s)",
		origin,
		args[0],
		seconds,
		seconds,
		origin,
	), nil
}

func irisTimestampLiteral(t time.Time) string {
	return fmt.Sprintf("{ts '%s'}", t.UTC().Format("2006-01-02 15:04:05"))
}

func intervalSeconds(raw string, queryInterval time.Duration) (int64, error) {
	interval := strings.TrimSpace(raw)
	if interval == "$__interval" {
		if queryInterval <= 0 {
			return 0, fmt.Errorf("unsupported $__timeGroup interval %q", raw)
		}
		return int64(queryInterval.Seconds()), nil
	}

	if strings.HasSuffix(interval, "d") {
		value, err := strconv.ParseInt(strings.TrimSuffix(interval, "d"), 10, 64)
		if err != nil || value <= 0 {
			return 0, fmt.Errorf("unsupported $__timeGroup interval %q", raw)
		}
		return value * 24 * 60 * 60, nil
	}

	duration, err := time.ParseDuration(interval)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("unsupported $__timeGroup interval %q", raw)
	}

	seconds := int64(duration.Seconds())
	if seconds <= 0 {
		return 0, fmt.Errorf("unsupported $__timeGroup interval %q", raw)
	}

	switch interval[len(interval)-1] {
	case 's', 'm', 'h':
		return seconds, nil
	default:
		return 0, fmt.Errorf("unsupported $__timeGroup interval %q; use s, m, h, or d", raw)
	}
}

func validateReadOnlySQL(rawSQL string) error {
	tokens, hasMultipleStatements, err := sqlTokens(rawSQL)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return fmt.Errorf("SQL query is required")
	}

	first := strings.ToUpper(tokens[0])
	if first != "SELECT" && first != "WITH" {
		return fmt.Errorf("only read-only SELECT or WITH queries are allowed")
	}
	if hasMultipleStatements {
		return fmt.Errorf("multiple SQL statements are not allowed")
	}

	for _, token := range tokens {
		if _, blocked := blockedReadOnlyTokens[strings.ToUpper(token)]; blocked {
			return fmt.Errorf("SQL keyword %q is not allowed in read-only mode", token)
		}
	}

	return nil
}

func sqlTokens(sql string) ([]string, bool, error) {
	var tokens []string
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false
	statementEnded := false
	hasMultipleStatements := false

	flush := func() {
		if token.Len() == 0 {
			return
		}
		tokens = append(tokens, token.String())
		token.Reset()
	}

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		switch {
		case inLineComment:
			if r == '\n' {
				inLineComment = false
			}
			continue
		case inBlockComment:
			if r == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		case inSingleQuote:
			if r == '\'' {
				if next == '\'' {
					i++
					continue
				}
				inSingleQuote = false
			}
			continue
		case inDoubleQuote:
			if r == '"' {
				if next == '"' {
					i++
					continue
				}
				inDoubleQuote = false
			}
			continue
		}

		if r == '-' && next == '-' {
			flush()
			inLineComment = true
			i++
			continue
		}
		if r == '/' && next == '*' {
			flush()
			inBlockComment = true
			i++
			continue
		}
		if r == '\'' {
			flush()
			inSingleQuote = true
			continue
		}
		if r == '"' {
			flush()
			inDoubleQuote = true
			continue
		}
		if r == ';' {
			flush()
			statementEnded = true
			continue
		}
		if statementEnded && !unicode.IsSpace(r) {
			hasMultipleStatements = true
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '%' || r == '$' {
			token.WriteRune(r)
			continue
		}
		flush()
	}

	if inBlockComment {
		return nil, false, fmt.Errorf("unterminated SQL block comment")
	}
	if inSingleQuote || inDoubleQuote {
		return nil, false, fmt.Errorf("unterminated SQL string or identifier")
	}

	flush()
	return tokens, hasMultipleStatements, nil
}
