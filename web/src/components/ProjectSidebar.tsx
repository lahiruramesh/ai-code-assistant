import React, { useState, useEffect, memo } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { 
  FolderPlus, 
  MessageSquare, 
  LogOut, 
  ChevronRight,
  Github,
  ExternalLink,
  Folder,
  Zap,
  Plus,
  MoreHorizontal,
  Trash2
} from 'lucide-react'
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuth } from './AuthProvider'
import { NewProjectModal } from './NewProjectModal'
import { DeleteProjectDialog } from './DeleteProjectDialog'
import { useProjects } from '@/hooks/useProjects'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import type { Project as ProjectType } from '@/types'

interface Project {
  id: string;
  name: string;
  template: string;
  status: string;
  created_at: string;
  docker_container?: string;
  port?: number;
}

interface ProjectSidebarProps {
  currentProjectId: string | null;
}

export const ProjectSidebar: React.FC<ProjectSidebarProps> = memo(({ 
  currentProjectId,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { projects, isLoading, refetch, deleteProject } = useProjects();
  const [isProjectsOpen, setIsProjectsOpen] = useState(true);
  const [isNewProjectModalOpen, setIsNewProjectModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [projectToDelete, setProjectToDelete] = useState<ProjectType | null>(null);
  const [isDeletingProject, setIsDeletingProject] = useState(false);

  const handleProjectClick = (project: Project) => {
    navigate(`/projects/${project.id}`);
  };

  const handleNewProject = () => {
    setIsNewProjectModalOpen(true);
  };

  const handleProjectCreated = (projectId: string) => {
    refetch();
    navigate(`/projects/${projectId}`);
  };

  const handleDeleteProject = (project: ProjectType) => {
    setProjectToDelete(project);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async (projectId: string) => {
    setIsDeletingProject(true);
    try {
      await deleteProject(projectId);
      setDeleteDialogOpen(false);
      setProjectToDelete(null);
      // Navigate away if we deleted the current project
      if (currentProjectId === projectId) {
        navigate('/');
      }
    } catch (error) {
      console.error('Failed to delete project:', error);
    } finally {
      setIsDeletingProject(false);
    }
  };

  const handleCloseDeleteDialog = () => {
    if (!isDeletingProject) {
      setDeleteDialogOpen(false);
      setProjectToDelete(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <>
      <Sidebar className="border-r">
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>Quick Actions</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton onClick={handleNewProject}>
                    <Plus className="w-4 h-4" />
                    New Project
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <Collapsible open={isProjectsOpen} onOpenChange={setIsProjectsOpen}>
              <SidebarGroupLabel asChild>
                <CollapsibleTrigger className="flex items-center justify-between w-full">
                  <span>Projects ({projects.length})</span>
                  <ChevronRight className={`w-4 h-4 transition-transform ${isProjectsOpen ? 'transform rotate-90' : ''}`} />
                </CollapsibleTrigger>
              </SidebarGroupLabel>
              <CollapsibleContent>
                <SidebarGroupContent>
                  <SidebarMenu>
                    {isLoading && projects.length == 0 ? (
                      <SidebarMenuItem>
                        <div className="p-2 text-sm text-muted-foreground">Loading projects...</div>
                      </SidebarMenuItem>
                    ) : !Array.isArray(projects) || projects.length === 0 ? (
                      <SidebarMenuItem>
                        <div className="p-2 text-sm text-muted-foreground">No projects yet</div>
                      </SidebarMenuItem>
                    ) : (
                      projects.map((project) => (
                        <SidebarMenuItem key={project.id}>
                          <div className="flex items-center w-full group">
                            <SidebarMenuButton
                              onClick={() => handleProjectClick(project)}
                              isActive={currentProjectId === project.id}
                              className="flex flex-col items-start gap-1 h-auto p-3 flex-1"
                            >
                              <div className="flex flex-col items-start justify-between w-full">
                                <div className="flex items-center gap-2">
                                  <Folder className="w-4 h-4" />
                                  <span className="font-medium truncate">{project.name}</span>
                                </div>
                              </div>
                              <div className="flex items-center justify-between w-full text-xs text-muted-foreground">
                                <span>{project.template}</span>
                                <span>{formatDate(project.created_at)}</span>
                              </div>
                              {project.port && (
                                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                                  <ExternalLink className="w-3 h-3" />
                                  <span>localhost:{project.port}</span>
                                </div>
                              )}
                            </SidebarMenuButton>
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="opacity-0 group-hover:opacity-100 transition-opacity h-8 w-8 p-0"
                                  onClick={(e) => e.stopPropagation()}
                                >
                                  <MoreHorizontal className="w-4 h-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem
                                  onClick={() => handleDeleteProject(project)}
                                  className="text-destructive focus:text-destructive"
                                >
                                  <Trash2 className="w-4 h-4 mr-2" />
                                  Delete Project
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
                        </SidebarMenuItem>
                      ))
                    )}
                  </SidebarMenu>
                </SidebarGroupContent>
              </CollapsibleContent>
            </Collapsible>
          </SidebarGroup>

          {user?.github_connected && (
            <SidebarGroup>
              <SidebarGroupLabel>Integrations</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton>
                      <Github className="w-4 h-4" />
                      <span>GitHub</span>
                      <Badge variant="secondary" className="text-green-700 bg-green-100 ml-auto">
                        Connected
                      </Badge>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  {user?.vercel_connected && (
                    <SidebarMenuItem>
                      <SidebarMenuButton>
                        <Zap className="w-4 h-4" />
                        <span>Vercel</span>
                        <Badge variant="secondary" className="text-green-700 bg-green-100 ml-auto">
                          Connected
                        </Badge>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          )}
        </SidebarContent>

        <SidebarFooter className="border-t">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton>
                <Avatar className="w-6 h-6">
                  <AvatarImage src={user?.avatar_url} alt={user?.name} />
                  <AvatarFallback>{user?.name?.charAt(0)}</AvatarFallback>
                </Avatar>
                <div className="flex flex-col items-start text-left">
                  <span className="text-sm font-medium">{user?.name}</span>
                  <span className="text-xs text-muted-foreground">{user?.email}</span>
                </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <SidebarMenuButton onClick={logout}>
                <LogOut className="w-4 h-4" />
                Sign Out
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </Sidebar>
      
      <NewProjectModal
        isOpen={isNewProjectModalOpen}
        onClose={() => setIsNewProjectModalOpen(false)}
        onSuccess={handleProjectCreated}
      />

      <DeleteProjectDialog
        project={projectToDelete}
        isOpen={deleteDialogOpen}
        onClose={handleCloseDeleteDialog}
        onConfirm={handleConfirmDelete}
        isDeleting={isDeletingProject}
      />
    </>
  );
});

ProjectSidebar.displayName = 'ProjectSidebar';
