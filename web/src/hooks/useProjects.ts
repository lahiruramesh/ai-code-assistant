import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '@/services/api';
import type { 
  Project, 
  CreateProjectData, 
  UseProjectsOptions, 
  UseProjectsReturn,
} from '@/types';

const PROJECTS_QUERY_KEY = ['projects'] as const;

export const useProjects = (options: UseProjectsOptions = {}): UseProjectsReturn => {
  const { enabled = true, refetchInterval } = options;
  const queryClient = useQueryClient();

  // Query for fetching projects
  const {
    data: projectsData,
    isLoading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: PROJECTS_QUERY_KEY,
    queryFn: () => projectsApi.getAll(),
    enabled,
    refetchInterval,
    select: (data) => data.projects || [],
  });

  // Mutation for creating projects
  const createProjectMutation = useMutation({
    mutationFn: (projectData: CreateProjectData) => projectsApi.create(projectData),
    onSuccess: (newProject) => {
      // Optimistically update the projects list
      queryClient.setQueryData(PROJECTS_QUERY_KEY, (old: { projects: Project[] } | undefined) => {
        if (!old) return { projects: [newProject] };
        return { projects: [newProject, ...old.projects] };
      });
    },
  });

  // Mutation for deleting projects
  const deleteProjectMutation = useMutation({
    mutationFn: (projectId: string) => projectsApi.delete(projectId),
    onSuccess: (_, projectId) => {
      // Optimistically update the projects list
      queryClient.setQueryData(PROJECTS_QUERY_KEY, (old: { projects: Project[] } | undefined) => {
        if (!old) return { projects: [] };
        return { projects: old.projects.filter(project => project.id !== projectId) };
      });
    },
  });

  const projects = projectsData || [];
  const error = queryError?.message || createProjectMutation.error?.message || deleteProjectMutation.error?.message || null;

  const createProject = async (projectData: CreateProjectData): Promise<Project> => {
    return createProjectMutation.mutateAsync(projectData);
  };

  const deleteProject = async (projectId: string): Promise<void> => {
    deleteProjectMutation.mutateAsync(projectId);
  };

  return {
    projects,
    isLoading,
    error,
    refetch: async () => {
      await refetch();
    },
    createProject,
    deleteProject,
  };
};

export const useProject = (projectId: string | null) => {
  const {
    data: project,
    isLoading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getById(projectId!),
    enabled: !!projectId,
  });

  const error = queryError?.message || null;

  return {
    project: project || null,
    isLoading,
    error,
    refetch: async () => {
      await refetch();
    },
  };
};
