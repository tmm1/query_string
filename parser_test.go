//  Copyright (c) 2020 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package querystr

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/blugelabs/bluge"
)

func baseQuery() *bluge.BooleanQuery {
	return bluge.NewBooleanQuery().SetMinShould(1)
}

func matchQuery(q string) *bluge.MatchQuery {
	return bluge.NewMatchQuery(q).SetOperator(bluge.MatchQueryOperatorAnd)
}

func TestQuerySyntaxParserValid(t *testing.T) {
	theDate, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input  string
		result bluge.Query
	}{
		{
			input: "test",
			result: baseQuery().
				AddMust(matchQuery("test")),
		},
		{
			input: "127.0.0.1",
			result: baseQuery().
				AddMust(matchQuery("127.0.0.1")),
		},
		{
			input: `"test phrase 1"`,
			result: baseQuery().
				AddMust(bluge.NewBooleanQuery().AddShould(
					bluge.NewMatchPhraseQuery("test phrase 1"),
					bluge.NewTermQuery("test phrase 1"),
				).SetMinShould(1)),
		},
		{
			input: "field:test",
			result: baseQuery().
				AddMust(matchQuery("test").SetField("field")),
		},
		// - is allowed inside a term, just not the start
		{
			input: "field:t-est",
			result: baseQuery().
				AddMust(matchQuery("t-est").SetField("field")),
		},
		// + is allowed inside a term, just not the start
		{
			input: "field:t+est",
			result: baseQuery().
				AddMust(matchQuery("t+est").SetField("field")),
		},
		// > is allowed inside a term, just not the start
		{
			input: "field:t>est",
			result: baseQuery().
				AddMust(matchQuery("t>est").SetField("field")),
		},
		// < is allowed inside a term, just not the start
		{
			input: "field:t<est",
			result: baseQuery().
				AddMust(matchQuery("t<est").SetField("field")),
		},
		// = is allowed inside a term, just not the start
		{
			input: "field:t=est",
			result: baseQuery().
				AddMust(matchQuery("t=est").SetField("field")),
		},
		{
			input: "+field1:test1",
			result: baseQuery().
				AddMust(matchQuery("test1").SetField("field1")),
		},
		{
			input: "-field2:test2",
			result: baseQuery().
				AddMustNot(matchQuery("test2").SetField("field2")),
		},
		{
			input: `field3:"test phrase 2"`,
			result: baseQuery().
				AddMust(bluge.NewMatchPhraseQuery("test phrase 2").SetField("field3")),
		},
		{
			input: `+field4:"test phrase 1"`,
			result: baseQuery().
				AddMust(bluge.NewMatchPhraseQuery("test phrase 1").SetField("field4")),
		},
		{
			input: `-field5:"test phrase 2"`,
			result: baseQuery().
				AddMustNot(bluge.NewMatchPhraseQuery("test phrase 2").SetField("field5")),
		},
		{
			input: `+field6:test3 -field7:test4 ?field8:test5`,
			result: baseQuery().
				AddMust(matchQuery("test3").SetField("field6")).
				AddShould(matchQuery("test5").SetField("field8")).
				AddMustNot(matchQuery("test4").SetField("field7")),
		},
		{
			input: "test^3",
			result: baseQuery().
				AddMust(matchQuery("test").SetBoost(3.0)),
		},
		{
			input: "test^3 other^6",
			result: baseQuery().
				AddMust(matchQuery("test").SetBoost(3.0)).
				AddMust(matchQuery("other").SetBoost(6.0)),
		},
		{
			input: "33",
			result: baseQuery().
				AddMust(
					bluge.NewBooleanQuery().SetMinShould(1).
						AddShould(bluge.NewMatchQuery("33")).
						AddShould(
							bluge.NewNumericRangeInclusiveQuery(33.0, 33.0,
								true, true))),
		},
		{
			input: "field:33",
			result: baseQuery().
				AddMust(
					bluge.NewBooleanQuery().SetMinShould(1).
						AddShould(bluge.NewMatchQuery("33").SetField("field")).
						AddShould(
							bluge.NewNumericRangeInclusiveQuery(33.0, 33.0,
								true, true).
								SetField("field"))),
		},
		{
			input: "cat-dog",
			result: baseQuery().
				AddMust(matchQuery("cat-dog")),
		},
		{
			input: "watex~",
			result: baseQuery().
				AddMust(bluge.NewMatchQuery("watex").SetFuzziness(1)),
		},
		{
			input: "watex~2",
			result: baseQuery().
				AddMust(bluge.NewMatchQuery("watex").SetFuzziness(2)),
		},
		{
			input: "watex~ 2",
			result: baseQuery().
				AddMust(bluge.NewMatchQuery("watex").SetFuzziness(1)).
				AddMust(bluge.NewBooleanQuery().SetMinShould(1).
					AddShould(bluge.NewMatchQuery("2")).
					AddShould(
						bluge.NewNumericRangeInclusiveQuery(2.0, 2.0, true, true))),
		},
		{
			input: "?field:watex~",
			result: baseQuery().
				AddShould(
					bluge.NewMatchQuery("watex").
						SetFuzziness(1).
						SetField("field")),
		},
		{
			input: "field:watex~2",
			result: baseQuery().
				AddMust(bluge.NewMatchQuery("watex").SetFuzziness(2).SetField("field")),
		},
		{
			input: `field:555c3bb06f7a127cda000005`,
			result: baseQuery().
				AddMust(matchQuery("555c3bb06f7a127cda000005").SetField("field")),
		},
		{
			input: `field:>5`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(5.0, bluge.MaxNumeric, false, true).
						SetField("field")),
		},
		{
			input: `field:>=5`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(5.0, bluge.MaxNumeric, true, true).
						SetField("field")),
		},
		{
			input: `field:<5`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(bluge.MinNumeric, 5.0, true, false).
						SetField("field")),
		},
		{
			input: `field:<=5`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(bluge.MinNumeric, 5.0, true, true).
						SetField("field")),
		},
		{
			input: `field:true`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(1.0, 1.0, true, true).
						SetField("field")),
		},
		{
			input: `field:false`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(0.0, 0.0, true, true).
						SetField("field")),
		},
		// new range tests with negative number
		{
			input: "field:-5",
			result: baseQuery().
				AddMust(
					bluge.NewBooleanQuery().SetMinShould(1).
						AddShould(
							bluge.NewMatchQuery("-5").SetField("field")).
						AddShould(
							bluge.NewNumericRangeInclusiveQuery(-5.0, -5.0, true, true).
								SetField("field"))),
		},
		{
			input: `field:>-5`,
			result: baseQuery().
				AddMust(
					bluge.NewNumericRangeInclusiveQuery(-5.0, bluge.MaxNumeric, false, true).
						SetField("field")),
		},
		{
			input: `field:>=-5`,
			result: baseQuery().
				AddMust(bluge.NewNumericRangeInclusiveQuery(-5.0, bluge.MaxNumeric, true, true).
					SetField("field")),
		},
		{
			input: `field:<-5`,
			result: baseQuery().
				AddMust(bluge.NewNumericRangeInclusiveQuery(bluge.MinNumeric, -5.0, true, false).
					SetField("field")),
		},
		{
			input: `field:<=-5`,
			result: baseQuery().
				AddMust(bluge.NewNumericRangeInclusiveQuery(bluge.MinNumeric, -5.0, true, true).
					SetField("field")),
		},
		{
			input: `field:>"2006-01-02T15:04:05Z"`,
			result: baseQuery().
				AddMust(bluge.NewDateRangeInclusiveQuery(theDate, time.Time{}, false, true).
					SetField("field")),
		},
		{
			input: `field:>="2006-01-02T15:04:05Z"`,
			result: baseQuery().
				AddMust(bluge.NewDateRangeInclusiveQuery(theDate, time.Time{}, true, true).
					SetField("field")),
		},
		{
			input: `field:<"2006-01-02T15:04:05Z"`,
			result: baseQuery().
				AddMust(bluge.NewDateRangeInclusiveQuery(time.Time{}, theDate, true, false).
					SetField("field")),
		},
		{
			input: `field:<="2006-01-02T15:04:05Z"`,
			result: baseQuery().
				AddMust(bluge.NewDateRangeInclusiveQuery(time.Time{}, theDate, true, true).
					SetField("field")),
		},
		{
			input: `/mar.*ty/`,
			result: baseQuery().
				AddMust(bluge.NewRegexpQuery("mar.*ty")),
		},
		{
			input: `name:/mar.*ty/`,
			result: baseQuery().
				AddMust(bluge.NewRegexpQuery("mar.*ty").
					SetField("name")),
		},
		{
			input: `mart*`,
			result: baseQuery().
				AddMust(bluge.NewWildcardQuery("mart*")),
		},
		{
			input: `name:mart*`,
			result: baseQuery().
				AddMust(bluge.NewWildcardQuery("mart*").
					SetField("name")),
		},

		// tests for escaping

		// escape : as field delimeter
		{
			input: `name\:marty`,
			result: baseQuery().
				AddMust(matchQuery("name:marty")),
		},
		// first colon delimiter, second escaped
		{
			input: `name:marty\:couchbase`,
			result: baseQuery().
				AddMust(matchQuery("marty:couchbase").
					SetField("name")),
		},
		// escape space, single arguemnt to match query
		{
			input: `marty\ couchbase`,
			result: baseQuery().
				AddMust(matchQuery("marty couchbase")),
		},
		// escape leading plus, not a must clause
		{
			input: `\+marty`,
			result: baseQuery().
				AddMust(matchQuery("+marty")),
		},
		// escape leading minus, not a must not clause
		{
			input: `\-marty`,
			result: baseQuery().
				AddMust(matchQuery("-marty")),
		},
		// escape quote inside of phrase
		{
			input: `field:"what does \"quote\" mean"`,
			result: baseQuery().
				AddMust(bluge.NewMatchPhraseQuery(`what does "quote" mean`).SetField("field")),
		},
		// escaping an unsupported character retains backslash
		{
			input: `can\ i\ escap\e`,
			result: baseQuery().
				AddMust(matchQuery(`can i escap\e`)),
		},
		// leading spaces
		{
			input: `   what`,
			result: baseQuery().
				AddMust(matchQuery(`what`)),
		},
		// no boost value defaults to 1
		{
			input: `term^`,
			result: baseQuery().
				AddMust(matchQuery(`term`).
					SetBoost(1.0)),
		},
		// weird lexer cases, something that starts like a number
		// but contains escape and ends up as string
		{
			input: `3.0\:`,
			result: baseQuery().
				AddMust(matchQuery(`3.0:`)),
		},
		{
			input: `3.0\a`,
			result: baseQuery().
				AddMust(matchQuery(`3.0\a`)),
		},
		// implicit phrases
		{
			input: "animated scifi",
			result: baseQuery().
				AddMust(matchQuery("animated scifi")),
		},
		{
			input: "animated scifi Tag:test comedy movies",
			result: baseQuery().AddMust(
				matchQuery("animated scifi"),
				matchQuery("test").SetField("Tag"),
				matchQuery("comedy movies")),
		},
		{
			input: "animated scifi ?Tag:test comedy movies",
			result: baseQuery().AddMust(
				matchQuery("animated scifi comedy movies"),
			).AddShould(
				matchQuery("test").SetField("Tag"),
			),
		},
	}

	opts := DefaultOptions().WithLogger(log.Default())
	// opts = opts.WithDebugLexer(true)
	// opts = opts.WithDebugParser(true)
	for _, test := range tests {
		q, err := ParseQueryString(test.input, opts)
		if err != nil {
			t.Errorf("error parsing query `%s`: %v", test.input, err)
		}
		if !reflect.DeepEqual(q, test.result) {
			t.Errorf("\nExpected: %s\n     got: %s\n     for: %s", queryInfo(test.result), queryInfo(q), test.input)
		}
	}
}

func querySliceInfo(s []bluge.Query) string {
	out := "["
	for _, q := range s {
		out += " " + queryInfo(q)
	}
	if len(s) > 0 {
		out += " "
	}
	out += "]"
	return out
}

func queryInfo(q bluge.Query) string {
	if bq, ok := q.(*bluge.BooleanQuery); ok {
		return fmt.Sprintf("BooleanQuery{ must=%s should=%s mustnot=%s minShould=%v }",
			querySliceInfo(bq.Musts()),
			querySliceInfo(bq.Shoulds()),
			querySliceInfo(bq.MustNots()),
			bq.MinShould(),
		)
	}
	return fmt.Sprintf("%#v", q)
}

func TestQuerySyntaxParserInvalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"^"},
		{"^5"},
		{"field:-text"},
		{"field:+text"},
		{"field:>text"},
		{"field:>=text"},
		{"field:<text"},
		{"field:<=text"},
		{"field:~text"},
		{"field:^text"},
		{"field::text"},
		{`"this is the time`},
		{`cat^3\:`},
		{`cat^3\0`},
		{`cat~3\:`},
		{`cat~3\0`},
		{strings.Repeat(`9`, 369)},
		{`field:` + strings.Repeat(`9`, 369)},
		{`field:>` + strings.Repeat(`9`, 369)},
		{`field:>=` + strings.Repeat(`9`, 369)},
		{`field:<` + strings.Repeat(`9`, 369)},
		{`field:<=` + strings.Repeat(`9`, 369)},
	}

	for _, test := range tests {
		_, err := ParseQueryString(test.input, DefaultOptions())
		if err == nil {
			t.Errorf("expected error, got nil for `%s`", test.input)
		}
	}
}

func TestQueryOptionTermFields(t *testing.T) {
	input := "+Field:term"
	opts := DefaultOptions().WithTermFields(map[string]bool{
		"Field": true,
	})
	q, err := ParseQueryString(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	result := baseQuery().AddMust(bluge.NewTermQuery("term").SetField("Field"))
	if !reflect.DeepEqual(q, result) {
		t.Errorf("\nExpected: %s\n     got: %s\n     for: %s", queryInfo(result), queryInfo(q), input)
	}
}

func TestQueryOptionLowerFields(t *testing.T) {
	input := "+Field:term"
	opts := DefaultOptions().WithLowercaseFields()
	q, err := ParseQueryString(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	result := baseQuery().AddMust(matchQuery("term").SetField("field"))
	if !reflect.DeepEqual(q, result) {
		t.Errorf("\nExpected: %s\n     got: %s\n     for: %s", queryInfo(result), queryInfo(q), input)
	}
}

var extTokenTypes []int
var extTokens []yySymType

func BenchmarkLexer(b *testing.B) {

	for n := 0; n < b.N; n++ {
		var tokenTypes []int
		var tokens []yySymType
		r := strings.NewReader(`+field4:"test phrase 1"`)
		l := newQueryStringLex(r, DefaultOptions())
		var lval yySymType
		rv := l.Lex(&lval)
		for rv > 0 {
			tokenTypes = append(tokenTypes, rv)
			tokens = append(tokens, lval)
			lval.s = ""
			lval.n = 0
			rv = l.Lex(&lval)
		}
		extTokenTypes = tokenTypes
		extTokens = tokens
	}

}
