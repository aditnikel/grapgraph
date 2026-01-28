package cypher

// Relationship type (event_type) cannot be parameterized in Cypher.
// It must be validated and safely interpolated by the caller.

const UpsertAggregatedEdgeTemplate = `
MERGE (u:User {user_id:$user_id})
MERGE (t:%s {%s:$target_key})
MERGE (u)-[r:%s]->(t)
ON CREATE SET
  r.event_count = 0,
  r.first_seen = $ts,
  r.last_seen = $ts,
  r.event_count_30d = 0,
  r.distinct_ip_count_30d = 0,
  r.window_start_30d = $ts,
  r.total_amount = 0.0,
  r.max_amount = 0.0
SET
  r.event_count = r.event_count + 1,
  r.first_seen = CASE WHEN r.first_seen > $ts THEN $ts ELSE r.first_seen END,
  r.last_seen = CASE WHEN r.last_seen < $ts THEN $ts ELSE r.last_seen END,
  r.event_count_30d = CASE WHEN ($ts - r.window_start_30d) > 2592000000 THEN 1 ELSE r.event_count_30d + 1 END,
  r.window_start_30d = CASE WHEN ($ts - r.window_start_30d) > 2592000000 THEN $ts ELSE r.window_start_30d END
RETURN u.user_id
`
