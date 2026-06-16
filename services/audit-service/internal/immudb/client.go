package immudb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/digital-twin/platform/services/audit-service/internal/events"
)

const (
	headKeyPrefix  = "audit:head"
	entryKeyPrefix = "audit:entry:"
)

type HeadState struct {
	LastSequence    int64  `json:"lastSequence"`
	LastPayloadHash string `json:"lastPayloadHash"`
}

type Ledger interface {
	Ping(ctx context.Context) error
	AppendEntry(ctx context.Context, entry events.AuditEntry) error
	GetEntry(ctx context.Context, entryID string) (events.AuditEntry, error)
	GetHead(ctx context.Context) (HeadState, error)
}

type Client struct {
	cli      client.ImmuClient
	database string
}

func Connect(ctx context.Context, host string, port int, database, user, password string) (*Client, error) {
	opts := client.DefaultOptions().
		WithAddress(host).
		WithPort(port)

	cli := client.NewClient()
	cli.WithOptions(opts)

	if err := cli.OpenSession(ctx, []byte(user), []byte(password), database); err != nil {
		_ = cli.CloseSession(ctx)
		_ = cli.Disconnect()
		if createErr := tryCreateDatabase(ctx, host, port, user, password, database); createErr != nil {
			return nil, fmt.Errorf("open session: %w (ensure db: %v)", err, createErr)
		}
		cli = client.NewClient()
		cli.WithOptions(opts)
		if err := cli.OpenSession(ctx, []byte(user), []byte(password), database); err != nil {
			_ = cli.CloseSession(ctx)
			_ = cli.Disconnect()
			return nil, fmt.Errorf("open session after create: %w", err)
		}
	}

	return &Client{cli: cli, database: database}, nil
}

func tryCreateDatabase(ctx context.Context, host string, port int, user, password, database string) error {
	opts := client.DefaultOptions().WithAddress(host).WithPort(port)
	cli := client.NewClient()
	cli.WithOptions(opts)
	defer cli.Disconnect()

	if err := cli.OpenSession(ctx, []byte(user), []byte(password), "defaultdb"); err != nil {
		return err
	}
	defer func() { _ = cli.CloseSession(ctx) }()

	_, err := cli.CreateDatabaseV2(ctx, database, &schema.DatabaseNullableSettings{})
	return err
}

func (c *Client) Close() error {
	if c == nil || c.cli == nil {
		return nil
	}
	ctx := context.Background()
	_ = c.cli.CloseSession(ctx)
	return c.cli.Disconnect()
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.cli == nil {
		return fmt.Errorf("immudb client not initialized")
	}
	return c.cli.HealthCheck(ctx)
}

func (c *Client) ResetHead(ctx context.Context) error {
	headBytes, err := json.Marshal(HeadState{})
	if err != nil {
		return err
	}
	_, err = c.cli.Set(ctx, []byte(headKeyPrefix), headBytes)
	return err
}

func (c *Client) GetHead(ctx context.Context) (HeadState, error) {
	entry, err := c.cli.Get(ctx, []byte(headKeyPrefix))
	if err != nil {
		return HeadState{}, nil
	}
	var head HeadState
	if err := json.Unmarshal(entry.Value, &head); err != nil {
		return HeadState{}, fmt.Errorf("decode head: %w", err)
	}
	return head, nil
}

func (c *Client) AppendEntry(ctx context.Context, entry events.AuditEntry) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	key := entryKeyPrefix + entry.EntryID
	if _, err := c.cli.Set(ctx, []byte(key), body); err != nil {
		return fmt.Errorf("set entry: %w", err)
	}
	head := HeadState{
		LastSequence:    entry.SequenceNumber,
		LastPayloadHash: entry.PayloadHash,
	}
	headBytes, err := json.Marshal(head)
	if err != nil {
		return err
	}
	if _, err := c.cli.Set(ctx, []byte(headKeyPrefix), headBytes); err != nil {
		return fmt.Errorf("set head: %w", err)
	}
	slog.Info("appended audit entry to immudb", "entryId", entry.EntryID, "sequence", entry.SequenceNumber)
	return nil
}

func (c *Client) GetEntry(ctx context.Context, entryID string) (events.AuditEntry, error) {
	entry, err := c.cli.Get(ctx, []byte(entryKeyPrefix+entryID))
	if err != nil {
		return events.AuditEntry{}, fmt.Errorf("get entry %s: %w", entryID, err)
	}
	var out events.AuditEntry
	if err := json.Unmarshal(entry.Value, &out); err != nil {
		return events.AuditEntry{}, err
	}
	return out, nil
}
