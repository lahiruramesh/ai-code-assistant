import { useState, useEffect } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';
import { ChatPanel } from '@/components/ChatPanel';
import PreviewPanel from '@/components/PreviewPanel';
import { Layout } from '@/components/Layout';
import { eventManager, EVENTS } from '@/services/eventManager';
import { useAuth } from '@/components/AuthProvider';
import { useProject } from '@/hooks/useProjects';

interface LocationState {
  initialMessage?: string;
  isNewChat?: boolean;
}

export const ProjectPage = () => {
  const { chatId, projectId } = useParams<{ chatId?: string; projectId?: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();
  
  const [mode, setMode] = useState<'preview' | 'edit'>('edit');
  const [initialMessage, setInitialMessage] = useState<string>('');
  const activeId = projectId || chatId;

  // Use the project hook to get project data
  const { project, isLoading: isLoadingProject, refetch } = useProject(projectId || null);

  const state = location.state as LocationState;

  useEffect(() => {
    // Handle initial message from homepage
    if (state?.initialMessage && state?.isNewChat) {
      setInitialMessage(state.initialMessage);
      // Clear the state to prevent re-triggering
      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [state, location.pathname, navigate]);

  // Emit project data for preview panel when project loads
  useEffect(() => {
    if (project) {
      eventManager.emit('PROJECT_CREATED', {
        projectId: project.id,
        projectName: project.name,
        projectPath: project.name,
        projectUrl: project.port ? `http://localhost:${project.port}` : undefined
      });
    }
  }, [project]);

  useEffect(() => {
    // Switch to preview mode when build starts
    const unsubscribe = eventManager.on(EVENTS.PROJECT_BUILD_START, () => {
      setMode('preview');
    });

    return unsubscribe;
  }, []);

  if (!activeId) {
    navigate('/');
    return null;
  }

  return (
    <Layout 
      showSidebar={true} 
      showTopNav={true} 
      currentProjectId={projectId || null}
      project={project}
      className="w-full"
    >
      <div className="h-[calc(100vh-4.34rem)] flex flex-col w-full">        
        {/* Main Content */}
        <div className="flex-1 overflow-hidden">
          {isLoadingProject ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center space-y-4">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
                <p className="text-muted-foreground">Loading project...</p>
              </div>
            </div>
          ) : (
            <ResizablePanelGroup direction="horizontal" className="h-full w-full">
              <ResizablePanel defaultSize={30} minSize={30} maxSize={60}>
                <ChatPanel 
                  chatId={chatId}
                  projectId={projectId || null}
                  initialMessage={initialMessage}
                />
              </ResizablePanel>
              
              <ResizableHandle className="bg-border hover:bg-primary/20 transition-colors duration-200 w-1" />
              
              <ResizablePanel defaultSize={70} minSize={40} maxSize={100}>
                <PreviewPanel mode={mode} onModeChange={setMode} projectId={projectId || null} />
              </ResizablePanel>
            </ResizablePanelGroup>
          )}
        </div>
      </div>
    </Layout>
  );
};
