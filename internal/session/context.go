package session

const charsPerToken = 2.5

func (sess *Session) PrepareContext(maxCtxTokens int) []Message {
	maxChars := int(float64(maxCtxTokens) * charsPerToken)
	n := len(sess.Messages)
	if n == 0 { return nil }
	var ctx []Message
	currentChars := 0
	for i := n - 1; i >= 0; i-- {
		msg := sess.Messages[i]
		chars := len(msg.Content)
		if msg.Role == "system" && i == 0 {
			ctx = append(ctx, msg)
			currentChars += chars
			continue
		}
		if currentChars+chars > maxChars { break }
		ctx = append(ctx, msg)
		currentChars += chars
	}
	for i, j := 0, len(ctx)-1; i < j; i, j = i+1, j-1 {
		ctx[i], ctx[j] = ctx[j], ctx[i]
	}
	return ctx
}
