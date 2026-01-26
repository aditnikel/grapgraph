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
  AND r.%s >= $min_count
  AND r.last_seen >= $from_ts AND r.last_seen <= $to_ts
RETURN
  'USER' AS from_type,
  u.user_id AS from_key,
  ` + NodeTypeCase + ` AS to_type,
  ` + NodeKeyCase + ` AS to_key,
  type(r) AS edge_type,
  coalesce(r.event_count, 0) AS event_count,
  coalesce(r.event_count_30d, 0) AS event_count_30d,
  coalesce(r.distinct_ip_count_30d, 0) AS distinct_ip_count_30d,
  coalesce(r.first_seen, 0) AS first_seen,
  coalesce(r.last_seen, 0) AS last_seen,
  coalesce(r.total_amount, 0.0) AS total_amount,
  coalesce(r.max_amount, 0.0) AS max_amount
ORDER BY r.%s DESC
LIMIT $limit
`

const EntityToUserTemplate = `
MATCH (n)<-[r]-(u:User)
WHERE id(n) = $entity_id
  AND type(r) IN [%s]
  AND r.%s >= $min_count
  AND r.last_seen >= $from_ts AND r.last_seen <= $to_ts
RETURN
  ` + NodeTypeCase + ` AS from_type,
  ` + NodeKeyCase + ` AS from_key,
  'USER' AS to_type,
  u.user_id AS to_key,
  type(r) AS edge_type,
  coalesce(r.event_count, 0) AS event_count,
  coalesce(r.event_count_30d, 0) AS event_count_30d,
  coalesce(r.distinct_ip_count_30d, 0) AS distinct_ip_count_30d,
  coalesce(r.first_seen, 0) AS first_seen,
  coalesce(r.last_seen, 0) AS last_seen,
  coalesce(r.total_amount, 0.0) AS total_amount,
  coalesce(r.max_amount, 0.0) AS max_amount,
  id(u) AS user_internal_id
ORDER BY r.%s DESC
LIMIT $limit
`

const EntityInternalIDByKey = `
MATCH (n)
WHERE
  (%s)
RETURN id(n) AS entity_id
LIMIT 1
`
