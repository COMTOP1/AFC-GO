import { useSearchParams } from 'react-router';

export function useSearchQuery(): string {
  const [params] = useSearchParams();
  return params.get('q') ?? '';
}
