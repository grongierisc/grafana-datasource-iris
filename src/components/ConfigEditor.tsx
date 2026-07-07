import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { DEFAULT_OPTIONS, IrisDataSourceOptions, IrisSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<IrisDataSourceOptions, IrisSecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const updateJsonData = (patch: Partial<IrisDataSourceOptions>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        ...patch,
      },
    });
  };

  const updateText = (key: keyof IrisDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ [key]: event.target.value } as Partial<IrisDataSourceOptions>);
  };

  const updateNumber = (key: keyof IrisDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.valueAsNumber;
    updateJsonData({ [key]: Number.isFinite(value) ? value : undefined } as Partial<IrisDataSourceOptions>);
  };

  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...secureJsonData,
        password: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Host" labelWidth={24} required>
        <Input
          id="config-editor-host"
          onChange={updateText('host')}
          value={jsonData.host ?? DEFAULT_OPTIONS.host}
          width={40}
        />
      </InlineField>
      <InlineField label="Port" labelWidth={24} required>
        <Input
          id="config-editor-port"
          onChange={updateNumber('port')}
          value={jsonData.port ?? DEFAULT_OPTIONS.port}
          width={16}
          type="number"
          min={1}
        />
      </InlineField>
      <InlineField label="Namespace" labelWidth={24} required>
        <Input
          id="config-editor-namespace"
          onChange={updateText('namespace')}
          value={jsonData.namespace ?? DEFAULT_OPTIONS.namespace}
          width={24}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={24} required>
        <Input id="config-editor-username" onChange={updateText('username')} value={jsonData.username ?? ''} width={40} />
      </InlineField>
      <InlineField label="Password" labelWidth={24} required>
        <SecretInput
          id="config-editor-password"
          isConfigured={secureJsonFields?.password ?? false}
          value={secureJsonData?.password ?? ''}
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
      <InlineField label="Query timeout" labelWidth={24} required>
        <Input
          id="config-editor-query-timeout"
          onChange={updateNumber('queryTimeoutSeconds')}
          value={jsonData.queryTimeoutSeconds ?? DEFAULT_OPTIONS.queryTimeoutSeconds}
          width={16}
          type="number"
          min={1}
          suffix="s"
        />
      </InlineField>
      <InlineField label="Row limit" labelWidth={24} required>
        <Input
          id="config-editor-row-limit"
          onChange={updateNumber('rowLimit')}
          value={jsonData.rowLimit ?? DEFAULT_OPTIONS.rowLimit}
          width={16}
          type="number"
          min={1}
        />
      </InlineField>
      <InlineField label="Max open connections" labelWidth={24} required>
        <Input
          id="config-editor-max-open-conns"
          onChange={updateNumber('maxOpenConns')}
          value={jsonData.maxOpenConns ?? DEFAULT_OPTIONS.maxOpenConns}
          width={16}
          type="number"
          min={1}
        />
      </InlineField>
      <InlineField label="Max idle connections" labelWidth={24} required>
        <Input
          id="config-editor-max-idle-conns"
          onChange={updateNumber('maxIdleConns')}
          value={jsonData.maxIdleConns ?? DEFAULT_OPTIONS.maxIdleConns}
          width={16}
          type="number"
          min={0}
        />
      </InlineField>
      <InlineField label="Connection lifetime" labelWidth={24} required>
        <Input
          id="config-editor-conn-lifetime"
          onChange={updateNumber('connMaxLifetimeSeconds')}
          value={jsonData.connMaxLifetimeSeconds ?? DEFAULT_OPTIONS.connMaxLifetimeSeconds}
          width={16}
          type="number"
          min={1}
          suffix="s"
        />
      </InlineField>
    </>
  );
}
