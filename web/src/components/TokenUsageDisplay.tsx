import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { 
  BarChart3, 
  TrendingUp, 
  Database, 
  RefreshCw,
  ArrowUp,
  ArrowDown,
  Activity
} from "lucide-react";

interface TokenUsageStats {
  total_requests: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
  by_provider_model: Array<{
    provider: string;
    model: string;
    requests: number;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
  }>;
}

interface SessionTokenUsage {
  session_id: string;
  total_input: number;
  total_output: number;
  total_tokens: number;
  usage_details: Array<{
    model: string;
    provider: string;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    request_type: string;
    created_at: string;
  }>;
  request_count: number;
}

interface TokenUsageDisplayProps {
  sessionId?: string;
  showGlobalStats?: boolean;
  className?: string;
}

const TokenUsageDisplay: React.FC<TokenUsageDisplayProps> = ({
  sessionId,
  showGlobalStats = false,
  className = "",
}) => {
  const [sessionUsage, setSessionUsage] = useState<SessionTokenUsage | null>(null);
  const [globalStats, setGlobalStats] = useState<TokenUsageStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  useEffect(() => {
    if (sessionId) {
      fetchSessionUsage();
    }
    if (showGlobalStats) {
      fetchGlobalStats();
    }
  }, [sessionId, showGlobalStats]);

  const fetchSessionUsage = async () => {
    if (!sessionId) return;
    
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/tokens/usage/${sessionId}`);
      const data = await response.json();
      setSessionUsage(data);
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to fetch session token usage:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchGlobalStats = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/tokens/stats');
      const data = await response.json();
      setGlobalStats(data);
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to fetch global token stats:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const calculateRatio = (input: number, output: number) => {
    const total = input + output;
    return total > 0 ? Math.round((input / total) * 100) : 50;
  };

  const getProviderColor = (provider: string) => {
    const colors = {
      ollama: 'bg-blue-100 text-blue-800',
      openrouter: 'bg-green-100 text-green-800',
      gemini: 'bg-purple-100 text-purple-800',
      anthropic: 'bg-orange-100 text-orange-800',
      bedrock: 'bg-red-100 text-red-800',
    };
    return colors[provider as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  const refresh = () => {
    if (sessionId) {
      fetchSessionUsage();
    }
    if (showGlobalStats) {
      fetchGlobalStats();
    }
  };

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Session Usage */}
      {sessionId && sessionUsage && (
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2 text-lg">
                <Activity className="h-5 w-5" />
                Session Token Usage
              </CardTitle>
              <Button
                variant="outline"
                size="sm"
                onClick={refresh}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Session Summary */}
            <div className="grid grid-cols-3 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-blue-600">
                  {formatNumber(sessionUsage.total_input)}
                </div>
                <div className="text-xs text-muted-foreground flex items-center justify-center gap-1">
                  <ArrowUp className="h-3 w-3" />
                  Input Tokens
                </div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600">
                  {formatNumber(sessionUsage.total_output)}
                </div>
                <div className="text-xs text-muted-foreground flex items-center justify-center gap-1">
                  <ArrowDown className="h-3 w-3" />
                  Output Tokens
                </div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold">
                  {formatNumber(sessionUsage.total_tokens)}
                </div>
                <div className="text-xs text-muted-foreground">
                  Total Tokens
                </div>
              </div>
            </div>

            {/* Token Distribution */}
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Input/Output Ratio</span>
                <span>{calculateRatio(sessionUsage.total_input, sessionUsage.total_output)}% input</span>
              </div>
              <Progress 
                value={calculateRatio(sessionUsage.total_input, sessionUsage.total_output)} 
                className="h-2"
              />
            </div>

            {/* Request Count */}
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Requests:</span>
              <Badge variant="secondary">{sessionUsage.request_count}</Badge>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Global Statistics */}
      {showGlobalStats && globalStats && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-lg">
              <BarChart3 className="h-5 w-5" />
              Global Token Statistics
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Overall Stats */}
            <div className="grid grid-cols-2 gap-4">
              <div className="text-center">
                <div className="text-xl font-bold">
                  {formatNumber(globalStats.total_tokens)}
                </div>
                <div className="text-xs text-muted-foreground">Total Tokens</div>
              </div>
              <div className="text-center">
                <div className="text-xl font-bold">
                  {formatNumber(globalStats.total_requests)}
                </div>
                <div className="text-xs text-muted-foreground">Total Requests</div>
              </div>
            </div>

            <Separator />

            {/* By Provider/Model */}
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Usage by Model</h4>
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {globalStats.by_provider_model
                  .sort((a, b) => b.total_tokens - a.total_tokens)
                  .slice(0, 5)
                  .map((item, index) => (
                    <div key={index} className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <Badge 
                          variant="secondary" 
                          className={`text-xs ${getProviderColor(item.provider)}`}
                        >
                          {item.provider}
                        </Badge>
                        <span className="font-mono text-xs truncate max-w-24">
                          {item.model}
                        </span>
                      </div>
                      <div className="text-right">
                        <div className="font-medium">{formatNumber(item.total_tokens)}</div>
                        <div className="text-xs text-muted-foreground">
                          {item.requests} req
                        </div>
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Last Updated */}
      {lastUpdated && (
        <div className="text-xs text-muted-foreground text-center">
          Last updated: {lastUpdated.toLocaleTimeString()}
        </div>
      )}
    </div>
  );
};

export default TokenUsageDisplay;
