import React, { useEffect, useState } from 'react';
import { Select } from '@grafana/ui';

import type { KwdbDataSource } from '../../../datasource';

interface Props {
  datasource: KwdbDataSource;
  value?: string;
  onChange: (table: string) => void;
}

export function TableSelector({ datasource, value, onChange }: Props) {
  const [state, setState] = useState<{ tables: string[]; loading: boolean }>({ tables: [], loading: true });

  useEffect(() => {
    let cancelled = false;
    datasource
      .getTables()
      .then((result) => {
        if (!cancelled) {
          setState({ tables: result ?? [], loading: false });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setState({ tables: [], loading: false });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [datasource]);

  const options = state.tables.map((table) => ({ label: table, value: table }));

  return (
    <Select<string>
      aria-label="Table"
      isLoading={state.loading}
      options={options}
      value={value ? { label: value, value } : undefined}
      onChange={(item) => onChange(item.value ?? '')}
      placeholder="Select table"
      isClearable
      menuShouldPortal={false}
      width={40}
    />
  );
}
