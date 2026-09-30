import { useSearchParams } from 'react-router';

/** The current value of a tab-like search param, or the default when absent/unknown. */
export function useTabParam(param: string, values: string[], defaultValue: string): string {
  const [params] = useSearchParams();
  const value = params.get(param);
  return value !== null && values.includes(value) ? value : defaultValue;
}
