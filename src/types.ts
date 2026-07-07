import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type QueryFormat = 'table' | 'time_series';

export interface IrisQuery extends DataQuery {
  rawSql: string;
  format: QueryFormat;
  rowLimit?: number;
}

export const DEFAULT_QUERY: Partial<IrisQuery> = {
  rawSql: 'SELECT 1 AS value',
  format: 'table',
};

export interface IrisDataSourceOptions extends DataSourceJsonData {
  host?: string;
  port?: number;
  namespace?: string;
  username?: string;
  queryTimeoutSeconds?: number;
  rowLimit?: number;
  maxOpenConns?: number;
  maxIdleConns?: number;
  connMaxLifetimeSeconds?: number;
}

export interface IrisSecureJsonData {
  password?: string;
}

type DefaultOptionKeys =
  | 'host'
  | 'port'
  | 'namespace'
  | 'queryTimeoutSeconds'
  | 'rowLimit'
  | 'maxOpenConns'
  | 'maxIdleConns'
  | 'connMaxLifetimeSeconds';

export const DEFAULT_OPTIONS: Required<Pick<IrisDataSourceOptions, DefaultOptionKeys>> = {
  host: 'localhost',
  port: 1972,
  namespace: 'USER',
  queryTimeoutSeconds: 30,
  rowLimit: 1000,
  maxOpenConns: 10,
  maxIdleConns: 5,
  connMaxLifetimeSeconds: 1800,
};
