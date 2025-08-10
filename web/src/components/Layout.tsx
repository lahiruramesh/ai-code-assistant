import React, { memo } from 'react';
import { ProjectSidebar } from './ProjectSidebar';
import { Header } from './Header';
import { Footer } from './Footer';
import { SidebarProvider } from '@/components/ui/sidebar';
import type { LayoutProps } from '@/types';
import type { Project } from '@/types';
import { cn } from '@/lib/utils';

interface ExtendedLayoutProps extends LayoutProps {
  currentProjectId?: string | null;
  project?: Project | null;
  className?: string
}

export const Layout: React.FC<ExtendedLayoutProps> = memo(({
  children,
  showSidebar = true,
  showTopNav = true,
  currentProjectId = null,
  project = null,
  className
}) => {
  if (!showSidebar && !showTopNav) {
    return (
      <div className="min-h-screen flex flex-col w-full">
        <main className="flex-1">
          {children}
        </main>
      </div>
    );
  }

  if (!showSidebar) {
    return (
      <div className="min-h-screen flex flex-col">
        {showTopNav && (
          <Header 
            project={project} 
            showIntegrations={!!project} 
          />
        )}
        <main className="flex-1">
          {children}
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <SidebarProvider>
      <div className={cn("min-h-screen flex", className)}>
        <ProjectSidebar currentProjectId={currentProjectId} />
        
        <div className="flex-1 flex flex-col min-w-0">
          {showTopNav && (
            <Header 
              project={project} 
              showIntegrations={!!project}
              showSidebarTrigger={true}
            />
          )}
          
          <main className="overflow-auto">
            {children}
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
});

Layout.displayName = 'Layout';