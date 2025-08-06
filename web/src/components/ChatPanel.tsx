import { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Send, User, Bot, Sparkles, Code2, Palette, Activity, AlertCircle, StopCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { eventManager, EVENTS } from '@/services/eventManager';

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
const WS_URL = import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8084/api/v1/chat/stream';

interface ChatPanelProps {
  chatId?: string;
  initialMessage?: string;
}

export const ChatPanel = ({ chatId, initialMessage }: ChatPanelProps) => {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      content: 'Hello! I\'m your AI assistant powered by multi-agent system. I can help you build React applications with authentication, file upload, and deployment. What would you like to create today?',
      sender: 'assistant',
      timestamp: new Date(),
      type: 'text'
    }
  ]);
  const [input, setInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [currentProvider, setCurrentProvider] = useState<string>('');
  const [sessionId] = useState(() => `session_${Date.now()}`);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  
  const wsRef = useRef<WebSocket | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Initialize WebSocket connection
  useEffect(() => {
    connectWebSocket();
    fetchProviderInfo();
    
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Handle initial message from homepage
  useEffect(() => {
    if (initialMessage && initialMessage.trim()) {
      // Set the input and send it after WebSocket connects
      setInput(initialMessage);
      const timer = setTimeout(() => {
        if (isConnected) {
          handleSend();
        }
      }, 1000);
      
      return () => clearTimeout(timer);
    }
  }, [initialMessage, isConnected]);

  // Load existing chat if chatId is provided
  useEffect(() => {
    if (chatId) {
      loadChatHistory(chatId);
    }
  }, [chatId]);

  const loadChatHistory = async (chatId: string) => {
    try {
      const response = await fetch(`${API_BASE}/chat/${chatId}`);
      if (response.ok) {
        const data = await response.json();
        if (data.messages && Array.isArray(data.messages)) {
          const loadedMessages = data.messages.map((msg: any) => ({
            id: msg.id,
            content: msg.content,
            sender: msg.type === 'user' ? 'user' : 'assistant',
            timestamp: new Date(msg.timestamp),
            type: msg.type || 'text',
            agent_type: msg.agent_type
          }));
          setMessages(loadedMessages);
        }
      }
    } catch (error) {
      console.error('Failed to load chat history:', error);
    }
  };

  const fetchProviderInfo = async () => {
    try {
      const response = await fetch(`${API_BASE}/models`);
      const data = await response.json();
      setCurrentProvider(data.provider || 'unknown');
    } catch (error) {
      console.error('Failed to fetch provider info:', error);
    }
  };

  const connectWebSocket = () => {
    try {
      const ws = new WebSocket(WS_URL);
      
      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        addStatusMessage('Connected to AI multi-agent system', 'connection');
      };

      ws.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);
          handleWebSocketMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        setIsTyping(false);
        addStatusMessage('Disconnected from server', 'error');
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setIsConnected(false);
        addStatusMessage('Connection error', 'error');
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
        break;
      case 'completion':
        setIsTyping(false);
        setIsProcessing(false);
        setCurrentSessionId(null);
        addStatusMessage(data.content, 'completed');
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
        break;
      case 'agent_response':
        setIsTyping(false);
        addMessage(data.content, 'assistant', data.agent_type);
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
      id: Date.now().toString(),
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
      id: `status_${Date.now()}`,
      content,
      sender: 'assistant',
      timestamp: new Date(),
      type: 'status',
      agent_type: status
    };
    setMessages(prev => [...prev, statusMessage]);
  };

  const handleSend = () => {
    if (!input.trim() || !isConnected || isProcessing) return;

    // Add user message
    addMessage(input, 'user');
    
    // Send to WebSocket
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const message = {
        message: input,
        session_id: sessionId,
        timestamp: new Date().toISOString()
      };
      
      wsRef.current.send(JSON.stringify(message));
      setIsTyping(true);
      setIsProcessing(true);
      setCurrentSessionId(sessionId);
      
      // Emit build start event to switch to preview mode
      eventManager.emit(EVENTS.PROJECT_BUILD_START);
    }

    setInput('');
  };

  const handleCancel = async () => {
    if (!currentSessionId) return;

    try {
      // Call the cancel endpoint
      const response = await fetch(`${API_BASE}/chat/${currentSessionId}/cancel`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        addStatusMessage('Request cancelled', 'cancelled');
        setIsProcessing(false);
        setIsTyping(false);
        setCurrentSessionId(null);
      }
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

  const quickActions = [
    { label: 'React todo app with auth', icon: Code2 },
    { label: 'Dashboard with file upload', icon: Palette },
    { label: 'E-commerce landing page', icon: Sparkles },
  ];

  const handleQuickAction = (label: string) => {
    setInput(label);
    // Auto-send after a brief delay
    setTimeout(() => handleSend(), 100);
  };

  return (
    <div className="h-full flex flex-col bg-background">
      {/* Chat Header - Fixed */}
      <div className="flex-shrink-0 p-4 border-b border-border/50 bg-card/50 backdrop-blur-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-primary" />
            <h2 className="font-semibold text-lg">AI Multi-Agent System</h2>
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
          </div>
        </div>
        <p className="text-sm text-muted-foreground mt-1">
          Supervisor • Code Editor • React Specialist • DevOps
        </p>
      </div>

      {/* Quick Actions */}
      <div className="flex-shrink-0 p-4 space-y-2">
        <p className="text-xs text-muted-foreground font-medium">Quick Actions</p>
        <div className="flex flex-wrap gap-2">
          {quickActions.map((action, index) => {
            const IconComponent = action.icon;
            return (
              <Button
                key={index}
                variant="outline"
                size="sm"
                className="text-xs h-8"
                onClick={() => handleQuickAction(action.label)}
              >
                <IconComponent className="w-3 h-3 mr-1" />
                {action.label}
              </Button>
            );
          })}
        </div>
      </div>

      {/* Messages - Scrollable */}
      <div className="flex-1 overflow-hidden flex flex-col">
        <ScrollArea className="flex-1" ref={scrollAreaRef}>
          <div className="p-4 space-y-4">
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
        <div className="flex gap-2">
          <Input
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder={isConnected ? "Describe what you want to build..." : "Connecting..."}
            disabled={!isConnected || isProcessing}
            className="flex-1"
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
            Check if the Go server is running on port 8084
          </div>
        )}
      </div>
    </div>
  );
};