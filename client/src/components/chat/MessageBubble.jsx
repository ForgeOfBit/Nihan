import React from 'react';

export default function MessageBubble({ message, isOwn }) {
  return (
    <div className={`flex w-full ${isOwn ? 'justify-end' : 'justify-start'} mb-4`}>
      <div 
        className={`max-w-[70%] rounded-2xl px-4 py-2.5 ${
          isOwn 
            ? 'bg-gradient-to-br from-primary to-accent text-white rounded-br-sm shadow-lg shadow-primary/20' 
            : 'bg-surface/80 border border-white/5 text-white/90 rounded-bl-sm shadow-md'
        }`}
      >
        <p className="text-[15px] leading-relaxed break-words">{message.content}</p>
        <div className={`flex items-center gap-1 mt-1 text-[10px] ${isOwn ? 'text-white/70 justify-end' : 'text-white/40'}`}>
          <span>{new Date(message.timestamp || Date.now()).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</span>
          {isOwn && (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          )}
        </div>
      </div>
    </div>
  );
}
