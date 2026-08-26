import type { KwdbVariableQuery } from './types';

const VARIABLE_QUERY_RE = /^(tag_values|tables|columns)\s*\((.*)\)$/i;

// Parses a template variable query into its structured form. `query` may be a
// structured KwdbVariableQuery, a raw string such as `tag_values(table, column)`,
// or null/undefined.
export function parseVariableQuery(query: unknown): KwdbVariableQuery {
  if (query && typeof query === 'object' && typeof (query as KwdbVariableQuery).queryType === 'string') {
    const parsed = query as KwdbVariableQuery;
    switch (parsed.queryType) {
      case 'columns':
        if (!parsed.table?.trim()) {
          throw new Error('Variable query "columns" requires a table');
        }
        break;
      case 'tagValues':
        if (!parsed.table?.trim() || !parsed.column?.trim()) {
          throw new Error('Variable query "tagValues" requires a table and a tag column');
        }
        break;
    }
    return parsed;
  }

  const text = String(query ?? '')
    .trim()
    .replace(/;+$/, '');
  const match = VARIABLE_QUERY_RE.exec(text);
  if (!match) {
    throw new Error('Unsupported variable query. Use tag_values(table, column), tables() or columns(table).');
  }
  const fn = match[1].toLowerCase();
  const args = match[2].trim() ? match[2].split(',').map((part) => stripQuotes(part.trim())) : [];
  switch (fn) {
    case 'tables':
      if (args.length !== 0) {
        throw new Error('Variable query tables() takes no arguments');
      }
      return { queryType: 'tables' };
    case 'columns':
      if (args.length !== 1 || !args[0]) {
        throw new Error('Variable query columns(table) requires exactly one table argument');
      }
      return { queryType: 'columns', table: args[0] };
    default:
      if (args.length !== 2 || !args[0] || !args[1]) {
        throw new Error('Variable query tag_values(table, column) requires exactly two arguments');
      }
      return { queryType: 'tagValues', table: args[0], column: args[1] };
  }
}

function stripQuotes(s: string): string {
  if (
    s.length >= 2 &&
    ((s.startsWith('"') && s.endsWith('"')) ||
      (s.startsWith("'") && s.endsWith("'")) ||
      (s.startsWith('`') && s.endsWith('`')))
  ) {
    return s.slice(1, -1);
  }
  return s;
}

// Renders the canonical raw query string shown as the variable definition in Grafana.
export function buildVariableQueryString(query: KwdbVariableQuery): string {
  switch (query.queryType) {
    case 'tables':
      return 'tables()';
    case 'columns':
      return `columns(${query.table ?? ''})`;
    case 'tagValues':
      return `tag_values(${query.table ?? ''}, ${query.column ?? ''})`;
  }
}
