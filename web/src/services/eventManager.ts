// Simple event manager for component communication

type EventCallback = (data?: any) => void;

class EventManager {
  private listeners: Map<string, EventCallback[]> = new Map();

  // Subscribe to an event
  on(event: string, callback: EventCallback): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    
    this.listeners.get(event)!.push(callback);
    
    // Return unsubscribe function
    return () => {
      const callbacks = this.listeners.get(event);
      if (callbacks) {
        const index = callbacks.indexOf(callback);
        if (index > -1) {
          callbacks.splice(index, 1);
        }
      }
    };
  }

  // Emit an event
  emit(event: string, data?: any): void {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event callback for ${event}:`, error);
        }
      });
    }
  }

  // Remove all listeners for an event
  off(event: string): void {
    this.listeners.delete(event);
  }

  // Remove all listeners
  clear(): void {
    this.listeners.clear();
  }
}

// Create singleton instance
export const eventManager = new EventManager();

// Event types
export const EVENTS = {
  PROJECT_BUILD_START: 'project:build:start',
  PROJECT_BUILD_PROGRESS: 'project:build:progress',
  PROJECT_BUILD_COMPLETE: 'project:build:complete',
  PROJECT_BUILD_ERROR: 'project:build:error',
  AGENT_ACTIVITY: 'agent:activity',
  MESSAGE_RECEIVED: 'message:received',
} as const;
