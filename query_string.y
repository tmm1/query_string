%{
package querystr

import(
    "github.com/blugelabs/bluge"
)
%}

%union {
s string
n int
f float64
q bluge.Query
pf *float64
}

%token tSTRING tPHRASE tPLUS tMINUS tCOLON tBOOST tNUMBER tGREATER tLESS tEQUAL tTILDE tQUESTION

%type <s>                tSTRING
%type <s>                tPHRASE
%type <s>                tNUMBER
%type <s>                posOrNegNumber
%type <s>                tTILDE
%type <s>                tBOOST
%type <q>                searchBase
%type <pf>               searchSuffix
%type <n>                searchPrefix

%%

input:
searchParts {
	yylex.(*lexerWrapper).logDebugGrammarf("INPUT")
};

searchParts:
searchPart searchParts {
	yylex.(*lexerWrapper).logDebugGrammarf("SEARCH PARTS")
}
|
searchPart {
	yylex.(*lexerWrapper).logDebugGrammarf("SEARCH PART")
};

searchPart:
searchPrefix searchBase searchSuffix {
    q := $2
    if $3 != nil {
        var err error
        q, err = queryStringSetBoost($2, *$3)
        if err != nil {
          yylex.(*lexerWrapper).lex.Error(err.Error())
        }
    }
	switch($1) {
		case queryShould:
			yylex.(*lexerWrapper).query.AddShould(q)
		case queryMust:
			yylex.(*lexerWrapper).query.AddMust(q)
		case queryMustNot:
			yylex.(*lexerWrapper).query.AddMustNot(q)
	}
};


searchPrefix:
/* empty */ {
	$$ = queryMust
}
|
tQUESTION {
	yylex.(*lexerWrapper).logDebugGrammarf("QUESTION")
	$$ = queryShould
}
|
tPLUS {
	yylex.(*lexerWrapper).logDebugGrammarf("PLUS")
	$$ = queryMust
}
|
tMINUS {
	yylex.(*lexerWrapper).logDebugGrammarf("MINUS")
	$$ = queryMustNot
};

searchBase:
tSTRING {
    yylex.(*lexerWrapper).logDebugGrammarf("STRING - %s", $1)
	$$ = queryStringStringToken(yylex, "", $1)
}
|
tSTRING tTILDE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FUZZY STRING - %s %s", $1, $2)
	q, err := queryStringStringTokenFuzzy("", $1, $2)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tSTRING tTILDE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - %s FUZZY STRING - %s %s", $1, $3, $4)
    q, err := queryStringStringTokenFuzzy(w.fieldname($1), $3, $4)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tNUMBER {
	yylex.(*lexerWrapper).logDebugGrammarf("STRING - %s", $1)
	q, err := queryStringNumberToken("", $1)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tPHRASE {
	yylex.(*lexerWrapper).logDebugGrammarf("PHRASE - %s", $1)
	$$ = queryStringPhraseToken(yylex, "", $1)
}
|
tSTRING tCOLON tSTRING {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - %s STRING - %s", $1, $3)
	$$ = queryStringStringToken(yylex, w.fieldname($1), $3)
}
|
tSTRING tCOLON posOrNegNumber {
	w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - %s STRING - %s", $1, $3)
	q, err := queryStringNumberToken(w.fieldname($1), $3)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tPHRASE {
	w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - %s PHRASE - %s", $1, $3)
	$$ = queryStringPhraseToken(yylex, w.fieldname($1), $3)
}
|
tSTRING tCOLON tGREATER posOrNegNumber {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - GREATER THAN %s", $4)
	q, err := queryStringNumericRangeGreaterThanOrEqual(w.fieldname($1), $4, false)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tGREATER tEQUAL posOrNegNumber {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - GREATER THAN OR EQUAL %s", $5)
    q, err := queryStringNumericRangeGreaterThanOrEqual(w.fieldname($1), $5, true)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
    $$ = q
}
|
tSTRING tCOLON tLESS posOrNegNumber {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - LESS THAN %s", $4)
    q, err := queryStringNumericRangeLessThanOrEqual(w.fieldname($1), $4, false)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
    $$ = q
}
|
tSTRING tCOLON tLESS tEQUAL posOrNegNumber {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - LESS THAN OR EQUAL %s", $5)
    q, err := queryStringNumericRangeLessThanOrEqual(w.fieldname($1), $5, true)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
    $$ = q
}
|
tSTRING tCOLON tGREATER tPHRASE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - GREATER THAN DATE %s", $4)
	q, err := queryStringDateRangeGreaterThanOrEqual(yylex, w.fieldname($1), $4, false)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tGREATER tEQUAL tPHRASE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - GREATER THAN OR EQUAL DATE %s", $5)
    q, err := queryStringDateRangeGreaterThanOrEqual(yylex, w.fieldname($1), $5, true)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tLESS tPHRASE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - LESS THAN DATE %s", $4)
    q, err := queryStringDateRangeLessThanOrEqual(yylex, w.fieldname($1), $4, false)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
}
|
tSTRING tCOLON tLESS tEQUAL tPHRASE {
    w := yylex.(*lexerWrapper)
    w.logDebugGrammarf("FIELD - LESS THAN OR EQUAL DATE %s", $5)
    q, err := queryStringDateRangeLessThanOrEqual(yylex, w.fieldname($1), $5, true)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    }
	$$ = q
};

searchSuffix:
/* empty */ {
	$$ = nil
}
|
tBOOST {
    $$ = nil
    yylex.(*lexerWrapper).logDebugGrammarf("BOOST %s", $1)
    boost, err := queryStringParseBoost($1)
    if err != nil {
      yylex.(*lexerWrapper).lex.Error(err.Error())
    } else {
        $$ = &boost
    }
};

posOrNegNumber:
tNUMBER {
	$$ = $1
}
|
tMINUS tNUMBER {
	$$ = "-" + $2
};
