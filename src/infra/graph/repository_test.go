package graph

import (
	"testing"

	"github.com/aditnikel/grapgraph/src/model"
)

func TestStableIDs(t *testing.T) {
	t.Run("node_id_format", func(t *testing.T) {
		nodeID := StableNodeID(model.NodeUser, "u1")
		if nodeID != "USER:u1" {
			t.Fatalf("unexpected node id: %s", nodeID)
		}
	})

	t.Run("edge_id_stability_and_format", func(t *testing.T) {
		edgeID1 := StableEdgeID("USER:u1", "MERCHANT:m1", "PAYMENT")
		edgeID2 := StableEdgeID("USER:u1", "MERCHANT:m1", "PAYMENT")
		edgeID3 := StableEdgeID("USER:u1", "MERCHANT:m1", "REFUND")
		if edgeID1 != edgeID2 {
			t.Fatalf("expected stable edge id")
		}
		if edgeID1 == edgeID3 {
			t.Fatalf("expected edge id to vary by type")
		}
		if len(edgeID1) != 18 || edgeID1[:2] != "e_" {
			t.Fatalf("unexpected edge id format: %s", edgeID1)
		}
	})
}

func TestQuoteEdgeTypes(t *testing.T) {
	t.Run("quotes_and_strips_single_quotes", func(t *testing.T) {
		got := QuoteEdgeTypes([]model.EventType{"PAYMENT", "O'CLOCK"})
		if got != "'PAYMENT','OCLOCK'" {
			t.Fatalf("unexpected quoted types: %s", got)
		}
	})
}

func TestInterpolate(t *testing.T) {
	t.Run("replaces_placeholders_and_escapes_strings", func(t *testing.T) {
		repo := &Repo{}
		query := "MATCH (n) WHERE n.name = $name AND n.age = $age RETURN n"
		out := repo.interpolate(query, map[string]any{
			"name": "O'Reilly",
			"age":  5,
		})
		want := "MATCH (n) WHERE n.name = 'O\\'Reilly' AND n.age = 5 RETURN n"
		if out != want {
			t.Fatalf("unexpected interpolate: %s", out)
		}
	})
}
