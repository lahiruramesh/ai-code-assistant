// Sound notification service for the React app

class SoundService {
  private audioContext: AudioContext | null = null;
  private enabled: boolean = true;

  constructor() {
    this.initializeAudioContext();
  }

  private initializeAudioContext() {
    try {
      this.audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    } catch (error) {
      console.warn('Web Audio API not supported:', error);
      this.enabled = false;
    }
  }

  private async resumeAudioContext() {
    if (this.audioContext && this.audioContext.state === 'suspended') {
      await this.audioContext.resume();
    }
  }

  // Create a simple tone using Web Audio API
  private createTone(frequency: number, duration: number, type: OscillatorType = 'sine') {
    if (!this.audioContext || !this.enabled) return;

    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.audioContext.destination);

    oscillator.frequency.setValueAtTime(frequency, this.audioContext.currentTime);
    oscillator.type = type;

    // Envelope for smooth sound
    gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
    gainNode.gain.linearRampToValueAtTime(0.1, this.audioContext.currentTime + 0.01);
    gainNode.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + duration);

    oscillator.start(this.audioContext.currentTime);
    oscillator.stop(this.audioContext.currentTime + duration);
  }

  // Success notification sound
  async playSuccess() {
    await this.resumeAudioContext();
    
    // Play a pleasant success chord
    setTimeout(() => this.createTone(523.25, 0.2), 0);    // C5
    setTimeout(() => this.createTone(659.25, 0.2), 100);  // E5
    setTimeout(() => this.createTone(783.99, 0.3), 200);  // G5
  }

  // Error notification sound
  async playError() {
    await this.resumeAudioContext();
    
    // Play a lower, more attention-grabbing sound
    setTimeout(() => this.createTone(220, 0.3, 'square'), 0);
    setTimeout(() => this.createTone(196, 0.3, 'square'), 150);
  }

  // Progress step completion sound
  async playStepComplete() {
    await this.resumeAudioContext();
    
    // Play a quick positive beep
    this.createTone(800, 0.1);
  }

  // Agent activity notification (subtle)
  async playAgentActivity() {
    await this.resumeAudioContext();
    
    // Play a very subtle notification
    this.createTone(1000, 0.05);
  }

  // Message received sound
  async playMessageReceived() {
    await this.resumeAudioContext();
    
    // Play a gentle notification
    this.createTone(600, 0.1);
  }

  // Project build complete celebration sound
  async playBuildComplete() {
    await this.resumeAudioContext();
    
    // Play a celebratory ascending sequence
    const notes = [523.25, 587.33, 659.25, 698.46, 783.99]; // C-D-E-F-G
    notes.forEach((note, index) => {
      setTimeout(() => this.createTone(note, 0.2), index * 100);
    });
  }

  // Enable/disable sounds
  setEnabled(enabled: boolean) {
    this.enabled = enabled;
  }

  // Check if sounds are enabled
  isEnabled(): boolean {
    return this.enabled && this.audioContext !== null;
  }
}

// Create a singleton instance
export const soundService = new SoundService();

// Hook for using sound service in React components
export const useSoundService = () => {
  return soundService;
};
