package graph

func SeedMoneyGraph(c *Client) error {
	query := `
	CREATE
	  // Users
	  (alice:User {id: 'u1', name: 'Alice'}),
	  (bob:User   {id: 'u2', name: 'Bob'}),
	  (carol:User {id: 'u3', name: 'Carol'}),
	  (dave:User  {id: 'u4', name: 'Dave'}),

	  // Accounts
	  (aliceChk:Account {id: 'a1', owner: 'Alice', type: 'checking'}),
	  (aliceSav:Account {id: 'a2', owner: 'Alice', type: 'savings'}),
	  (bobChk:Account   {id: 'a3', owner: 'Bob',   type: 'checking'}),
	  (carolChk:Account {id: 'a4', owner: 'Carol', type: 'checking'}),
	  (daveChk:Account  {id: 'a5', owner: 'Dave',  type: 'checking'}),

	  (shopAcc:Account  {id: 'a6', owner: 'Acme Store', type: 'merchant'}),
	  (casinoAcc:Account {id: 'a7', owner: 'Lucky Casino', type: 'merchant'}),

	  // Merchants
	  (shop:Merchant {id: 'm1', name: 'Acme Store'}),
	  (casino:Merchant {id: 'm2', name: 'Lucky Casino'}),

	  // Ownership
	  (alice)-[:OWNS]->(aliceChk),
	  (alice)-[:OWNS]->(aliceSav),
	  (bob)-[:OWNS]->(bobChk),
	  (carol)-[:OWNS]->(carolChk),
	  (dave)-[:OWNS]->(daveChk),

	  (shopAcc)-[:BELONGS_TO]->(shop),
	  (casinoAcc)-[:BELONGS_TO]->(casino),

	  // Transfers — normal behavior
	  (aliceChk)-[:TRANSFER {
	    tx_id: 'tx2001',
	    amount: 500.00,
	    currency: 'USD',
	    ts: 1700000000
	  }]->(aliceSav),

	  (aliceChk)-[:TRANSFER {
	    tx_id: 'tx2002',
	    amount: 120.00,
	    currency: 'USD',
	    ts: 1700000300
	  }]->(shopAcc),

	  (bobChk)-[:TRANSFER {
	    tx_id: 'tx2003',
	    amount: 75.00,
	    currency: 'USD',
	    ts: 1700000500
	  }]->(shopAcc),

	  // Peer-to-peer
	  (aliceSav)-[:TRANSFER {
	    tx_id: 'tx2004',
	    amount: 200.00,
	    currency: 'USD',
	    ts: 1700000800
	  }]->(bobChk),

	  (bobChk)-[:TRANSFER {
	    tx_id: 'tx2005',
	    amount: 200.00,
	    currency: 'USD',
	    ts: 1700000900
	  }]->(carolChk),

	  // Suspicious pass-through
	  (carolChk)-[:TRANSFER {
	    tx_id: 'tx2006',
	    amount: 200.00,
	    currency: 'USD',
	    ts: 1700001000
	  }]->(daveChk),

	  // Merchant + gambling
	  (daveChk)-[:TRANSFER {
	    tx_id: 'tx2007',
	    amount: 190.00,
	    currency: 'USD',
	    ts: 1700001200
	  }]->(casinoAcc),

	  (casinoAcc)-[:TRANSFER {
	    tx_id: 'tx2008',
	    amount: 180.00,
	    currency: 'USD',
	    ts: 1700002000
	  }]->(daveChk)
	`

	_, err := c.Query(query)
	return err
}
