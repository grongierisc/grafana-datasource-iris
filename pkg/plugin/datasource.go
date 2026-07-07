package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/caretdev/go-irisnative"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"github.com/grongierisc/grafana-datasource-iris/pkg/models"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type sqlDB interface {
	PingContext(context.Context) error
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	Close() error
}

type dbFactory func(*models.PluginSettings) (*sql.DB, error)

func NewDatasource(_ context.Context, instanceSettings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadPluginSettings(instanceSettings)
	if err != nil {
		return nil, err
	}

	db, err := openIRISDB(settings)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		settings: settings,
		db:       db,
		openDB:   openIRISDB,
	}, nil
}

type Datasource struct {
	mu       sync.Mutex
	settings *models.PluginSettings
	db       sqlDB
	openDB   dbFactory
}

func (d *Datasource) Dispose() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		_ = d.db.Close()
		d.db = nil
	}
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, req.PluginContext, q)
	}

	return response, nil
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	settings, db, err := d.ensureConnection(pCtx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	model, err := parseQuery(query)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	expandedSQL, err := interpolateSQL(model.RawSQL, query)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}
	if err := validateReadOnlySQL(expandedSQL); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.QueryTimeoutSeconds)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(queryCtx, expandedSQL)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}
	defer func() {
		_ = rows.Close()
	}()

	rowLimit := settings.RowLimit
	if model.RowLimit != nil && *model.RowLimit > 0 {
		rowLimit = *model.RowLimit
	}

	frame, err := sqlutil.FrameFromRows(rows, int64(rowLimit))
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	frame.Name = query.RefID
	frame.Meta = &data.FrameMeta{
		ExecutedQueryString:    expandedSQL,
		PreferredVisualization: preferredVisualization(model.Format),
	}

	return backend.DataResponse{Frames: data.Frames{frame}}
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	settings, db, err := d.ensureConnection(req.PluginContext)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	if err := settings.Validate(); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.QueryTimeoutSeconds)*time.Second)
	defer cancel()

	if err := db.PingContext(checkCtx); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	var ok int
	if err := db.QueryRowContext(checkCtx, "SELECT 1 AS ok").Scan(&ok); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	if ok != 1 {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "unexpected health check response"}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected to InterSystems IRIS",
	}, nil
}

func (d *Datasource) ensureConnection(pCtx backend.PluginContext) (*models.PluginSettings, sqlDB, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.settings != nil && d.db != nil {
		return d.settings, d.db, nil
	}
	if pCtx.DataSourceInstanceSettings == nil {
		return nil, nil, errors.New("datasource settings are missing")
	}

	settings, err := models.LoadPluginSettings(*pCtx.DataSourceInstanceSettings)
	if err != nil {
		return nil, nil, err
	}

	factory := d.openDB
	if factory == nil {
		factory = openIRISDB
	}

	db, err := factory(settings)
	if err != nil {
		return nil, nil, err
	}

	d.settings = settings
	d.db = db

	return settings, db, nil
}

func openIRISDB(settings *models.PluginSettings) (*sql.DB, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("iris", settings.DSN())
	if err != nil {
		return nil, fmt.Errorf("open IRIS connection: %w", err)
	}

	db.SetMaxOpenConns(settings.MaxOpenConnections)
	db.SetMaxIdleConns(settings.MaxIdleConnections)
	db.SetConnMaxLifetime(settings.ConnectionLifetime())

	return db, nil
}

type queryModel struct {
	RawSQL   string `json:"rawSql"`
	Format   string `json:"format"`
	RowLimit *int   `json:"rowLimit,omitempty"`
}

func parseQuery(query backend.DataQuery) (queryModel, error) {
	var model queryModel
	if err := json.Unmarshal(query.JSON, &model); err != nil {
		return queryModel{}, fmt.Errorf("json unmarshal: %w", err)
	}
	if model.RawSQL == "" {
		return queryModel{}, errors.New("SQL query is required")
	}
	if model.Format == "" {
		model.Format = "table"
	}
	return model, nil
}

func preferredVisualization(format string) data.VisType {
	if format == "time_series" {
		return data.VisTypeGraph
	}
	return data.VisTypeTable
}
