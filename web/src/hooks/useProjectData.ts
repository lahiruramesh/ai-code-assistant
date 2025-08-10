import { useQuery } from '@tanstack/react-query';
import { projectsApi } from '@/services/api';

export const useProjectFiles = (projectName: string | undefined) => {
  return useQuery({
    queryKey: ['project-files', projectName],
    queryFn: () => projectsApi.getFiles(projectName!),
    enabled: !!projectName,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useProjectFileContent = (projectName: string | undefined, filePath: string | undefined) => {
  return useQuery({
    queryKey: ['project-file-content', projectName, filePath],
    queryFn: () => projectsApi.getFileContent(projectName!, filePath!),
    enabled: !!projectName && !!filePath,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
};

export const useProjectPreview = (projectName: string | undefined) => {
  return useQuery({
    queryKey: ['project-preview', projectName],
    queryFn: () => projectsApi.getPreview(projectName!),
    enabled: !!projectName,
    staleTime: 30 * 1000, // 30 seconds
  });
};
