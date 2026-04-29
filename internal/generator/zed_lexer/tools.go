package zedlexer

func filterComments(inputTokens []Token) (outputTokens []Token) {
	for _, t := range inputTokens {
		if t.Type == COMMENT {
			continue
		}
		outputTokens = append(outputTokens, t)
	}

	return
}

func filterCaveats(inputTokens []Token) []Token {
	// TODO: add filter to delete Caveat Definitions from tokens set (internal)
	return inputTokens
}

func haveIllegal(inputTokens []Token) (bool, Token) {
	for _, t := range inputTokens {
		if t.Type == ILLEGAL {
			return true, t
		}
	}

	return false, Token{}
}
