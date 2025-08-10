import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import {
  Plus,
  Search,
  Settings,
  FolderOpen,
  Sparkles,
  X,
  Home
} from 'lucide-react';

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8084/api/v1';

interface Project {
  id: string;
  name: string;
  template: string;
  status: string;
  created_at: string;
  url?: string;
}

interface ProjectSidebarProps {
  currentProjectId?: string | null;
  onClose?: () => void;
}

export const ProjectSidebar = ({ currentProjectId, onClose }: ProjectSidebarProps) => {
  const navigate = useNavigate();
  const [projects, setProjects] = useState<Project[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [dataLoaded, setDataLoaded] = useState(false);

  useEffect(() => {
    if (!dataLoaded) {
      loadData();
    }
  }, [dataLoaded]);

  const loadData = async () => {
    try {
      setLoading(true);
      
      // Load projects with trailing slash to avoid redirects
      const projectResponse = await fetch(`${API_BASE}/projects`);
      if (projectResponse.ok) {
        const projectData = await projectResponse.json();
        setProjects(projectData.projects || []);
      } else {
        console.warn('Failed to load projects:', projectResponse.status, projectResponse.statusText);
      }
      
      setDataLoaded(true);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleProjectSelect = (project: Project) => {
    navigate(`/projects/${project.id}`);
    onClose?.();
  };


  const filteredProjects = projects.filter(project =>
    project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    project.template.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="w-80 h-full border-r bg-background flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center">
              <Sparkles className="h-4 w-4 text-white" />
            </div>
            <span className="font-bold text-lg">AI App Builder</span>
          </div>
          {onClose && (
            <Button variant="ghost" size="sm" onClick={onClose} className="lg:hidden">
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
        
        <div className="space-y-2 mt-4">
          
          <Button 
            onClick={() => navigate('/')}
            className="w-full justify-start"
            variant="ghost"
          >
            <Home className="h-4 w-4 mr-2" />
            Home
          </Button>
        </div>
        
        <div className="relative mt-4">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search projects..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-4">


          {/* Projects */}
          <div>
            <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2">
              Projects
            </h3>
            {loading ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : filteredProjects.length === 0 ? (
              <div className="text-sm text-muted-foreground">No projects yet</div>
            ) : (
              <div className="space-y-1">
                {filteredProjects.map((project) => (
                  <Button
                    key={project.id}
                    variant={project.id === currentProjectId ? "secondary" : "ghost"}
                    className="w-full justify-start h-auto p-3"
                    onClick={() => handleProjectSelect(project)}
                  >
                    <FolderOpen className="h-4 w-4 mr-3 flex-shrink-0" />
                    <div className="flex-1 min-w-0 text-left">
                      <div className="truncate text-sm font-medium">
                        {project.name}
                      </div>
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>{project.template}</span>
                        <Badge 
                          variant={project.status === 'active' ? 'default' : 'secondary'}
                          className="text-xs"
                        >
                          {project.status}
                        </Badge>
                      </div>
                    </div>
                  </Button>
                ))}
              </div>
            )}
          </div>
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="p-4 border-t">
        <div className="flex items-center justify-between">
          <div className="text-xs text-muted-foreground">
            {projects.length} projects
          </div>
          <Button variant="ghost" size="sm">
            <Settings className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};
