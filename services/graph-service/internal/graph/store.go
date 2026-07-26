package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Node struct {
	EntityID  string  `json:"entityId"`
	Name      string  `json:"name"`
	LCR       float64 `json:"lcr,omitempty"`
	CET1Ratio float64 `json:"cet1Ratio,omitempty"`
}

type Edge struct {
	FromEntityID string  `json:"fromEntityId"`
	ToEntityID   string  `json:"toEntityId"`
	ExposureType string  `json:"exposureType"`
	NotionalEur  float64 `json:"notionalEur"`
	Layer        string  `json:"layer"`
	InstrumentID string  `json:"instrumentId,omitempty"`
}

type Summary struct {
	NodeCount int `json:"nodeCount"`
	EdgeCount int `json:"edgeCount"`
}

type Store struct {
	driver   neo4j.DriverWithContext
	tenantID string
}

func NewStore(uri, user, password, tenantID string) (*Store, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	return &Store{driver: driver, tenantID: tenantID}, nil
}

func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

func (s *Store) VerifyConnectivity(ctx context.Context) error {
	return s.driver.VerifyConnectivity(ctx)
}

func (s *Store) UpsertInstitution(ctx context.Context, entityID, name string, lcr, cet1 float64) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MERGE (e:LegalEntity {tenantId: $tenantId, entityId: $entityId})
			SET e.name = $name, e.lcr = $lcr, e.cet1Ratio = $cet1, e.updatedAt = datetime()
		`, map[string]any{
			"tenantId": s.tenantID,
			"entityId": entityID,
			"name":     name,
			"lcr":      lcr,
			"cet1":     cet1,
		})
		return nil, err
	})
	return err
}

func (s *Store) UpsertExposure(ctx context.Context, in ExposureInput) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MERGE (from:LegalEntity {tenantId: $tenantId, entityId: $fromId})
			MERGE (to:LegalEntity {tenantId: $tenantId, entityId: $toId})
			MERGE (from)-[r:EXPOSURE {tenantId: $tenantId, edgeKey: $edgeKey}]->(to)
			SET r.exposureType = $exposureType,
			    r.notionalEur = $notionalEur,
			    r.layer = $layer,
			    r.instrumentId = $instrumentId,
			    r.updatedAt = datetime()
		`, map[string]any{
			"tenantId":     s.tenantID,
			"fromId":       in.FromEntityID,
			"toId":         in.ToEntityID,
			"edgeKey":      in.EdgeKey,
			"exposureType": in.ExposureType,
			"notionalEur":  in.NotionalEur,
			"layer":        in.Layer,
			"instrumentId": in.InstrumentID,
		})
		return nil, err
	})
	return err
}

type ExposureInput struct {
	FromEntityID string
	ToEntityID   string
	EdgeKey      string
	ExposureType string
	NotionalEur  float64
	Layer        string
	InstrumentID string
}

func (s *Store) Summary(ctx context.Context) (Summary, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		rec, err := tx.Run(ctx, `
			MATCH (e:LegalEntity {tenantId: $tenantId})
			WITH count(e) AS nodes
			MATCH (:LegalEntity {tenantId: $tenantId})-[r:EXPOSURE {tenantId: $tenantId}]->(:LegalEntity {tenantId: $tenantId})
			RETURN nodes, count(r) AS edges
		`, map[string]any{"tenantId": s.tenantID})
		if err != nil {
			return Summary{}, err
		}
		if rec.Next(ctx) {
			row := rec.Record()
			nodes, _ := row.Get("nodes")
			edges, _ := row.Get("edges")
			return Summary{
				NodeCount: int(asInt64(nodes)),
				EdgeCount: int(asInt64(edges)),
			}, nil
		}
		return Summary{}, rec.Err()
	})
	if err != nil {
		return Summary{}, err
	}
	return result.(Summary), nil
}

func (s *Store) ListNodes(ctx context.Context, nameQuery string, limit int) ([]Node, error) {
	if limit <= 0 {
		limit = 100
	}
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (e:LegalEntity {tenantId: $tenantId})
			WHERE $nameQuery = '' OR toLower(e.name) CONTAINS toLower($nameQuery)
			RETURN e.entityId AS entityId, e.name AS name, coalesce(e.lcr, 0.0) AS lcr, coalesce(e.cet1Ratio, 0.0) AS cet1Ratio
			ORDER BY e.name
			LIMIT $limit
		`
		rec, err := tx.Run(ctx, query, map[string]any{
			"tenantId":  s.tenantID,
			"nameQuery": nameQuery,
			"limit":     limit,
		})
		if err != nil {
			return nil, err
		}
		var nodes []Node
		for rec.Next(ctx) {
			row := rec.Record()
			entityID, _ := row.Get("entityId")
			name, _ := row.Get("name")
			lcr, _ := row.Get("lcr")
			cet1, _ := row.Get("cet1Ratio")
			nodes = append(nodes, Node{
				EntityID:  fmt.Sprint(entityID),
				Name:      fmt.Sprint(name),
				LCR:       asFloat64(lcr),
				CET1Ratio: asFloat64(cet1),
			})
		}
		return nodes, rec.Err()
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []Node{}, nil
	}
	return result.([]Node), nil
}

func (s *Store) ListEdges(ctx context.Context, layer string, limit int) ([]Edge, error) {
	if limit <= 0 {
		limit = 500
	}
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (from:LegalEntity {tenantId: $tenantId})-[r:EXPOSURE {tenantId: $tenantId}]->(to:LegalEntity {tenantId: $tenantId})
			WHERE $layer = '' OR r.layer = $layer
			RETURN from.entityId AS fromEntityId, to.entityId AS toEntityId,
			       r.exposureType AS exposureType, r.notionalEur AS notionalEur,
			       r.layer AS layer, coalesce(r.instrumentId, '') AS instrumentId
			LIMIT $limit
		`
		rec, err := tx.Run(ctx, query, map[string]any{
			"tenantId": s.tenantID,
			"layer":    layer,
			"limit":    limit,
		})
		if err != nil {
			return nil, err
		}
		var edges []Edge
		for rec.Next(ctx) {
			row := rec.Record()
			fromID, _ := row.Get("fromEntityId")
			toID, _ := row.Get("toEntityId")
			expType, _ := row.Get("exposureType")
			notional, _ := row.Get("notionalEur")
			layerVal, _ := row.Get("layer")
			instID, _ := row.Get("instrumentId")
			edges = append(edges, Edge{
				FromEntityID: fmt.Sprint(fromID),
				ToEntityID:   fmt.Sprint(toID),
				ExposureType: fmt.Sprint(expType),
				NotionalEur:  asFloat64(notional),
				Layer:        fmt.Sprint(layerVal),
				InstrumentID: fmt.Sprint(instID),
			})
		}
		return edges, rec.Err()
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []Edge{}, nil
	}
	return result.([]Edge), nil
}

type PathResult struct {
	FromEntityID string   `json:"fromEntityId"`
	ToEntityID   string   `json:"toEntityId"`
	Hops         int      `json:"hops"`
	NodeIDs      []string `json:"nodeIds"`
}

type CentralityRow struct {
	EntityID string  `json:"entityId"`
	Name     string  `json:"name"`
	Degree   int     `json:"degree"`
	Score    float64 `json:"score"`
}

// ShortestPath returns an undirected exposure path (max 6 hops) between two institutions.
func (s *Store) ShortestPath(ctx context.Context, fromID, toID string) (PathResult, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		rec, err := tx.Run(ctx, `
			MATCH (a:LegalEntity {tenantId: $tenantId, entityId: $fromId})
			MATCH (b:LegalEntity {tenantId: $tenantId, entityId: $toId})
			MATCH p = shortestPath((a)-[:EXPOSURE*..6]-(b))
			RETURN [n IN nodes(p) | n.entityId] AS nodeIds, length(p) AS hops
			LIMIT 1
		`, map[string]any{
			"tenantId": s.tenantID,
			"fromId":   fromID,
			"toId":     toID,
		})
		if err != nil {
			return PathResult{}, err
		}
		if !rec.Next(ctx) {
			return PathResult{}, fmt.Errorf("no path")
		}
		row := rec.Record()
		rawIDs, _ := row.Get("nodeIds")
		hops, _ := row.Get("hops")
		var ids []string
		if arr, ok := rawIDs.([]any); ok {
			for _, v := range arr {
				ids = append(ids, fmt.Sprint(v))
			}
		}
		return PathResult{
			FromEntityID: fromID,
			ToEntityID:   toID,
			Hops:         int(asInt64(hops)),
			NodeIDs:      ids,
		}, rec.Err()
	})
	if err != nil {
		return PathResult{}, err
	}
	return result.(PathResult), nil
}

// DegreeCentrality returns degree centrality over the exposure graph (no GDS plugin required).
func (s *Store) DegreeCentrality(ctx context.Context, limit int) ([]CentralityRow, error) {
	if limit <= 0 {
		limit = 20
	}
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		rec, err := tx.Run(ctx, `
			MATCH (e:LegalEntity {tenantId: $tenantId})
			OPTIONAL MATCH (e)-[r:EXPOSURE {tenantId: $tenantId}]-(:LegalEntity {tenantId: $tenantId})
			RETURN e.entityId AS entityId, coalesce(e.name, '') AS name, count(r) AS degree
			ORDER BY degree DESC, entityId
			LIMIT $limit
		`, map[string]any{"tenantId": s.tenantID, "limit": limit})
		if err != nil {
			return nil, err
		}
		var out []CentralityRow
		maxDeg := 0
		for rec.Next(ctx) {
			row := rec.Record()
			id, _ := row.Get("entityId")
			name, _ := row.Get("name")
			degRaw, _ := row.Get("degree")
			deg := int(asInt64(degRaw))
			if deg > maxDeg {
				maxDeg = deg
			}
			out = append(out, CentralityRow{
				EntityID: fmt.Sprint(id),
				Name:     fmt.Sprint(name),
				Degree:   deg,
			})
		}
		if err := rec.Err(); err != nil {
			return nil, err
		}
		if maxDeg == 0 {
			maxDeg = 1
		}
		for i := range out {
			out[i].Score = float64(out[i].Degree) / float64(maxDeg)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []CentralityRow{}, nil
	}
	return result.([]CentralityRow), nil
}

func asFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

func asInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	default:
		return 0
	}
}
