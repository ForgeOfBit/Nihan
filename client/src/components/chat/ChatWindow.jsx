import React, { useState } from 'react';
import MessageBubble from './MessageBubble';

export default function ChatWindow() {
  const [inputText, setInputText] = useState('');
  
  // Dummy messages for UI
  const messages = [
    { id: 1, content: "Selam, E2EE test mesajı.", isOwn: false, timestamp: new Date(Date.now() - 3600000) },
    { id: 2, content: "Selam! Şifreleme harika çalışıyor.", isOwn: true, timestamp: new Date(Date.now() - 3500000) },
    { id: 3, content: "Aynen, libsodium çok hızlı.", isOwn: false, timestamp: new Date(Date.now() - 3400000) }
  ];

  const handleSend = (e) => {
    e.preventDefault();
    if (!inputText.trim()) return;
    // Send logic will go here
    setInputText('');
  };

  return (
    <div className="flex-1 flex flex-col h-full bg-gradient-to-b from-transparent to-black/20">
      {/* Chat Header */}
      <div className="h-16 border-b border-white/5 flex items-center px-6 bg-surface/50 backdrop-blur-md z-10 shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary/40 to-accent/40 flex items-center justify-center font-bold text-white relative">
            S
            <span className="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#1E1E24] bg-green-500"></span>
          </div>
          <div>
            <h3 className="font-semibold text-white">Samet<span className="text-white/30 text-sm font-normal ml-1">#0432</span></h3>
            <p className="text-xs text-green-400">Çevrimiçi</p>
          </div>
        </div>
        
        {/* Header Actions */}
        <div className="ml-auto flex items-center gap-4 text-white/50">
          <button className="hover:text-white transition-colors" title="Sesli Arama (Yakında)">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path d="M2 3a1 1 0 011-1h2.153a1 1 0 01.986.836l.74 4.435a1 1 0 01-.54 1.06l-1.548.773a11.037 11.037 0 006.105 6.105l.774-1.548a1 1 0 011.059-.54l4.435.74a1 1 0 01.836.986V17a1 1 0 01-1 1h-2C7.82 18 2 12.18 2 5V3z" /></svg>
          </button>
          <div className="w-px h-5 bg-white/10"></div>
          <button className="hover:text-white transition-colors flex items-center gap-1.5 px-2 py-1 rounded border border-white/10 text-xs">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5 text-primary" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" /></svg>
            E2EE
          </button>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-6 scroll-smooth">
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border border-white/10 text-xs text-white/40">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" /></svg>
            Mesajlar uçtan uca şifrelenmektedir
          </div>
        </div>
        
        {messages.map(msg => (
          <MessageBubble key={msg.id} message={msg} isOwn={msg.isOwn} />
        ))}
      </div>

      {/* Input Area */}
      <div className="p-4 bg-surface/80 backdrop-blur-xl border-t border-white/5 shrink-0">
        <form onSubmit={handleSend} className="flex items-end gap-2 max-w-4xl mx-auto">
          <button type="button" className="p-3 text-white/40 hover:text-white hover:bg-white/5 rounded-xl transition-all">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" /></svg>
          </button>
          
          <div className="flex-1 relative">
            <input
              type="text"
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              placeholder="Şifreli bir mesaj yazın..."
              className="w-full bg-black/20 border border-white/10 text-white placeholder-white/30 rounded-2xl py-3 px-4 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all"
            />
          </div>

          <button 
            type="submit"
            disabled={!inputText.trim()}
            className="p-3 bg-gradient-to-br from-primary to-accent text-white rounded-xl shadow-lg shadow-primary/20 hover:shadow-primary/40 disabled:opacity-50 disabled:cursor-not-allowed transition-all transform hover:-translate-y-0.5 disabled:transform-none"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 transform rotate-90" viewBox="0 0 20 20" fill="currentColor"><path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" /></svg>
          </button>
        </form>
      </div>
    </div>
  );
}
