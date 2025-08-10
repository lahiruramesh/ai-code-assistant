import { useQuery } from '@tanstack/react-query';
import { chatApi } from '@/services/api';

export const useChatMessages = (projectId: string | undefined) => {
  return useQuery({
    queryKey: ['chat-messages', projectId],
    queryFn: () => chatApi.getMessages(projectId!),
    enabled: !!projectId,
  });
};

export const useChat = (chatId: string | undefined) => {
  return useQuery({
    queryKey: ['chat', chatId],
    queryFn: () => chatApi.getChatById(chatId!),
    enabled: !!chatId,
  });
};
