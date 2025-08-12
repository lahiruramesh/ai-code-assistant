import React, { useState, useEffect, useCallback } from 'react';
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
import { useProject, useProjectPreview, useProjectFileContent } from '@/hooks';

// Type definitions
interface PreviewPanelProps {
  mode: 'preview' | 'edit';
  onModeChange: (mode: 'preview' | 'edit') => void;
  projectId: string;
}

interface ActiveFile {
  content: string;
  fileName: string;
}

const PreviewPanel: React.FC<PreviewPanelProps> = ({ mode, onModeChange, projectId }) => {
  // State for project management
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [projectPath, setProjectPath] = useState<string>('');

  // State for code editing
  const [activeFile, setActiveFile] = useState<ActiveFile | null>(null);
  const [viewMode, setViewMode] = useState<'desktop' | 'tablet' | 'mobile'>('desktop');

  // State for project building progress
  const [isBuildingProject, setIsBuildingProject] = useState(false);
  const [buildProgress, setBuildProgress] = useState(0);
  const [currentBuildStep, setCurrentBuildStep] = useState('');
  const [completedSteps, setCompletedSteps] = useState<string[]>([]);
  const [soundEnabled, setSoundEnabled] = useState(true);

  // Use hooks for data fetching
  const { project: activeProject, isLoading: projectLoading } = useProject(projectId);
  const { data: previewData, refetch: refetchPreview } = useProjectPreview(activeProject?.name);
  const { data: fileContent } = useProjectFileContent(
    activeProject?.name, 
    activeFile?.fileName
  );

  const totalBuildSteps = ['Project Setup', 'Component Creation', 'Styling', 'Build & Preview'];

  // Update preview URL when preview data changes
  useEffect(() => {
    if (previewData) {
      setPreviewUrl(previewData.preview_url);
      setProjectPath(previewData.host_path);
    }
  }, [previewData]);

  // Update file content when it changes
  useEffect(() => {
    if (fileContent && activeFile) {
      setActiveFile(prev => prev ? {
        ...prev,
        content: fileContent.content
      } : null);
    }
  }, [fileContent]);

  // Setup event listeners for project building
  useEffect(() => {
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

    const unsubscribeBuildComplete = eventManager.on(EVENTS.PROJECT_BUILD_COMPLETE, () => {
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
          refetchPreview();
        }
      }, 2000);
    });

    // Subscribe to project creation events
    const unsubscribeProjectCreated = eventManager.on('PROJECT_CREATED', (data) => {
      console.log('Project created:', data);

      // Validate data before using it
      if (data && data.projectUrl && data.projectPath) {
        setPreviewUrl(data.projectUrl);
        setProjectPath(data.projectPath);
      }
    });

    // Cleanup event listeners
    return () => {
      unsubscribeBuildStart();
      unsubscribeBuildProgress();
      unsubscribeBuildComplete();
      unsubscribeProjectCreated();
    };
  }, [soundEnabled, activeProject, refetchPreview, currentBuildStep, completedSteps]);

  const getPreviousStep = (currentStep: string): string | null => {
    const stepIndex = totalBuildSteps.indexOf(currentStep);
    return stepIndex > 0 ? totalBuildSteps[stepIndex - 1] : null;
  };

  const handleFileSelect = useCallback((filePath: string) => {
    setActiveFile({
      fileName: filePath,
      content: '' // Content will be loaded by the hook
    });
  }, []);

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
      case 'created':
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
              <Badge variant={activeProject.status === 'running' ? 'default' : 'secondary'}>
                <div className={cn("w-2 h-2 rounded-full mr-1", getStatusColor(activeProject.status))} />
                {activeProject.status}
              </Badge>
            )}
          </div>

          <div className="flex items-center gap-2">
            <Tabs value={mode} onValueChange={onModeChange as any} className="w-auto">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="preview" className="flex items-center gap-1">
                  <Eye className="w-3 h-3" />
                  Preview
                </TabsTrigger>
                <TabsTrigger value="edit" className="flex items-center gap-1">
                  <Code2 className="w-3 h-3" />
                  Edit
                </TabsTrigger>
              </TabsList>
            </Tabs>

            {mode === 'preview' && (
              <div className="flex items-center gap-1 ml-2">
                <Button
                  variant={viewMode === 'desktop' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('desktop')}
                >
                  <Monitor className="w-4 h-4" />
                </Button>
                <Button
                  variant={viewMode === 'tablet' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('tablet')}
                >
                  <Tablet className="w-4 h-4" />
                </Button>
                <Button
                  variant={viewMode === 'mobile' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('mobile')}
                >
                  <Smartphone className="w-4 h-4" />
                </Button>
              </div>
            )}

            <Button
              variant="outline"
              size="sm"
              onClick={() => refetchPreview()}
              disabled={projectLoading}
            >
              <RefreshCw className={cn("w-3 h-3 mr-1", projectLoading && "animate-spin")} />
              Refresh
            </Button>
            <a href={`http://localhost:${activeProject.port}`} target='_blank'>
              <Button
                variant="outline"
                size="sm"
              >
                View
              </Button>
            </a>
          </div>
        </div>

        {/* Project Info */}
        {activeProject && (
          <div className="flex items-center gap-4 text-sm">
            {activeProject.port && (
              <div className="flex items-center gap-1">
                <Server className="w-3 h-3 text-muted-foreground" />
                <span>Port {activeProject.port}</span>
              </div>
            )}
            {activeProject.name && (
              <div className="flex items-center gap-1">
                <FolderOpen className="w-3 h-3 text-muted-foreground" />
                <span>{activeProject.name}</span>
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
            {previewUrl ? (
              <div className="flex-1 p-4">
                <div className={cn("w-full h-full border rounded-lg overflow-hidden", getViewportClasses())}>
                  <iframe
                    src={previewUrl}
                    className="w-full h-full"
                    title="Project Preview"
                  />
                </div>
              </div>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center space-y-4">
                  <Globe className="w-16 h-16 text-muted-foreground mx-auto" />
                  <div>
                    <h3 className="text-lg font-semibold">No Preview Available</h3>
                    <p className="text-muted-foreground">
                      Your project is still being built or is not running yet.
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="h-full flex">
            <FileExplorer
              activeFile={activeFile?.fileName || ''}
              onFileSelect={handleFileSelect}
              projectName={activeProject?.name}
            />
            <div className="flex-1">
              {activeFile ? (
                <CodeEditor
                  activeFile={activeFile.fileName}
                  content={activeFile.content}
                  onContentChange={(value) => setActiveFile(prev => prev ? { ...prev, content: value } : null)}
                />
              ) : (
                <div className="h-full flex items-center justify-center">
                  <div className="text-center space-y-4">
                    <Code2 className="w-16 h-16 text-muted-foreground mx-auto" />
                    <div>
                      <h3 className="text-lg font-semibold">Select a File</h3>
                      <p className="text-muted-foreground">
                        Choose a file from the explorer to start editing.
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default PreviewPanel;
