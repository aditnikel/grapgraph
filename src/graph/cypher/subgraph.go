package cypher

const NodeTypeCase = `
CASE
  WHEN n:Merchant THEN 'MERCHANT'
  WHEN n:Exchange THEN 'EXCHANGE'
  WHEN n:Wallet THEN 'WALLET'
  WHEN n:PaymentMethod THEN 'PAYMENT_METHOD'
  WHEN n:Bank THEN 'BANK'
  WHEN n:Device THEN 'DEVICE'
  ELSE 'UNKNOWN'
END
`

const NodeKeyCase = `
CASE
  WHEN n:Merchant THEN n.merchant_id_mpan
  WHEN n:Exchange THEN n.exchange
  WHEN n:Wallet THEN n.wallet_address
  WHEN n:PaymentMethod THEN n.payment_method
  WHEN n:Bank THEN n.issuing_bank
  WHEN n:Device THEN n.device_id
  ELSE ''
END
`

const UserToEntityTemplate = `
MATCH (u:User {user_id:$user_id})-[r]->(n)
WHERE type(r) IN [%s]
RETURN
  'USER' AS from_type,
  u.user_id AS from_key,
  ` + NodeTypeCase + ` AS to_type,
  ` + NodeKeyCase + ` AS to_key,
  type(r) AS edge_type
LIMIT $limit
`

const EntityToUserTemplate = `
MATCH (n)<-[r]-(u:User)
WHERE id(n) = $entity_id
  AND type(r) IN [%s]
RETURN
  ` + NodeTypeCase + ` AS from_type,
  ` + NodeKeyCase + ` AS from_key,
  'USER' AS to_type,
  u.user_id AS to_key,
  type(r) AS edge_type,
  id(u) AS user_internal_id
LIMIT $limit
`

const EntityInternalIDByKey = `
MATCH (n)
WHERE
  (%s)
RETURN id(n) AS entity_id
LIMIT 1
`
