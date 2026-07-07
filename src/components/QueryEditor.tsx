import React, { ChangeEvent } from 'react';
import { InlineField, Input, RadioButtonGroup, TextArea } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { IrisDataSourceOptions, IrisQuery, QueryFormat } from '../types';

type Props = QueryEditorProps<DataSource, IrisQuery, IrisDataSourceOptions>;

const formatOptions: Array<SelectableValue<QueryFormat>> = [
  { label: 'Table', value: 'table' },
  { label: 'Time series', value: 'time_series' },
];

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onRawSqlChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    onChange({ ...query, rawSql: event.target.value });
  };

  const onFormatChange = (value: QueryFormat) => {
    onChange({ ...query, format: value });
    onRunQuery();
  };

  const onRowLimitChange = (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.valueAsNumber;
    onChange({ ...query, rowLimit: Number.isFinite(value) && value > 0 ? value : undefined });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <InlineField label="Format" labelWidth={14}>
        <RadioButtonGroup<QueryFormat>
          id="query-editor-format"
          options={formatOptions}
          value={query.format ?? 'table'}
          onChange={onFormatChange}
        />
      </InlineField>
      <InlineField label="Row limit" labelWidth={14}>
        <Input
          id="query-editor-row-limit"
          aria-label="Row limit"
          onChange={onRowLimitChange}
          value={query.rowLimit ?? ''}
          width={16}
          type="number"
          min={1}
          onBlur={onRunQuery}
        />
      </InlineField>
      <InlineField label="SQL" labelWidth={14} grow required>
        <TextArea
          id="query-editor-raw-sql"
          aria-label="SQL"
          onBlur={onRunQuery}
          onChange={onRawSqlChange}
          value={query.rawSql ?? ''}
          rows={8}
        />
      </InlineField>
    </div>
  );
}
