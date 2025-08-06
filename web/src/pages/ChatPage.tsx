import { useState, useEffect } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';
import { ChatPanel } from '@/components/ChatPanel';
import PreviewPanel from '@/components/PreviewPanel';
import { ProjectSidebar } from '@/components/ProjectSidebar';
import { Button } from '@/components/ui/button';
import { Menu, X } from 'lucide-react';
import { eventManager, EVENTS } from '@/services/eventManager';

interface LocationState {
  initialMessage?: string;
  isNewChat?: boolean;
}

export const ChatPage = () => {
  const { chatId } = useParams<{ chatId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  
  const [mode, setMode] = useState<'preview' | 'edit'>('edit');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [initialMessage, setInitialMessage] = useState<string>('');

  const state = location.state as LocationState;

  useEffect(() => {
    // Handle initial message from homepage
    if (state?.initialMessage && state?.isNewChat) {
      setInitialMessage(state.initialMessage);
      // Clear the state to prevent re-triggering
      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [state, location.pathname, navigate]);

  useEffect(() => {
    // Switch to preview mode when build starts
    const unsubscribe = eventManager.on(EVENTS.PROJECT_BUILD_START, () => {
      setMode('preview');
    });

    return unsubscribe;
  }, []);

  if (!chatId) {
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
            currentChatId={chatId} 
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
          
          <div className="flex-1">
            <h1 className="text-lg font-semibold">
              Chat {chatId?.slice(0, 8)}...
            </h1>
          </div>
        </header>
        
        {/* Chat and Preview Panels */}
        <div className="flex-1 overflow-hidden">
          <ResizablePanelGroup direction="horizontal" className="h-full">
            <ResizablePanel defaultSize={40} minSize={30} maxSize={60}>
              <ChatPanel 
                chatId={chatId}
                initialMessage={initialMessage}
              />
            </ResizablePanel>
            
            <ResizableHandle className="bg-border hover:bg-primary/20 transition-colors duration-200 w-1" />
            
            <ResizablePanel defaultSize={60} minSize={40} maxSize={100}>
              <PreviewPanel mode={mode} onModeChange={setMode} />
            </ResizablePanel>
          </ResizablePanelGroup>
        </div>
      </div>
    </div>
  );
};
