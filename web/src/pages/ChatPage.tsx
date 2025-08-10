import { useState, useEffect } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';
import { ChatPanel } from '@/components/ChatPanel';
import PreviewPanel from '@/components/PreviewPanel';
import { ProjectSidebar } from '@/components/ProjectSidebar';
import { Button } from '@/components/ui/button';
import { Menu, X, RefreshCw } from 'lucide-react';
import { eventManager, EVENTS } from '@/services/eventManager';

interface LocationState {
  initialMessage?: string;
  isNewChat?: boolean;
}

export const ChatPage = () => {
  const { chatId, projectId } = useParams<{ chatId?: string; projectId?: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  
  const [mode, setMode] = useState<'preview' | 'edit'>('edit');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [initialMessage, setInitialMessage] = useState<string>('');
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null);
  const [projectData, setProjectData] = useState<any>(null);
  const [isLoadingProject, setIsLoadingProject] = useState(false);

  const state = location.state as LocationState;
  const activeId = projectId || chatId;

  useEffect(() => {
    // Handle initial message from homepage
    if (state?.initialMessage && state?.isNewChat) {
      setInitialMessage(state.initialMessage);
      // Clear the state to prevent re-triggering
      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [state, location.pathname, navigate]);

  // Load project data if we have a projectId
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId);
      loadProjectData(projectId);
    }
  }, [projectId]);

  const loadProjectData = async (id: string) => {
    setIsLoadingProject(true);
    try {
      const response = await fetch(`http://localhost:8084/api/v1/projects/${id}`);
      if (response.ok) {
        const data = await response.json();
        setProjectData(data);
        
        // Emit project data for preview panel
        eventManager.emit('PROJECT_CREATED', {
          projectId: data.id,
          projectName: data.name,
          projectPath: data.name,
          projectUrl: data.url || `http://localhost:${data.port}`
        });
        
        // If container was started, show a brief message
        if (data.container_started) {
          console.log('Container was started automatically');
        }
      } else {
        console.error('Failed to load project:', response.status);
      }
    } catch (error) {
      console.error('Failed to load project data:', error);
    } finally {
      setIsLoadingProject(false);
    }
  };

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
    <div className="h-screen bg-gradient-subtle flex">
        {/* Sidebar */}
        <div className={`fixed inset-y-0 left-0 z-50 w-80 transform transition-transform duration-300 ease-in-out lg:relative lg:translate-x-0 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}>
          <ProjectSidebar 
            currentProjectId={currentProjectId}
            onClose={() => setSidebarOpen(false)}
          />
        </div>

        {/* Sidebar overlay for mobile */}
        {sidebarOpen && (
          <div 
            className="fixed inset-0 z-40 bg-black/50 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        {/* Main content */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Header */}
          <header className="flex-shrink-0 sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
            <Button
              variant="ghost"
              size="sm"
              className="lg:hidden"
              onClick={() => setSidebarOpen(true)}
            >
            <Menu className="h-4 w-4" />
          </Button>
          
          <div className="flex-1 flex items-center justify-between">
            <h1 className="text-lg font-semibold">
              {projectData ? projectData.name : `Chat ${activeId?.slice(0, 8)}...`}
            </h1>
            
            {projectId && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => loadProjectData(projectId)}
                disabled={isLoadingProject}
                className="ml-2"
              >
                <RefreshCw className={`h-4 w-4 ${isLoadingProject ? 'animate-spin' : ''}`} />
              </Button>
            )}
          </div>
        </header>
        
        {/* Chat and Preview Panels */}
        <div className="flex-1 overflow-hidden">
          {isLoadingProject ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center space-y-4">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
                <p className="text-muted-foreground">Loading project...</p>
              </div>
            </div>
          ) : (
            <ResizablePanelGroup direction="horizontal" className="h-full">
              <ResizablePanel defaultSize={40} minSize={30} maxSize={60}>
                <ChatPanel 
                  chatId={chatId}
                  projectId={currentProjectId}
                  initialMessage={initialMessage}
                />
              </ResizablePanel>
              
              <ResizableHandle className="bg-border hover:bg-primary/20 transition-colors duration-200 w-1" />
              
              <ResizablePanel defaultSize={60} minSize={40} maxSize={100}>
                <PreviewPanel mode={mode} onModeChange={setMode} projectId={currentProjectId} />
              </ResizablePanel>
            </ResizablePanelGroup>
          )}
        </div>
      </div>
    </div>
  );
};
