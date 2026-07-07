import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { IrisQuery, IrisDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<IrisQuery, IrisDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<IrisDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<IrisQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: IrisQuery, scopedVars: ScopedVars): IrisQuery {
    return {
      ...query,
      rawSql: getTemplateSrv().replace(query.rawSql, scopedVars),
    };
  }

  filterQuery(query: IrisQuery): boolean {
    return Boolean(query.rawSql?.trim());
  }
}
