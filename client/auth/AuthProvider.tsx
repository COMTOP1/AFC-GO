import { useQueryClient } from '@tanstack/react-query';
import { useMemo, type ReactNode } from 'react';

import { queryKeys, useMe } from '../api/queries';
import { AuthContext, type AuthState } from './context';

export function AuthProvider({ children }: { children: ReactNode }) {
  const me = useMe();
  const queryClient = useQueryClient();

  const value = useMemo<AuthState>(
    () => ({
      user: me.data ?? null,
      isLoading: me.isPending,
      refresh: () => queryClient.invalidateQueries({ queryKey: queryKeys.me }),
    }),
    [me.data, me.isPending, queryClient],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
