import React, { useState, useEffect } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Brain, Cpu, Settings, Zap } from "lucide-react";
import { useModels } from '@/hooks';

interface ModelSelectorProps {
  onModelChange: (provider: string, model: string, autoMode: boolean) => void;
  currentProvider?: string;
  currentModel?: string;
  autoMode?: boolean;
}

interface ModelGroup {
  [provider: string]: {
    [category: string]: string[];
  };
}

const ModelSelector: React.FC<ModelSelectorProps> = ({
  onModelChange,
  currentProvider = "ollama",
  currentModel = "cogito:8b",
  autoMode = false,
}) => {
  const [selectedProvider, setSelectedProvider] = useState(currentProvider);
  const [selectedModel, setSelectedModel] = useState(currentModel);
  const [isAutoMode, setIsAutoMode] = useState(autoMode);
  const [allModels, setAllModels] = useState<ModelGroup>({});

  const { data: modelsData, isLoading: loading, error: modelsError } = useModels();

  useEffect(() => {
    if (modelsData?.models) {
      setAllModels(modelsData.models);
    } else if (modelsError) {
      // Fallback to environment variables if API fails
      setAllModels(getModelsFromEnv());
    }
  }, [modelsData, modelsError]);

  const getModelsFromEnv = (): ModelGroup => {
    return {
      ollama: {
        local: (import.meta.env.VITE_OLLAMA_MODELS || "").split(",").filter(Boolean)
      },
      openrouter: {
        ai: (import.meta.env.VITE_OPENROUTER_MODELS || "").split(",").filter(Boolean)
      },
      gemini: {
        google: (import.meta.env.VITE_GEMINI_MODELS || "").split(",").filter(Boolean)
      },
      anthropic: {
        claude: (import.meta.env.VITE_ANTHROPIC_MODELS || "").split(",").filter(Boolean)
      },
      bedrock: {
        aws: (import.meta.env.VITE_BEDROCK_MODELS || "").split(",").filter(Boolean)
      }
    };
  };

  const handleProviderChange = (provider: string) => {
    setSelectedProvider(provider);
    
    // Auto-select the first model for the provider
    const providerModels = allModels[provider];
    if (providerModels) {
      const firstCategory = Object.keys(providerModels)[0];
      const firstModel = providerModels[firstCategory]?.[0];
      if (firstModel) {
        setSelectedModel(firstModel);
        onModelChange(provider, firstModel, isAutoMode);
      }
    }
  };

  const handleModelChange = (model: string) => {
    setSelectedModel(model);
    onModelChange(selectedProvider, model, isAutoMode);
  };

  const handleAutoModeToggle = (checked: boolean) => {
    setIsAutoMode(checked);
    onModelChange(selectedProvider, selectedModel, checked);
  };

  const getProviderIcon = (provider: string) => {
    switch (provider) {
      case 'ollama':
        return <Cpu className="h-4 w-4" />;
      case 'openrouter':
        return <Zap className="h-4 w-4" />;
      case 'gemini':
        return <Brain className="h-4 w-4" />;
      case 'anthropic':
        return <Brain className="h-4 w-4" />;
      default:
        return <Settings className="h-4 w-4" />;
    }
  };

  const getProviderModels = () => {
    const providerModels = allModels[selectedProvider];
    if (!providerModels) return [];

    const models: { value: string; label: string; category: string }[] = [];
    Object.entries(providerModels).forEach(([category, modelList]) => {
      modelList.forEach((model) => {
        models.push({
          value: model,
          label: model,
          category: category,
        });
      });
    });
    return models;
  };

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Settings className="h-5 w-5" />
          Model Configuration
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Auto Mode Toggle */}
        <div className="flex items-center justify-between">
          <Label htmlFor="auto-mode" className="text-sm font-medium">
            Auto Mode
          </Label>
          <Switch
            id="auto-mode"
            checked={isAutoMode}
            onCheckedChange={handleAutoModeToggle}
          />
        </div>
        
        <Separator />

        {/* Provider Selection */}
        <div className="space-y-2">
          <Label className="text-sm font-medium">Provider</Label>
          <Select value={selectedProvider} onValueChange={handleProviderChange}>
            <SelectTrigger>
              <SelectValue>
                <div className="flex items-center gap-2">
                  {getProviderIcon(selectedProvider)}
                  <span className="capitalize">{selectedProvider}</span>
                </div>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {Object.keys(allModels).map((provider) => (
                <SelectItem key={provider} value={provider}>
                  <div className="flex items-center gap-2">
                    {getProviderIcon(provider)}
                    <span className="capitalize">{provider}</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Model Selection */}
        <div className="space-y-2">
          <Label className="text-sm font-medium">Model</Label>
          <Select value={selectedModel} onValueChange={handleModelChange}>
            <SelectTrigger>
              <SelectValue placeholder="Select a model" />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(allModels[selectedProvider] || {}).map(([category, models]) => (
                <div key={category}>
                  <div className="px-2 py-1">
                    <Badge variant="secondary" className="text-xs">
                      {category}
                    </Badge>
                  </div>
                  {models.map((model) => (
                    <SelectItem key={model} value={model} className="ml-4">
                      {model}
                    </SelectItem>
                  ))}
                </div>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Current Selection Display */}
        <div className="rounded-lg bg-muted p-3 text-sm">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Current:</span>
            <div className="flex items-center gap-1">
              {getProviderIcon(selectedProvider)}
              <span className="font-medium">{selectedModel}</span>
            </div>
          </div>
          {isAutoMode && (
            <div className="mt-1 text-xs text-muted-foreground">
              Auto mode enabled - system will select optimal models
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

export default ModelSelector;
