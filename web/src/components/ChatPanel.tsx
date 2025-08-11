import { useState, useRef, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import { Send, User, Bot, Activity, AlertCircle, StopCircle, Settings } from 'lucide-react';
import { cn } from '@/lib/utils';
import { eventManager, EVENTS } from '@/services/eventManager';
import { soundService } from '@/services/soundService';
import { chatApi, modelsApi } from '@/services/api';
import ModelSelector from '@/components/ModelSelector';
import TokenUsageDisplay from '@/components/TokenUsageDisplay';
import { Textarea } from './ui/textarea';

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'assistant';
  timestamp: Date;
  type?: 'text' | 'code' | 'image' | 'status';
  agent_type?: string;
}

interface WebSocketMessage {
  type: string;
  content: string;
  session_id: string;
  timestamp: string;
  agent_type?: string;
  status?: string;
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8084/api/v1';
const WS_BASE = import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8084/api/v1/chat';

interface ChatPanelProps {
  chatId?: string;
  projectId?: string | null;
  initialMessage?: string;
}

export const ChatPanel = ({ chatId, projectId, initialMessage }: ChatPanelProps) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [currentProvider, setCurrentProvider] = useState<string>('');
  const [sessionId] = useState(() => `session_${Date.now()}`);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);

  // Project-related state
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null);

  // Handler for model changes
  const handleModelChange = (provider: string, model: string) => {
    setCurrentProvider(provider);
    addStatusMessage(`Switched to ${provider}/${model}`, 'info');
  };

  const wsRef = useRef<WebSocket | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);

  // Initialize WebSocket connection only when we have a project
  useEffect(() => {
    fetchProviderInfo();

    return () => {
      if (wsRef.current) {
        wsRef.current.close(1000, 'Component unmounting');
      }
    };
  }, []);

  // Load existing chat if chatId or projectId is provided
  useEffect(() => {
    if (projectId) {
      loadProjectChatHistory(projectId);
      setCurrentProjectId(projectId);
      // Auto-connect WebSocket for project
      connectWebSocket(projectId);
    }
  }, [projectId]);

  const loadProjectChatHistory = useCallback(async (projectId: string) => {
    setIsLoadingHistory(true);
    try {
      const data = await chatApi.getMessages(projectId);
      if (data.messages && Array.isArray(data.messages)) {
        const loadedMessages = data.messages.map((msg: any) => ({
          id: msg.id.toString(),
          content: msg.content,
          sender: (msg.role === 'user' ? 'user' : 'assistant') as 'user' | 'assistant',
          timestamp: new Date(msg.created_at),
          type: 'text' as const,
          agent_type: msg.role === 'assistant' ? 'react' : undefined
        }));

        setMessages(loadedMessages);
      }
    } catch (error) {
      console.error('Failed to load project chat history:', error);
    } finally {
      setIsLoadingHistory(false);
    }
  }, [setMessages]);
  console.log('Current Project ID:', messages);

  const fetchProviderInfo = async () => { 
    try {
      const data = await modelsApi.getProviderInfo();
      setCurrentProvider(data.provider || 'unknown');
    } catch (error) {
      console.error('Failed to fetch provider info:', error);
    }
  };

  const createNewProject = async (message: string) => {
    // Prevent multiple simultaneous project creation attempts
    if (isProcessing) {
      console.log('Project creation already in progress, skipping...');
      return;
    }

    try {
      setIsProcessing(true);
      addStatusMessage('Creating new project...', 'processing');

      const data = await chatApi.createSession(message);
      setCurrentProjectId(data.project_id);
      setCurrentSessionId(data.session_id);

      addStatusMessage(`Project created successfully!`, 'completed');

      // Emit project events
      eventManager.emit(EVENTS.PROJECT_BUILD_START, {
        step: 'Project Setup',
        progress: 0
      });

      // Redirect to project-based URL
      window.history.replaceState({}, '', `/projects/${data.project_id}`);

      // Connect WebSocket to the new project
      connectWebSocket(data.project_id);

      // Send the initial message automatically
      if (data.session_id) {
        setTimeout(() => {
          sendMessageToAgent(message, data.session_id);
        }, 1500);
      }
    } catch (error) {
      console.error('Failed to create project:', error);
      addStatusMessage('Failed to create project - check if backend server is running', 'error');
    } finally {
      setIsProcessing(false);
    }
  };

  const connectWebSocket = (projectId?: string) => {
    if (!projectId && !currentProjectId) {
      console.log('No project ID available for WebSocket connection');
      return;
    }

    const targetProjectId = projectId || currentProjectId;
    const wsUrl = `${WS_BASE}/stream/${targetProjectId}`;

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connected to project', targetProjectId);
        setIsConnected(true);
        // Don't add status message for cleaner UX
      };

      ws.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);
          handleWebSocketMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected', event.code, event.reason);
        setIsConnected(false);
        setIsTyping(false);

        // Only show disconnect message if it wasn't intentional
        if (event.code !== 1000) {
          addStatusMessage('Connection lost - trying to reconnect...', 'error');

          // Auto-reconnect after a delay
          setTimeout(() => {
            if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED) {
              connectWebSocket(targetProjectId);
            }
          }, 3000);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setIsConnected(false);
        
        // Check if this is a connection refused error (API server down)
        if (ws.readyState === WebSocket.CONNECTING) {
          addStatusMessage('Unable to connect to project. Please check if the backend is running.', 'error');
        } else {
          addStatusMessage('Connection error occurred', 'error');
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      setIsConnected(false);
    }
  };

  const handleWebSocketMessage = (data: WebSocketMessage) => {
    // Detect build-related activities
    detectBuildActivity(data);

    switch (data.type) {
      case 'connection':
        // Connection acknowledgment handled in onopen
        break;
      case 'status':
        setIsTyping(true);
        addStatusMessage(data.content, 'processing');
        // Play agent activity sound
        soundService.playAgentActivity();
        break;
      case 'completion':
        setIsTyping(false);
        setIsProcessing(false);
        setCurrentSessionId(null);
        addStatusMessage(data.content, 'completed');
        // Play completion sound
        soundService.playSuccess();
        break;
      case 'cancelled':
        setIsTyping(false);
        setIsProcessing(false);
        setCurrentSessionId(null);
        addStatusMessage(data.content, 'cancelled');
        break;
      case 'error':
        setIsTyping(false);
        setIsProcessing(false);
        setCurrentSessionId(null);
        addStatusMessage(data.content, 'error');
        // Play error sound
        soundService.playError();
        break;
      case 'agent_response':
        setIsTyping(false);
        addMessage(data.content, 'assistant', data.agent_type);
        // Play message received sound
        soundService.playMessageReceived();
        break;
      case 'agent_chunk':
        // Handle streaming chunks - append to last message or create new one
        setIsTyping(false);
        addMessage(data.content, 'assistant', data.agent_type);
        break;
      case 'response_complete':
        setIsTyping(false);
        setIsProcessing(false);
        addStatusMessage('Response completed', 'completed');
        // Play build complete sound for final completion
        soundService.playBuildComplete();
        break;
      case 'message_received':
        addStatusMessage('Message received, processing...', 'processing');
        soundService.playStepComplete();
        break;
      case 'debug':
        // Only show debug messages in development
        if (import.meta.env.DEV) {
          addStatusMessage(data.content, 'debug');
        }
        break;
      default:
        // Handle any other message types
        if (data.content) {
          addMessage(data.content, 'assistant', data.agent_type);
        }
    }
  };

  const detectBuildActivity = (data: WebSocketMessage) => {
    const content = data.content?.toLowerCase() || '';
    const agentType = data.agent_type;

    // Detect project building start
    if (content.includes('creating') || content.includes('setting up') ||
      content.includes('building') || content.includes('generating') ||
      (agentType === 'supervisor' && (content.includes('coordinate') || content.includes('breakdown')))) {
      eventManager.emit(EVENTS.PROJECT_BUILD_START, {
        step: 'Project Setup',
        progress: 0
      });
    }

    // Detect progress based on agent activity
    if (agentType === 'code_editing' || content.includes('file') || content.includes('component')) {
      eventManager.emit(EVENTS.PROJECT_BUILD_PROGRESS, {
        step: 'Component Creation',
        progress: 25,
        agent: agentType
      });
    }

    if (agentType === 'react' || content.includes('style') || content.includes('css')) {
      eventManager.emit(EVENTS.PROJECT_BUILD_PROGRESS, {
        step: 'Styling',
        progress: 50,
        agent: agentType
      });
    }

    if (content.includes('complete') || content.includes('finished') ||
      content.includes('done') || content.includes('ready')) {
      eventManager.emit(EVENTS.PROJECT_BUILD_COMPLETE, {
        step: 'Build & Preview',
        progress: 100
      });
    }

    // General agent activity
    eventManager.emit(EVENTS.AGENT_ACTIVITY, {
      agent: agentType,
      content: data.content
    });
  };

  const addMessage = (content: string, sender: 'user' | 'assistant', agentType?: string) => {
    const newMessage: Message = {
      id: `${Date.now()}_${Math.random().toString(36).substring(2, 11)}`,
      content,
      sender,
      timestamp: new Date(),
      type: 'text',
      agent_type: agentType
    };
    setMessages(prev => [...prev, newMessage]);
  };

  const addStatusMessage = (content: string, status: string) => {
    const statusMessage: Message = {
      id: `status_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`,
      content,
      sender: 'assistant',
      timestamp: new Date(),
      type: 'status',
      agent_type: status
    };
    setMessages(prev => [...prev, statusMessage]);
  };

  const sendMessageToAgent = (message: string, sessionId?: string) => {
    if (!message.trim() || isProcessing) return;

    // If we don't have a WebSocket connection, try to connect
    if (!isConnected) {
      connectWebSocket();
      addStatusMessage('Please try sending your message again once connected', 'info');
      return;
    }

    // Send to WebSocket
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const messagePayload = {
        message: message,
        session_id: sessionId || currentSessionId || sessionId,
        timestamp: new Date().toISOString()
      };

      wsRef.current.send(JSON.stringify(messagePayload));
      setIsTyping(true);
      setIsProcessing(true);

      // Emit build start event to switch to preview mode
      eventManager.emit(EVENTS.PROJECT_BUILD_START);
    }
  };

  const handleSend = () => {
    if (!input.trim() || isProcessing) return;

    // If we don't have a project yet, create one first
    if (!currentProjectId) {
      createNewProject(input);
      return;
    }

    // Add user message
    addMessage(input, 'user');

    // Send the message
    sendMessageToAgent(input);

    setInput('');
  };

  const handleCancel = async () => {
    if (!currentSessionId) return;

    try {
      // Call the cancel endpoint
      await chatApi.cancelSession(currentSessionId);
      addStatusMessage('Request cancelled', 'cancelled');
      setIsProcessing(false);
      setIsTyping(false);
      setCurrentSessionId(null);
    } catch (error) {
      console.error('Failed to cancel request:', error);
      addStatusMessage('Failed to cancel request', 'error');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  useEffect(() => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTo({ top: scrollAreaRef.current.scrollHeight, behavior: 'smooth' });
    }
  }, [messages]);

  return (
    <div className="h-full flex flex-col bg-background min-h-0">
      {/* Chat Header - Fixed */}
      <div className="flex-shrink-0 p-4 border-b border-border/50 bg-card/50 backdrop-blur-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-primary" />
            <h2 className="font-semibold text-lg">Chat with AI</h2>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant={isConnected ? "default" : "destructive"} className="text-xs">
              <Activity className="w-3 h-3 mr-1" />
              {isConnected ? 'Connected' : 'Disconnected'}
            </Badge>
            {currentProvider && (
              <Badge variant="outline" className="text-xs">
                {currentProvider}
              </Badge>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSettingsOpen(!settingsOpen)}
              className="h-8 w-8 p-0"
            >
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Collapsible Settings Panel */}
        <Collapsible open={settingsOpen} onOpenChange={setSettingsOpen}>
          <CollapsibleContent className="mt-4 space-y-4">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {/* Model Selector */}
              <div>
                <h3 className="text-sm font-medium mb-2">Model Configuration</h3>
                <ModelSelector onModelChange={handleModelChange} />
              </div>

              {/* Token Usage */}
              <div>
                <h3 className="text-sm font-medium mb-2">Token Usage</h3>
                <TokenUsageDisplay
                  sessionId={sessionId}
                  showGlobalStats={true}
                  className="max-h-48"
                />
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>

      {/* Messages - Scrollable */}
      <div className="flex-1 overflow-hidden flex flex-col min-h-0">
        <ScrollArea className="flex-1 h-full" ref={scrollAreaRef}>
          <div className="p-4 space-y-4">
            {isLoadingHistory ? (
              <div className="flex items-center justify-center py-8">
                <div className="text-center space-y-2">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div>
                  <p className="text-sm text-muted-foreground">Loading chat history...</p>
                </div>
              </div>
            ) : messages.length === 0 ? (
              <div className="flex items-center justify-center py-8">
                <p className="text-muted-foreground text-center">
                  Start a conversation by typing a message below
                </p>
              </div>
            ) : null}

            {messages.map((message) => (
              <div
                key={message.id}
                className={cn(
                  "flex gap-3 max-w-full",
                  message.sender === 'user' ? "justify-end" : "justify-start"
                )}
              >
                {message.sender === 'assistant' && (
                  <Avatar className="h-8 w-8 shrink-0 ring-2 ring-primary/10">
                    <AvatarFallback className="bg-primary text-primary-foreground text-xs">
                      {message.type === 'status' ? (
                        message.agent_type === 'error' ? <AlertCircle className="w-4 h-4" /> :
                          message.agent_type === 'processing' ? <Activity className="w-4 h-4" /> :
                            <Bot className="w-4 h-4" />
                      ) : (
                        message.agent_type?.charAt(0)?.toUpperCase() || 'AI'
                      )}
                    </AvatarFallback>
                  </Avatar>
                )}

                <div
                  className={cn(
                    "rounded-lg px-4 py-2 max-w-[80%] break-words",
                    message.sender === 'user'
                      ? "bg-primary text-primary-foreground ml-auto"
                      : message.type === 'status'
                        ? "bg-muted border border-border"
                        : "bg-muted"
                  )}
                >
                  {message.agent_type && message.sender === 'assistant' && message.type !== 'status' && (
                    <div className="text-xs text-muted-foreground mb-1 font-medium">
                      {message.agent_type.charAt(0).toUpperCase() + message.agent_type.slice(1)} Agent
                    </div>
                  )}

                  <div className={cn(
                    "text-sm leading-relaxed",
                    message.type === 'status' && "text-muted-foreground text-xs"
                  )}>
                    {message.content}
                  </div>

                  <div className={cn(
                    "text-xs mt-1 opacity-70",
                    message.sender === 'user' ? "text-primary-foreground/70" : "text-muted-foreground"
                  )}>
                    {message.timestamp.toLocaleTimeString()}
                  </div>
                </div>

                {message.sender === 'user' && (
                  <Avatar className="h-8 w-8 shrink-0 ring-2 ring-muted/30">
                    <AvatarFallback className="bg-muted text-xs">
                      <User className="w-4 h-4" />
                    </AvatarFallback>
                  </Avatar>
                )}
              </div>
            ))}

            {/* Typing indicator */}
            {isTyping && (
              <div className="flex gap-3 justify-start">
                <Avatar className="h-8 w-8 shrink-0 ring-2 ring-primary/10">
                  <AvatarFallback className="bg-primary text-primary-foreground text-xs">
                    <Activity className="w-4 h-4 animate-pulse" />
                  </AvatarFallback>
                </Avatar>
                <div className="bg-muted rounded-lg px-4 py-2 max-w-[80%]">
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                    <div className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                    <div className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"></div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Input - Fixed */}
      <div className="flex-shrink-0 p-4 border-t border-border/50 bg-card/50 backdrop-blur-sm">
        <div className="flex gap-2 items-center">
          <Textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder={isConnected ? "Describe what you want to build..." : "Connecting..."}
            disabled={!isConnected || isProcessing}
            className="flex-1"
            rows={4}
          />
          {isProcessing ? (
            <Button
              onClick={handleCancel}
              size="icon"
              variant="destructive"
              className="shrink-0"
            >
              <StopCircle className="w-4 h-4" />
            </Button>
          ) : (
            <Button
              onClick={handleSend}
              disabled={!input.trim() || !isConnected}
              size="icon"
              className="shrink-0"
            >
              <Send className="w-4 h-4" />
            </Button>
          )}
        </div>

        {!isConnected && (
          <div className="mt-2 text-xs text-muted-foreground flex items-center gap-1">
            <AlertCircle className="w-3 h-3" />
            Check if the server is running on port 8084
          </div>
        )}
      </div>
    </div>
  );
};