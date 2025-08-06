import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import {
  MessageSquare,
  Plus,
  Search,
  Settings,
  Calendar,
  FolderOpen,
  Sparkles,
  X,
  Home
} from 'lucide-react';

interface ChatSession {
  id: string;
  project_id?: string;
  created_at: string;
  last_activity: string;
  message_count: number;
  title?: string;
}

interface Project {
  id: number;
  name: string;
  template: string;
  status: string;
  created_at: string;
}

interface ProjectSidebarProps {
  currentChatId: string;
  onClose?: () => void;
}

export const ProjectSidebar = ({ currentChatId, onClose }: ProjectSidebarProps) => {
  const navigate = useNavigate();
  const [chatSessions, setChatSessions] = useState<ChatSession[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      
      // Load chat sessions
      const chatResponse = await fetch('/api/v1/chat/sessions');
      if (chatResponse.ok) {
        const chatData = await chatResponse.json();
        setChatSessions(chatData.sessions || []);
      }

      // Load projects
      const projectResponse = await fetch('/api/v1/projects');
      if (projectResponse.ok) {
        const projectData = await projectResponse.json();
        setProjects(projectData.projects || []);
      }
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNewChat = () => {
    navigate('/');
    onClose?.();
  };

  const handleChatSelect = (chatId: string) => {
    navigate(`/chat/${chatId}`);
    onClose?.();
  };

  const handleProjectSelect = (project: Project) => {
    // Find a chat session for this project or create new one
    const projectChat = chatSessions.find(chat => chat.project_id === String(project.id));
    if (projectChat) {
      navigate(`/chat/${projectChat.id}`);
    } else {
      // Create new chat for this project
      const newChatId = crypto.randomUUID();
      navigate(`/chat/${newChatId}`, { 
        state: { 
          projectId: project.id,
          projectName: project.name 
        } 
      });
    }
    onClose?.();
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffHours < 1) return 'Just now';
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  const filteredSessions = chatSessions.filter(session =>
    session.id.includes(searchQuery) ||
    (session.title && session.title.toLowerCase().includes(searchQuery.toLowerCase()))
  );

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
            <span className="font-bold text-lg">React Builder</span>
          </div>
          {onClose && (
            <Button variant="ghost" size="sm" onClick={onClose} className="lg:hidden">
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
        
        <div className="space-y-2 mt-4">
          <Button 
            onClick={handleNewChat}
            className="w-full justify-start"
            variant="default"
          >
            <Plus className="h-4 w-4 mr-2" />
            New Chat
          </Button>
          
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
            placeholder="Search chats and projects..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-4">
          {/* Recent Chats */}
          <div>
            <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2">
              Recent Chats
            </h3>
            {loading ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : filteredSessions.length === 0 ? (
              <div className="text-sm text-muted-foreground">No chats yet</div>
            ) : (
              <div className="space-y-1">
                {filteredSessions.slice(0, 10).map((session) => (
                  <Button
                    key={session.id}
                    variant={session.id === currentChatId ? "secondary" : "ghost"}
                    className="w-full justify-start h-auto p-3"
                    onClick={() => handleChatSelect(session.id)}
                  >
                    <MessageSquare className="h-4 w-4 mr-3 flex-shrink-0" />
                    <div className="flex-1 min-w-0 text-left">
                      <div className="truncate text-sm font-medium">
                        {session.title || `Chat ${session.id.slice(0, 8)}...`}
                      </div>
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>{formatDate(session.last_activity)}</span>
                        <Badge variant="secondary" className="text-xs">
                          {session.message_count}
                        </Badge>
                      </div>
                    </div>
                  </Button>
                ))}
              </div>
            )}
          </div>

          <Separator />

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
                    variant="ghost"
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
            {chatSessions.length} chats, {projects.length} projects
          </div>
          <Button variant="ghost" size="sm">
            <Settings className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};
