import React, { memo } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import {
    ExternalLink,
    Rocket,
    Link,
    AlertCircle,
} from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { useIntegrations } from '@/hooks/useIntegrations';
import type { Project } from '@/types';

interface HeaderProps {
    project?: Project | null;
    showIntegrations?: boolean;
    showSidebarTrigger?: boolean;
}

export const Header: React.FC<HeaderProps> = memo(({
    project,
    showIntegrations = false,
    showSidebarTrigger = false
}) => {
    const { user } = useAuth();
    const {
        isLoading,
        error,
        connectGitHub,
        connectVercel,
        deployToVercel,
        pushToGitHub,
    } = useIntegrations();

    const handleGitHubAction = async () => {
        if (!user?.github_connected) {
            await connectGitHub();
        } else if (project) {
            try {
                const repoUrl = await pushToGitHub(project.id);
                window.open(repoUrl, '_blank');
            } catch (err) {
                console.error('Failed to push to GitHub:', err);
            }
        }
    };

    const handleVercelAction = async () => {
        if (!user?.vercel_connected) {
            await connectVercel();
        } else if (project) {
            try {
                const deploymentUrl = await deployToVercel(project.id);
                window.open(deploymentUrl, '_blank');
            } catch (err) {
                console.error('Failed to deploy to Vercel:', err);
            }
        }
    };

    const openGitHubRepo = () => {
        if (project?.github_repo) {
            window.open(`https://github.com/${project.github_repo}`, '_blank');
        }
    };

    const openVercelProject = () => {
        if (project?.vercel_project) {
            window.open(`https://vercel.com/dashboard/${project.vercel_project}`, '_blank');
        }
    };

    if (!showIntegrations) {
        return (
            <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                <a href='/'>
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-lg bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-between">
                            <span className="text-white text-xs font-bold">CA</span>
                        </div>
                        <span className="font-semibold">Code Agent</span>
                    </div>
                </a>
            </header>
        );
    }

    return (
        <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
            <div className="flex h-14 items-center justify-between w-full px-2">
                <div className="flex items-center gap-4">
                    {showSidebarTrigger && <SidebarTrigger />}
                    <a href='/'>
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 rounded-lg bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-between">
                                <span className="text-white text-xs font-bold">CA</span>
                            </div>
                            <span className="font-semibold">Code Agent</span>
                        </div>
                    </a>

                    {project && (
                        <>
                            <Separator orientation="vertical" className="h-6" />
                            <div className="flex items-center gap-2">
                                <h1 className="text-lg font-semibold">{project.name}</h1>
                                <Badge variant="outline" className="text-xs">
                                    {project.template}
                                </Badge>
                            </div>
                        </>
                    )}
                </div>

                <div className="flex items-center gap-2">
                    {error && (
                        <div className="flex items-center gap-1 text-red-600 text-sm">
                            <AlertCircle className="w-4 h-4" />
                            <span>Connection failed</span>
                        </div>
                    )}

                    {/* GitHub Integration */}
                    <div className="flex items-center gap-2">
                        {user?.github_connected ? (
                            <div className="flex items-center gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={handleGitHubAction}
                                    disabled={isLoading}
                                    className="flex items-center gap-2"
                                >
                                    <Github className="w-4 h-4" />
                                    {project?.github_repo ? 'Push to GitHub' : 'Create Repo'}
                                </Button>

                                {project?.github_repo && (
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={openGitHubRepo}
                                        className="flex items-center gap-1"
                                    >
                                        <ExternalLink className="w-3 h-3" />
                                        View Repo
                                    </Button>
                                )}
                            </div>
                        ) : (
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleGitHubAction}
                                disabled={isLoading}
                                className="flex items-center gap-2"
                            >
                                <Github className="w-4 h-4" />
                                <Link className="w-3 h-3" />
                                Connect GitHub
                            </Button>
                        )}
                    </div>

                    <Separator orientation="vertical" className="h-6" />

                    {/* Vercel Integration */}
                    <div className="flex items-center gap-2">
                        {user?.vercel_connected ? (
                            <div className="flex items-center gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={handleVercelAction}
                                    disabled={isLoading}
                                    className="flex items-center gap-2"
                                >
                                    <Rocket className="w-4 h-4" />
                                    Deploy to Vercel
                                </Button>

                                {project?.vercel_project && (
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={openVercelProject}
                                        className="flex items-center gap-1"
                                    >
                                        <ExternalLink className="w-3 h-3" />
                                        View Project
                                    </Button>
                                )}
                            </div>
                        ) : (
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleVercelAction}
                                disabled={isLoading}
                                className="flex items-center gap-2"
                            >
                                <Rocket className="w-4 h-4" />
                                <Link className="w-3 h-3" />
                                Connect Vercel
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </header>
    );
});

Header.displayName = 'Header';
