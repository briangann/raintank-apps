package migrations

import (
	"fmt"

	"github.com/grafana/grafana/pkg/services/sqlstore/migrator"
)

func addAgentMetricMigrations(mg *migrator.Migrator) {
	agentMetricV1 := migrator.Table{
		Name: "agent_metric",
		Columns: []*migrator.Column{
			{Name: "id", Type: migrator.DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "agent_id", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "owner", Type: migrator.DB_NVarchar, Length: 255},
			{Name: "metric_id", Type: migrator.DB_NVarchar, Length: 255},
			{Name: "created", Type: migrator.DB_DateTime},
		},
		Indices: []*migrator.Index{
			{Cols: []string{"owner", "agent_id"}},
			{Cols: []string{"owner", "metric_id"}},
		},
	}
	mg.AddMigration("create agent_metric table v1", migrator.NewAddTableMigration(agentMetricV1))
	for _, index := range agentMetricV1.Indices {
		migrationId := fmt.Sprintf("create index %s - %s", index.XName(agentMetricV1.Name), "v1")
		mg.AddMigration(migrationId, migrator.NewAddIndexMigration(agentMetricV1, index))
	}
}
