import { useCallback, useRef, useState } from 'react';

export interface URLStateField<T> {
  default: T;
  parse: (raw: string) => T;
  serialize: (value: T) => string;
  isValid?: (value: T) => boolean;
}

export type URLStateSchema = { [key: string]: URLStateField<any> };

export type URLStateValues<S extends URLStateSchema> = {
  [K in keyof S]: S[K]['default'];
};

/**
 * Reads a schema's keys from a query string once, falling back to each
 * field's default when the value is missing, unparsable, or fails isValid.
 * A plain function (not a hook) so it can be read once on mount via useRef.
 */
export function readURLState<S extends URLStateSchema>(
  schema: S,
  search: string = window.location.search
): URLStateValues<S> {
  const params = new URLSearchParams(search);
  const result = {} as URLStateValues<S>;

  for (const key of Object.keys(schema)) {
    const field = schema[key];
    const raw = params.get(key);

    if (raw === null) {
      result[key as keyof S] = field.default;
      continue;
    }

    try {
      const parsed = field.parse(raw);
      result[key as keyof S] = field.isValid && !field.isValid(parsed) ? field.default : parsed;
    } catch {
      result[key as keyof S] = field.default;
    }
  }

  return result;
}

function writeURLState<S extends URLStateSchema>(schema: S, values: URLStateValues<S>) {
  const params = new URLSearchParams(window.location.search);

  for (const key of Object.keys(schema)) {
    const field = schema[key];
    const value = values[key as keyof S];

    if (value === field.default) {
      params.delete(key);
    } else {
      params.set(key, field.serialize(value));
    }
  }

  const query = params.toString();
  const newUrl = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash}`;
  window.history.replaceState(null, '', newUrl);
}

/**
 * Shared URL-backed state. Every patchState() call re-reads window.location.search
 * fresh and only touches this schema's own keys, so independent useURLState
 * instances (different schemas) never clobber each other's query params.
 */
export function useURLState<S extends URLStateSchema>(
  schema: S
): [URLStateValues<S>, (patch: Partial<URLStateValues<S>>) => void] {
  const initial = useRef(readURLState(schema)).current;
  const [state, setState] = useState<URLStateValues<S>>(initial);

  const patchState = useCallback(
    (patch: Partial<URLStateValues<S>>) => {
      const current = readURLState(schema);
      const merged = { ...current, ...patch };
      writeURLState(schema, merged);
      setState(merged);
    },
    [schema]
  );

  return [state, patchState];
}
