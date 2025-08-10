import { useRef, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Code2, Copy, Download, MoreHorizontal } from 'lucide-react';
import { useTheme } from 'next-themes';

interface CodeEditorProps {
  activeFile: string;
  content: string;
  onContentChange: (content: string) => void;
}

export const CodeEditor = ({ activeFile, content, onContentChange }: CodeEditorProps) => {
  const editorRef = useRef<any>(null);
  const { theme } = useTheme();

  const getLanguage = (filename: string): string => {
    const ext = filename?.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'tsx':
      case 'ts':
        return 'typescript';
      case 'jsx':
      case 'js':
        return 'javascript';
      case 'css':
        return 'css';
      case 'html':
        return 'html';
      case 'json':
        return 'json';
      case 'md':
        return 'markdown';
      default:
        return 'typescript';
    }
  };

  const getFileType = (filename: string): string => {
    const ext = filename.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'tsx':
        return 'TypeScript React';
      case 'ts':
        return 'TypeScript';
      case 'jsx':
        return 'JavaScript React';
      case 'js':
        return 'JavaScript';
      case 'css':
        return 'CSS';
      case 'html':
        return 'HTML';
      case 'json':
        return 'JSON';
      case 'md':
        return 'Markdown';
      default:
        return 'Text';
    }
  };

  const handleEditorDidMount = (editor: any, monaco: any) => {
    editorRef.current = editor;
    
    // Configure editor
    editor.updateOptions({
      fontSize: 14,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Monaco', 'Consolas', monospace",
      lineHeight: 22,
      padding: { top: 16, bottom: 16 },
      scrollBeyondLastLine: false,
      minimap: { enabled: false },
      wordWrap: 'on',
      renderLineHighlight: 'gutter',
      bracketPairColorization: { enabled: true },
    });

    // Configure Monaco themes
    monaco.editor.defineTheme('custom-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6B7280' },
        { token: 'keyword', foreground: 'A855F7' },
        { token: 'string', foreground: '10B981' },
        { token: 'number', foreground: 'F59E0B' },
        { token: 'type', foreground: '3B82F6' },
      ],
      colors: {
        'editor.background': '#0F0F0F',
        'editor.lineHighlightBackground': '#1F1F1F',
        'editorLineNumber.foreground': '#6B7280',
        'editorLineNumber.activeForeground': '#A855F7',
      }
    });

    monaco.editor.defineTheme('custom-light', {
      base: 'vs',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6B7280' },
        { token: 'keyword', foreground: '7C3AED' },
        { token: 'string', foreground: '059669' },
        { token: 'number', foreground: 'D97706' },
        { token: 'type', foreground: '2563EB' },
      ],
      colors: {
        'editor.background': '#FFFFFF',
        'editor.lineHighlightBackground': '#F8FAFC',
        'editorLineNumber.foreground': '#9CA3AF',
        'editorLineNumber.activeForeground': '#7C3AED',
      }
    });
  };

  const copyContent = () => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(content);
    }
  };

  return (
    <div className="flex-1 flex flex-col h-full">
      {/* Editor Header */}
      <div className="p-3 border-b border-border/50 bg-card/30 backdrop-blur-sm flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h3 className="font-medium text-sm truncate">{activeFile}</h3>
            <Badge variant="outline" className="text-xs">
              {getFileType(activeFile)}
            </Badge>
          </div>
          
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={copyContent} className="h-7 w-7 p-0">
              <Copy className="w-3 h-3" />
            </Button>
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0">
              <Download className="w-3 h-3" />
            </Button>
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0">
              <MoreHorizontal className="w-3 h-3" />
            </Button>
          </div>
        </div>
      </div>
      
      {/* Monaco Editor */}
      <div className="flex-1 overflow-hidden">
        <Editor
          height="100%"
          language={getLanguage(activeFile)}
          value={content}
          onChange={(value) => onContentChange(value || '')}
          onMount={handleEditorDidMount}
          theme={theme === 'dark' ? 'custom-dark' : 'custom-light'}
          options={{
            automaticLayout: true,
            tabSize: 2,
            insertSpaces: true,
            detectIndentation: false,
            renderWhitespace: 'selection',
            smoothScrolling: true,
            cursorBlinking: 'smooth',
            cursorSmoothCaretAnimation: 'on',
            contextmenu: true,
            mouseWheelZoom: true,
            quickSuggestions: true,
            suggestOnTriggerCharacters: true,
            acceptSuggestionOnEnter: 'on',
            accessibilitySupport: 'auto',
          }}
        />
      </div>
    </div>
  );
};