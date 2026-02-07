// src/components/VoiceLink.tsx - Neural Voice Link with Web Speech API
import React, { useState, useEffect } from 'react';
import { Mic, MicOff, Send, Volume2 } from 'lucide-react';

declare global {
  interface Window {
    SpeechRecognition: any;
    webkitSpeechRecognition: any;
  }
}

interface VoiceResponse {
  query: string;
  response: string;
  timestamp: number;
}

export const VoiceLink: React.FC = () => {
  const [isListening, setIsListening] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [response, setResponse] = useState('');
  const [history, setHistory] = useState<VoiceResponse[]>([]);
  const [recognition, setRecognition] = useState<any>(null);

  // Initialize Speech Recognition
  useEffect(() => {
    if ('SpeechRecognition' in window || 'webkitSpeechRecognition' in window) {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
      const recognitionInstance = new SpeechRecognition();
      
      recognitionInstance.continuous = false;
      recognitionInstance.interimResults = false;
      recognitionInstance.lang = 'en-US';

      recognitionInstance.onresult = async (event: any) => {
        const query = event.results[0][0].transcript;
        setTranscript(query);
        await processVoiceQuery(query);
      };

      recognitionInstance.onerror = (event: any) => {
        console.error('Speech recognition error:', event.error);
        setIsListening(false);
      };

      recognitionInstance.onend = () => {
        setIsListening(false);
      };

      setRecognition(recognitionInstance);
    }
  }, []);

  const processVoiceQuery = async (query: string) => {
    try {
      const res = await fetch('http://localhost:5000/voice_query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
      });

      if (res.ok) {
        const json = await res.json();
        setResponse(json.response);
        
        // Add to history
        const newEntry: VoiceResponse = {
          query,
          response: json.response,
          timestamp: json.timestamp || Date.now()
        };
        setHistory(prev => [newEntry, ...prev].slice(0, 10)); // Keep last 10

        // Speak response
        if ('speechSynthesis' in window) {
          const utterance = new SpeechSynthesisUtterance(json.response);
          utterance.rate = 1.1;
          utterance.pitch = 1.0;
          window.speechSynthesis.speak(utterance);
        }
      }
    } catch (error) {
      console.error('Failed to process voice query:', error);
      setResponse('Error processing query');
    }
  };

  const startListening = () => {
    if (recognition) {
      setTranscript('');
      setResponse('');
      recognition.start();
      setIsListening(true);
    }
  };

  const stopListening = () => {
    if (recognition) {
      recognition.stop();
      setIsListening(false);
    }
  };

  const handleTextSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (transcript.trim()) {
      await processVoiceQuery(transcript);
    }
  };

  return (
    <div className="container mx-auto p-6 max-w-5xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="bg-black/50 backdrop-blur-lg border border-blue-500/50 rounded-lg p-6">
          <h2 className="text-3xl font-bold text-blue-400 mb-2 flex items-center gap-3">
            <Mic size={32} />
            Neural Voice Link
          </h2>
          <p className="text-gray-400">
            Voice-controlled interface to the Sovereign Maps neural mesh. 
            Ask questions about network status, security, or request threat scans.
          </p>
        </div>

        {/* Voice Control Panel */}
        <div className="bg-black/50 backdrop-blur-lg border border-blue-500/50 rounded-lg p-6">
          <div className="flex flex-col items-center gap-6">
            {/* Microphone Button */}
            <button
              onClick={isListening ? stopListening : startListening}
              className={`w-32 h-32 rounded-full flex items-center justify-center transition-all transform hover:scale-105 ${
                isListening
                  ? 'bg-red-500 hover:bg-red-600 animate-pulse'
                  : 'bg-blue-500 hover:bg-blue-600'
              }`}
            >
              {isListening ? <MicOff size={48} /> : <Mic size={48} />}
            </button>

            <p className="text-lg font-semibold">
              {isListening ? (
                <span className="text-red-400 animate-pulse">● LISTENING...</span>
              ) : (
                <span className="text-gray-400">Click to speak</span>
              )}
            </p>

            {/* Manual Input */}
            <form onSubmit={handleTextSubmit} className="w-full flex gap-2">
              <input
                type="text"
                value={transcript}
                onChange={(e) => setTranscript(e.target.value)}
                placeholder="Or type your query here..."
                className="flex-1 bg-gray-800 border border-blue-500/30 rounded-lg px-4 py-3 focus:outline-none focus:border-blue-500 text-white"
              />
              <button
                type="submit"
                className="bg-blue-500 hover:bg-blue-600 px-6 py-3 rounded-lg flex items-center gap-2 transition"
              >
                <Send size={20} />
                Send
              </button>
            </form>
          </div>
        </div>

        {/* Current Response */}
        {response && (
          <div className="bg-black/50 backdrop-blur-lg border border-green-500/50 rounded-lg p-6">
            <div className="flex items-start gap-3">
              <Volume2 size={24} className="text-green-400 mt-1" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-green-400 mb-2">Mesh Response</h3>
                <p className="text-white text-lg">{response}</p>
              </div>
            </div>
          </div>
        )}

        {/* Query History */}
        {history.length > 0 && (
          <div className="bg-black/50 backdrop-blur-lg border border-blue-500/50 rounded-lg p-6">
            <h3 className="text-xl font-semibold text-blue-400 mb-4">Query History</h3>
            <div className="space-y-3">
              {history.map((entry, index) => (
                <div
                  key={index}
                  className="bg-gray-800/50 border border-gray-700 rounded-lg p-4"
                >
                  <div className="flex items-start gap-2 mb-2">
                    <Mic size={16} className="text-blue-400 mt-1" />
                    <p className="text-gray-300 italic">"{entry.query}"</p>
                  </div>
                  <div className="flex items-start gap-2 pl-6">
                    <Volume2 size={16} className="text-green-400 mt-1" />
                    <p className="text-white">{entry.response}</p>
                  </div>
                  <p className="text-xs text-gray-500 mt-2 pl-6">
                    {new Date(entry.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Sample Commands */}
        <div className="bg-black/50 backdrop-blur-lg border border-blue-500/50 rounded-lg p-6">
          <h3 className="text-xl font-semibold text-blue-400 mb-4">Sample Commands</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {[
              'Scan for threats',
              'What is the current stake?',
              'How many enclaves are active?',
              'Show network status',
              'Run security check',
              'What is the mesh health?'
            ].map((cmd, i) => (
              <button
                key={i}
                onClick={() => {
                  setTranscript(cmd);
                  processVoiceQuery(cmd);
                }}
                className="bg-gray-800 hover:bg-gray-700 border border-blue-500/30 hover:border-blue-500 rounded-lg p-3 text-left transition"
              >
                <p className="text-sm text-gray-400">Try saying:</p>
                <p className="text-white font-medium">"{cmd}"</p>
              </button>
            ))}
          </div>
        </div>

        {/* Browser Compatibility */}
        {!recognition && (
          <div className="bg-yellow-500/20 border border-yellow-500 rounded-lg p-4">
            <p className="text-yellow-200">
              ⚠️ Web Speech API is not supported in your browser. 
              Please use Chrome, Edge, or Safari for voice features.
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
