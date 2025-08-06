import React from 'react';
import { Loader2, Code2, Palette, FileText, CheckCircle2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';

interface ProjectLoadingScreenProps {
  currentStep: string;
  progress: number;
  completedSteps: string[];
  totalSteps: string[];
}

const stepIcons = {
  'Project Setup': FileText,
  'Component Creation': Code2,
  'Styling': Palette,
  'Build & Preview': CheckCircle2,
};

export const ProjectLoadingScreen: React.FC<ProjectLoadingScreenProps> = ({
  currentStep,
  progress,
  completedSteps,
  totalSteps
}) => {
  return (
    <div className="h-full flex items-center justify-center bg-gradient-to-br from-background via-muted/20 to-background">
      <Card className="w-full max-w-md mx-4 bg-card/50 backdrop-blur-sm border border-border/50">
        <CardContent className="p-8 space-y-6">
          {/* Main Loading Animation */}
          <div className="text-center space-y-4">
            <div className="relative inline-block">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <div className="absolute inset-0 h-12 w-12 rounded-full border-2 border-primary/20 animate-pulse" />
            </div>
            
            <div>
              <h3 className="text-xl font-semibold mb-2">Building Your Application</h3>
              <p className="text-muted-foreground text-sm">
                Our AI agents are working together to create your React app
              </p>
            </div>
          </div>

          {/* Progress Bar */}
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Progress</span>
              <span className="font-medium">{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="h-2" />
          </div>

          {/* Current Step */}
          <div className="space-y-3">
            <div className="flex items-center gap-3 p-3 rounded-lg bg-primary/5 border border-primary/20">
              <Loader2 className="h-4 w-4 animate-spin text-primary" />
              <div className="flex-1">
                <div className="text-sm font-medium">Currently Working On</div>
                <div className="text-xs text-muted-foreground">{currentStep}</div>
              </div>
            </div>
          </div>

          {/* Steps List */}
          <div className="space-y-2">
            <div className="text-sm font-medium mb-3">Build Steps</div>
            {totalSteps.map((step, index) => {
              const isCompleted = completedSteps.includes(step);
              const isCurrent = currentStep === step;
              const Icon = stepIcons[step as keyof typeof stepIcons] || FileText;
              
              return (
                <div
                  key={step}
                  className={`flex items-center gap-3 p-2 rounded-md transition-colors ${
                    isCompleted
                      ? 'bg-green-50 text-green-700 border border-green-200'
                      : isCurrent
                      ? 'bg-blue-50 text-blue-700 border border-blue-200'
                      : 'text-muted-foreground'
                  }`}
                >
                  {isCompleted ? (
                    <CheckCircle2 className="h-4 w-4 text-green-600" />
                  ) : isCurrent ? (
                    <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
                  ) : (
                    <Icon className="h-4 w-4" />
                  )}
                  <span className="text-sm">{step}</span>
                </div>
              );
            })}
          </div>

          {/* Estimated Time */}
          <div className="text-center pt-4 border-t border-border/50">
            <p className="text-xs text-muted-foreground">
              Estimated time: 2-5 minutes
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
