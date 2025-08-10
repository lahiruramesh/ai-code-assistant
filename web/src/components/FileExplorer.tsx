import { useState, useEffect } from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { 
  Folder, 
  FolderOpen, 
  FileText, 
  ChevronRight, 
  ChevronDown,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileNode {
  name: string;
  type: 'file' | 'folder';
  path: string;
  size?: string;
  children?: FileNode[];
  expanded?: boolean;
}

interface FileExplorerProps {
  activeFile: string;
  onFileSelect: (path: string) => void;
  projectName?: string;
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8084/api/v1';

export const FileExplorer = ({ activeFile, onFileSelect, projectName }: FileExplorerProps) => {
  const [fileTree, setFileTree] = useState<FileNode[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load project files
  useEffect(() => {
    if (!projectName) return;
    
    const fetchFiles = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        // Updated to use the new aiagent path
        const response = await fetch(`${API_BASE}/projects/${projectName}/files?source=aiagent`);
        if (!response.ok) {
          throw new Error(`Failed to fetch files: ${response.statusText}`);
        }
        
        const data = await response.json();
        setFileTree(data.files || []);
      } catch (err) {
        console.error('Failed to fetch files:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch files');
        // Fallback to mock data if API fails
        setFileTree(getMockFileTree());
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchFiles();
  }, [projectName]);

  // Mock data fallback
  const getMockFileTree = (): FileNode[] => [
    {
      name: 'src',
      type: 'folder',
      path: 'src',
      expanded: true,
      children: [
        {
          name: 'components',
          type: 'folder',
          path: 'src/components',
          expanded: true,
          children: [
            { name: 'Header.tsx', type: 'file', path: 'src/components/Header.tsx', size: '1.8 KB' },
            { name: 'Hero.tsx', type: 'file', path: 'src/components/Hero.tsx', size: '3.2 KB' },
            { name: 'Button.tsx', type: 'file', path: 'src/components/Button.tsx', size: '2.1 KB' },
          ]
        },
        { name: 'App.tsx', type: 'file', path: 'src/App.tsx', size: '2.4 KB' },
      ]
    },
    { name: 'package.json', type: 'file', path: 'package.json', size: '1.8 KB' },
  ];

  const toggleFolder = (path: string) => {
    const updateNode = (nodes: FileNode[]): FileNode[] => {
      return nodes.map(node => {
        if (node.path === path && node.type === 'folder') {
          return { ...node, expanded: !node.expanded };
        }
        if (node.children) {
          return { ...node, children: updateNode(node.children) };
        }
        return node;
      });
    };
    setFileTree(updateNode(fileTree));
  };

  const renderFileNode = (node: FileNode, depth = 0) => {
    const isActive = activeFile === node.path;
    const paddingLeft = `${depth * 1.5 + 0.5}rem`;

    return (
      <div key={node.path}>
        <button
          onClick={() => {
            if (node.type === 'folder') {
              toggleFolder(node.path);
            } else {
              onFileSelect(node.path);
            }
          }}
          className={cn(
            'w-full text-left p-1.5 rounded-md text-sm transition-colors interactive-hover flex items-center gap-2',
            isActive && node.type === 'file' 
              ? 'bg-primary/10 text-primary border border-primary/20' 
              : 'hover:bg-muted/50'
          )}
          style={{ paddingLeft }}
        >
          {node.type === 'folder' ? (
            <>
              {node.expanded ? (
                <ChevronDown className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              ) : (
                <ChevronRight className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              )}
              {node.expanded ? (
                <FolderOpen className="w-4 h-4 text-primary flex-shrink-0" />
              ) : (
                <Folder className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              )}
            </>
          ) : (
            <>
              <div className="w-4 h-4 flex-shrink-0" /> {/* Spacer for alignment */}
              <FileText className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            </>
          )}
          
          <span className="flex-1 truncate">{node.name}</span>
          
          {node.type === 'file' && node.size && (
            <Badge variant="secondary" className="text-xs flex-shrink-0">
              {node.size}
            </Badge>
          )}
        </button>

        {node.type === 'folder' && node.expanded && node.children && (
          <div>
            {node.children.map(child => renderFileNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="w-64 border-r border-border/50 bg-card/30 backdrop-blur-sm flex flex-col h-full">
      {/* File Explorer Header */}
      <div className="p-3 border-b border-border/50 flex-shrink-0">
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-medium text-sm flex items-center gap-2">
            <Folder className="w-4 h-4 text-primary" />
            Explorer
          </h3>
        </div>
      </div>

      {/* File Tree */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-0.5">
          {fileTree.map(node => renderFileNode(node))}
        </div>
      </ScrollArea>
    </div>
  );
};