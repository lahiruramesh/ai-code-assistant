import React, { memo, useCallback, useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import {
  Loader2,
  Rocket
} from 'lucide-react';
import { useProjects } from '@/hooks/useProjects';
import type { CreateProjectData, ProjectTemplate } from '@/types';

interface NewProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (projectId: string) => void;
}

const PROJECT_TEMPLATES: ProjectTemplate[] = [
  {
    id: 'react-vite',
    name: 'React + Vite',
    description: 'Modern React app with Vite, TypeScript, and Tailwind CSS',
    tags: ['React', 'TypeScript', 'Vite', 'Tailwind'],
    icon: '⚛️',
    popular: true,
  },
  {
    id: 'next-js',
    name: 'Next.js',
    description: 'Full-stack React framework with SSR and API routes',
    tags: ['Next.js', 'React', 'TypeScript', 'API Routes'],
    icon: '▲',
    popular: true,
  },
];

export const NewProjectModal: React.FC<NewProjectModalProps> = memo(({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const { createProject } = useProjects();
  const [selectedTemplate, setSelectedTemplate] = useState<ProjectTemplate | null>(null);
  const [projectName, setProjectName] = useState('');
  const [message, setMessage] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Set projectName when template changes and projectName is empty
  useEffect(() => {
    if (selectedTemplate && !projectName) {
      setProjectName(`my-${selectedTemplate.id.replace(/-/g, '-')}-project`);
    }
    // Clear error if all fields are now valid
    if (selectedTemplate && projectName.trim()) {
      setError(null);
    }
  }, [selectedTemplate, projectName]);

  const handleTemplateSelect = useCallback((template: ProjectTemplate) => {
    setSelectedTemplate(template);
  }, []);

  const handleCreateProject = useCallback(async () => {
    if (!selectedTemplate || !projectName.trim()) {
      setError('Please select a template and enter a project name');
      return;
    }

    setIsCreating(true);
    setError(null);

    try {
      const projectData: CreateProjectData = {
        name: projectName.trim(),
        template: selectedTemplate.id,
        message: message.trim() || undefined,
      };

      const project = await createProject(projectData);
      if (project.id) {
        onSuccess(project.id);
      }
      onClose();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create project';
      setError(errorMessage);
    } finally {
      setIsCreating(false);
    }
  }, [selectedTemplate, projectName, message, createProject, onSuccess, onClose]);

  const handleClose = useCallback(() => {
    if (!isCreating) {
      setSelectedTemplate(null);
      setProjectName('');
      setMessage('');
      setError(null);
      onClose();
    }
  }, [
    setSelectedTemplate,
    setProjectName,
    setMessage,
    setError
  ]);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Rocket className="w-5 h-5" />
            Create New Project
          </DialogTitle>
          <DialogDescription>
            Choose a template and configure your new project. We'll set everything up for you.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Template Selection */}
          <div className="space-y-4">
            <Label className="text-base font-medium">Choose a Template</Label>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {PROJECT_TEMPLATES.map((template) => (
                <Card
                  key={template.id}
                  className={`cursor-pointer transition-all hover:shadow-md ${selectedTemplate?.id === template.id
                      ? 'ring-2 ring-primary border-primary'
                      : ''
                    }`}
                  onClick={() => handleTemplateSelect(template)}
                >
                  <CardHeader className="pb-3">
                    <CardTitle className="flex items-center gap-2 text-lg">
                      <span className="text-xl">{template.icon}</span>
                      {template.name}
                      {template.popular && (
                        <Badge variant="secondary" className="text-xs">
                          Popular
                        </Badge>
                      )}
                    </CardTitle>
                    <CardDescription className="text-sm">
                      {template.description}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="flex flex-wrap gap-1">
                      {template.tags.map((tag) => (
                        <Badge key={tag} variant="outline" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          {/* Project Configuration */}
          {selectedTemplate && (
            <div className="space-y-4 border-t pt-6">
              <Label className="text-base font-medium">Project Configuration</Label>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="project-name">Project Name *</Label>
                  <Input
                    id="project-name"
                    value={projectName}
                    onChange={e => {
                      setProjectName(e.target.value);
                      if (error) setError(null);
                    }}
                    placeholder="my-awesome-project"
                    disabled={isCreating}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Selected Template</Label>
                  <div className="flex items-center gap-2 p-2 border rounded-md bg-muted/50">
                    <span>{selectedTemplate.icon}</span>
                    <span className="font-medium">{selectedTemplate.name}</span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="project-description">Message</Label>
                <Textarea
                  id="project-description"
                  value={message}
                  onChange={e => {
                    setMessage(e.target.value);
                    if (error) setError(null);
                  }}
                  placeholder="Describe your project..."
                  rows={3}
                  disabled={isCreating}
                  required
                />
              </div>
            </div>
          )}

          {/* Error Display */}
          {error && (
            <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-6 border-t">
            <Button
              variant="outline"
              onClick={handleClose}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateProject}
              disabled={!selectedTemplate || !projectName.trim() || isCreating}
            >
              {isCreating ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating Project...
                </>
              ) : (
                <>
                  <Rocket className="w-4 h-4 mr-2" />
                  Create Project
                </>
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
});

NewProjectModal.displayName = 'NewProjectModal';
