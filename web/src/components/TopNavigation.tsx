import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Github, Database, Zap, Cloud, Settings, Sparkles } from 'lucide-react';

export const TopNavigation = () => {
  const integrations = [
    { name: 'Supabase', icon: Database, status: 'connected', color: 'bg-success' },
    { name: 'GitHub', icon: Github, status: 'disconnected', color: 'bg-muted' },
    { name: 'Vercel', icon: Zap, status: 'connected', color: 'bg-primary' },
    { name: 'Cloud', icon: Cloud, status: 'pending', color: 'bg-warning' },
  ];

  return (
    <header className="h-16 bg-card/80 backdrop-blur-sm border-b border-border/50 flex items-center justify-between px-6 animate-slide-up">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-gradient-primary rounded-lg flex items-center justify-center animate-glow-pulse">
            <Sparkles className="w-5 h-5 text-white" />
          </div>
          <h1 className="text-xl font-bold bg-gradient-primary bg-clip-text text-transparent">
            AI Builder
          </h1>
        </div>
        
        <div className="flex items-center gap-2">
          {integrations.map((integration) => (
            <Tooltip key={integration.name}>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 px-2 interactive-hover">
                  <integration.icon className="w-4 h-4 mr-1" />
                  <span className="text-xs">{integration.name}</span>
                  <Badge 
                    className={`ml-1 h-2 w-2 rounded-full p-0 ${integration.color}`}
                    variant="secondary"
                  />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{integration.name} - {integration.status}</p>
              </TooltipContent>
            </Tooltip>
          ))}
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" className="interactive-hover">
          <Settings className="w-4 h-4" />
        </Button>
        <Button className="btn-gradient">
          <Sparkles className="w-4 h-4 mr-2" />
          Deploy
        </Button>
      </div>
    </header>
  );
};