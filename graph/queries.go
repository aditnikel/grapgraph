package graph

func OutgoingTransfers(c *Client, accountID string) (any, error) {
	query := `
	MATCH (a:Account {id: '` + accountID + `'})-[t:TRANSFER]->(b:Account)
	RETURN t.tx_id, t.amount, b.id, t.ts
	ORDER BY t.ts
	`
	return c.Query(query)
}

func TraceMoney(c *Client, accountID string) (any, error) {
	query := `
	MATCH path = (a:Account {id: '` + accountID + `'})-[:TRANSFER*1..4]->(b)
	RETURN path
	`
	return c.Query(query)
}
