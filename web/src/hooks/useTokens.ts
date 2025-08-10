import { useQuery } from '@tanstack/react-query';
import { tokensApi } from '@/services/api';

export const useSessionTokenUsage = (sessionId: string | undefined) => {
  return useQuery({
    queryKey: ['session-token-usage', sessionId],
    queryFn: () => tokensApi.getSessionUsage(sessionId!),
    enabled: !!sessionId,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
};

export const useGlobalTokenStats = () => {
  return useQuery({
    queryKey: ['global-token-stats'],
    queryFn: () => tokensApi.getGlobalStats(),
    refetchInterval: 60000, // Refetch every minute
  });
};
