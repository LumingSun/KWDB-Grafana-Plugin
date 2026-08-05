import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, Select, VerticalGroup } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';

import { KwdbDataSourceOptions, KwdbSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<KwdbDataSourceOptions, KwdbSecureJsonData>;

const SSL_MODE_OPTIONS: Array<{ label: string; value: NonNullable<KwdbDataSourceOptions['sslMode']> }> = [
  { label: 'disable', value: 'disable' },
  { label: 'require', value: 'require' },
  { label: 'verify-ca', value: 'verify-ca' },
  { label: 'verify-full', value: 'verify-full' },
];

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const updateJsonData = (patch: Partial<KwdbDataSourceOptions>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        ...patch,
      },
    });
  };

  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ host: event.target.value });
  };

  const onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    const port = Number.parseInt(event.target.value, 10);
    updateJsonData({ port: Number.isNaN(port) ? undefined : port });
  };

  const onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ database: event.target.value });
  };

  const onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ user: event.target.value });
  };

  const onSslModeChange = (sslMode: KwdbDataSourceOptions['sslMode']) => {
    updateJsonData({ sslMode: sslMode ?? 'disable' });
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
    <VerticalGroup spacing="md">
      <InlineField label="Host" labelWidth={14}>
        <Input
          id="config-editor-host"
          onChange={onHostChange}
          value={jsonData.host ?? ''}
          placeholder="localhost"
          width={40}
        />
      </InlineField>
      <InlineField label="Port" labelWidth={14}>
        <Input
          id="config-editor-port"
          onChange={onPortChange}
          value={jsonData.port ?? 26257}
          type="number"
          placeholder="26257"
          width={40}
        />
      </InlineField>
      <InlineField label="Database" labelWidth={14}>
        <Input
          id="config-editor-database"
          onChange={onDatabaseChange}
          value={jsonData.database ?? ''}
          placeholder="defaultdb"
          width={40}
        />
      </InlineField>
      <InlineField label="User" labelWidth={14}>
        <Input id="config-editor-user" onChange={onUserChange} value={jsonData.user ?? ''} placeholder="root" width={40} />
      </InlineField>
      <InlineField label="SSL Mode" labelWidth={14}>
        <Select<'disable' | 'require' | 'verify-ca' | 'verify-full'>
          aria-label="SSL Mode"
          options={SSL_MODE_OPTIONS}
          value={jsonData.sslMode ?? 'disable'}
          onChange={(item) => onSslModeChange(item.value)}
          width={40}
          menuShouldPortal={false}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={14}>
        <SecretInput
          id="config-editor-password"
          isConfigured={Boolean(secureJsonFields?.password)}
          value={secureJsonData?.password ?? ''}
          placeholder="Password"
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
    </VerticalGroup>
  );
}
