import React, { useState, useEffect } from 'react';
import { 
  RefreshCw, 
  Monitor, 
  Tablet, 
  Smartphone, 
  Globe, 
  ExternalLink, 
  Eye, 
  Code2,
  Server,
  FolderOpen,
  Volume2,
  VolumeX
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Switch } from '@/components/ui/switch';
import { FileExplorer } from './FileExplorer';
import { CodeEditor } from './CodeEditor';
import { cn } from '@/lib/utils';
import { ProjectLoadingScreen } from './ProjectLoadingScreen';
import { soundService } from '@/services/soundService';
import { eventManager, EVENTS } from '@/services/eventManager';

// Type definitions
interface Project {
  name: string;
  status: 'running' | 'stopped' | 'building';
  port?: number;
}

interface ProjectPreview {
  preview_url: string;
  host_path: string;
}

interface PreviewPanelProps {
  mode: 'preview' | 'edit';
  onModeChange: (mode: 'preview' | 'edit') => void;
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8084/api/v1';

const PreviewPanel: React.FC<PreviewPanelProps> = ({ mode, onModeChange }) => {
  // State for project management
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProject, setActiveProject] = useState<Project | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [projectPath, setProjectPath] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);

  // State for code editing
  const [activeFile, setActiveFile] = useState<string>('src/App.tsx');
  const [viewMode, setViewMode] = useState<'desktop' | 'tablet' | 'mobile'>('desktop');
  
  // State for project building progress
  const [isBuildingProject, setIsBuildingProject] = useState(false);
  const [buildProgress, setBuildProgress] = useState(0);
  const [currentBuildStep, setCurrentBuildStep] = useState('');
  const [completedSteps, setCompletedSteps] = useState<string[]>([]);
  const [soundEnabled, setSoundEnabled] = useState(true);
  
  const totalBuildSteps = ['Project Setup', 'Component Creation', 'Styling', 'Build & Preview'];
  
  const [fileContents, setFileContents] = useState<Record<string, string>>({
    'src/App.tsx': `import React from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export const App = () => {
  return (
    <div className="min-h-screen bg-gradient-hero flex items-center justify-center">
      <Card className="max-w-2xl mx-auto card-glass">
        <CardHeader className="text-center">
          <CardTitle className="text-4xl font-bold bg-gradient-primary bg-clip-text text-transparent">
            Welcome to AI Builder
          </CardTitle>
        </CardHeader>
        <CardContent className="text-center space-y-6">
          <p className="text-lg text-muted-foreground">
            Build amazing applications with the power of AI
          </p>
          <div className="flex gap-4 justify-center">
            <Button className="btn-gradient">
              Get Started
            </Button>
            <Button variant="outline" className="btn-glass">
              Learn More
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};`,
    'package.json': `{
  "name": "ai-generated-app",
  "version": "0.1.0",
  "private": true,
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "test": "react-scripts test",
    "eject": "react-scripts eject"
  }
}`
  });

  // Fetch projects on component mount and setup event listeners
  useEffect(() => {
    fetchProjects();
    
    // Configure sound service
    soundService.setEnabled(soundEnabled);

    // Subscribe to build events
    const unsubscribeBuildStart = eventManager.on(EVENTS.PROJECT_BUILD_START, (data) => {
      setIsBuildingProject(true);
      setBuildProgress(data.progress || 0);
      setCurrentBuildStep(data.step || 'Project Setup');
      setCompletedSteps([]);
      
      if (soundEnabled) {
        soundService.playAgentActivity();
      }
    });

    const unsubscribeBuildProgress = eventManager.on(EVENTS.PROJECT_BUILD_PROGRESS, (data) => {
      setBuildProgress(data.progress || 0);
      setCurrentBuildStep(data.step || currentBuildStep);
      
      // Add to completed steps if not already there
      if (data.step && !completedSteps.includes(data.step)) {
        const prevStep = getPreviousStep(data.step);
        if (prevStep && !completedSteps.includes(prevStep)) {
          setCompletedSteps(prev => [...prev, prevStep]);
        }
      }
      
      if (soundEnabled) {
        soundService.playStepComplete();
      }
    });

    const unsubscribeBuildComplete = eventManager.on(EVENTS.PROJECT_BUILD_COMPLETE, (data) => {
      setBuildProgress(100);
      setCompletedSteps(totalBuildSteps);
      setCurrentBuildStep('Complete');
      
      // Hide loading screen after a short delay
      setTimeout(() => {
        setIsBuildingProject(false);
        
        if (soundEnabled) {
          soundService.playBuildComplete();
        }
        
        // Refresh project preview
        if (activeProject) {
          fetchProjectPreview(activeProject.name);
        }
      }, 2000);
    });

    // Cleanup event listeners
    return () => {
      unsubscribeBuildStart();
      unsubscribeBuildProgress();
      unsubscribeBuildComplete();
    };
  }, [soundEnabled, activeProject, completedSteps, currentBuildStep]);

  const getPreviousStep = (currentStep: string): string | null => {
    const stepIndex = totalBuildSteps.indexOf(currentStep);
    return stepIndex > 0 ? totalBuildSteps[stepIndex - 1] : null;
  };

  const fetchProjects = async () => {
    try {
      const response = await fetch(`${API_BASE}/projects`);
      const data = await response.json();
      setProjects(data.projects || []);
      
      if (data.projects && data.projects.length > 0) {
        setActiveProject(data.projects[0]);
        await fetchProjectPreview(data.projects[0].name);
      }
    } catch (error) {
      console.error('Failed to fetch projects:', error);
    }
  };

  const fetchProjectPreview = async (projectName: string) => {
    try {
      setIsLoading(true);
      const response = await fetch(`${API_BASE}/projects/${projectName}/preview`);
      const data: ProjectPreview = await response.json();
      
      setPreviewUrl(data.preview_url);
      setProjectPath(data.host_path);
    } catch (error) {
      console.error('Failed to fetch project preview:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileSelect = async (filePath: string) => {
    setActiveFile(filePath);
    
    // Fetch file content if not already loaded
    if (!fileContents[filePath] && activeProject) {
      try {
        // Updated to use the aiagent source for Monaco editor
        const response = await fetch(`${API_BASE}/projects/${activeProject.name}/files/${filePath}?source=aiagent`);
        if (response.ok) {
          const data = await response.json();
          setFileContents(prev => ({
            ...prev,
            [filePath]: data.content
          }));
        } else {
          console.error('Failed to fetch file content:', response.statusText);
          setFileContents(prev => ({
            ...prev,
            [filePath]: `// Failed to load file: ${filePath}`
          }));
        }
      } catch (error) {
        console.error('Error fetching file content:', error);
        setFileContents(prev => ({
          ...prev,
          [filePath]: `// Error loading file: ${filePath}`
        }));
      }
    }
  };

  const getViewportClasses = () => {
    switch (viewMode) {
      case 'mobile':
        return 'max-w-sm mx-auto';
      case 'tablet':
        return 'max-w-2xl mx-auto';
      default:
        return 'w-full';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'bg-green-500';
      case 'building':
        return 'bg-yellow-500';
      case 'stopped':
        return 'bg-red-500';
      default:
        return 'bg-gray-500';
    }
  };

  return (
    <div className="h-full flex flex-col bg-background">
      {/* Header */}
      <div className="flex-shrink-0 p-4 border-b border-border/50 bg-card/50 backdrop-blur-sm">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <h2 className="font-semibold text-lg">Project Preview</h2>
            {activeProject && (
              <Badge variant="outline" className="text-xs">
                {activeProject.name}
              </Badge>
            )}
          </div>
          
          <div className="flex items-center gap-2">
            <Tabs value={mode} onValueChange={onModeChange as any} className="w-auto">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="preview" className="text-xs">
                  <Eye className="w-3 h-3 mr-1" />
                  Preview
                </TabsTrigger>
                <TabsTrigger value="edit" className="text-xs">
                  <Code2 className="w-3 h-3 mr-1" />
                  Code
                </TabsTrigger>
              </TabsList>
            </Tabs>
            
            {/* Sound Toggle */}
            <div className="flex items-center gap-2 ml-4">
              {soundEnabled ? (
                <Volume2 className="w-4 h-4 text-muted-foreground" />
              ) : (
                <VolumeX className="w-4 h-4 text-muted-foreground" />
              )}
              <Switch
                checked={soundEnabled}
                onCheckedChange={setSoundEnabled}
                className="scale-75"
              />
            </div>
          </div>
        </div>

        {/* Project Info */}
        {activeProject && (
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${getStatusColor(activeProject.status)}`} />
              <span className="text-muted-foreground">Status:</span>
              <span className="font-medium">{activeProject.status}</span>
            </div>
            {activeProject.port && (
              <div className="flex items-center gap-2">
                <Server className="w-3 h-3 text-muted-foreground" />
                <span className="text-muted-foreground">Port:</span>
                <span className="font-medium">{activeProject.port}</span>
              </div>
            )}
            {projectPath && (
              <div className="flex items-center gap-2">
                <FolderOpen className="w-3 h-3 text-muted-foreground" />
                <span className="text-muted-foreground">Path:</span>
                <span className="font-mono text-xs">{projectPath}</span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        {isBuildingProject ? (
          <ProjectLoadingScreen
            currentStep={currentBuildStep}
            progress={buildProgress}
            completedSteps={completedSteps}
            totalSteps={totalBuildSteps}
          />
        ) : mode === 'preview' ? (
          <div className="h-full flex flex-col">
            {/* Preview Controls */}
            <div className="flex-shrink-0 p-4 border-b border-border/50 bg-muted/30">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {/* Viewport Controls */}
                  <div className="flex items-center border border-border rounded-lg p-1">
                    <Button
                      variant={viewMode === 'desktop' ? 'default' : 'ghost'}
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => setViewMode('desktop')}
                    >
                      <Monitor className="w-4 h-4" />
                    </Button>
                    <Button
                      variant={viewMode === 'tablet' ? 'default' : 'ghost'}
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => setViewMode('tablet')}
                    >
                      <Tablet className="w-4 h-4" />
                    </Button>
                    <Button
                      variant={viewMode === 'mobile' ? 'default' : 'ghost'}
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => setViewMode('mobile')}
                    >
                      <Smartphone className="w-4 h-4" />
                    </Button>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {previewUrl && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => window.open(previewUrl, '_blank')}
                    >
                      <ExternalLink className="w-3 h-3 mr-1" />
                      Open in New Tab
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => fetchProjectPreview(activeProject?.name || '')}
                    disabled={isLoading}
                  >
                    <RefreshCw className={cn("w-3 h-3 mr-1", isLoading && "animate-spin")} />
                    Refresh
                  </Button>
                </div>
              </div>
            </div>

            {/* Preview Content */}
            <div className="flex-1 overflow-auto p-4 bg-muted/20">
              <div className={cn("transition-all duration-300", getViewportClasses())}>
                {previewUrl ? (
                  <div className="bg-white rounded-lg shadow-lg overflow-hidden border border-border">
                    <iframe
                      src={previewUrl}
                      className="w-full h-[600px]"
                      title="App Preview"
                      onError={() => {
                        console.error('Failed to load preview iframe');
                      }}
                    />
                  </div>
                ) : (
                  <Card className="text-center p-8">
                    <CardContent className="space-y-4">
                      <Globe className="w-12 h-12 mx-auto text-muted-foreground" />
                      <div>
                        <h3 className="font-semibold mb-2">No Preview Available</h3>
                        <p className="text-sm text-muted-foreground">
                          Start building an application to see the preview here.
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            </div>
          </div>
        ) : (
          <div className="h-full flex">
            {/* File Explorer */}
            <div className="w-1/3 border-r border-border/50">
              <FileExplorer 
                onFileSelect={handleFileSelect}
                activeFile={activeFile}
                projectName={activeProject?.name}
              />
            </div>
            
            {/* Code Editor */}
            <div className="flex-1">
                            <CodeEditor
                activeFile={activeFile || 'untitled.txt'}
                content={fileContents[activeFile] || ''}
                onContentChange={(value) => {
                  setFileContents(prev => ({
                    ...prev,
                    [activeFile]: value
                  }));
                }}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default PreviewPanel;
