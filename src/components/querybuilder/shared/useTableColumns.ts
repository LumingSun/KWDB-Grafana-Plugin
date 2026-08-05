import { useEffect, useState } from 'react';

import type { KwdbDataSource } from '../../../datasource';
import type { ColumnInfo } from '../../../types';

export function useTableColumns(datasource: KwdbDataSource, table?: string): ColumnInfo[] {
  const [loaded, setLoaded] = useState<{ table?: string; columns: ColumnInfo[] }>({ columns: [] });

  useEffect(() => {
    let cancelled = false;
    if (table) {
      datasource
        .getColumns(table)
        .then((result) => {
          if (!cancelled) {
            setLoaded({ table, columns: result ?? [] });
          }
        })
        .catch(() => {
          if (!cancelled) {
            setLoaded({ table, columns: [] });
          }
        });
    }
    return () => {
      cancelled = true;
    };
  }, [datasource, table]);

  return table && loaded.table === table ? loaded.columns : [];
}
