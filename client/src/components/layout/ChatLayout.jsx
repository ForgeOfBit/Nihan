import React from 'react';
import Sidebar from './Sidebar';

export default function ChatLayout({ children }) {
  return (
    <div className="flex h-screen bg-background text-white overflow-hidden">
      <Sidebar />
      <main className="flex-1 flex flex-col relative bg-surface/30 backdrop-blur-3xl border-l border-white/5">
        {children}
      </main>
    </div>
  );
}
