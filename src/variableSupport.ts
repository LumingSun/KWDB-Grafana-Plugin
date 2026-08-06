import { StandardVariableQuery, StandardVariableSupport } from '@grafana/data';

import type { KwdbDataSource } from './datasource';
import type { KwdbQuery } from './types';

export class KwdbVariableSupport extends StandardVariableSupport<KwdbDataSource> {
  toDataQuery(query: StandardVariableQuery): KwdbQuery {
    return {
      refId: query.refId || 'variable-query',
      mode: 'raw',
      format: 'table',
      rawSql: query.query,
    };
  }
}
