import { useState, useEffect } from 'react';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable';
import { ChatPanel } from './ChatPanel';
import PreviewPanel from './PreviewPanel';
import { TopNavigation } from './TopNavigation';
import { eventManager, EVENTS } from '@/services/eventManager';

export const AIBuilder = () => {
  const [mode, setMode] = useState<'preview' | 'edit'>('edit');

  useEffect(() => {
    // Switch to preview mode when build starts
    const unsubscribe = eventManager.on(EVENTS.PROJECT_BUILD_START, () => {
      setMode('preview');
    });

    return unsubscribe;
  }, []);

  return (
    <div className="h-screen bg-gradient-subtle flex flex-col">
      {/* Fixed Header */}
      <div className="flex-shrink-0 sticky top-0 z-50">
        <TopNavigation />
      </div>
      
      {/* Main Content Area with proper height */}
      <div className="flex-1 h-full overflow-hidden">
        <ResizablePanelGroup direction="horizontal" className="h-full">
          <ResizablePanel defaultSize={40} minSize={30} maxSize={60}>
            <ChatPanel />
          </ResizablePanel>
          
          <ResizableHandle className="bg-border hover:bg-primary/20 transition-colors duration-200 w-1" />
          
          <ResizablePanel defaultSize={60} minSize={40} maxSize={100}>
            <PreviewPanel mode={mode} onModeChange={setMode} />
          </ResizablePanel>
        </ResizablePanelGroup>
      </div>
    </div>
  );
};