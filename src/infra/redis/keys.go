package redis

func GraphKey(ledger string) string  { return "graph:" + ledger }
func StreamKey(ledger string) string { return "stream:" + ledger }
