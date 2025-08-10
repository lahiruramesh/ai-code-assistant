import { useQuery, useMutation } from '@tanstack/react-query';
import { modelsApi } from '@/services/api';

export const useModels = () => {
  return useQuery({
    queryKey: ['models'],
    queryFn: () => modelsApi.getAll(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useProviderInfo = () => {
  return useQuery({
    queryKey: ['provider-info'],
    queryFn: () => modelsApi.getProviderInfo(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};
