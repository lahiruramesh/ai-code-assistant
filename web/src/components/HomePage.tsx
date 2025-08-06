import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  Sparkles, 
  Zap, 
  Rocket, 
  Code, 
  Palette, 
  Database,
  ArrowRight,
  Github,
  Heart
} from 'lucide-react';

interface Suggestion {
  id: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  category: string;
  prompt: string;
}

const suggestions: Suggestion[] = [
  {
    id: '1',
    title: 'Todo App with React',
    description: 'Build a modern todo app with React, TypeScript, and local storage',
    icon: <Code className="h-5 w-5" />,
    category: 'React',
    prompt: 'Create a todo app with React and TypeScript that has add, edit, delete, and mark complete functionality. Use modern React hooks and local storage for persistence.'
  },
  {
    id: '2',
    title: 'Landing Page Template',
    description: 'Create a responsive landing page with hero section and features',
    icon: <Palette className="h-5 w-5" />,
    category: 'Design',
    prompt: 'Build a modern landing page with a hero section, features grid, testimonials, and contact form. Make it responsive and use beautiful gradients.'
  },
  {
    id: '3',
    title: 'Dashboard with Charts',
    description: 'Interactive dashboard with data visualization and charts',
    icon: <Database className="h-5 w-5" />,
    category: 'Dashboard',
    prompt: 'Create an admin dashboard with charts, data tables, and real-time stats. Include sidebar navigation and responsive design.'
  },
  {
    id: '4',
    title: 'E-commerce Store',
    description: 'Full-featured online store with cart and checkout',
    icon: <Rocket className="h-5 w-5" />,
    category: 'E-commerce',
    prompt: 'Build an e-commerce store with product catalog, shopping cart, user authentication, and checkout process. Use modern UI components.'
  },
  {
    id: '5',
    title: 'Blog Platform',
    description: 'Content management system with rich text editor',
    icon: <Zap className="h-5 w-5" />,
    category: 'CMS',
    prompt: 'Create a blog platform with post creation, editing, categories, tags, and commenting system. Include admin panel and SEO features.'
  },
  {
    id: '6',
    title: 'Portfolio Website',
    description: 'Professional portfolio with projects showcase',
    icon: <Sparkles className="h-5 w-5" />,
    category: 'Portfolio',
    prompt: 'Build a professional portfolio website with projects showcase, about section, contact form, and smooth animations.'
  }
];

export const HomePage = () => {
  const [input, setInput] = useState('');
  const navigate = useNavigate();

  const handleSubmit = (prompt?: string) => {
    const message = prompt || input;
    if (message.trim()) {
      // Generate a new chat/project ID and navigate to chat page
      const chatId = crypto.randomUUID();
      navigate(`/chat/${chatId}`, { 
        state: { 
          initialMessage: message,
          isNewChat: true 
        } 
      });
    }
  };

  const handleSuggestionClick = (suggestion: Suggestion) => {
    handleSubmit(suggestion.prompt);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-background via-background to-muted/30">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-14 items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-2">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center">
                <Sparkles className="h-4 w-4 text-white" />
              </div>
              <span className="font-bold text-xl">React Builder</span>
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <Button variant="ghost" size="sm">
              <Github className="h-4 w-4 mr-2" />
              GitHub
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container max-w-4xl mx-auto px-4 py-12">
        {/* Hero Section */}
        <div className="text-center space-y-6 mb-12">
          <div className="space-y-4">
            <div className="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium bg-muted">
              <Sparkles className="h-4 w-4 mr-2 text-purple-500" />
              AI-Powered Development
            </div>
            <h1 className="text-4xl font-bold tracking-tight sm:text-6xl bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
              Build React Apps with AI
            </h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Describe your idea and watch as AI generates a complete React application with modern UI components, routing, and functionality.
            </p>
          </div>
        </div>

        {/* Chat Input */}
        <div className="max-w-2xl mx-auto mb-12">
          <Card className="border-2 border-dashed border-muted-foreground/25 bg-background/50 backdrop-blur">
            <CardContent className="p-6">
              <div className="space-y-4">
                <div className="flex space-x-2">
                  <Input
                    placeholder="Describe the React app you want to build..."
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSubmit()}
                    className="flex-1 text-base"
                  />
                  <Button 
                    onClick={() => handleSubmit()}
                    disabled={!input.trim()}
                    size="lg"
                    className="px-6"
                  >
                    <ArrowRight className="h-4 w-4 mr-2" />
                    Build
                  </Button>
                </div>
                <p className="text-sm text-muted-foreground text-center">
                  Or choose from one of the suggestions below
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Suggestions Grid */}
        <div className="space-y-6">
          <div className="text-center">
            <h2 className="text-2xl font-semibold mb-2">Popular Templates</h2>
            <p className="text-muted-foreground">
              Get started quickly with these pre-made project ideas
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {suggestions.map((suggestion) => (
              <Card 
                key={suggestion.id}
                className="cursor-pointer transition-all duration-200 hover:shadow-lg hover:scale-105 group border-muted-foreground/20 hover:border-primary/50"
                onClick={() => handleSuggestionClick(suggestion)}
              >
                <CardContent className="p-6">
                  <div className="space-y-4">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center space-x-3">
                        <div className="p-2 rounded-lg bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-colors">
                          {suggestion.icon}
                        </div>
                        <div>
                          <Badge variant="secondary" className="text-xs">
                            {suggestion.category}
                          </Badge>
                        </div>
                      </div>
                      <ArrowRight className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                    </div>
                    
                    <div className="space-y-2">
                      <h3 className="font-semibold text-lg group-hover:text-primary transition-colors">
                        {suggestion.title}
                      </h3>
                      <p className="text-sm text-muted-foreground leading-relaxed">
                        {suggestion.description}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        {/* Features Section */}
        <div className="mt-16 text-center space-y-6">
          <h2 className="text-2xl font-semibold">Why Choose React Builder?</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-8">
            <div className="space-y-2">
              <div className="mx-auto w-12 h-12 rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                <Zap className="h-6 w-6 text-white" />
              </div>
              <h3 className="font-semibold">Lightning Fast</h3>
              <p className="text-sm text-muted-foreground">
                Generate complete React apps in seconds, not hours
              </p>
            </div>
            <div className="space-y-2">
              <div className="mx-auto w-12 h-12 rounded-full bg-gradient-to-r from-green-500 to-emerald-500 flex items-center justify-center">
                <Code className="h-6 w-6 text-white" />
              </div>
              <h3 className="font-semibold">Modern Stack</h3>
              <p className="text-sm text-muted-foreground">
                Built with React, TypeScript, and the latest best practices
              </p>
            </div>
            <div className="space-y-2">
              <div className="mx-auto w-12 h-12 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center">
                <Heart className="h-6 w-6 text-white" />
              </div>
              <h3 className="font-semibold">Ready to Deploy</h3>
              <p className="text-sm text-muted-foreground">
                Get production-ready code that you can customize and deploy
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t bg-muted/30 mt-16">
        <div className="container mx-auto px-4 py-8 text-center text-sm text-muted-foreground">
          <p>Built with ❤️ using React, TypeScript, and AI</p>
        </div>
      </footer>
    </div>
  );
};
