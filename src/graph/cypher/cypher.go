package cypher

const AddTransfer = `
MERGE (from:Account {id:$from})
MERGE (to:Account {id:$to})
MERGE (t:Transaction {id:$tx})
ON CREATE SET
  t.amount = $amount,
  t.ts = $ts
MERGE (from)-[:SENT]->(t)-[:RECEIVED]->(to)
`

const WindowedBFS = `
MATCH p=(a:Account {id:$start})-[:SENT]->(:Transaction)-[:RECEIVED]->(b)
WHERE ALL(rel IN relationships(p)
  WHERE rel.ts >= $startTs AND rel.ts <= $endTs)
RETURN p
LIMIT $limit
`

const PruneOldEdges = `
MATCH ()-[r]->()
WHERE r.ts < $cutoff
DELETE r
`
