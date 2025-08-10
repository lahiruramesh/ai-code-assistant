import { useMutation } from '@tanstack/react-query';
import { chatApi } from '@/services/api';
import type { CreateSessionData, CreateSessionResponse, UseCreateSessionOptions } from '@/types';

export const useCreateSession = (options?: UseCreateSessionOptions) => {
  return useMutation({
    mutationFn: (data: CreateSessionData) => chatApi.createSession(data.message),
    onSuccess: options?.onSuccess,
    onError: options?.onError,
  });
};

export const useChatSession = () => {
  const createSessionMutation = useCreateSession();

  const createSession = async (message: string): Promise<CreateSessionResponse> => {
    return createSessionMutation.mutateAsync({ message });
  };

  return {
    createSession,
    isCreating: createSessionMutation.isPending,
    error: createSessionMutation.error,
    reset: createSessionMutation.reset,
  };
};
